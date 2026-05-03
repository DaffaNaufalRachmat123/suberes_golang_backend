package services

import (
	"fmt"
	"log"
	"strings"
	"suberes_golang/dtos"
	"suberes_golang/models"
	"suberes_golang/queue"
	"suberes_golang/repositories"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type XenditService struct {
	UserRepo         *repositories.UserRepository
	TransactionRepo  *repositories.TransactionRepository
	OrderRepo        *repositories.OrderTransactionRepository
	OrderRepeatsRepo *repositories.OrderTransactionRepeatsRepository
	DB               *gorm.DB
}

func NewXenditService(db *gorm.DB) *XenditService {
	return &XenditService{
		UserRepo:         &repositories.UserRepository{DB: db},
		TransactionRepo:  &repositories.TransactionRepository{DB: db},
		OrderRepo:        &repositories.OrderTransactionRepository{DB: db},
		OrderRepeatsRepo: &repositories.OrderTransactionRepeatsRepository{DB: db},
		DB:               db,
	}
}

func (s *XenditService) HandleEwalletCapture(data *dtos.XenditEwalletData) error {
	if strings.HasPrefix(data.ReferenceID, "Topup") {
		return s.handleEwalletCaptureTopup(data)
	}
	return s.handleEwalletCaptureOrder(data)
}

func (s *XenditService) handleEwalletCaptureTopup(data *dtos.XenditEwalletData) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	transaction, err := s.TransactionRepo.FindTransactionByExternalID(tx, data.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction not found: %w", err)
	}

	user, err := s.UserRepo.FindUserForTransaction(tx, transaction)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("user not found: %w", err)
	}

	if data.Status == "SUCCEEDED" {
		topupAmount := transaction.TransactionAmount - transaction.TransactionFee
		if err := s.TransactionRepo.UpdateTransactionStatus(tx, transaction.ID, "success"); err != nil {
			tx.Rollback()
			return err
		}
		if err := s.UserRepo.UpdateUserBalance(tx, user.ID, topupAmount); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := s.TransactionRepo.UpdateTransactionStatus(tx, transaction.ID, "failed"); err != nil {
			tx.Rollback()
			return err
		}
	}

	log.Printf("Push notification placeholder for topup status %s to user %s", data.Status, user.ID)
	return tx.Commit().Error
}

func (s *XenditService) handleEwalletCaptureOrder(data *dtos.XenditEwalletData) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	orderData, err := s.OrderRepo.FindOrderByPaymentID(tx, data.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("order not found: %w", err)
	}

	if data.Status == "SUCCEEDED" {
		if err := s.handleSuccessfulPayment(tx, orderData); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		updates := map[string]interface{}{"order_status": "CANCELED_FAILED_PAYMENT"}
		if err := s.OrderRepo.UpdateOrderStatus(tx, orderData.ID, "WAITING_PAYMENT", updates); err != nil {
			tx.Rollback()
			return err
		}
	}

	log.Printf("Push notification placeholder for order payment status %s to customer %s", data.Status, orderData.CustomerID)
	return tx.Commit().Error
}

func (s *XenditService) handleSuccessfulPayment(tx *gorm.DB, orderData *models.OrderTransaction) error {
	updates := map[string]interface{}{"is_paid_customer": "1"}
	switch orderData.OrderType {
	case "now":
		updates["order_status"] = "FINDING_MITRA"
		updates["order_time"] = time.Now().UTC()
		if err := s.OrderRepo.UpdateOrderStatus(tx, orderData.ID, "WAITING_PAYMENT", updates); err != nil {
			return err
		}
		taskPayload, err := queue.NewOrderQueueVATask(orderData.ID)
		if err != nil {
			return err
		}
		_, err = queue.AsynqClient.Enqueue(asynq.NewTask(queue.TypeOrderQueueVA, taskPayload), asynq.Queue("critical"))
		if err != nil {
			return err
		}
	case "repeat":
		if err := s.OrderRepeatsRepo.UpdateRepeatOrderStatus(tx, orderData.ID, "WAIT_SCHEDULE"); err != nil {
			return err
		}
	case "coming soon":
		log.Printf("Scheduling reminder for 'coming soon' order %s", orderData.ID)
		// Placeholder for Asynq ProcessAt
	}
	return nil
}

func (s *XenditService) HandleEwalletVoid(data *dtos.XenditEwalletData) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	orderData, err := s.OrderRepo.FindVoidableOrder(tx, data.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("voidable order not found: %w", err)
	}

	voidStatus := fmt.Sprintf("VOID_%s", data.VoidStatus)
	if err := s.OrderRepo.UpdateVoidStatus(tx, orderData.ID, voidStatus); err != nil {
		tx.Rollback()
		return err
	}

	if orderData.OrderType == "repeat" {
		if err := s.OrderRepeatsRepo.UpdateRepeatOrderStatus(tx, orderData.ID, "CANCELED_VOID"); err != nil {
			tx.Rollback()
			return err
		}
	}

	log.Printf("Push notification placeholder for ewallet void status %s to customer %s", data.VoidStatus, orderData.CustomerID)
	return tx.Commit().Error
}
