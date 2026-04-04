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
	"suberes_golang/service"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// HandleOrderComingSoonRunTask fires at order_time for coming_soon/repeat orders.
// Sends FCM to both mitra and customer to start the order.
func HandleOrderComingSoonRunTask(ctx context.Context, t *asynq.Task) error {
	var p OrderComingSoonRunPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("HandleOrderComingSoonRun: order_id=%s", p.OrderID)

	var orderData models.OrderTransaction
	if err := config.DB.
		Where("id = ? AND order_status = ?", p.OrderID, "WAIT_SCHEDULE").
		First(&orderData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("ComingSoonRun: order %s not in WAIT_SCHEDULE – skipping", p.OrderID)
			return nil
		}
		return fmt.Errorf("failed to find order %s: %v", p.OrderID, err)
	}

	var customerData models.User
	if err := config.DB.
		Where("id = ? AND user_type = ?", orderData.CustomerID, "customer").
		First(&customerData).Error; err != nil {
		return fmt.Errorf("customer not found: %v", err)
	}

	var mitraData models.User
	if err := config.DB.
		Where("id = ? AND user_type = ?", orderData.MitraID, "mitra").
		First(&mitraData).Error; err != nil {
		return fmt.Errorf("mitra not found: %v", err)
	}

	// Notify mitra to start order
	if mitraData.FirebaseToken != nil && *mitraData.FirebaseToken != "" {
		payloadMitra := map[string]interface{}{
			"data": map[string]string{
				"notification_type": "COMING_SOON_ORDER_RUN_NOTIFICATION",
				"title":             "Order Terjadwal",
				"message":           fmt.Sprintf("Halo %s harap jalankan orderan mu sekarang", mitraData.CompleteName),
				"order_id":          orderData.ID,
				"mitra_id":          helpers.DerefStr(orderData.MitraID),
				"sub_order_id":      "-1",
				"customer_id":       orderData.CustomerID,
				"notif_type":        "order",
			},
			"tokens": []string{*mitraData.FirebaseToken},
		}
		if _, err := service.SendMulticast(config.DB, "mitra", payloadMitra); err != nil {
			log.Printf("ComingSoonRun: mitra push error: %v", err)
		}
	}

	// Notify customer that mitra should now be running
	if customerData.FirebaseToken != nil && *customerData.FirebaseToken != "" {
		payloadCustomer := map[string]interface{}{
			"data": map[string]string{
				"notification_type": "COMING_SOON_ORDER_NOTIFICATION",
				"title":             "Order Terjadwal",
				"message":           fmt.Sprintf("Mitra %s seharusnya menjalankan orderan mu sekarang", mitraData.CompleteName),
				"order_id":          orderData.ID,
				"mitra_id":          helpers.DerefStr(orderData.MitraID),
				"sub_order_id":      "-1",
				"customer_id":       orderData.CustomerID,
				"notif_type":        "order",
			},
			"tokens": []string{*customerData.FirebaseToken},
		}
		if _, err := service.SendMulticast(config.DB, "customer", payloadCustomer); err != nil {
			log.Printf("ComingSoonRun: customer push error: %v", err)
		}
	}

	log.Printf("ComingSoonRun: notifications sent for order %s", p.OrderID)
	return nil
}

// HandleOrderComingSoonWarningTask fires 3 minutes after order_time.
// If the mitra has not started (is_busy=no) and the order is still WAIT_SCHEDULE,
// the order is automatically cancelled and both parties are notified.
func HandleOrderComingSoonWarningTask(ctx context.Context, t *asynq.Task) error {
	var p OrderComingSoonWarningPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("HandleOrderComingSoonWarning: order_id=%s", p.OrderID)

	var orderData models.OrderTransaction
	if err := config.DB.
		Where("id = ?", p.OrderID).
		First(&orderData).Error; err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	var paymentData models.Payment
	config.DB.Where("id = ?", orderData.PaymentID).First(&paymentData)

	var mitraData models.User
	if err := config.DB.
		Where("id = ? AND user_type = ?", orderData.MitraID, "mitra").
		First(&mitraData).Error; err != nil {
		return fmt.Errorf("mitra not found: %v", err)
	}

	var customerData models.User
	if err := config.DB.
		Where("id = ? AND user_type = ?", orderData.CustomerID, "customer").
		First(&customerData).Error; err != nil {
		return fmt.Errorf("customer not found: %v", err)
	}

	// Only cancel if mitra hasn't started and order is still WAIT_SCHEDULE
	if mitraData.IsBusy == "no" && orderData.OrderStatus == "WAIT_SCHEDULE" {
		tx := config.DB.Begin()
		if tx.Error != nil {
			return fmt.Errorf("begin tx failed: %v", tx.Error)
		}

		if err := tx.Model(&models.OrderTransaction{}).
			Where("id = ? AND mitra_id = ? AND customer_id = ?",
				orderData.ID, orderData.MitraID, orderData.CustomerID).
			Updates(map[string]interface{}{
				"order_status": "CANCELED",
				"updated_at":   time.Now(),
			}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to cancel order: %v", err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("commit failed: %v", err)
		}

		// Notify mitra – performance warning
		if mitraData.FirebaseToken != nil && *mitraData.FirebaseToken != "" {
			payloadMitra := map[string]interface{}{
				"data": map[string]string{
					"notification_type": "CANCELED_ORDER_COMING_SOON_NOTIFICATION",
					"title":             "Perhatian !",
					"message":           fmt.Sprintf("Halo %s performa kamu diturunkan karena tidak menjalankan order", mitraData.CompleteName),
					"order_id":          orderData.ID,
					"sub_order_id":      "-1",
					"mitra_id":          helpers.DerefStr(orderData.MitraID),
					"customer_id":       orderData.CustomerID,
					"notif_type":        "order",
				},
				"tokens": []string{*mitraData.FirebaseToken},
			}
			if _, err := service.SendMulticast(config.DB, "mitra", payloadMitra); err != nil {
				log.Printf("ComingSoonWarning: mitra push error: %v", err)
			}
		}

		// Notify customer
		if customerData.FirebaseToken != nil && *customerData.FirebaseToken != "" {
			msg := fmt.Sprintf("Order dibatalkan otomatis karena mitra %s tidak menjalankan order", mitraData.CompleteName)
			if paymentData.Type == "virtual account" || paymentData.Type == "ewallet" {
				msg += " dan uang kamu akan dikembalikan segera"
			}
			payloadCustomer := map[string]interface{}{
				"data": map[string]string{
					"notification_type": "CANCELED_ORDER_LATE_NOTIFICATION",
					"title":             "Order Terlambat & Dibatalkan",
					"message":           msg,
					"order_id":          orderData.ID,
					"sub_order_id":      "-1",
					"mitra_id":          helpers.DerefStr(orderData.MitraID),
					"customer_id":       orderData.CustomerID,
					"notif_type":        "order",
				},
				"tokens": []string{*customerData.FirebaseToken},
			}
			if _, err := service.SendMulticast(config.DB, "customer", payloadCustomer); err != nil {
				log.Printf("ComingSoonWarning: customer push error: %v", err)
			}
		}

		log.Printf("ComingSoonWarning: order %s auto-canceled", p.OrderID)
	} else {
		log.Printf("ComingSoonWarning: order %s already running or not in WAIT_SCHEDULE – no action", p.OrderID)
	}

	return nil
}
