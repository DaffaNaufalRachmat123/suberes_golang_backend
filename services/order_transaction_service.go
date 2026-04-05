package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/queue"
	"suberes_golang/repositories"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type OrderTransactionService struct {
	DB                          *gorm.DB
	OrderTransactionRepo        *repositories.OrderTransactionRepository
	OrderTransactionRepeatsRepo *repositories.OrderTransactionRepeatsRepository
	OrderRepo                   *repositories.OrderRepository
	UserRepo                    *repositories.UserRepository
	SubServiceRepo              *repositories.SubServiceRepository
	TransactionRepo             *repositories.TransactionRepository
}

// ---------- 1. FindAllByStatusWithPagination ----------
func (s *OrderTransactionService) FindAllByStatusWithPagination(status string, page, limit int, search string) ([]models.OrderTransaction, int64, error) {
	return s.OrderTransactionRepo.FindAllByStatusWithPagination(status, page, limit, search)
}

// ---------- 2. GetPaymentStatus ----------
func (s *OrderTransactionService) GetPaymentStatus(idTransaction string) (*models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindByIDTransaction(idTransaction)
}

// ---------- 2. GetAdminDashboard ----------
func (s *OrderTransactionService) GetAdminDashboard() (*repositories.AdminDashboardStats, error) {
	return s.OrderTransactionRepo.GetAdminDashboard()
}

// ---------- 3. GetTimezoneCode ----------
type TimezoneResponse struct {
	Status   string `json:"status"`
	ZoneName string `json:"zoneName"`
	Message  string `json:"message"`
}

func (s *OrderTransactionService) GetTimezoneCode(lat, lng string) (*TimezoneResponse, error) {
	apiKey := os.Getenv("TIMEZONE_API_KEY")
	url := fmt.Sprintf(
		"https://api.timezonedb.com/v2.1/get-time-zone?key=%s&format=json&by=position&lat=%s&lng=%s",
		apiKey, lat, lng,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call timezonedb: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var tz TimezoneResponse
	if err := json.Unmarshal(body, &tz); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	return &tz, nil
}

// ---------- 4. SelectMitra ----------
type SelectMitraRequest struct {
	MitraIDs          []string `json:"mitra_ids" binding:"required"`
	NotificationTitle string   `json:"notification_title"`
	NotificationBody  string   `json:"notification_body"`
}

func (s *OrderTransactionService) SelectMitra(orderID string, req SelectMitraRequest) error {
	order, err := s.OrderTransactionRepo.FindById(orderID)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}
	if order.OrderStatus != "WAITING_FOR_SELECTED_MITRA" {
		return fmt.Errorf("order is not in WAITING_FOR_SELECTED_MITRA status")
	}

	tempID := uuid.New().String()
	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"order_status": "FINDING_MITRA",
			"temp_id":      tempID,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var offers []models.OrderOffer
	var selectedMitras []models.OrderSelectedMitra
	var firebaseTokens []string

	for _, mitraID := range req.MitraIDs {
		offers = append(offers, models.OrderOffer{
			TempID:     tempID,
			OrderID:    orderID,
			CustomerID: order.CustomerID,
			MitraID:    mitraID,
		})
		selectedMitras = append(selectedMitras, models.OrderSelectedMitra{
			OrderID:     orderID,
			MitraID:     mitraID,
			OfferStatus: "SELECTED",
		})
		var mitra models.User
		if err := s.DB.Select("firebase_token").Where("id = ?", mitraID).First(&mitra).Error; err == nil {
			if mitra.FirebaseToken != nil && *mitra.FirebaseToken != "" {
				firebaseTokens = append(firebaseTokens, *mitra.FirebaseToken)
			}
		}
	}

	if len(offers) > 0 {
		if err := tx.Create(&offers).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if len(selectedMitras) > 0 {
		if err := tx.Create(&selectedMitras).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: Send FCM notification to selected mitras
	log.Printf("SelectMitra: sending notifications to tokens: %v, title: %s", firebaseTokens, req.NotificationTitle)

	return nil
}

// ---------- 5. GetSelectedMitra ----------
func (s *OrderTransactionService) GetSelectedMitra(orderID string, page, limit int) ([]repositories.SelectedMitraResult, int64, error) {
	order, err := s.OrderTransactionRepo.FindById(orderID)
	if err != nil {
		return nil, 0, fmt.Errorf("order not found: %v", err)
	}
	offset := (page - 1) * limit
	return s.OrderTransactionRepo.FindSelectedMitraPaginated(orderID, order.CustomerLatitude, order.CustomerLongitude, limit, offset)
}

// ---------- 6. GetAdminOrderDetail ----------
func (s *OrderTransactionService) GetAdminOrderDetail(orderID string) (*models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindForAdminDetail(orderID)
}

// ---------- 7. GetComingSoonOrdersForMitra ----------
func (s *OrderTransactionService) GetComingSoonOrdersForMitra(mitraID string, limit, offset int) ([]models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindComingSoonForMitra(mitraID, limit, offset)
}

// ---------- 8. GetRunningOrdersForCustomer ----------
func (s *OrderTransactionService) GetRunningOrdersForCustomer(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindRunningForCustomer(customerID, limit, offset)
}

// ---------- 9. GetCanceledOrdersForCustomer ----------
func (s *OrderTransactionService) GetCanceledOrdersForCustomer(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindCanceledForCustomer(customerID, limit, offset)
}

// ---------- 10. GetDoneOrdersForCustomer ----------
func (s *OrderTransactionService) GetDoneOrdersForCustomer(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindDoneForCustomer(customerID, limit, offset)
}

// ---------- 11. GetDoneOrdersRangeDate ----------
func (s *OrderTransactionService) GetDoneOrdersRangeDate(customerID, startDate, endDate string, limit, offset int) ([]models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindDoneRangeDateForCustomer(customerID, startDate, endDate, limit, offset)
}

// ---------- 12. GetRepeatOrders ----------
func (s *OrderTransactionService) GetRepeatOrders(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindRepeatForCustomer(customerID, limit, offset)
}

// ---------- 13. GetRepeatOrdersSearch ----------
func (s *OrderTransactionService) GetRepeatOrdersSearch(customerID, completeName string, limit, offset int) ([]models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindRepeatSearchForCustomer(customerID, completeName, limit, offset)
}

// ---------- 14. GetRunningOrderDetail ----------
func (s *OrderTransactionService) GetRunningOrderDetail(orderID string, subID int, customerID, mitraID, orderType string) (interface{}, error) {
	if orderType == "repeat" && subID > 0 {
		return s.OrderTransactionRepeatsRepo.FindByOrderAndSubID(orderID, subID)
	}
	return s.OrderTransactionRepo.FindRunningOrderDetail(orderID, customerID, mitraID)
}

// ---------- 15. GetVirtualAccountOrders ----------
func (s *OrderTransactionService) GetVirtualAccountOrders(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	return s.OrderTransactionRepo.FindVirtualAccountOrders(customerID, limit, offset)
}

// ---------- 16. GetOrderDetailFull ----------
func (s *OrderTransactionService) GetOrderDetailFull(orderID string, subID int, customerID, mitraID string, loadAllRepeat bool) (interface{}, error) {
	if subID > 0 {
		return s.OrderTransactionRepeatsRepo.FindByOrderAndSubID(orderID, subID)
	}
	return s.OrderTransactionRepo.FindOrderDetailFull(orderID, customerID, mitraID, loadAllRepeat)
}

// ---------- 17. GetOrderDetailCustomer ----------
func (s *OrderTransactionService) GetOrderDetailCustomer(orderID string, subID int, customerID, mitraID, orderType string, isLoadRepeatList bool) (interface{}, error) {
	if orderType == "repeat" && subID > 0 {
		return s.OrderTransactionRepeatsRepo.FindByOrderAndSubID(orderID, subID)
	}
	return s.OrderTransactionRepo.FindOrderDetailFull(orderID, customerID, mitraID, isLoadRepeatList)
}

// ---------- 18. UpdateToOnProgress ----------
func (s *OrderTransactionService) UpdateToOnProgress(id, customerID, mitraID string) error {
	order, err := s.OrderTransactionRepo.FindFullForUpdateOnProgress(id)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}
	if order.OrderStatus != "OTW" {
		return fmt.Errorf("order is not in OTW status")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()
	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"order_status":        "ON_PROGRESS",
			"order_progress_time": now,
			"updated_at":          now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.User{}).Where("id = ?", mitraID).
		Updates(map[string]interface{}{
			"user_status": "on progress",
			"updated_at":  now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var subService models.SubService
	s.DB.Where("id = ?", order.SubServiceID).First(&subService)

	minutes := subService.MinutesSubServices
	if minutes <= 0 {
		minutes = 60
	}

	delay := time.Duration(minutes) * time.Minute

	taskPayload, err := queue.NewOrderOnProgressToFinishTask(id, customerID, mitraID, order.ServiceID, order.SubServiceID)
	if err == nil {
		taskInfo, enqErr := queue.AsynqClient.Enqueue(
			asynq.NewTask(queue.TypeOrderOnProgressToFinish, taskPayload),
			asynq.ProcessIn(delay),
		)
		if enqErr != nil {
			log.Printf("could not enqueue on_progress_to_finish task: %v", enqErr)
		} else {
			// Save job ID to order
			tx.Model(&models.OrderTransaction{}).Where("id = ?", id).
				Update("on_progress_job_id", taskInfo.ID)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: Send FCM notification to customer
	// TODO: Emit socket.io to admin rooms
	if order.Customer != nil {
		log.Printf("UpdateToOnProgress: notify customer %s order %s is now ON_PROGRESS", customerID, id)
	}

	return nil
}

// ---------- 19. UpdateToOnProgressRepeat ----------
func (s *OrderTransactionService) UpdateToOnProgressRepeat(id string, subID int, customerID, mitraID string) error {
	repeat, err := s.OrderTransactionRepeatsRepo.FindByOrderAndSubID(id, subID)
	if err != nil {
		return fmt.Errorf("repeat order not found: %v", err)
	}
	if repeat.OrderStatus != "OTW" {
		return fmt.Errorf("repeat order is not in OTW status")
	}

	now := time.Now()
	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := s.OrderTransactionRepeatsRepo.UpdateRepeatByOrderAndSubID(tx, id, subID, map[string]interface{}{
		"order_status": "ON_PROGRESS",
		"updated_at":   now,
	}); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.User{}).Where("id = ?", mitraID).
		Updates(map[string]interface{}{
			"user_status": "on progress",
			"updated_at":  now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: FCM notify customer
	log.Printf("UpdateToOnProgressRepeat: order %s sub %d is now ON_PROGRESS", id, subID)
	return nil
}

// ---------- 20. UpdateToFinish ----------
func (s *OrderTransactionService) UpdateToFinish(id, customerID, mitraID string) error {
	order, err := s.OrderTransactionRepo.FindFullForFinish(id)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}
	if order.OrderStatus != "ON_PROGRESS" && order.OrderStatus != "OTW" {
		return fmt.Errorf("order cannot be finished from status %s", order.OrderStatus)
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"order_status": "FINISH",
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.User{}).Where("id = ?", mitraID).
		Updates(map[string]interface{}{
			"user_status":  "stay",
			"is_busy":      "no",
			"today_order":  gorm.Expr("today_order + 1"),
			"total_order":  gorm.Expr("total_order + 1"),
			"today_income": gorm.Expr("today_income + ?", order.GrossAmountMitra),
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	mitraTxID := uuid.New().String()
	mitraTx := models.Transaction{
		ID:                     mitraTxID,
		MitraID:                mitraID,
		CustomerID:             customerID,
		OrderID:                id,
		UserType:               "mitra",
		TransactionName:        "Pendapatan Order",
		TransactionAmount:      order.GrossAmountMitra,
		TransactionType:        "transaction_in",
		TransactionTypeFor:     "order_finish_mitra",
		TransactionFor:         "order",
		TransactionStatus:      "success",
		TransactionDescription: "Pendapatan dari order yang telah selesai",
		TimezoneCode:           order.TimezoneCode,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if err := tx.Create(&mitraTx).Error; err != nil {
		tx.Rollback()
		return err
	}

	if order.PaymentType == "balance" {
		if err := tx.Model(&models.User{}).Where("id = ?", customerID).
			Update("account_balance", gorm.Expr("account_balance - ?", order.GrossAmount)).Error; err != nil {
			tx.Rollback()
			return err
		}
		customerTxID := uuid.New().String()
		customerTx := models.Transaction{
			ID:                     customerTxID,
			MitraID:                mitraID,
			CustomerID:             customerID,
			OrderID:                id,
			UserType:               "customer",
			TransactionName:        "Pembayaran Order",
			TransactionAmount:      order.GrossAmount,
			TransactionType:        "transaction_out",
			TransactionTypeFor:     "order_finish_customer",
			TransactionFor:         "order",
			TransactionStatus:      "success",
			TransactionDescription: "Pembayaran order yang telah selesai",
			TimezoneCode:           order.TimezoneCode,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := tx.Create(&customerTx).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Delete the on_progress job if it exists
	if order.OnProgressJobID != "" {
		if err := queue.Inspector.DeleteTask("default", order.OnProgressJobID); err != nil {
			log.Printf("could not delete on_progress_job %s: %v", order.OnProgressJobID, err)
		}
	}

	notifID := uuid.New().String()
	notif := models.Notification{
		ID:                  notifID,
		CustomerID:          customerID,
		MitraID:             mitraID,
		OrderID:             id,
		ServiceID:           order.ServiceID,
		SubServiceID:        order.SubServiceID,
		UserType:            "customer",
		NotificationType:    "ORDER_FINISH",
		NotificationTitle:   "Pesanan Selesai",
		NotificationMessage: "Pesanan Anda telah selesai dikerjakan",
		NotifType:           "order",
		IsRead:              "0",
	}
	_ = tx.Create(&notif).Error

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: FCM notify customer
	// TODO: Socket.io emit admin
	log.Printf("UpdateToFinish: order %s is now FINISH", id)
	return nil
}

// ---------- 21. UpdateToFinishRepeat ----------
func (s *OrderTransactionService) UpdateToFinishRepeat(id string, subID int, customerID, mitraID string) error {
	repeat, err := s.OrderTransactionRepeatsRepo.FindByOrderAndSubID(id, subID)
	if err != nil {
		return fmt.Errorf("repeat order not found: %v", err)
	}
	if repeat.OrderStatus != "ON_PROGRESS" && repeat.OrderStatus != "OTW" {
		return fmt.Errorf("repeat order cannot be finished from status %s", repeat.OrderStatus)
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	if err := s.OrderTransactionRepeatsRepo.UpdateRepeatByOrderAndSubID(tx, id, subID, map[string]interface{}{
		"order_status": "FINISH",
		"updated_at":   now,
	}); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.User{}).Where("id = ?", mitraID).
		Updates(map[string]interface{}{
			"user_status":  "stay",
			"is_busy":      "no",
			"today_order":  gorm.Expr("today_order + 1"),
			"today_income": gorm.Expr("today_income + ?", repeat.GrossAmountMitra),
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	mitraTxID := uuid.New().String()
	mitraTx := models.Transaction{
		ID:                     mitraTxID,
		MitraID:                mitraID,
		CustomerID:             customerID,
		OrderID:                id,
		SubOrderID:             subID,
		UserType:               "mitra",
		TransactionName:        "Pendapatan Order Repeat",
		TransactionAmount:      repeat.GrossAmountMitra,
		TransactionType:        "transaction_in",
		TransactionTypeFor:     "order_finish_mitra",
		TransactionFor:         "order",
		TransactionStatus:      "success",
		TransactionDescription: "Pendapatan dari order repeat yang telah selesai",
		TimezoneCode:           repeat.OrderTransaction.TimezoneCode,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if err := tx.Create(&mitraTx).Error; err != nil {
		tx.Rollback()
		return err
	}

	if repeat.OrderTransaction != nil && repeat.OrderTransaction.PaymentType == "balance" {
		if err := tx.Model(&models.User{}).Where("id = ?", customerID).
			Update("account_balance", gorm.Expr("account_balance - ?", repeat.GrossAmount)).Error; err != nil {
			tx.Rollback()
			return err
		}
		customerTxID := uuid.New().String()
		customerTx := models.Transaction{
			ID:                     customerTxID,
			MitraID:                mitraID,
			CustomerID:             customerID,
			OrderID:                id,
			SubOrderID:             subID,
			UserType:               "customer",
			TransactionName:        "Pembayaran Order Repeat",
			TransactionAmount:      repeat.GrossAmount,
			TransactionType:        "transaction_out",
			TransactionTypeFor:     "order_finish_customer",
			TransactionFor:         "order",
			TransactionStatus:      "success",
			TransactionDescription: "Pembayaran order repeat yang telah selesai",
			TimezoneCode:           repeat.OrderTransaction.TimezoneCode,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := tx.Create(&customerTx).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: FCM notify customer
	log.Printf("UpdateToFinishRepeat: order %s sub %d is now FINISH", id, subID)
	return nil
}

// ---------- 22. CancelBlast ----------
func (s *OrderTransactionService) CancelBlast(orderID string) error {
	order, err := s.OrderTransactionRepo.FindById(orderID)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	allowedStatuses := []string{"FINDING_MITRA", "WAITING_FOR_SELECTED_MITRA", "WAITING_PAYMENT", "PROCESSING_PAYMENT"}
	isAllowed := false
	for _, st := range allowedStatuses {
		if order.OrderStatus == st {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return fmt.Errorf("order cannot be canceled from status %s", order.OrderStatus)
	}

	// Collect mitra firebase tokens from order_offers for FCM multicast
	var offers []models.OrderOffer
	s.DB.Where("order_id = ?", orderID).Find(&offers)
	var mitraTokens []string
	for _, offer := range offers {
		var mitra models.User
		if err := s.DB.Select("firebase_token").Where("id = ?", offer.MitraID).First(&mitra).Error; err == nil {
			if mitra.FirebaseToken != nil && *mitra.FirebaseToken != "" {
				mitraTokens = append(mitraTokens, *mitra.FirebaseToken)
			}
		}
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	// Refund if balance
	if order.PaymentType == "balance" {
		if err := tx.Model(&models.User{}).Where("id = ?", order.CustomerID).
			Update("account_balance", gorm.Expr("account_balance + ?", order.GrossAmount)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Delete all order_offers for this order (matches JS: order_offers.destroy by temp_id/order_id)
	if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderOffer{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update order status
	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"order_status": "CANCELED",
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Delete asynq jobs
	if order.OfferExpiredJobID != "" {
		_ = queue.Inspector.DeleteTask("default", order.OfferExpiredJobID)
	}
	if order.OfferSelectedJobID != "" {
		_ = queue.Inspector.DeleteTask("default", order.OfferSelectedJobID)
	}
	if order.EwalletNotifyJobID != "" {
		_ = queue.Inspector.DeleteTask("default", order.EwalletNotifyJobID)
	}

	// TODO: Send FCM multicast to mitras
	log.Printf("CancelBlast: order %s canceled, notifying %d mitras", orderID, len(mitraTokens))
	// TODO: Socket.io emit to admin rooms

	return nil
}

// ---------- 23. RejectOrder ----------
func (s *OrderTransactionService) RejectOrder(customerID, mitraID string, serviceID, subServiceID int) error {
	rejected := models.OrderRejected{
		CustomerID:   customerID,
		MitraID:      mitraID,
		ServiceID:    serviceID,
		SubServiceID: subServiceID,
	}
	return s.DB.Create(&rejected).Error
}

// ---------- 24. AdminCancelOrder ----------
type AdminCancelRequest struct {
	CanceledReason string `json:"canceled_reason"`
}

func (s *OrderTransactionService) AdminCancelOrder(id string, canceledReason string) error {
	order, err := s.OrderTransactionRepo.FindFullForFinish(id)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	allowedStatuses := []string{"OTW", "ON_PROGRESS", "WAIT_SCHEDULE", "FINDING_MITRA", "WAITING_FOR_SELECTED_MITRA"}
	isAllowed := false
	for _, st := range allowedStatuses {
		if order.OrderStatus == st {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return fmt.Errorf("order cannot be canceled from status %s", order.OrderStatus)
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	// Refund if balance payment
	if order.PaymentType == "balance" {
		if err := tx.Model(&models.User{}).Where("id = ?", order.CustomerID).
			Update("account_balance", gorm.Expr("account_balance + ?", order.GrossAmount)).Error; err != nil {
			tx.Rollback()
			return err
		}
		refundTxID := uuid.New().String()
		refundTx := models.Transaction{
			ID:                     refundTxID,
			CustomerID:             order.CustomerID,
			MitraID:                helpers.DerefStr(order.MitraID),
			OrderID:                id,
			UserType:               "customer",
			TransactionName:        "Refund Cancel Admin",
			TransactionAmount:      order.GrossAmount,
			TransactionType:        "transaction_in",
			TransactionTypeFor:     "order_cancel_admin",
			TransactionFor:         "order",
			TransactionStatus:      "success",
			TransactionDescription: "Refund dari pembatalan oleh admin",
			TimezoneCode:           order.TimezoneCode,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := tx.Create(&refundTx).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// If mitra was assigned, reset mitra status
	if order.MitraID != nil && *order.MitraID != "" && (order.OrderStatus == "OTW" || order.OrderStatus == "ON_PROGRESS") {
		if err := tx.Model(&models.User{}).Where("id = ?", *order.MitraID).
			Updates(map[string]interface{}{
				"user_status": "stay",
				"is_busy":     "no",
				"updated_at":  now,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"order_status":    "CANCELED",
			"canceled_user":   "admin",
			"canceled_reason": canceledReason,
			"updated_at":      now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Delete on_progress job if exists
	if order.OnProgressJobID != "" {
		_ = queue.Inspector.DeleteTask("default", order.OnProgressJobID)
	}

	// TODO: Send FCM to customer and mitra
	// TODO: Socket.io emit to admin rooms
	log.Printf("AdminCancelOrder: order %s canceled by admin", id)
	return nil
}

// ---------- 25. CancelOrder ----------
type CancelOrderRequest struct {
	CanceledReason string `json:"canceled_reason"`
}

func (s *OrderTransactionService) CancelOrder(id, customerID, mitraID, canceledUser string, canceledReason string) error {
	order, err := s.OrderTransactionRepo.FindFullForFinish(id)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	if order.OrderStatus != "OTW" && order.OrderStatus != "ON_PROGRESS" {
		return fmt.Errorf("order cannot be canceled from status %s", order.OrderStatus)
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	// Refund if balance
	if order.PaymentType == "balance" {
		if err := tx.Model(&models.User{}).Where("id = ?", order.CustomerID).
			Update("account_balance", gorm.Expr("account_balance + ?", order.GrossAmount)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Model(&models.User{}).Where("id = ?", mitraID).
		Updates(map[string]interface{}{
			"user_status": "stay",
			"is_busy":     "no",
			"updated_at":  now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"order_status":    "CANCELED",
			"canceled_user":   canceledUser,
			"canceled_reason": canceledReason,
			"updated_at":      now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Delete on_progress job if exists
	if order.OnProgressJobID != "" {
		_ = queue.Inspector.DeleteTask("default", order.OnProgressJobID)
	}

	// TODO: FCM - if canceled by customer, notify mitra; if by mitra, notify customer
	log.Printf("CancelOrder: order %s canceled by %s", id, canceledUser)
	// TODO: Socket.io emit to admin rooms
	return nil
}

// ---------- 26. CancelRepeatOrder ----------
func (s *OrderTransactionService) CancelRepeatOrder(id string, subID int, customerID, mitraID, canceledUser, canceledReason string) error {
	_, err := s.OrderTransactionRepeatsRepo.FindByOrderAndSubID(id, subID)
	if err != nil {
		return fmt.Errorf("repeat order not found: %v", err)
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	if err := s.OrderTransactionRepeatsRepo.UpdateRepeatByOrderAndSubID(tx, id, subID, map[string]interface{}{
		"order_status":    "CANCELED",
		"canceled_user":   canceledUser,
		"canceled_reason": canceledReason,
		"updated_at":      now,
	}); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.User{}).Where("id = ?", mitraID).
		Updates(map[string]interface{}{
			"user_status": "stay",
			"is_busy":     "no",
			"updated_at":  now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: FCM conditional notification
	log.Printf("CancelRepeatOrder: order %s sub %d canceled by %s", id, subID, canceledUser)
	return nil
}

// ---------- 27. StartRepeatRunOrder ----------
func (s *OrderTransactionService) StartRepeatRunOrder(orderID string, subID int, customerID, mitraID string) error {
	repeat, err := s.OrderTransactionRepeatsRepo.FindByOrderAndSubID(orderID, subID)
	if err != nil {
		return fmt.Errorf("repeat order not found: %v", err)
	}
	if repeat.OrderStatus != "WAIT_SCHEDULE" {
		return fmt.Errorf("repeat order is not in WAIT_SCHEDULE status")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	if err := s.OrderTransactionRepeatsRepo.UpdateRepeatByOrderAndSubID(tx, orderID, subID, map[string]interface{}{
		"order_status": "OTW",
		"updated_at":   now,
	}); err != nil {
		tx.Rollback()
		return err
	}

	// Also update the main order
	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"order_status": "OTW",
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: FCM notify customer
	log.Printf("StartRepeatRunOrder: order %s sub %d is now OTW", orderID, subID)
	return nil
}

// ---------- 28. StartRunOrder (coming soon) ----------
func (s *OrderTransactionService) StartRunOrder(orderID, customerID, mitraID string) error {
	order, err := s.OrderTransactionRepo.FindById(orderID)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}
	if order.OrderStatus != "WAIT_SCHEDULE" {
		return fmt.Errorf("order is not in WAIT_SCHEDULE status")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"order_status": "OTW",
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: FCM notify customer
	log.Printf("StartRunOrder: order %s is now OTW", orderID)
	return nil
}

// ---------- 29. IsAutoBid ----------
func (s *OrderTransactionService) IsAutoBid(orderID, mitraID string) (*models.OrderOffer, error) {
	return s.OrderTransactionRepo.FindIsAutoBid(orderID, mitraID)
}

// ---------- 30. GetDirections ----------
type DirectionsResult struct {
	Cached   bool        `json:"cached"`
	Response interface{} `json:"response"`
}

func (s *OrderTransactionService) GetDirections(orderID string) (*DirectionsResult, error) {
	order, err := s.OrderTransactionRepo.FindWithDirection(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %v", err)
	}

	if order.DirectionResponse != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(order.DirectionResponse), &parsed); err == nil {
			return &DirectionsResult{Cached: true, Response: parsed}, nil
		}
	}

	apiKey := os.Getenv("DIRECTION_API_KEY")
	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/directions/json?origin=%s,%s&destination=%s,%s&key=%s",
		strconv.FormatFloat(order.MitraLatitude, 'f', 8, 64),
		strconv.FormatFloat(order.MitraLongitude, 'f', 8, 64),
		strconv.FormatFloat(order.CustomerLatitude, 'f', 8, 64),
		strconv.FormatFloat(order.CustomerLongitude, 'f', 8, 64),
		apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call directions API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Save to DB
	_ = s.OrderTransactionRepo.UpdateDirectionResponse(s.DB, orderID, string(body))

	var parsed interface{}
	_ = json.Unmarshal(body, &parsed)

	return &DirectionsResult{Cached: false, Response: parsed}, nil
}
