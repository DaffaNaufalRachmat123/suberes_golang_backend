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
	"golang.org/x/sync/errgroup"
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

	log.Printf("Handling order queue cash task for order_id=%s", p.OrderID)

	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ?", p.OrderID).First(&orderData).Error; err != nil {
		return fmt.Errorf("failed to find order transaction with id %s: %v", p.OrderID, err)
	}

	// Fetch customer, sub_service, service in parallel — they are independent once we have orderData.
	var customerData models.User
	var subServiceData models.SubService
	var serviceData models.Service

	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return config.DB.Where("id = ? AND user_type = ?", orderData.CustomerID, "customer").First(&customerData).Error
	})
	eg.Go(func() error {
		return config.DB.Where("id = ?", orderData.SubServiceID).First(&subServiceData).Error
	})
	eg.Go(func() error {
		return config.DB.Where("id = ?", orderData.ServiceID).First(&serviceData).Error
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to fetch order dependencies: %v", err)
	}

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
		ServiceDuration:      subServiceData.MinutesSubServices,
		CustomerTimezoneCode: orderData.TimezoneCode,
		CustomerTimeOrder:    orderData.OrderTime.String(),
		GrossAmountCompany:   float64(orderData.GrossAmountCompany),
		IsWithTime:           serviceData.ServiceType == "Durasi",
		IsCash:               orderData.PaymentType == "tunai",
		Limit:                10,
		Page:                 0,
	}

	result, err := GetNearestMitraProduction(params)
	if err != nil {
		return fmt.Errorf("failed to get nearest mitra: %v", err)
	}

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
			return fmt.Errorf("failed to update order transaction temp_id: %v", err)
		}

		if err := tx.Create(&orderOffersPayload).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create order offers: %v", err)
		}
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit order offers transaction: %v", err)
		}

		log.Printf("Sending push notifications to tokens: %v", registrationTokenList)

		// --- Kirim push notification ke mitra ---
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
			}
			if result.IsAutoBid {
				msgData["mitra_id"] = result.AutoBidMitraID
			}
			msg := map[string]interface{}{
				"data":   msgData,
				"tokens": registrationTokenList,
			}
			if _, err := service.SendMulticast(config.DB, "mitra", msg); err != nil {
				log.Printf("failed to send push notification to mitra: %v", err)
			} else {
				log.Printf("Successfully sent push notification to mitra")
			}
		} else {
			log.Printf("No tokens to send push notification to")
		}

		// ---

		timeoutCanTakeOrder := envMinutes("TIMEOUT_CAN_TAKE_ORDER", 1)
		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)

		offerExpiredTaskPayload, err := NewOrderOfferExpiredTask(orderData.ID, tempID, orderData.CustomerID, orderData.NotificationID, timeoutFindingOrder.String())
		if err != nil {
			return fmt.Errorf("failed to create offer expired task payload: %v", err)
		}
		if _, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferExpired, offerExpiredTaskPayload), asynq.ProcessIn(timeoutCanTakeOrder)); err != nil {
			return fmt.Errorf("could not enqueue offer expired task: %v", err)
		}
	} else {
		if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Update("order_status", "WAITING_FOR_SELECTED_MITRA").Error; err != nil {
			return fmt.Errorf("failed to update order status: %v", err)
		}
		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)
		orderSelectedExpiredTaskPayload, err := NewOrderSelectedExpiredTask(orderData.ID)
		if err == nil {
			_, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, orderSelectedExpiredTaskPayload), asynq.ProcessIn(timeoutFindingOrder))
			if err != nil {
				log.Printf("could not enqueue task: %v", err)
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

	log.Printf("Handling order queue VA task for order_id=%s", p.OrderID)

	// 1. Fetch order, customer, service, sub_service, repeats
	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ?", p.OrderID).First(&orderData).Error; err != nil {
		return fmt.Errorf("failed to find order transaction with id %s: %v", p.OrderID, err)
	}

	// Fetch customer, service, sub_service, repeats in parallel.
	var customerData models.User
	var serviceData models.Service
	var subServiceData models.SubService
	var repeatsData []models.OrderTransactionRepeat

	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return config.DB.Where("id = ? AND user_type = ?", orderData.CustomerID, "customer").First(&customerData).Error
	})
	eg.Go(func() error {
		return config.DB.Where("id = ?", orderData.ServiceID).First(&serviceData).Error
	})
	eg.Go(func() error {
		return config.DB.Where("id = ?", orderData.SubServiceID).First(&subServiceData).Error
	})
	eg.Go(func() error {
		return config.DB.Where("order_id = ?", p.OrderID).Find(&repeatsData).Error
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to fetch order dependencies: %v", err)
	}

	initialRange, _ := strconv.ParseFloat(os.Getenv("INITIAL_RANGE_CUSTOMER"), 64)
	if initialRange == 0 {
		initialRange = 1
	}
	maxRange, _ := strconv.ParseFloat(os.Getenv("MAX_RANGE_CUSTOMER"), 64)
	if maxRange == 0 {
		maxRange = 10
	}

	// 2. Cari mitra terdekat (pakai core logic yang sudah ada)
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
		ServiceDuration:      subServiceData.MinutesSubServices,
		CustomerTimezoneCode: orderData.TimezoneCode,
		CustomerTimeOrder:    orderData.OrderTime.String(),
		JsonOrderTimes:       repeatsData,
		GrossAmountCompany:   float64(orderData.GrossAmountCompany),
		IsWithTime:           serviceData.ServiceType == "Durasi",
		IsCash:               orderData.PaymentType == "tunai",
		Limit:                10,
		Page:                 0,
	}

	result, err := GetNearestMitraProduction(params)
	if err != nil {
		return fmt.Errorf("failed to get nearest mitra: %v", err)
	}

	tempID := uuid.New().String()
	var orderOffersPayload []models.OrderOffer
	var registrationTokenList []string

	if len(result.PayloadMitra) > 0 {
		// 3. Ada mitra ditemukan
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
			return fmt.Errorf("failed to update order transaction: %v", err)
		}
		if err := tx.Create(&orderOffersPayload).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create order offers: %v", err)
		}
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit order offers transaction: %v", err)
		}

		// Push notification ke mitra
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
			}
			if result.IsAutoBid {
				msgData["mitra_id"] = result.AutoBidMitraID
			}
			msg := map[string]interface{}{
				"data":   msgData,
				"tokens": registrationTokenList,
			}
			if _, err := service.SendMulticast(config.DB, "mitra", msg); err != nil {
				log.Printf("failed to send push notification to mitra: %v", err)
			}
		}

		// Delay untuk offer expired dan selected expired
		timeoutCanTakeOrder := envMinutes("TIMEOUT_CAN_TAKE_ORDER", 1)
		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)

		offerExpiredTaskPayload, err := NewOrderOfferExpiredTask(orderData.ID, tempID, orderData.CustomerID, orderData.NotificationID, timeoutFindingOrder.String())
		if err != nil {
			return fmt.Errorf("failed to create offer expired task payload: %v", err)
		}
		if _, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferExpired, offerExpiredTaskPayload), asynq.ProcessIn(timeoutCanTakeOrder)); err != nil {
			return fmt.Errorf("could not enqueue offer expired task: %v", err)
		}
	} else {
		// 4. Tidak ada mitra ditemukan
		if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Updates(map[string]interface{}{
			"order_time":   time.Now(),
			"temp_id":      tempID,
			"order_radius": result.TriedRange,
			"order_status": "WAITING_FOR_SELECTED_MITRA",
		}).Error; err != nil {
			return fmt.Errorf("failed to update order status: %v", err)
		}

		log.Printf("No mitra found for order %s, notifying admin", orderData.ID)

		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)
		orderSelectedExpiredTaskPayload, err := NewOrderSelectedExpiredTask(orderData.ID)
		if err == nil {
			_, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, orderSelectedExpiredTaskPayload), asynq.ProcessIn(timeoutFindingOrder))
			if err != nil {
				log.Printf("could not enqueue selected expired task: %v", err)
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

	log.Printf("Handling order offer expired task for order_id=%s", p.OrderID)

	// 1. Collect order data and mitra FCM tokens BEFORE deleting offers
	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ?", p.OrderID).First(&orderData).Error; err != nil {
		log.Printf("failed to fetch order for offer expired task: %v", err)
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
		log.Printf("failed to delete order offers: %v", err)
	}

	if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ? AND order_status = ?", p.OrderID, "FINDING_MITRA").Update("order_status", "WAITING_FOR_SELECTED_MITRA").Error; err != nil {
		log.Printf("failed to update order transaction status: %v", err)
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
					log.Printf("failed to send expired push notification to mitra: %v", err)
				}
			}
		}
	}

	log.Printf("Notifying mitras about expired offer for order %s", p.OrderID)
	log.Printf("Notifying admin about order %s waiting for selection", p.OrderID)

	orderSelectedExpiredTaskPayload, err := NewOrderSelectedExpiredTask(p.OrderID)
	if err == nil {
		delay, _ := time.ParseDuration(p.MinuteDifferenceSelected)
		_, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, orderSelectedExpiredTaskPayload), asynq.ProcessIn(delay))
		if err != nil {
			log.Printf("could not enqueue task: %v", err)
		}
	}

	return nil
}

func HandleOrderSelectedExpiredTask(ctx context.Context, t *asynq.Task) error {
	var p OrderSelectedExpiredPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Handling order selected expired task for order_id=%s", p.OrderID)

	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ? AND order_status = ?", p.OrderID, "WAITING_FOR_SELECTED_MITRA").First(&orderData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Order %s not found or not in WAITING_FOR_SELECTED_MITRA state", p.OrderID)
			return nil
		}
		return fmt.Errorf("failed to find order transaction: %v", err)
	}

	if err := config.DB.Model(&orderData).Update("order_status", "CANCELED_CANT_FIND_MITRA").Error; err != nil {
		log.Printf("failed to update order status to canceled: %v", err)
	}

	if orderData.PaymentType == "balance" {
		if err := config.DB.Model(&models.User{}).Where("id = ?", orderData.CustomerID).Update("account_balance", gorm.Expr("account_balance + ?", orderData.GrossAmount)).Error; err != nil {
			log.Printf("failed to refund customer balance: %v", err)
		}
	}

	log.Printf("Notifying customer about order cancellation for order %s", p.OrderID)
	log.Printf("Notifying admin about order timeout for order %s", p.OrderID)

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
				log.Printf("failed to send FCM to customer %s: %v", orderData.CustomerID, err)
			}
		}
	} else {
		log.Printf("failed to fetch customer firebase token for order selected expired: %v", err)
	}

	return nil
}
