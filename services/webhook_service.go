package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/queue"
	"suberes_golang/repositories"
	"suberes_golang/service"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// WebhookService handles all inbound Xendit webhook events.
// Each public method maps to exactly one Xendit event type / webhook URL:
//
//	HandleVACreate    → POST /api/webhook/va/create    (FVA created/updated)
//	HandleVAPaid      → POST /api/webhook/va/paid      (FVA payment received)
//	HandleDisbursement→ POST /api/webhook/disbursement (disbursement completed/failed)
//	HandleEwallet     → POST /api/webhook/ewallet      (ewallet.capture / ewallet.void)
type WebhookService struct {
	DB                         *gorm.DB
	UserRepo                   *repositories.UserRepository
	OrderTransactionRepo       *repositories.OrderTransactionRepository
	OrderTransactionRepeatRepo *repositories.OrderTransactionRepeatsRepository
	TransactionRepo            *repositories.TransactionRepository
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{
		DB:                         db,
		UserRepo:                   &repositories.UserRepository{DB: db},
		OrderTransactionRepo:       &repositories.OrderTransactionRepository{DB: db},
		OrderTransactionRepeatRepo: &repositories.OrderTransactionRepeatsRepository{DB: db},
		TransactionRepo:            &repositories.TransactionRepository{DB: db},
	}
}

// ─── VA Create ────────────────────────────────────────────────────────────────

// HandleVACreate handles the Xendit "FVA Created/Updated" webhook event.
// Route: POST /api/webhook/va/create
//
// Routing by external_id prefix:
//   - "Order-"        → update order PROCESSING_PAYMENT → WAITING_PAYMENT, notify customer
//   - "Topup-"        → skip (VA activation requires no state change for topups)
//   - "Disbursement-" → skip
func (s *WebhookService) HandleVACreate(body map[string]interface{}) (int, error) {
	externalID, _ := body["external_id"].(string)
	if externalID == "" {
		externalID, _ = body["id"].(string)
	}
	name, _ := body["name"].(string)

	if strings.HasPrefix(externalID, "Topup-") || strings.HasPrefix(externalID, "Disbursement-") {
		return http.StatusOK, nil
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var orderData models.OrderTransaction
	if err := s.DB.
		Where("external_id = ? AND name = ?", externalID, name).
		First(&orderData).Error; err != nil {
		tx.Rollback()
		return http.StatusNotFound, errors.New("order data not found")
	}

	var customerData models.User
	if err := s.DB.
		Where("id = ? AND user_type = ?", orderData.CustomerID, "customer").
		First(&customerData).Error; err != nil {
		tx.Rollback()
		return http.StatusNotFound, errors.New("customer not found")
	}

	if err := tx.Model(&models.OrderTransaction{}).
		Where("va_id = ? AND external_id = ? AND name = ?", externalID, externalID, name).
		Updates(map[string]interface{}{
			"xendit_status": body["status"],
			"order_status":  "WAITING_PAYMENT",
			"order_time":    orderData.OrderTimeTemp,
		}).Error; err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, err
	}

	if customerData.FirebaseToken != nil && *customerData.FirebaseToken != "" {
		service.SendMulticast(s.DB, "customer", map[string]interface{}{
			"data": map[string]string{
				"notification_type": "PAYMENT_PROCEED",
				"title":             "Pembayaran telah diproses",
				"message":           "Kamu bisa melakukan pembayaran sekarang",
				"order_id":          orderData.ID,
				"customer_id":       orderData.CustomerID,
				"notif_type":        "order",
			},
			"tokens": []string{*customerData.FirebaseToken},
		})
	}

	tx.Create(&models.Notification{
		ID:                  uuid.New().String(),
		CustomerID:          orderData.CustomerID,
		OrderID:             orderData.ID,
		NotificationType:    "PAYMENT_PROCEED",
		NotificationTitle:   "Pembayaran telah diproses",
		NotificationMessage: "Kamu bisa melakukan pembayaran sekarang",
	})

	if err := tx.Commit().Error; err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// ─── VA Paid ──────────────────────────────────────────────────────────────────

// HandleVAPaid handles the Xendit "FVA Payment Received" webhook event.
// Route: POST /api/webhook/va/paid
//
// Routing by external_id prefix:
//   - "Topup-"  → credit user balance (mitra or customer)
//   - otherwise → order WAITING_PAYMENT → FINDING_MITRA (+ queue + FCM)
func (s *WebhookService) HandleVAPaid(body map[string]interface{}) (int, error) {
	externalID, _ := body["external_id"].(string)
	if externalID == "" {
		externalID, _ = body["id"].(string)
	}

	if strings.HasPrefix(externalID, "Topup-") {
		return s.handleVATopupPaid(body, externalID)
	}

	// Order VA paid flow
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	logTime := now.Format("2006-01-02 15:04:05")

	var orderData models.OrderTransaction
	if err := s.DB.
		Where("external_id = ? AND xendit_status = ?", externalID, "ACTIVE").
		First(&orderData).Error; err != nil {
		tx.Rollback()
		return http.StatusNotFound, errors.New("order data not found")
	}

	var customerData models.User
	if err := s.DB.Where("id = ?", orderData.CustomerID).First(&customerData).Error; err != nil {
		tx.Rollback()
		return http.StatusNotFound, errors.New("customer not found")
	}

	if err := tx.Model(&models.OrderTransaction{}).
		Where("external_id = ? AND order_status = ?", externalID, "WAITING_PAYMENT").
		Updates(map[string]interface{}{
			"order_status":     "FINDING_MITRA",
			"is_paid_customer": "1",
			"xendit_status":    "INACTIVE",
		}).Error; err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, err
	}

	switch orderData.OrderType {
	case "repeat":
		tx.Model(&models.OrderTransactionRepeat{}).
			Where("order_id = ?", orderData.ID).
			Update("order_status", "WAIT_SCHEDULE")

		var repeats []models.OrderTransactionRepeat
		s.DB.Where("order_id = ?", orderData.ID).Find(&repeats)
		for _, rep := range repeats {
			runAt := rep.OrderTime
			warningAt := runAt.Add(3 * time.Minute)
			rp, _ := queue.NewOrderComingSoonRunTask(orderData.ID)
			queue.ScheduleOnceAt(queue.TypeOrderComingSoonRun, rp, runAt)
			wp, _ := queue.NewOrderComingSoonWarningTask(orderData.ID)
			queue.ScheduleOnceAt(queue.TypeOrderComingSoonWarning, wp, warningAt)
		}
		if customerData.FirebaseToken != nil && *customerData.FirebaseToken != "" {
			service.SendMulticast(s.DB, "customer", map[string]interface{}{
				"data": map[string]string{
					"notification_type": "REPEAT_VIRTUAL_ACCOUNT_ORDER_PAID_NOTIFICATION",
					"title":             "Pembayaran Order Berulang Berhasil",
					"message":           fmt.Sprintf("Pembayaran order berulang dengan ID Transaksi %s berhasil", orderData.IDTransaction),
					"order_id":          orderData.ID,
					"sub_order_id":      "-1",
					"customer_id":       orderData.CustomerID,
					"notif_type":        "order",
				},
				"tokens": []string{*customerData.FirebaseToken},
			})
		}

	case "coming soon":
		scheduleAt := orderData.OrderTime
		warningAt := scheduleAt.Add(3 * time.Minute)
		rp, _ := queue.NewOrderComingSoonRunTask(orderData.ID)
		queue.ScheduleOnceAt(queue.TypeOrderComingSoonRun, rp, scheduleAt)
		wp, _ := queue.NewOrderComingSoonWarningTask(orderData.ID)
		queue.ScheduleOnceAt(queue.TypeOrderComingSoonWarning, wp, warningAt)

		if customerData.FirebaseToken != nil && *customerData.FirebaseToken != "" {
			service.SendMulticast(s.DB, "customer", map[string]interface{}{
				"data": map[string]string{
					"notification_type": "COMING_SOON_VIRTUAL_ACCOUNT_ORDER_PAID_NOTIFICATION",
					"title":             "Pembayaran Order Terjadwal Berhasil",
					"message":           fmt.Sprintf("Pembayaran order terjadwal dengan ID Transaksi : %s berhasil", orderData.IDTransaction),
					"order_id":          orderData.ID,
					"sub_order_id":      "-1",
					"customer_id":       customerData.ID,
					"notif_type":        "order",
				},
				"tokens": []string{*customerData.FirebaseToken},
			})
		}

	case "now":
		if customerData.FirebaseToken != nil && *customerData.FirebaseToken != "" {
			service.SendMulticast(s.DB, "customer", map[string]interface{}{
				"data": map[string]string{
					"notification_type": "NOW_VIRTUAL_ACCOUNT_ORDER_PAID_NOTIFICATION",
					"title":             "Pembayaran Order Berhasil",
					"message":           fmt.Sprintf("Pembayaran order dengan ID Transaksi : %s berhasil dan kami sedang mencarikan mitra untukmu", orderData.IDTransaction),
					"order_id":          orderData.ID,
					"sub_order_id":      "-1",
					"customer_id":       orderData.CustomerID,
					"notif_type":        "order",
				},
				"tokens": []string{*customerData.FirebaseToken},
			})
		}
	}

	taskPayload, _ := queue.NewOrderQueueVATask(orderData.ID)
	queue.AsynqClient.Enqueue(asynq.NewTask(queue.TypeOrderQueueVA, taskPayload), asynq.Queue("critical"))

	s.DB.Create(&models.SuberesLogs{
		LogName: "Notification Paid VA Order",
		LogType: "Paid VA Order",
		LogURL:  fmt.Sprintf("%+v", body),
		LogTime: logTime,
	})

	if err := tx.Commit().Error; err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// handleVATopupPaid processes a VA paid event where external_id has a "Topup-" prefix.
// Validates bank_code, amount, account_number, currency; credits user balance; sends FCM.
// Supports both mitra and customer user types.
func (s *WebhookService) handleVATopupPaid(body map[string]interface{}, externalID string) (int, error) {
	bankCode, _ := body["bank_code"].(string)
	amount, _ := body["amount"].(float64)
	accountNumber, _ := body["account_number"].(string)
	currency, _ := body["currency"].(string)

	trx, err := s.TransactionRepo.FindTopupTransactionByExternalIDForCallback(externalID)
	if err != nil {
		log.Printf("[WebhookService] handleVATopupPaid: transaction not found for externalID=%s: %v", externalID, err)
		return http.StatusNotFound, fmt.Errorf("transaction not found: %w", err)
	}

	if bankCode != "" && !strings.EqualFold(bankCode, trx.BankCode) {
		return http.StatusBadRequest, errors.New("bank code not same")
	}
	if amount != 0 && int64(amount) != trx.TransactionAmount {
		return http.StatusBadRequest, errors.New("transaction amount not same")
	}
	if accountNumber != "" && accountNumber != trx.AccountNumber {
		return http.StatusBadRequest, errors.New("account number not same")
	}
	if currency != "" && currency != "IDR" {
		return http.StatusBadRequest, errors.New("currency not IDR")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return http.StatusInternalServerError, tx.Error
	}

	netAmount := trx.TransactionAmount - trx.TransactionFee

	var user *models.User
	if trx.UserType == "mitra" {
		user, err = s.UserRepo.FindMitraById(helpers.DerefStr(trx.MitraID))
	} else if trx.UserType == "customer" {
		user, err = s.UserRepo.FindCustomerById(helpers.DerefStr(trx.CustomerID))
	} else {
		tx.Rollback()
		return http.StatusBadRequest, errors.New("unknown user type")
	}
	if err != nil {
		tx.Rollback()
		log.Printf("[WebhookService] handleVATopupPaid: user not found (type=%s): %v", trx.UserType, err)
		return http.StatusNotFound, fmt.Errorf("%s not found: %w", trx.UserType, err)
	}

	if err := s.UserRepo.UpdateUserBalance(tx, user.ID, netAmount); err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, err
	}

	lastAmount := user.AccountBalance + trx.TransactionAmount
	if err := s.TransactionRepo.UpdateTopupSuccess(tx, externalID, lastAmount); err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		return http.StatusInternalServerError, err
	}

	if user.FirebaseToken != nil && *user.FirebaseToken != "" {
		var fcmData map[string]string
		if trx.UserType == "mitra" {
			fcmData = map[string]string{
				"notification_type":  "TOPUP_NOTIFICATION",
				"topup_status":       "TOPUP_SUCCESSFUL",
				"title":              "Status Top Up",
				"id":                 trx.ID,
				"mitra_id":           user.ID,
				"transaction_status": "success",
				"transaction_id":     trx.ID,
				"transaction_amount": fmt.Sprintf("%d", trx.TransactionAmount),
				"idempotency_key":    trx.IdempotencyKey,
				"message":            fmt.Sprintf("Berhasil Top Up dengan %s Sebesar Rp %d", trx.BankName, netAmount),
				"notif_type":         "topup",
			}
		} else {
			fcmData = map[string]string{
				"notification_type":  "TOPUP_NOTIFICATION",
				"topup_status":       "TOPUP_SUCCESSFUL",
				"title":              "Status Top Up",
				"id":                 trx.ID,
				"customer_id":        user.ID,
				"transaction_id":     trx.ID,
				"transaction_status": "success",
				"transaction_amount": fmt.Sprintf("%d", netAmount),
				"idempotency_key":    trx.IdempotencyKey,
				"message":            fmt.Sprintf("Berhasil Top Up dengan %s Sebesar Rp %d", trx.BankName, netAmount),
				"notif_type":         "topup",
				"mitra_id":           user.ID,
			}
		}
		if _, err := service.SendToDevice(s.DB, trx.UserType, *user.FirebaseToken, map[string]interface{}{"data": fcmData}); err != nil {
			log.Printf("[FCM] TOPUP_NOTIFICATION error for %s %s: %v", trx.UserType, user.ID, err)
		}
	}

	log.Printf("[WebhookService] handleVATopupPaid: success for trxID=%s", trx.ID)
	return http.StatusOK, nil
}

// ─── Disbursement ─────────────────────────────────────────────────────────────

// HandleDisbursement handles the Xendit disbursement (outgoing bank transfer) result webhook.
// Route: POST /api/webhook/disbursement
//
//   - status "COMPLETED" → mark transaction success
//   - status "FAILED"    → refund user balance, mark transaction failed, send FCM
func (s *WebhookService) HandleDisbursement(payload *dtos.DisbursementCallbackPayload) error {
	payloadBytes, _ := json.MarshalIndent(payload, "", "  ")
	log.Printf("[WebhookService] HandleDisbursement PAYLOAD: %s", string(payloadBytes))
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		log.Printf("[WebhookService] HandleDisbursement: tx begin error: %v", tx.Error)
		return tx.Error
	}

	trx, err := s.TransactionRepo.FindPendingDisbursementByExternalID(payload.ID)
	if err != nil {
		tx.Rollback()
		log.Printf("[WebhookService] HandleDisbursement: transaction not found for id=%s", payload.ID)
		return nil // not found → return 200, Xendit will not retry
	}

	var user *models.User
	if trx.UserType == "mitra" {
		user, err = s.UserRepo.FindMitraById(helpers.DerefStr(trx.MitraID))
	} else if trx.UserType == "customer" {
		user, err = s.UserRepo.FindCustomerById(helpers.DerefStr(trx.CustomerID))
	}
	if err != nil {
		tx.Rollback()
		log.Printf("[WebhookService] HandleDisbursement: %s not found: %v", trx.UserType, err)
		return fmt.Errorf("%s not found: %w", trx.UserType, err)
	}

	if payload.Status == "FAILED" {
		refundAmount := trx.TransactionAmount - trx.TransactionFee
		if err := s.TransactionRepo.UpdateDisbursementFailure(tx, trx.ID, payload.FailureCode, refundAmount); err != nil {
			tx.Rollback()
			return err
		}
		if err := s.UserRepo.UpdateUserBalance(tx, user.ID, refundAmount); err != nil {
			tx.Rollback()
			return err
		}
	} else if payload.Status == "COMPLETED" {
		if err := s.TransactionRepo.UpdateDisbursementStatus(tx, trx.ID, "success"); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	var customerBalance int64
	if trx.UserType == "customer" && payload.Status == "FAILED" {
		if updatedCustomer, err := s.UserRepo.FindCustomerById(user.ID); err == nil {
			customerBalance = updatedCustomer.AccountBalance
		}
	}

	if user.FirebaseToken != nil && *user.FirebaseToken != "" {
		titleMsg := "Tarik Tunai Kamu Berhasil"
		if payload.Status != "COMPLETED" {
			titleMsg = "Tarik Tunai Kamu Gagal"
		}
		txStatusStr := "success"
		if payload.Status == "FAILED" {
			txStatusStr = "cancelled"
		}

		var fcmData map[string]string
		if trx.UserType == "mitra" {
			var msgMsg string
			if payload.Status == "COMPLETED" {
				msgMsg = fmt.Sprintf("Tarik tunai kamu ke rekening %s sebesar Rp.%d berhasil", trx.BankName, trx.TransactionAmount)
			} else {
				msgMsg = fmt.Sprintf("Tarik tunai kamu ke rekening %s sebesar Rp.%d gagal", trx.BankName, trx.TransactionAmount)
			}
			fcmData = map[string]string{
				"notification_type":          "DISBURSEMENT_STATUS",
				"title":                      titleMsg,
				"message":                    msgMsg,
				"transaction_status":         payload.Status,
				"disbursement_status":        payload.Status,
				"disbursement_failed_status": txStatusStr,
				"id":                         trx.ID,
				"mitra_id":                   user.ID,
				"transaction_id":             trx.ID,
				"disbursement_id":            trx.DisbursementID,
				"idempotency_key":            trx.IdempotencyKey,
				"notif_type":                 "disbursement",
			}
		} else if trx.UserType == "customer" {
			var msgMsg string
			if payload.Status == "COMPLETED" {
				msgMsg = fmt.Sprintf("Tarik tunai kamu ke rekening %s sebesar Rp.%d berhasil", trx.BankName, trx.TransactionAmount)
			} else {
				msgMsg = fmt.Sprintf("Tarik tunai kamu ke rekening %s gagal", trx.BankName)
			}
			fcmData = map[string]string{
				"notification_type":          "DISBURSEMENT_STATUS",
				"title":                      titleMsg,
				"message":                    msgMsg,
				"disbursement_status":        txStatusStr,
				"disbursement_failed_status": txStatusStr,
				"id":                         trx.ID,
				"customer_id":                user.ID,
				"transaction_id":             trx.ID,
				"account_balance":            fmt.Sprintf("%d", customerBalance),
				"transaction_status":         payload.Status,
				"disbursement_id":            trx.DisbursementID,
				"notif_type":                 "disbursement",
			}
		}
		if fcmData != nil {
			if _, err := service.SendToDevice(s.DB, trx.UserType, *user.FirebaseToken, map[string]interface{}{"data": fcmData}); err != nil {
				log.Printf("[FCM] DISBURSEMENT_STATUS error for %s %s: %v", trx.UserType, user.ID, err)
			}
		}
	}

	log.Printf("[WebhookService] HandleDisbursement: success for trxID=%s", trx.ID)
	return nil
}

// ─── eWallet ──────────────────────────────────────────────────────────────────

// HandleEwallet handles the Xendit eWallet charge webhook event.
// Route: POST /api/webhook/ewallet
//
//   - event "ewallet.capture" → order: WAITING_PAYMENT → FINDING_MITRA (or WAIT_SCHEDULE)
//   - event "ewallet.void"    → order: → CANCELED_VOID, repeat orders updated, FCM sent
func (s *WebhookService) HandleEwallet(payload dtos.XenditCallbackPayload) (int, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	logTime := now.Format("2006-01-02 15:04:05")

	log.Printf("[WebhookService] HandleEwallet: event=%s data=%s", payload.Event, func() string {
		b, _ := json.Marshal(payload.Data)
		return string(b)
	}())

	switch payload.Event {
	case "ewallet.capture":
		orderData, err := s.OrderTransactionRepo.FindDynamicOrderTransactionMap(
			nil,
			nil,
			"payment_id_pay = ?",
			[]interface{}{payload.Data.ID},
		)
		if err != nil || orderData == nil {
			tx.Rollback()
			return http.StatusNotFound, errors.New("order not found")
		}

		orderID := fmt.Sprintf("%v", orderData["id"])
		customerID := fmt.Sprintf("%v", orderData["customer_id"])
		orderType := fmt.Sprintf("%v", orderData["order_type"])

		if payload.Data.Status == "SUCCEEDED" {
			updates := map[string]interface{}{"is_paid_customer": "1"}

			switch orderType {
			case "now":
				updates["order_status"] = "FINDING_MITRA"
				updates["order_time"] = now.UTC()
				if err := tx.Model(&models.OrderTransaction{}).
					Where("payment_id_pay = ? AND order_status = ?", payload.Data.ID, "WAITING_PAYMENT").
					Updates(updates).Error; err != nil {
					tx.Rollback()
					return http.StatusInternalServerError, err
				}
				taskPayload, _ := queue.NewOrderQueueVATask(orderID)
				queue.AsynqClient.Enqueue(asynq.NewTask(queue.TypeOrderQueueVA, taskPayload), asynq.Queue("critical"))
				s.webhookPushFCM("customer", customerID,
					"NOW_EWALLET_ORDER_PAID_NOTIFICATION",
					"Pembayaran Berhasil",
					"Pembayaran order kamu berhasil dan mitra sedang dicarikan",
					orderID, customerID)

			case "repeat":
				updates["order_status"] = "WAIT_SCHEDULE"
				if err := tx.Model(&models.OrderTransaction{}).
					Where("payment_id_pay = ? AND order_status = ?", payload.Data.ID, "WAITING_PAYMENT").
					Updates(updates).Error; err != nil {
					tx.Rollback()
					return http.StatusInternalServerError, err
				}
				tx.Model(&models.OrderTransactionRepeat{}).
					Where("order_id = ? AND order_status = ?", orderID, "WAITING_PAYMENT").
					Update("order_status", "WAIT_SCHEDULE")

				var repeats []models.OrderTransactionRepeat
				s.DB.Where("order_id = ?", orderID).Find(&repeats)
				for _, rep := range repeats {
					runAt := rep.OrderTime
					warningAt := runAt.Add(3 * time.Minute)
					rp, _ := queue.NewOrderComingSoonRunTask(orderID)
					queue.ScheduleOnceAt(queue.TypeOrderComingSoonRun, rp, runAt)
					wp, _ := queue.NewOrderComingSoonWarningTask(orderID)
					queue.ScheduleOnceAt(queue.TypeOrderComingSoonWarning, wp, warningAt)
				}

				taskPayload, _ := queue.NewOrderQueueVATask(orderID)
				queue.AsynqClient.Enqueue(asynq.NewTask(queue.TypeOrderQueueVA, taskPayload), asynq.Queue("critical"))
				s.webhookPushFCM("customer", customerID,
					"REPEAT_EWALLET_ORDER_PAID_NOTIFICATION",
					"Pembayaran Order Berulang Berhasil",
					"Pembayaran order berulang berhasil",
					orderID, customerID)

			case "coming soon":
				updates["order_status"] = "WAIT_SCHEDULE"
				if err := tx.Model(&models.OrderTransaction{}).
					Where("payment_id_pay = ? AND order_status = ?", payload.Data.ID, "WAITING_PAYMENT").
					Updates(updates).Error; err != nil {
					tx.Rollback()
					return http.StatusInternalServerError, err
				}
				var orderFull models.OrderTransaction
				s.DB.Where("id = ?", orderID).First(&orderFull)
				scheduleAt := orderFull.OrderTime
				warningAt := scheduleAt.Add(3 * time.Minute)
				runPayload, _ := queue.NewOrderComingSoonRunTask(orderID)
				queue.ScheduleOnceAt(queue.TypeOrderComingSoonRun, runPayload, scheduleAt)
				warnPayload, _ := queue.NewOrderComingSoonWarningTask(orderID)
				queue.ScheduleOnceAt(queue.TypeOrderComingSoonWarning, warnPayload, warningAt)

				taskPayload, _ := queue.NewOrderQueueVATask(orderID)
				queue.AsynqClient.Enqueue(asynq.NewTask(queue.TypeOrderQueueVA, taskPayload), asynq.Queue("critical"))
				s.webhookPushFCM("customer", customerID,
					"COMING_SOON_EWALLET_ORDER_PAID_NOTIFICATION",
					"Pembayaran Order Terjadwal Berhasil",
					"Pembayaran order terjadwal berhasil",
					orderID, customerID)
			}
		} else {
			if err := tx.Model(&models.OrderTransaction{}).
				Where("payment_id_pay = ? AND order_status = ?", payload.Data.ID, "WAITING_PAYMENT").
				Update("order_status", "CANCELED_FAILED_PAYMENT").Error; err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, err
			}
			s.webhookPushFCM("customer", customerID,
				"EWALLET_PAYMENT_FAILED",
				"Pembayaran Gagal",
				"Pembayaran ewallet kamu gagal",
				orderID, customerID)
		}

		s.DB.Create(&models.SuberesLogs{
			LogName: "Notification Paid Ewallet Order",
			LogType: "Paid Ewallet Order",
			LogURL:  fmt.Sprintf("%+v", payload),
			LogTime: logTime,
		})

	case "ewallet.void":
		orderData, err := s.OrderTransactionRepo.FindDynamicOrderTransactionMap(
			nil,
			nil,
			"payment_id_pay = ?",
			[]interface{}{payload.Data.ID},
		)
		if err != nil || orderData == nil {
			tx.Rollback()
			return http.StatusNotFound, errors.New("order not found for void")
		}

		orderID := fmt.Sprintf("%v", orderData["id"])
		orderType := fmt.Sprintf("%v", orderData["order_type"])
		customerID := fmt.Sprintf("%v", orderData["customer_id"])
		grossAmount := fmt.Sprintf("%v", orderData["gross_amount"])

		voidStatus := fmt.Sprintf("VOID_%s", payload.Data.VoidStatus)
		if err := tx.Model(&models.OrderTransaction{}).
			Where("id = ?", orderID).
			Updates(map[string]interface{}{
				"void_status":  voidStatus,
				"order_status": "CANCELED_VOID",
			}).Error; err != nil {
			tx.Rollback()
			return http.StatusInternalServerError, err
		}

		if orderType == "repeat" {
			tx.Model(&models.OrderTransactionRepeat{}).
				Where("order_id = ?", orderID).
				Update("order_status", "CANCELED_VOID")
		}

		var subPayment models.SubPayment
		subPaymentID := fmt.Sprintf("%v", orderData["sub_payment_id"])
		s.DB.Where("id = ?", subPaymentID).First(&subPayment)

		notifType, notifTitle, notifMsg := webhookVoidNotifMessage(payload.Data.VoidStatus, grossAmount, subPayment.TitlePayment)
		s.webhookPushFCM("customer", customerID, notifType, notifTitle, notifMsg, orderID, customerID)

	default:
		return http.StatusOK, nil
	}

	if err := tx.Commit().Error; err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// ─── Private helpers ──────────────────────────────────────────────────────────

// webhookPushFCM sends an FCM notification to a single user by userType + userID.
func (s *WebhookService) webhookPushFCM(userType, userID, notifType, title, message, orderID, customerID string) {
	var user models.User
	if err := s.DB.Where("id = ? AND user_type = ?", userID, userType).First(&user).Error; err != nil {
		return
	}
	if user.FirebaseToken == nil || *user.FirebaseToken == "" {
		return
	}
	if _, err := service.SendMulticast(s.DB, userType, map[string]interface{}{
		"data": map[string]string{
			"notification_type": notifType,
			"title":             title,
			"message":           message,
			"order_id":          orderID,
			"customer_id":       customerID,
			"notif_type":        "order",
		},
		"tokens": []string{*user.FirebaseToken},
	}); err != nil {
		log.Printf("[WebhookService] webhookPushFCM error: %v", err)
	}
}

// webhookVoidNotifMessage returns the FCM notification type, title and message for ewallet void events.
func webhookVoidNotifMessage(voidStatus, grossAmount, paymentTitle string) (string, string, string) {
	amount := fmt.Sprintf("Rp. %s", grossAmount)
	switch voidStatus {
	case "PENDING":
		return "NOW_EWALLET_ORDER_VOID_PENDING_NOTIFICATION",
			"Pengembalian Saldo diproses",
			fmt.Sprintf("Pengembalian Saldo sebesar %s ke akun %s sedang diproses", amount, paymentTitle)
	case "SUCCEEDED":
		return "NOW_EWALLET_ORDER_VOID_SUCCEEDED_NOTIFICATION",
			"Pengembalian Saldo berhasil",
			fmt.Sprintf("Pengembalian Saldo sebesar %s ke akun %s berhasil", amount, paymentTitle)
	default:
		return "NOW_EWALLET_ORDER_VOID_FAILED_NOTIFICATION",
			"Pengembalian Saldo gagal",
			fmt.Sprintf("Pengembalian Saldo sebesar %s ke akun %s gagal, Tim Suberes akan segera memeriksa nya", amount, paymentTitle)
	}
}
