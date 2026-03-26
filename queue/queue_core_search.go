package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"suberes_golang/config"
	"suberes_golang/models"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

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

	var customerData models.User
	if err := config.DB.Where("id = ? AND user_type = ?", orderData.CustomerID, "customer").First(&customerData).Error; err != nil {
		return fmt.Errorf("failed to find customer with id %s: %v", orderData.CustomerID, err)
	}

	var subServiceData models.SubService
	if err := config.DB.Where("id = ?", orderData.SubServiceID).First(&subServiceData).Error; err != nil {
		return fmt.Errorf("failed to find sub service with id %d: %v", orderData.SubServiceID, err)
	}

	var serviceData models.Service
	if err := config.DB.Where("id = ?", orderData.ServiceID).First(&serviceData).Error; err != nil {
		return fmt.Errorf("failed to find service with id %d: %v", orderData.ServiceID, err)
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
		tx.Commit()

		log.Printf("Sending push notifications to tokens: %v", registrationTokenList)

		timeoutCanTakeOrder, _ := time.ParseDuration(os.Getenv("TIMEOUT_CAN_TAKE_ORDER") + "m")
		if timeoutCanTakeOrder == 0 {
			timeoutCanTakeOrder = 1 * time.Minute
		}
		timeoutFindingOrder, _ := time.ParseDuration(os.Getenv("TIMEOUT_FINDING_ORDER") + "m")
		if timeoutFindingOrder == 0 {
			timeoutFindingOrder = 2 * time.Minute
		}

		offerExpiredTaskPayload, err := NewOrderOfferExpiredTask(orderData.ID, tempID, orderData.CustomerID, orderData.NotificationID, timeoutFindingOrder.String())
		if err == nil {
			_, err = AsynqClient.Enqueue(asynq.NewTask(TypeOrderOfferExpired, offerExpiredTaskPayload), asynq.ProcessIn(timeoutCanTakeOrder))
			if err != nil {
				log.Printf("could not enqueue task: %v", err)
			}
		}
	} else {
		if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ?", orderData.ID).Update("order_status", "WAITING_FOR_SELECTED_MITRA").Error; err != nil {
			return fmt.Errorf("failed to update order status: %v", err)
		}
		timeoutFindingOrder, _ := time.ParseDuration(os.Getenv("TIMEOUT_FINDING_ORDER") + "m")
		if timeoutFindingOrder == 0 {
			timeoutFindingOrder = 2 * time.Minute
		}
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

func HandleOrderOfferExpiredTask(ctx context.Context, t *asynq.Task) error {
	var p OrderOfferExpiredPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Handling order offer expired task for order_id=%s", p.OrderID)

	if err := config.DB.Where("temp_id = ?", p.TempID).Delete(&models.OrderOffer{}).Error; err != nil {
		log.Printf("failed to delete order offers: %v", err)
	}

	if err := config.DB.Model(&models.OrderTransaction{}).Where("id = ? AND order_status = ?", p.OrderID, "FINDING_MITRA").Update("order_status", "WAITING_FOR_SELECTED_MITRA").Error; err != nil {
		log.Printf("failed to update order transaction status: %v", err)
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

	return nil
}
