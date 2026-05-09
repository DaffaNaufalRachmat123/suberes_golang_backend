package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"suberes_golang/config"
	"suberes_golang/constants"
	"suberes_golang/models"
	"suberes_golang/service"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// envMinutes reads an env var expected to be a plain integer (number of minutes)
// and returns it as time.Duration. Falls back to defaultMinutes if unset or invalid.
func envMinutes(key string, defaultMinutes int) time.Duration {
	n, err := strconv.Atoi(os.Getenv(key))
	if err != nil || n <= 0 {
		n = defaultMinutes
	}
	return time.Duration(n) * time.Minute
}

func HandleOrderQueueCashTask(ctx context.Context, t *asynq.Task) error {
	var p OrderQueueCashPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Backfill FirstEnqueued for tasks created before this field existed
	if p.FirstEnqueued == 0 {
		p.FirstEnqueued = time.Now().Unix()
	}

	asynqRetry, _ := asynq.GetRetryCount(ctx)
	log.Printf("[QUEUE][CORE_SEARCH_CASH] START order_id=%s | asynq_retry=%d | enqueued_ago=%s",
		p.OrderID, asynqRetry, time.Since(time.Unix(p.FirstEnqueued, 0)).Truncate(time.Second))

	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ?", p.OrderID).First(&orderData).Error; err != nil {
		return fmt.Errorf("failed to find order transaction with id %s: %v", p.OrderID, err)
	}

	// Guard: only proceed if order is in FINDING_MITRA status
	if orderData.OrderStatus != "FINDING_MITRA" {
		log.Printf("[QUEUE][CORE_SEARCH_CASH] Skipping order_id=%s | current_status=%s (not FINDING_MITRA)", orderData.ID, orderData.OrderStatus)
		return nil
	}

	// IsWithTime and ServiceDuration are pre-fetched at enqueue time — no DB round-trip needed here.

	initialRange, _ := strconv.ParseFloat(os.Getenv("INITIAL_RANGE_CUSTOMER"), 64)
	if initialRange == 0 {
		initialRange = 1
	}
	maxRange, _ := strconv.ParseFloat(os.Getenv("MAX_RANGE_CUSTOMER"), 64)
	if maxRange == 0 {
		maxRange = 10
	}

	params := GetNearestMitraProductionParams{
		CustomerID:           orderData.CustomerID,
		Latitude:             orderData.CustomerLatitude,
		Longitude:            orderData.CustomerLongitude,
		InitialRange:         initialRange,
		MaxRange:             maxRange,
		UserGender:           orderData.MitraGender,
		OrderType:            orderData.OrderType,
		SubPaymentID:         orderData.SubPaymentID,
		IsAutoBid:            "yes",
		ServiceDuration:      p.ServiceDuration,
		CustomerTimezoneCode: orderData.TimezoneCode,
		CustomerTimeOrder:    orderData.OrderTime.String(),
		GrossAmountCompany:   float64(orderData.GrossAmountCompany),
		IsWithTime:           p.IsWithTime,
		IsCash:               orderData.PaymentType == "tunai",
		Limit:                10,
		Page:                 0,
	}

	queryStart := time.Now()
	result, err := GetNearestMitraProduction(params)
	queryElapsed := time.Since(queryStart).Truncate(time.Millisecond)
	if err != nil {
		// Spatial query errors (e.g. missing column, PostGIS issue) are permanent — SkipRetry
		// so Asynq does not pile up exponential-backoff delay on top of a schema-level bug.
		log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR spatial query order_id=%s | elapsed=%s | %v", p.OrderID, queryElapsed, err)
		return fmt.Errorf("failed to get nearest mitra: %v: %w", err, asynq.SkipRetry)
	}

	// Log mitra search result
	mitraIDs := make([]string, 0, len(result.PayloadMitra))
	for _, m := range result.PayloadMitra {
		mitraIDs = append(mitraIDs, m.ID)
	}
	log.Printf("[QUEUE][CORE_SEARCH_CASH] order_id=%s | order_type=%s | lat=%.6f | lng=%.6f | max_range=%.1f | mitra_found=%d | mitra_ids=%v | is_auto_bid=%v | query_elapsed=%s",
		orderData.ID, orderData.OrderType, orderData.CustomerLatitude, orderData.CustomerLongitude,
		maxRange, len(result.PayloadMitra), mitraIDs, result.IsAutoBid, queryElapsed)

	if len(result.PayloadMitra) > 0 {
		tempID := uuid.New().String()
		var orderOffersPayload []models.OrderOffer
		var registrationTokenList []string

		for _, mitra := range result.PayloadMitra {
			orderOffersPayload = append(orderOffersPayload, models.OrderOffer{
				TempID:     tempID,
				OrderID:    orderData.ID,
				CustomerID: orderData.CustomerID,
				MitraID:    mitra.ID,
			})
			if mitra.FirebaseToken != nil {
				registrationTokenList = append(registrationTokenList, *mitra.FirebaseToken)
			}
		}

		tx := config.DB.Begin()
		if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Update("temp_id", tempID).Error; err != nil {
			tx.Rollback()
			log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR update temp_id order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to update order transaction temp_id: %v: %w", err, asynq.SkipRetry)
		}

		if err := tx.Create(&orderOffersPayload).Error; err != nil {
			tx.Rollback()
			log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR create order_offers order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to create order offers: %v: %w", err, asynq.SkipRetry)
		}
		if err := tx.Commit().Error; err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR tx commit order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to commit order offers transaction: %v: %w", err, asynq.SkipRetry)
		}

		// --- Kirim push notification ke mitra ---
		timeoutCanTakeOrder := envMinutes("TIMEOUT_CAN_TAKE_ORDER", 1)
		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)

		if len(registrationTokenList) > 0 {
			notifType := constants.ORDER_BROADCAST
			titleOrder := "Orderan Masuk"
			messageOrder := "Ada orderan masuk nih"
			if result.IsAutoBid {
				notifType = constants.ORDER_AUTO_BID
				messageOrder = "Orderan Otomatis Masuk Ke Akun Mu"
			}
			msgData := map[string]interface{}{
				"temp_id":           tempID,
				"notification_type": notifType,
				"title":             titleOrder,
				"message":           messageOrder,
				"order_id":          orderData.ID,
				"order_type":        orderData.OrderType,
				"payment_type":      orderData.PaymentType,
				"customer_id":       orderData.CustomerID,
				"notification_id":   strconv.Itoa(orderData.NotificationID),
				"notif_type":        "order",
				"push_time":         time.Now().UTC().Format("2006-01-02 15:04:05"),
				"push_time_expired": strconv.Itoa(int(timeoutCanTakeOrder.Seconds())),
			}
			if result.IsAutoBid {
				msgData["mitra_id"] = result.AutoBidMitraID
			}
			msg := map[string]interface{}{
				"data":   msgData,
				"tokens": registrationTokenList,
			}
			if _, err := service.SendMulticast(config.DB, "mitra", msg); err != nil {
			} else {
			}
		} else {
		}

		// Mark search_time_complete — countdown starts from NOW for mitra
		now := time.Now()
		config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Update("search_time_complete", now)

		// ---

		// PENTING: Setelah TX commit, jangan return error — akan menyebabkan Asynq retry
		// dan membuat offer duplikat dengan tempID baru. Log + lanjutkan saja.
		offerExpiredTaskPayload, err := NewOrderOfferExpiredTask(orderData.ID, tempID, orderData.CustomerID, orderData.NotificationID, timeoutFindingOrder.String())
		if err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_CASH] WARN marshal offer_expired payload failed order_id=%s | %v", orderData.ID, err)
		} else if _, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferExpired, offerExpiredTaskPayload), asynq.ProcessIn(timeoutCanTakeOrder)); err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_CASH] WARN enqueue offer_expired failed order_id=%s | %v (offers tetap ada di DB, mitra bisa accept)", orderData.ID, err)
		}
		// Enqueue per-mitra expiry cleanup
		for _, offer := range orderOffersPayload {
			perMitraPayload, perMitraErr := NewOrderOfferMitraExpiredTask(orderData.ID, offer.MitraID, tempID)
			if perMitraErr == nil {
				_, _ = AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferMitraExpired, perMitraPayload), asynq.ProcessIn(timeoutCanTakeOrder))
			}
		}
	} else {
		// Tidak ada mitra ditemukan — langsung ubah ke WAITING_FOR_SELECTED_MITRA
		log.Printf("[QUEUE][CORE_SEARCH_CASH] No mitra found | order_id=%s | updating status to WAITING_FOR_SELECTED_MITRA", orderData.ID)
		if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Updates(map[string]interface{}{
			"order_status": "WAITING_FOR_SELECTED_MITRA",
			"order_time":   time.Now(),
		}).Error; err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR update to WAITING_FOR_SELECTED_MITRA order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to update order status: %v: %w", err, asynq.SkipRetry)
		}
		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)
		orderSelectedExpiredTaskPayload, err := NewOrderSelectedExpiredTask(orderData.ID)
		if err == nil {
			_, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, orderSelectedExpiredTaskPayload), asynq.ProcessIn(timeoutFindingOrder))
			if err != nil {
				log.Printf("[QUEUE][CORE_SEARCH_CASH] Failed to enqueue selected expired task | order_id=%s | error=%v", orderData.ID, err)
			}
		}
	}

	return nil
}

func HandleOrderQueueVATask(ctx context.Context, t *asynq.Task) error {
	var p OrderQueueVAPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Backfill FirstEnqueued for tasks created before this field existed
	if p.FirstEnqueued == 0 {
		p.FirstEnqueued = time.Now().Unix()
	}

	asynqRetry, _ := asynq.GetRetryCount(ctx)
	log.Printf("[QUEUE][CORE_SEARCH_VA] START order_id=%s | asynq_retry=%d | enqueued_ago=%s",
		p.OrderID, asynqRetry, time.Since(time.Unix(p.FirstEnqueued, 0)).Truncate(time.Second))

	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ?", p.OrderID).First(&orderData).Error; err != nil {
		log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR fetch order order_id=%s | %v", p.OrderID, err)
		return fmt.Errorf("failed to find order transaction with id %s: %v: %w", p.OrderID, err, asynq.SkipRetry)
	}

	// Guard: only proceed if order is in FINDING_MITRA status
	if orderData.OrderStatus != "FINDING_MITRA" {
		log.Printf("[QUEUE][CORE_SEARCH_VA] Skipping order_id=%s | current_status=%s (not FINDING_MITRA)", orderData.ID, orderData.OrderStatus)
		return nil
	}

	// IsWithTime and ServiceDuration are pre-fetched at enqueue time — no DB round-trip needed here.

	initialRange, _ := strconv.ParseFloat(os.Getenv("INITIAL_RANGE_CUSTOMER"), 64)
	if initialRange == 0 {
		initialRange = 1
	}
	maxRange, _ := strconv.ParseFloat(os.Getenv("MAX_RANGE_CUSTOMER"), 64)
	if maxRange == 0 {
		maxRange = 10
	}

	// Cari mitra terdekat
	params := GetNearestMitraProductionParams{
		CustomerID:           orderData.CustomerID,
		Latitude:             orderData.CustomerLatitude,
		Longitude:            orderData.CustomerLongitude,
		InitialRange:         initialRange,
		MaxRange:             maxRange,
		UserGender:           orderData.MitraGender,
		OrderType:            orderData.OrderType,
		SubPaymentID:         orderData.SubPaymentID,
		IsAutoBid:            "yes",
		ServiceDuration:      p.ServiceDuration,
		CustomerTimezoneCode: orderData.TimezoneCode,
		CustomerTimeOrder:    orderData.OrderTime.String(),
		GrossAmountCompany:   float64(orderData.GrossAmountCompany),
		IsWithTime:           p.IsWithTime,
		IsCash:               orderData.PaymentType == "tunai",
		Limit:                10,
		Page:                 0,
	}

	queryStart := time.Now()
	result, err := GetNearestMitraProduction(params)
	queryElapsed := time.Since(queryStart).Truncate(time.Millisecond)
	if err != nil {
		log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR spatial query order_id=%s | elapsed=%s | %v", p.OrderID, queryElapsed, err)
		return fmt.Errorf("failed to get nearest mitra: %v: %w", err, asynq.SkipRetry)
	}

	// Log mitra search result
	vaSearchMitraIDs := make([]string, 0, len(result.PayloadMitra))
	for _, m := range result.PayloadMitra {
		vaSearchMitraIDs = append(vaSearchMitraIDs, m.ID)
	}
	log.Printf("[QUEUE][CORE_SEARCH_VA] order_id=%s | order_type=%s | lat=%.6f | lng=%.6f | max_range=%.1f | mitra_found=%d | mitra_ids=%v | is_auto_bid=%v | query_elapsed=%s",
		orderData.ID, orderData.OrderType, orderData.CustomerLatitude, orderData.CustomerLongitude,
		maxRange, len(result.PayloadMitra), vaSearchMitraIDs, result.IsAutoBid, queryElapsed)

	tempID := uuid.New().String()
	var orderOffersPayload []models.OrderOffer
	var registrationTokenList []string

	if len(result.PayloadMitra) > 0 {
		// Ada mitra ditemukan
		for _, mitra := range result.PayloadMitra {
			orderOffersPayload = append(orderOffersPayload, models.OrderOffer{
				TempID:     tempID,
				OrderID:    orderData.ID,
				CustomerID: orderData.CustomerID,
				MitraID:    mitra.ID,
			})
			if mitra.FirebaseToken != nil {
				registrationTokenList = append(registrationTokenList, *mitra.FirebaseToken)
			}
		}

		tx := config.DB.Begin()
		if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Updates(map[string]interface{}{
			"temp_id":      tempID,
			"order_radius": result.InitRange,
			"created_at":   time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR update temp_id order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to update order transaction: %v: %w", err, asynq.SkipRetry)
		}
		if err := tx.Create(&orderOffersPayload).Error; err != nil {
			tx.Rollback()
			log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR create order_offers order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to create order offers: %v: %w", err, asynq.SkipRetry)
		}
		if err := tx.Commit().Error; err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR tx commit order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to commit order offers transaction: %v: %w", err, asynq.SkipRetry)
		}

		// Push notification ke mitra
		timeoutCanTakeOrder := envMinutes("TIMEOUT_CAN_TAKE_ORDER", 1)
		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)

		if len(registrationTokenList) > 0 {
			notifType := constants.ORDER_BROADCAST
			titleOrder := "Orderan Masuk"
			messageOrder := "Ada orderan masuk nih"
			if result.IsAutoBid {
				notifType = constants.ORDER_AUTO_BID
				messageOrder = "Orderan Otomatis Masuk Ke Akun Mu"
			}
			msgData := map[string]interface{}{
				"temp_id":           tempID,
				"notification_type": notifType,
				"title":             titleOrder,
				"message":           messageOrder,
				"order_id":          orderData.ID,
				"order_type":        orderData.OrderType,
				"payment_type":      orderData.PaymentType,
				"customer_id":       orderData.CustomerID,
				"notification_id":   strconv.Itoa(orderData.NotificationID),
				"notif_type":        "order",
				"push_time":         time.Now().UTC().Format("2006-01-02 15:04:05"),
				"push_time_expired": strconv.Itoa(int(timeoutCanTakeOrder.Seconds())),
			}
			if result.IsAutoBid {
				msgData["mitra_id"] = result.AutoBidMitraID
			}
			msg := map[string]interface{}{
				"data":   msgData,
				"tokens": registrationTokenList,
			}
			if _, err := service.SendMulticast(config.DB, "mitra", msg); err != nil {
			}
		}

		// Mark search_time_complete — countdown starts from NOW for mitra
		now := time.Now()
		config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Update("search_time_complete", now)

		// PENTING: Setelah TX commit, jangan return error — akan menyebabkan Asynq retry
		// dan membuat offer duplikat dengan tempID baru. Log + lanjutkan saja.
		offerExpiredTaskPayload, err := NewOrderOfferExpiredTask(orderData.ID, tempID, orderData.CustomerID, orderData.NotificationID, timeoutFindingOrder.String())
		if err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_VA] WARN marshal offer_expired payload failed order_id=%s | %v", orderData.ID, err)
		} else if _, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferExpired, offerExpiredTaskPayload), asynq.ProcessIn(timeoutCanTakeOrder)); err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_VA] WARN enqueue offer_expired failed order_id=%s | %v (offers tetap ada di DB, mitra bisa accept)", orderData.ID, err)
		}
		// Enqueue per-mitra expiry cleanup
		for _, offer := range orderOffersPayload {
			perMitraPayload, perMitraErr := NewOrderOfferMitraExpiredTask(orderData.ID, offer.MitraID, tempID)
			if perMitraErr == nil {
				_, _ = AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferMitraExpired, perMitraPayload), asynq.ProcessIn(timeoutCanTakeOrder))
			}
		}
	} else {
		// Tidak ada mitra ditemukan — langsung ubah ke WAITING_FOR_SELECTED_MITRA
		log.Printf("[QUEUE][CORE_SEARCH_VA] No mitra found | order_id=%s | updating status to WAITING_FOR_SELECTED_MITRA", orderData.ID)
		if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Updates(map[string]interface{}{
			"order_time":   time.Now(),
			"temp_id":      tempID,
			"order_radius": result.TriedRange,
			"order_status": "WAITING_FOR_SELECTED_MITRA",
		}).Error; err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR update to WAITING_FOR_SELECTED_MITRA order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to update order status: %v: %w", err, asynq.SkipRetry)
		}

		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)
		orderSelectedExpiredTaskPayload, err := NewOrderSelectedExpiredTask(orderData.ID)
		if err == nil {
			_, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, orderSelectedExpiredTaskPayload), asynq.ProcessIn(timeoutFindingOrder))
			if err != nil {
				log.Printf("[QUEUE][CORE_SEARCH_VA] Failed to enqueue selected expired task | order_id=%s | error=%v", orderData.ID, err)
			}
		}
	}

	return nil
}

func HandleOrderOfferMitraExpiredTask(ctx context.Context, t *asynq.Task) error {
	var p OrderOfferMitraExpiredPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Delete only this mitra's offer if it still exists (temp_id ensures it's the right round)
	result := config.DB.Where("order_id = ? AND mitra_id = ? AND temp_id = ?", p.OrderID, p.MitraID, p.TempID).
		Delete(&models.OrderOffer{})
	if result.Error != nil {
		return result.Error
	}

	// Kirim CANCEL_BROADCAST ke mitra agar dismiss notif broadcast yang sudah diterima
	if result.RowsAffected > 0 {
		var mitra models.User
		if err := config.DB.Select("id, firebase_token, complete_name").
			Where("id = ? AND user_type = ?", p.MitraID, "mitra").
			First(&mitra).Error; err == nil {
			if mitra.FirebaseToken != nil && *mitra.FirebaseToken != "" {
				var orderData models.OrderTransaction
				_ = config.DB.Select("id, temp_id, customer_id, notification_id").
					Where("id = ?", p.OrderID).First(&orderData)

				fcmPayload := map[string]interface{}{
					"data": map[string]interface{}{
						"notification_type": "CANCEL_BROADCAST",
						"title":             "Order dibatalin",
						"message":           "Waktu untuk ambil orderan ini sudah habis",
						"order_id":          p.OrderID,
						"order_temp_id":     p.TempID,
						"customer_id":       orderData.CustomerID,
						"notification_id":   strconv.Itoa(orderData.NotificationID),
						"notif_type":        "order",
					},
					"tokens": []string{*mitra.FirebaseToken},
				}
				if _, err := service.SendMulticast(config.DB, "mitra", fcmPayload); err != nil {
					log.Printf("[HandleOrderOfferMitraExpiredTask] SendMulticast failed mitra_id=%s: %v", p.MitraID, err)
				}
			}
		}
	}

	return nil
}

func HandleOrderOfferExpiredTask(ctx context.Context, t *asynq.Task) error {
	var p OrderOfferExpiredPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// 1. Collect order data and mitra FCM tokens BEFORE deleting offers
	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ?", p.OrderID).First(&orderData).Error; err != nil {
	}

	var mitraIDs []string
	var orderOffers []models.OrderOffer
	if err := config.DB.Where("temp_id = ?", p.TempID).Find(&orderOffers).Error; err == nil {
		for _, offer := range orderOffers {
			mitraIDs = append(mitraIDs, offer.MitraID)
		}
	}

	// 2. Now safe to delete offers and transition order status
	if err := config.DB.Where("temp_id = ?", p.TempID).Delete(&models.OrderOffer{}).Error; err != nil {
	}

	if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ? AND order_status = ?", p.OrderID, "FINDING_MITRA").Update("order_status", "WAITING_FOR_SELECTED_MITRA").Error; err != nil {
	}

	// 3. Send expired notifications to mitras using IDs collected before deletion
	if len(mitraIDs) > 0 {
		var mitras []models.User
		if err := config.DB.Select("firebase_token").Where("id IN ?", mitraIDs).Find(&mitras).Error; err == nil {
			var firebaseTokenArray []string
			for _, m := range mitras {
				if m.FirebaseToken != nil && *m.FirebaseToken != "" && *m.FirebaseToken != "null" {
					firebaseTokenArray = append(firebaseTokenArray, *m.FirebaseToken)
				}
			}
			if len(firebaseTokenArray) > 0 {
				customerData := models.User{}
				_ = config.DB.Where("id = ?", orderData.CustomerID).First(&customerData)
				payloadMessage := map[string]interface{}{
					"data": map[string]interface{}{
						"notification_type": "ORDER_OFFER_EXPIRED",
						"title":             "Ada tawaran order yang kadaluarsa",
						"message":           fmt.Sprintf("Tawaran order dari customer %s telah kadaluarsa", customerData.CompleteName),
						"order_temp_id":     orderData.TempID,
						"order_id":          orderData.ID,
						"customer_id":       orderData.CustomerID,
						"notification_id":   strconv.Itoa(orderData.NotificationID),
						"notif_type":        "order",
					},
					"tokens": firebaseTokenArray,
				}
				if _, err := service.SendMulticast(config.DB, "mitra", payloadMessage); err != nil {
				}
			}
		}
	}

	orderSelectedExpiredTaskPayload, err := NewOrderSelectedExpiredTask(p.OrderID)
	if err == nil {
		delay, _ := time.ParseDuration(p.MinuteDifferenceSelected)
		_, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, orderSelectedExpiredTaskPayload), asynq.ProcessIn(delay))
		if err != nil {
		}
	}

	return nil
}

func HandleOrderSelectedExpiredTask(ctx context.Context, t *asynq.Task) error {
	var p OrderSelectedExpiredPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ? AND order_status = ?", p.OrderID, "WAITING_FOR_SELECTED_MITRA").First(&orderData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("failed to find order transaction: %v", err)
	}

	if err := config.DB.Model(&orderData).Update("order_status", "CANCELED_CANT_FIND_MITRA").Error; err != nil {
	}

	if orderData.PaymentType == "balance" {
		if err := config.DB.Model(&models.User{}).Where("id = ?", orderData.CustomerID).Update("account_balance", gorm.Expr("account_balance + ?", orderData.GrossAmount)).Error; err != nil {
		}
	}

	// Send FCM notification to customer
	var customerData models.User
	if err := config.DB.Select("id, firebase_token").Where("id = ? AND user_type = ?", orderData.CustomerID, "customer").First(&customerData).Error; err == nil {
		if customerData.FirebaseToken != nil && *customerData.FirebaseToken != "" {
			fcmPayload := map[string]interface{}{
				"data": map[string]string{
					"notification_type": "MITRA_NOT_FOUND_NOTIFICATION",
					"title":             "Tidak bisa dapat mitra",
					"message":           "Sayang sekali, orderanmu dibatalin karena tidak ada mitra disekitar mu :(",
					"order_temp_id":     orderData.TempID,
					"order_id":          orderData.ID,
					"customer_id":       orderData.CustomerID,
					"notification_id":   strconv.Itoa(orderData.NotificationID),
					"notif_type":        "order",
				},
			}
			if _, err := service.SendToDevice(config.DB, "customer", *customerData.FirebaseToken, fcmPayload); err != nil {
			}
		}
	} else {
	}

	return nil
}
