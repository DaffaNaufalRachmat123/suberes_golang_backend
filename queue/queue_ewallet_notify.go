package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"suberes_golang/config"
	"suberes_golang/models"
	"suberes_golang/service"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// HandleOrderEwalletNotifyExpiredTask cancels an order that has not been paid
// within the ewallet payment window (TIMEOUT_COMING_SOON_VA_PAYMENT minutes).
func HandleOrderEwalletNotifyExpiredTask(ctx context.Context, t *asynq.Task) error {
	var p OrderEwalletNotifyExpiredPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("HandleOrderEwalletNotifyExpired: order_id=%s customer_id=%s", p.OrderID, p.CustomerID)

	var orderData models.OrderTransaction
	if err := config.DB.
		Where("id = ? AND customer_id = ? AND order_status = ?", p.OrderID, p.CustomerID, "WAITING_PAYMENT").
		First(&orderData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Already paid or canceled – nothing to do.
			log.Printf("EwalletNotify: order %s not in WAITING_PAYMENT state – skipping", p.OrderID)
			return nil
		}
		return fmt.Errorf("failed to find order %s: %v", p.OrderID, err)
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin tx failed: %v", tx.Error)
	}

	now := time.Now()

	if err := tx.Model(&models.OrderTransaction{}).
		Where("id = ? AND customer_id = ? AND order_status = ?", p.OrderID, p.CustomerID, "WAITING_PAYMENT").
		Updates(map[string]interface{}{
			"order_status":          "CANCELED_LATE_PAYMENT",
			"payment_id_pay":        "",
			"mobile_ewallet":        "",
			"checkout_url_ewallet":  "",
			"ewallet_notify_job_id": "",
			"updated_at":            now,
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit failed: %v", err)
	}

	// Push notification to customer
	var customerData models.User
	if err := config.DB.
		Where("id = ? AND user_type = ?", orderData.CustomerID, "customer").
		First(&customerData).Error; err == nil {
		if customerData.FirebaseToken != nil && *customerData.FirebaseToken != "" {
			payloadMsg := map[string]interface{}{
				"data": map[string]string{
					"notification_type": "CANCELED_LATE_PAYMENT",
					"title":             "Pembayaran Gagal",
					"message":           "Maaf, order kamu dibatalkan karena batas waktu pembayaran habis",
					"order_id":          orderData.ID,
					"customer_id":       orderData.CustomerID,
					"notif_type":        "order",
				},
				"tokens": []string{*customerData.FirebaseToken},
			}
			if _, err := service.SendMulticast(config.DB, "customer", payloadMsg); err != nil {
				log.Printf("EwalletNotify: push notification error: %v", err)
			}
		}
	}

	log.Printf("EwalletNotify: order %s canceled due to late payment", p.OrderID)
	return nil
}
