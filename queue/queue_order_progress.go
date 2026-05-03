package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"suberes_golang/config"
	"suberes_golang/helpers"
	"suberes_golang/models"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// HandleOrderOnProgressToFinishTask is triggered after service duration expires.
// It automatically marks the order as FINISH and updates balances / creates transaction records.
func HandleOrderOnProgressToFinishTask(ctx context.Context, t *asynq.Task) error {
	var p OrderOnProgressToFinishPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("HandleOrderOnProgressToFinish: order_id=%s", p.ID)

	var orderData models.OrderTransaction
	if err := config.DB.Where("id = ? AND order_status = ?", p.ID, "ON_PROGRESS").
		Preload("SubService").
		First(&orderData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Order %s not found or not ON_PROGRESS – already handled", p.ID)
			return nil
		}
		return fmt.Errorf("failed to find order %s: %v", p.ID, err)
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin tx failed: %v", tx.Error)
	}

	now := time.Now()

	// 1. Update order status to FINISH
	if err := tx.Model(&models.OrderTransaction{}).
		Where("id = ? AND order_status = ?", p.ID, "ON_PROGRESS").
		Updates(map[string]interface{}{
			"order_status": "FINISH",
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %v", err)
	}

	// 2. Update mitra user_status back to 'stay'
	if err := tx.Model(&models.User{}).
		Where("id = ?", p.MitraID).
		Updates(map[string]interface{}{
			"user_status":  "stay",
			"is_busy":      "no",
			"today_order":  gorm.Expr("today_order + 1"),
			"total_order":  gorm.Expr("total_order + 1"),
			"today_income": gorm.Expr("today_income + ?", orderData.GrossAmountMitra),
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update mitra status: %v", err)
	}

	// 3. Create transaction record for mitra (transaction_in)
	mitraTransactionID := uuid.New().String()
	mitraTransaction := models.Transaction{
		ID:                     mitraTransactionID,
		MitraID:                helpers.StringPtr(p.MitraID),
		CustomerID:             helpers.StringPtr(p.CustomerID),
		OrderID:                helpers.StringPtr(p.ID),
		UserType:               "mitra",
		TransactionName:        "Pendapatan Order",
		TransactionAmount:      orderData.GrossAmountMitra,
		TransactionType:        "transaction_in",
		TransactionTypeFor:     "order_finish_mitra",
		TransactionFor:         "order",
		TransactionStatus:      "success",
		TransactionDescription: "Pendapatan dari order yang telah selesai",
		TimezoneCode:           orderData.TimezoneCode,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if err := tx.Create(&mitraTransaction).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create mitra transaction: %v", err)
	}

	// 4. Handle customer payment type balance deduction
	if orderData.PaymentType == "balance" {
		if err := tx.Model(&models.User{}).
			Where("id = ?", p.CustomerID).
			Update("account_balance", gorm.Expr("account_balance - ?", orderData.GrossAmount)).
			Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to deduct customer balance: %v", err)
		}

		customerTransactionID := uuid.New().String()
		customerTransaction := models.Transaction{
			ID:                     customerTransactionID,
			MitraID:                helpers.StringPtr(p.MitraID),
			CustomerID:             helpers.StringPtr(p.CustomerID),
			OrderID:                helpers.StringPtr(p.ID),
			UserType:               "customer",
			TransactionName:        "Pembayaran Order",
			TransactionAmount:      orderData.GrossAmount,
			TransactionType:        "transaction_out",
			TransactionTypeFor:     "order_finish_customer",
			TransactionFor:         "order",
			TransactionStatus:      "success",
			TransactionDescription: "Pembayaran order yang telah selesai",
			TimezoneCode:           orderData.TimezoneCode,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := tx.Create(&customerTransaction).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create customer transaction: %v", err)
		}
	}

	// 5. Create notification records
	notifID := uuid.New().String()
	notification := models.Notification{
		ID:                  notifID,
		CustomerID:          p.CustomerID,
		MitraID:             p.MitraID,
		OrderID:             p.ID,
		ServiceID:           p.ServiceID,
		SubServiceID:        p.SubServiceID,
		UserType:            "customer",
		NotificationType:    "ORDER_FINISH",
		NotificationTitle:   "Pesanan Selesai",
		NotificationMessage: "Pesanan Anda telah selesai dikerjakan",
		NotifType:           "order",
		IsRead:              "0",
	}
	if err := tx.Create(&notification).Error; err != nil {
		log.Printf("failed to create notification: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("commit failed: %v", err)
	}

	log.Printf("Order %s auto-finished by on_progress_to_finish task", p.ID)
	// TODO: Send FCM notification to customer using their firebase_token
	// TODO: Emit socket.io event to admin rooms

	return nil
}
