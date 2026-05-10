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

// HandleOrderQueueCashTask adalah handler untuk pencarian mitra pada orderan tunai.
// Algoritma estafet (sequential):
//  1. Load semua mitra dalam radius 8 km
//  2. Hitung skor dan urutkan (dua tier berdasarkan total_order)
//  3. Simpan daftar kandidat di order_transaction.candidate_mitra_ids
//  4. Kirim offer + FCM hanya ke kandidat #1 (skor tertinggi)
//  5. Enqueue timer 3 menit per-mitra; jika timeout → lanjut ke kandidat berikutnya
func HandleOrderQueueCashTask(ctx context.Context, t *asynq.Task) error {
	var p OrderQueueCashPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

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

	if orderData.OrderStatus != "FINDING_MITRA" {
		log.Printf("[QUEUE][CORE_SEARCH_CASH] Skipping order_id=%s | status=%s (not FINDING_MITRA)", orderData.ID, orderData.OrderStatus)
		return nil
	}

	params := GetNearestMitraProductionParams{
		CustomerID:         orderData.CustomerID,
		Latitude:           orderData.CustomerLatitude,
		Longitude:          orderData.CustomerLongitude,
		UserGender:         orderData.MitraGender,
		OrderType:          orderData.OrderType,
		SubPaymentID:       orderData.SubPaymentID,
		ServiceDuration:    p.ServiceDuration,
		GrossAmountCompany: float64(orderData.GrossAmountCompany),
		IsWithTime:         p.IsWithTime,
		IsCash:             orderData.PaymentType == "tunai",
	}

	queryStart := time.Now()
	result, err := GetNearestMitraProduction(params)
	if err != nil {
		log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR scoring query order_id=%s | elapsed=%s | %v",
			p.OrderID, time.Since(queryStart).Truncate(time.Millisecond), err)
		return fmt.Errorf("failed to get nearest mitra: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[QUEUE][CORE_SEARCH_CASH] order_id=%s | candidates=%d | query_elapsed=%s",
		orderData.ID, len(result.ScoredCandidates), time.Since(queryStart).Truncate(time.Millisecond))

	if len(result.ScoredCandidates) > 0 {
		candidateIDs := ExtractCandidateIDs(result)
		candidateJSON := MarshalCandidateIDs(candidateIDs)
		firstMitra := result.ScoredCandidates[0]
		tempID := uuid.New().String()
		now := time.Now()

		// Simpan kandidat dan kirim offer ke mitra #1
		tx := config.DB.Begin()
		if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Updates(map[string]interface{}{
			"temp_id":               tempID,
			"candidate_mitra_ids":   candidateJSON,
			"current_candidate_idx": 0,
			"search_time_complete":  now,
		}).Error; err != nil {
			tx.Rollback()
			log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR update candidate list order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to update order: %v: %w", err, asynq.SkipRetry)
		}
		offer := models.OrderOffer{TempID: tempID, OrderID: orderData.ID, CustomerID: orderData.CustomerID, MitraID: firstMitra.ID}
		if err := tx.Create(&offer).Error; err != nil {
			tx.Rollback()
			log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR create order_offer order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to create order offer: %v: %w", err, asynq.SkipRetry)
		}
		if err := tx.Commit().Error; err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR tx commit order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to commit: %v: %w", err, asynq.SkipRetry)
		}

		// Kirim FCM ke mitra #1
		timeoutCanTakeOrder := envMinutes("TIMEOUT_CAN_TAKE_ORDER", 3)
		sendOfferFCM(orderData, firstMitra, tempID, result.IsAutoBid, result.AutoBidMitraID, now, timeoutCanTakeOrder)

		// Enqueue timer per-mitra 3 menit; saat timeout → lanjut ke kandidat berikutnya
		if payload, err := NewOrderOfferMitraExpiredTask(orderData.ID, firstMitra.ID, tempID); err == nil {
			if _, err := AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferMitraExpired, payload), asynq.ProcessIn(timeoutCanTakeOrder)); err != nil {
				log.Printf("[QUEUE][CORE_SEARCH_CASH] WARN enqueue per-mitra timer failed order_id=%s mitra_id=%s: %v",
					orderData.ID, firstMitra.ID, err)
			}
		}

		log.Printf("[QUEUE][CORE_SEARCH_CASH] offer → mitra_id=%s (rank 1/%d) | timeout=%s | order_id=%s",
			firstMitra.ID, len(candidateIDs), timeoutCanTakeOrder, orderData.ID)
	} else {
		// Tidak ada mitra ditemukan dalam 8 km → WAITING_FOR_SELECTED_MITRA
		log.Printf("[QUEUE][CORE_SEARCH_CASH] No mitra found in 8km | order_id=%s → WAITING_FOR_SELECTED_MITRA", orderData.ID)
		if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Updates(map[string]interface{}{
			"order_status": "WAITING_FOR_SELECTED_MITRA",
			"order_time":   time.Now(),
		}).Error; err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_CASH] ERROR update status order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to update order status: %v: %w", err, asynq.SkipRetry)
		}
		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)
		if payload, err := NewOrderSelectedExpiredTask(orderData.ID); err == nil {
			if _, err := AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, payload), asynq.ProcessIn(timeoutFindingOrder)); err != nil {
				log.Printf("[QUEUE][CORE_SEARCH_CASH] Failed enqueue selected_expired | order_id=%s | %v", orderData.ID, err)
			}
		}
	}

	return nil
}

// HandleOrderQueueVATask adalah handler untuk pencarian mitra pada orderan VA dan Ewallet.
// Menggunakan algoritma estafet yang sama seperti HandleOrderQueueCashTask.
func HandleOrderQueueVATask(ctx context.Context, t *asynq.Task) error {
	var p OrderQueueVAPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

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

	if orderData.OrderStatus != "FINDING_MITRA" {
		log.Printf("[QUEUE][CORE_SEARCH_VA] Skipping order_id=%s | status=%s (not FINDING_MITRA)", orderData.ID, orderData.OrderStatus)
		return nil
	}

	params := GetNearestMitraProductionParams{
		CustomerID:         orderData.CustomerID,
		Latitude:           orderData.CustomerLatitude,
		Longitude:          orderData.CustomerLongitude,
		UserGender:         orderData.MitraGender,
		OrderType:          orderData.OrderType,
		SubPaymentID:       orderData.SubPaymentID,
		ServiceDuration:    p.ServiceDuration,
		GrossAmountCompany: float64(orderData.GrossAmountCompany),
		IsWithTime:         p.IsWithTime,
		IsCash:             orderData.PaymentType == "tunai",
	}

	queryStart := time.Now()
	result, err := GetNearestMitraProduction(params)
	if err != nil {
		log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR scoring query order_id=%s | elapsed=%s | %v",
			p.OrderID, time.Since(queryStart).Truncate(time.Millisecond), err)
		return fmt.Errorf("failed to get nearest mitra: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[QUEUE][CORE_SEARCH_VA] order_id=%s | candidates=%d | query_elapsed=%s",
		orderData.ID, len(result.ScoredCandidates), time.Since(queryStart).Truncate(time.Millisecond))

	if len(result.ScoredCandidates) > 0 {
		candidateIDs := ExtractCandidateIDs(result)
		candidateJSON := MarshalCandidateIDs(candidateIDs)
		firstMitra := result.ScoredCandidates[0]
		tempID := uuid.New().String()
		now := time.Now()

		tx := config.DB.Begin()
		if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Updates(map[string]interface{}{
			"temp_id":               tempID,
			"order_radius":          maxSearchRadiusMeters / 1000,
			"candidate_mitra_ids":   candidateJSON,
			"current_candidate_idx": 0,
			"search_time_complete":  now,
		}).Error; err != nil {
			tx.Rollback()
			log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR update order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to update order: %v: %w", err, asynq.SkipRetry)
		}
		offer := models.OrderOffer{TempID: tempID, OrderID: orderData.ID, CustomerID: orderData.CustomerID, MitraID: firstMitra.ID}
		if err := tx.Create(&offer).Error; err != nil {
			tx.Rollback()
			log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR create order_offer order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to create order offer: %v: %w", err, asynq.SkipRetry)
		}
		if err := tx.Commit().Error; err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR tx commit order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to commit: %v: %w", err, asynq.SkipRetry)
		}

		timeoutCanTakeOrder := envMinutes("TIMEOUT_CAN_TAKE_ORDER", 3)
		sendOfferFCM(orderData, firstMitra, tempID, result.IsAutoBid, result.AutoBidMitraID, now, timeoutCanTakeOrder)

		if payload, err := NewOrderOfferMitraExpiredTask(orderData.ID, firstMitra.ID, tempID); err == nil {
			if _, err := AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferMitraExpired, payload), asynq.ProcessIn(timeoutCanTakeOrder)); err != nil {
				log.Printf("[QUEUE][CORE_SEARCH_VA] WARN enqueue per-mitra timer failed order_id=%s mitra_id=%s: %v",
					orderData.ID, firstMitra.ID, err)
			}
		}

		log.Printf("[QUEUE][CORE_SEARCH_VA] offer → mitra_id=%s (rank 1/%d) | timeout=%s | order_id=%s",
			firstMitra.ID, len(candidateIDs), timeoutCanTakeOrder, orderData.ID)
	} else {
		log.Printf("[QUEUE][CORE_SEARCH_VA] No mitra found in 8km | order_id=%s → WAITING_FOR_SELECTED_MITRA", orderData.ID)
		tempID := uuid.New().String()
		if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Updates(map[string]interface{}{
			"order_time":   time.Now(),
			"temp_id":      tempID,
			"order_radius": maxSearchRadiusMeters / 1000,
			"order_status": "WAITING_FOR_SELECTED_MITRA",
		}).Error; err != nil {
			log.Printf("[QUEUE][CORE_SEARCH_VA] ERROR update status order_id=%s | %v", orderData.ID, err)
			return fmt.Errorf("failed to update order status: %v: %w", err, asynq.SkipRetry)
		}
		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)
		if payload, err := NewOrderSelectedExpiredTask(orderData.ID); err == nil {
			if _, err := AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, payload), asynq.ProcessIn(timeoutFindingOrder)); err != nil {
				log.Printf("[QUEUE][CORE_SEARCH_VA] Failed enqueue selected_expired | order_id=%s | %v", orderData.ID, err)
			}
		}
	}

	return nil
}

// sendOfferFCM mengirim push notification ke satu mitra yang sedang di-offer.
// Dipakai oleh HandleOrderQueueCashTask, HandleOrderQueueVATask, dan AdvanceToNextMitraCandidate.
func sendOfferFCM(orderData models.OrderTransaction, mitra MitraCandidate, tempID string, isAutoBid bool, autoBidMitraID string, pushTime time.Time, timeout time.Duration) {
	if mitra.FirebaseToken == nil || *mitra.FirebaseToken == "" {
		return
	}
	notifType := constants.ORDER_BROADCAST
	titleOrder := "Orderan Masuk"
	messageOrder := "Ada orderan masuk nih"
	if isAutoBid {
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
		"push_time":         pushTime.UTC().Format("2006-01-02 15:04:05"),
		"push_time_expired": strconv.Itoa(int(timeout.Seconds())),
	}
	if isAutoBid {
		msgData["mitra_id"] = autoBidMitraID
	}
	msg := map[string]interface{}{
		"data":   msgData,
		"tokens": []string{*mitra.FirebaseToken},
	}
	if _, err := service.SendMulticast(config.DB, "mitra", msg); err != nil {
		log.Printf("[sendOfferFCM] WARN FCM failed mitra_id=%s order_id=%s: %v", mitra.ID, orderData.ID, err)
	}
}

// AdvanceToNextMitraCandidate menggeser antrian ke kandidat mitra berikutnya.
//
// Dipanggil dari dua tempat:
//  1. HandleOrderOfferMitraExpiredTask — saat timer 3 menit mitra habis tanpa accept
//  2. UpdateRejectionOrderCount (mitra_service) — saat mitra aktif menolak offer
//
// Fungsi ini menggunakan optimistic locking (WHERE current_candidate_idx = expectedIdx)
// untuk mencegah double-advance jika keduanya terjadi bersamaan.
func AdvanceToNextMitraCandidate(orderID string, expectedCurrentIdx int) {
	// Load order dengan lock optimistik: WHERE id=? AND order_status='FINDING_MITRA' AND current_candidate_idx=?
	var order models.OrderTransaction
	if err := config.DB.
		Where("id = ? AND order_status = 'FINDING_MITRA' AND current_candidate_idx = ?", orderID, expectedCurrentIdx).
		Select("id, order_type, payment_type, customer_id, notification_id, candidate_mitra_ids, current_candidate_idx, mitra_gender").
		First(&order).Error; err != nil {
		// Order tidak ditemukan dengan idx yang diharapkan → sudah di-advance atau selesai
		log.Printf("[AdvanceToNextMitraCandidate] order_id=%s idx=%d: order tidak ditemukan atau sudah di-advance, skip",
			orderID, expectedCurrentIdx)
		return
	}

	candidateIDs := UnmarshalCandidateIDs(order.CandidateMitraIDs)
	nextIdx := expectedCurrentIdx + 1

	if nextIdx >= len(candidateIDs) {
		// Semua kandidat sudah dicoba → WAITING_FOR_SELECTED_MITRA
		log.Printf("[AdvanceToNextMitraCandidate] order_id=%s: semua %d kandidat habis → WAITING_FOR_SELECTED_MITRA",
			orderID, len(candidateIDs))

		result := config.DB.Model(&models.OrderTransaction{}).
			Where("id = ? AND order_status = 'FINDING_MITRA' AND current_candidate_idx = ?", orderID, expectedCurrentIdx).
			Updates(map[string]interface{}{
				"order_status":          "WAITING_FOR_SELECTED_MITRA",
				"current_candidate_idx": nextIdx,
			})
		if result.RowsAffected == 0 {
			log.Printf("[AdvanceToNextMitraCandidate] order_id=%s: update WAITING gagal (sudah di-handle), skip", orderID)
			return
		}

		timeoutFindingOrder := envMinutes("TIMEOUT_FINDING_ORDER", 2)
		if payload, err := NewOrderSelectedExpiredTask(orderID); err == nil {
			if _, err := AsynqClient.Enqueue(asynq.NewTask(TypeOrderSelectedExpired, payload), asynq.ProcessIn(timeoutFindingOrder)); err != nil {
				log.Printf("[AdvanceToNextMitraCandidate] WARN enqueue selected_expired order_id=%s: %v", orderID, err)
			}
		}
		return
	}

	// Ada kandidat berikutnya
	nextMitraID := candidateIDs[nextIdx]
	newTempID := uuid.New().String()
	now := time.Now()

	// Load data mitra berikutnya (firebase_token, is_auto_bid)
	var nextMitra models.User
	if err := config.DB.Select("id, firebase_token, is_auto_bid").
		Where("id = ?", nextMitraID).First(&nextMitra).Error; err != nil {
		log.Printf("[AdvanceToNextMitraCandidate] ERROR load mitra_id=%s order_id=%s: %v", nextMitraID, orderID, err)
		// Lanjut ke idx berikutnya
		AdvanceToNextMitraCandidate(orderID, nextIdx)
		return
	}

	// Atomic update: hanya berhasil jika idx masih sama (optimistic lock)
	result := config.DB.Model(&models.OrderTransaction{}).
		Where("id = ? AND order_status = 'FINDING_MITRA' AND current_candidate_idx = ?", orderID, expectedCurrentIdx).
		Updates(map[string]interface{}{
			"temp_id":               newTempID,
			"current_candidate_idx": nextIdx,
			"search_time_complete":  now,
		})

	if result.Error != nil {
		log.Printf("[AdvanceToNextMitraCandidate] ERROR update order_id=%s: %v", orderID, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		log.Printf("[AdvanceToNextMitraCandidate] order_id=%s idx=%d: update tidak efektif (sudah di-advance), skip",
			orderID, expectedCurrentIdx)
		return
	}

	// Buat offer baru untuk mitra berikutnya
	newOffer := models.OrderOffer{
		TempID:     newTempID,
		OrderID:    orderID,
		CustomerID: order.CustomerID,
		MitraID:    nextMitraID,
	}
	if err := config.DB.Create(&newOffer).Error; err != nil {
		log.Printf("[AdvanceToNextMitraCandidate] ERROR create offer order_id=%s mitra_id=%s: %v", orderID, nextMitraID, err)
		return
	}

	// Kirim FCM ke mitra berikutnya
	timeoutCanTakeOrder := envMinutes("TIMEOUT_CAN_TAKE_ORDER", 3)
	mitraCandidate := MitraCandidate{ID: nextMitra.ID, FirebaseToken: nextMitra.FirebaseToken, IsAutoBid: nextMitra.IsAutoBid}
	isAutoBidNext := nextMitra.IsAutoBid == "yes"
	autoBidIDNext := ""
	if isAutoBidNext {
		autoBidIDNext = nextMitra.ID
	}
	sendOfferFCM(order, mitraCandidate, newTempID, isAutoBidNext, autoBidIDNext, now, timeoutCanTakeOrder)

	// Enqueue timer per-mitra 3 menit untuk kandidat berikutnya
	if payload, err := NewOrderOfferMitraExpiredTask(orderID, nextMitraID, newTempID); err == nil {
		if _, err := AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferMitraExpired, payload), asynq.ProcessIn(timeoutCanTakeOrder)); err != nil {
			log.Printf("[AdvanceToNextMitraCandidate] WARN enqueue timer order_id=%s mitra_id=%s: %v", orderID, nextMitraID, err)
		}
	}

	log.Printf("[AdvanceToNextMitraCandidate] order_id=%s: offer → mitra_id=%s (rank %d/%d) | timeout=%s",
		orderID, nextMitraID, nextIdx+1, len(candidateIDs), timeoutCanTakeOrder)
}

// HandleOrderOfferMitraExpiredTask dipanggil saat timer 3 menit per-mitra habis
// tanpa konfirmasi accept dari mitra.
//
// Alur:
//  1. Hapus offer mitra ini (jika masih ada, artinya mitra tidak accept)
//  2. Kirim CANCEL_BROADCAST ke mitra agar dismiss notif di app-nya
//  3. Panggil AdvanceToNextMitraCandidate untuk menggeser ke kandidat berikutnya
func HandleOrderOfferMitraExpiredTask(ctx context.Context, t *asynq.Task) error {
	var p OrderOfferMitraExpiredPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Hapus offer mitra ini. Jika RowsAffected == 0 berarti mitra sudah accept → tidak perlu lanjut.
	deleteResult := config.DB.Where("order_id = ? AND mitra_id = ? AND temp_id = ?", p.OrderID, p.MitraID, p.TempID).
		Delete(&models.OrderOffer{})
	if deleteResult.Error != nil {
		return deleteResult.Error
	}

	if deleteResult.RowsAffected == 0 {
		// Offer sudah dihapus sebelumnya (mitra accept atau reject manual) → skip advance
		log.Printf("[HandleOrderOfferMitraExpiredTask] offer sudah tidak ada order_id=%s mitra_id=%s | skip", p.OrderID, p.MitraID)
		return nil
	}

	// Kirim CANCEL_BROADCAST ke mitra agar dismiss notif broadcast yang sudah diterima
	var mitra models.User
	if err := config.DB.Select("id, firebase_token").Where("id = ?", p.MitraID).First(&mitra).Error; err == nil {
		if mitra.FirebaseToken != nil && *mitra.FirebaseToken != "" {
			var orderData models.OrderTransaction
			_ = config.DB.Select("id, temp_id, customer_id, notification_id").Where("id = ?", p.OrderID).First(&orderData)
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
				log.Printf("[HandleOrderOfferMitraExpiredTask] WARN FCM failed mitra_id=%s: %v", p.MitraID, err)
			}
		}
	}

	// Load current_candidate_idx untuk optimistic lock
	var order models.OrderTransaction
	if err := config.DB.Select("current_candidate_idx").Where("id = ?", p.OrderID).First(&order).Error; err != nil {
		log.Printf("[HandleOrderOfferMitraExpiredTask] ERROR load order order_id=%s: %v", p.OrderID, err)
		return nil // Jangan return error — task sudah selesai secara semantik
	}

	// Estafet ke kandidat berikutnya
	AdvanceToNextMitraCandidate(p.OrderID, order.CurrentCandidateIdx)

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
