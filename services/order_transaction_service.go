package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/queue"
	"suberes_golang/repositories"
	"suberes_golang/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type OrderTransactionService struct {
	DB                          *gorm.DB
	OrderTransactionRepo        *repositories.OrderTransactionRepository
	OrderTransactionRepeatsRepo *repositories.OrderTransactionRepeatsRepository
	OrderRepo                   *repositories.OrderRepository
	PaymentRepo                 *repositories.PaymentRepository
	SubPaymentRepo              *repositories.SubPaymentRepository
	ServiceRepo                 *repositories.ServiceRepository
	SubServiceRepo              *repositories.SubServiceRepository
	TransactionRepo             *repositories.TransactionRepository
	UserRepo                    *repositories.UserRepository
}

// ---------- 1. FindAllByStatusWithPagination ----------
func (s *OrderTransactionService) FindAllByStatusWithPagination(status string, page, limit int, search string) ([]models.OrderTransaction, int64, error) {
	return s.OrderTransactionRepo.FindAllByStatusWithPagination(status, page, limit, search)
}

// ---------- 2. GetPaymentStatus ----------
func (s *OrderTransactionService) GetPaymentStatus(idTransaction string) (*dtos.PaymentStatusResponse, error) {
	data, err := s.OrderTransactionRepo.FindByIDTransaction(idTransaction)
	if err != nil {
		return nil, err
	}

	amountFormatted := helpers.FormatRupiah(data.GrossAmount)

	statusStr := "Gagal"
	descSuffix := "Kami akan langsung balikin uang nya ke E-Wallet yang kamu pakai untuk bayar order ini"
	if data.IsPaidCustomer == "1" {
		statusStr = "Berhasil"
		descSuffix = "Kami akan langsung cariin Mitra Suberes buat kamu."
	}

	title := fmt.Sprintf("Pembayaran Order %s\nsebesar %s\n%s", data.IDTransaction, amountFormatted, statusStr)
	description := fmt.Sprintf("Pembayaran Order dengan ID Transaksi %s sebesar %s %s. %s", data.IDTransaction, amountFormatted, statusStr, descSuffix)

	supportMail := os.Getenv("SUPPORT_EMAIL")
	officeAddress := os.Getenv("SUPPORT_OFFICE_ADDRESS")

	return &dtos.PaymentStatusResponse{
		Title:          title,
		Description:    description,
		IsPaidCustomer: data.IsPaidCustomer,
		SupportMail:    supportMail,
		OfficeAddress:  officeAddress,
	}, nil
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
	MitraIDs          []string `json:"mitra_ids" binding:"required,min=1"`
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

	// Timeout for mitra to take the order
	timeoutCanTakeMinutes, _ := strconv.Atoi(os.Getenv("TIMEOUT_CAN_TAKE_ORDER"))
	if timeoutCanTakeMinutes == 0 {
		timeoutCanTakeMinutes = 1
	}
	timeoutCanTakeOrder := time.Duration(timeoutCanTakeMinutes) * time.Minute

	// Check remaining customer wait time from the OrderSelectedExpired task.
	// If customer's remaining time < mitra timeout, extend it.
	timeoutFindingMinutes, _ := strconv.Atoi(os.Getenv("TIMEOUT_FINDING_ORDER"))
	if timeoutFindingMinutes == 0 {
		timeoutFindingMinutes = 2
	}

	// Calculate remaining seconds for customer wait based on search_time_complete or order_time
	var referenceTime time.Time
	if order.SearchTimeComplete != nil {
		referenceTime = *order.SearchTimeComplete
	} else {
		referenceTime = order.OrderTime
	}
	elapsed := time.Since(referenceTime)
	totalCustomerWait := time.Duration(timeoutFindingMinutes) * time.Minute
	remainingCustomerWait := totalCustomerWait - elapsed

	// If remaining customer wait < mitra timeout, we need to extend
	if remainingCustomerWait < timeoutCanTakeOrder {
		deficit := timeoutCanTakeOrder - remainingCustomerWait
		totalCustomerWait = totalCustomerWait + deficit
		// Re-enqueue OrderSelectedExpired with extended time
		orderSelectedExpiredPayload, taskErr := queue.NewOrderSelectedExpiredTask(orderID)
		if taskErr == nil {
			_, _ = queue.AsynqClient.Enqueue(
				asynq.NewTask(queue.TypeOrderSelectedExpired, orderSelectedExpiredPayload),
				asynq.ProcessIn(remainingCustomerWait+deficit),
			)
		}
	}

	now := time.Now()
	tempID := uuid.New().String()
	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"order_status":         "FINDING_MITRA",
			"temp_id":              tempID,
			"search_time_complete": now,
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

	// Enqueue per-mitra expiry cleanup
	for _, offer := range offers {
		perMitraPayload, perMitraErr := queue.NewOrderOfferMitraExpiredTask(orderID, offer.MitraID, tempID)
		if perMitraErr == nil {
			_, _ = queue.AsynqClient.Enqueue(asynq.NewTask(queue.TypeOrderOfferMitraExpired, perMitraPayload), asynq.ProcessIn(timeoutCanTakeOrder))
		}
	}

	// Send FCM notification to selected mitras
	if len(firebaseTokens) > 0 {
		msgData := map[string]interface{}{
			"temp_id":           tempID,
			"notification_type": "ORDER_BROADCAST",
			"title":             "Orderan Masuk",
			"message":           "Ada orderan masuk dari admin",
			"order_id":          orderID,
			"order_type":        order.OrderType,
			"payment_type":      order.PaymentType,
			"customer_id":       order.CustomerID,
			"notification_id":   strconv.Itoa(order.NotificationID),
			"notif_type":        "order",
			"push_time":         now.UTC().Format("2006-01-02 15:04:05"),
			"push_time_expired": strconv.Itoa(int(timeoutCanTakeOrder.Seconds())),
		}
		msg := map[string]interface{}{
			"data":   msgData,
			"tokens": firebaseTokens,
		}
		if _, err := service.SendMulticast(s.DB, "mitra", msg); err != nil {
		}
	}

	return nil
}

// ---------- 5. GetSelectedMitra ----------
func (s *OrderTransactionService) GetSelectedMitra(orderID string, page, limit int, search string) ([]repositories.SelectedMitraResult, int64, error) {
	order, err := s.OrderTransactionRepo.FindById(orderID)
	if err != nil {
		return nil, 0, fmt.Errorf("order not found: %v", err)
	}
	maxRange, _ := strconv.ParseFloat(os.Getenv("MAX_RANGE_CUSTOMER_BY_SYSTEM"), 64)
	if maxRange == 0 {
		maxRange = 7
	}
	offset := (page - 1) * limit
	return s.OrderTransactionRepo.FindSelectedMitraPaginated(orderID, order.CustomerLatitude, order.CustomerLongitude, order.MitraGender, maxRange, search, limit, offset)
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

	// delay := time.Duration(minutes) * time.Minute

	// taskPayload, err := queue.NewOrderOnProgressToFinishTask(id, customerID, mitraID, order.ServiceID, order.SubServiceID)
	// if err == nil {
	// 	taskInfo, enqErr := queue.AsynqClient.Enqueue(
	// 		asynq.NewTask(queue.TypeOrderOnProgressToFinish, taskPayload),
	// 		asynq.ProcessIn(delay),
	// 	)
	// 	if enqErr != nil {
	// 	} else {
	// 		// Save job ID to order
	// 		tx.Model(&models.OrderTransaction{}).Where("id = ?", id).
	// 			Update("on_progress_job_id", taskInfo.ID)
	// 	}
	// }

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// TODO: Send FCM notification to customer
	if order.Customer != nil {

		// Kirim push notification ke customer jika ada firebase_token
		if order.Customer.FirebaseToken != nil && *order.Customer.FirebaseToken != "" {
			msgPayload := map[string]interface{}{
				"data": map[string]interface{}{
					"notification_type": "ON_PROGRESS_RECEIVER",
					"title":             "Order lagi dikerjain",
					"message":           "Mitra lagi ngerjain orderan mu",
					"order_id":          id,
					"customer_id":       customerID,
					"mitra_id":          mitraID,
					"notif_type":        "order",
				},
			}
			_, _ = service.SendToDevice(s.DB, "customer", *order.Customer.FirebaseToken, msgPayload)
		}
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
	return nil
}

func (s *OrderTransactionService) UpdateToFinish(id, customerID, mitraID string) error {
	order, err := s.OrderTransactionRepo.FindFullForFinish(id)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	if order.OrderStatus != "ON_PROGRESS" && order.OrderStatus != "OTW" {
		return fmt.Errorf("order cannot be finished from status %s", order.OrderStatus)
	}

	// 🔴 VALIDASI
	customer, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil {
		return fmt.Errorf("customer not found")
	}

	mitra, err := s.UserRepo.FindMitraById(mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found")
	}

	if order.ServiceID == 0 {
		return fmt.Errorf("order missing service_id")
	}

	serviceData, err := s.ServiceRepo.FindByID(order.ServiceID)
	if err != nil {
		return fmt.Errorf("service not found")
	}

	payment, err := s.PaymentRepo.FindById(order.PaymentID)
	if err != nil {
		return fmt.Errorf("payment not found")
	}

	subPayment, _ := s.SubPaymentRepo.FindById(order.SubPaymentID)

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	// =========================
	// 🔴 UPDATE ORDER
	// =========================
	if err := tx.Model(&models.OrderTransaction{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"order_status": "FINISH",
			"shared_prime": nil,
			"updated_at":   now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// =========================
	// 🔴 UPDATE SERVICE COUNT
	// =========================
	if err := tx.Model(&models.Service{}).
		Where("id = ?", serviceData.ID).
		Update("service_count", gorm.Expr("service_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	// =========================
	// 🔴 UPDATE MITRA BASE STATE
	// =========================
	mitraUpdate := map[string]interface{}{
		"user_status":            "stay",
		"is_busy":                "no",
		"order_id_running":       nil,
		"sub_order_id_running":   nil,
		"service_id_running":     nil,
		"sub_service_id_running": nil,
		"total_order":            gorm.Expr("total_order + 1"),
		"updated_at":             now,
	}

	// =========================
	// 🔴 PAYMENT LOGIC (FINAL)
	// =========================

	currentMitraBalance := mitra.AccountBalance
	currentCustomerBalance := customer.AccountBalance

	var transactions []models.Transaction

	// handle null sub payment
	bankName := ""
	if subPayment != nil {
		bankName = subPayment.Title
	}

	t1 := now
	t2 := now.Add(1 * time.Millisecond)
	t3 := now.Add(2 * time.Millisecond)

	// =========================
	// 🔴 1. PEMBAYARAN ORDER SELESAI (LOG ONLY)
	// =========================
	transactions = append(transactions, models.Transaction{
		ID:                     uuid.New().String(),
		MitraID:                helpers.StringPtr(mitraID),
		CustomerID:             helpers.StringPtr(customerID),
		OrderID:                helpers.StringPtr(id),
		OrderIDTransaction:     order.IDTransaction,
		UserType:               "mitra",
		BankName:               bankName,
		LastAmount:             currentMitraBalance, // saldo awal
		TransactionName:        "Pembayaran order selesai",
		TransactionAmount:      order.GrossAmount,
		TransactionType:        "transaction_in",
		TransactionTypeFor:     "Uang Order",
		TransactionFor:         "order",
		TransactionStatus:      "success",
		TransactionDescription: "Pembayaran dari customer",
		TimezoneCode:           order.TimezoneCode,
		CreatedAt:              t1,
		UpdatedAt:              t1,
	})

	// =========================
	// 🔴 2. PENDAPATAN MITRA (+)
	// =========================
	currentMitraBalance += order.GrossAmountMitra

	transactions = append(transactions, models.Transaction{
		ID:                     uuid.New().String(),
		MitraID:                helpers.StringPtr(mitraID),
		CustomerID:             helpers.StringPtr(customerID),
		OrderID:                helpers.StringPtr(id),
		OrderIDTransaction:     order.IDTransaction,
		UserType:               "mitra",
		BankName:               bankName,
		LastAmount:             currentMitraBalance, // setelah penambahan
		TransactionName:        "Pendapatan dari order selesai",
		TransactionAmount:      order.GrossAmountMitra,
		TransactionType:        "transaction_in",
		TransactionTypeFor:     "Uang Order",
		TransactionFor:         "order",
		TransactionStatus:      "success",
		TransactionDescription: "Pendapatan mitra",
		TimezoneCode:           order.TimezoneCode,
		CreatedAt:              t2,
		UpdatedAt:              t2,
	})

	// =========================
	// 🔴 3. POTONGAN (KHUSUS TUNAI)
	// =========================
	if payment.Type == "tunai" {

		currentMitraBalance -= order.GrossAmountCompany

		transactions = append(transactions, models.Transaction{
			ID:                     uuid.New().String(),
			MitraID:                helpers.StringPtr(mitraID),
			CustomerID:             helpers.StringPtr(customerID),
			OrderID:                helpers.StringPtr(id),
			OrderIDTransaction:     order.IDTransaction,
			UserType:               "mitra",
			BankName:               bankName,
			LastAmount:             currentMitraBalance, // setelah potongan
			TransactionName:        "Potongan order suberes",
			TransactionAmount:      order.GrossAmountCompany,
			TransactionType:        "transaction_out",
			TransactionTypeFor:     "Potongan Suberes",
			TransactionFor:         "order",
			TransactionStatus:      "success",
			TransactionDescription: "Potongan karena pembayaran customer tunai",
			TimezoneCode:           order.TimezoneCode,
			CreatedAt:              t3,
			UpdatedAt:              t3,
		})

	} else if payment.Type == "balance" {

		// =========================
		// 🔴 CUSTOMER BALANCE (NON-TUNAI)
		// =========================
		transactions = append(transactions, models.Transaction{
			ID:                     uuid.New().String(),
			MitraID:                helpers.StringPtr(mitraID),
			CustomerID:             helpers.StringPtr(customerID),
			OrderID:                helpers.StringPtr(id),
			OrderIDTransaction:     order.IDTransaction,
			UserType:               "customer",
			BankName:               bankName,
			LastAmount:             currentCustomerBalance,
			TransactionName:        "Pembayaran order",
			TransactionAmount:      order.GrossAmount,
			TransactionType:        "transaction_out",
			TransactionTypeFor:     "Uang Order",
			TransactionFor:         "order",
			TransactionStatus:      "success",
			TransactionDescription: "Pembayaran order selesai",
			TimezoneCode:           order.TimezoneCode,
			CreatedAt:              now,
			UpdatedAt:              now,
		})

		if err := tx.Model(&models.User{}).
			Where("id = ? AND account_balance >= ?", customerID, order.GrossAmount).
			Update("account_balance", gorm.Expr("account_balance - ?", order.GrossAmount)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// =========================
	// 🔴 UPDATE SALDO MITRA (FINAL - ATOMIC)
	// =========================
	var mitraBalanceExpr *gorm.DB
	if payment.Type == "tunai" {
		mitraBalanceExpr = tx.Model(&models.User{}).
			Where("id = ?", mitraID).
			Update("account_balance", gorm.Expr("account_balance + ? - ?", order.GrossAmountMitra, order.GrossAmountCompany))
	} else {
		mitraBalanceExpr = tx.Model(&models.User{}).
			Where("id = ?", mitraID).
			Update("account_balance", gorm.Expr("account_balance + ?", order.GrossAmountMitra))
	}
	if mitraBalanceExpr.Error != nil {
		tx.Rollback()
		return mitraBalanceExpr.Error
	}

	// =========================
	// 🔴 APPLY UPDATE MITRA STATE
	// =========================
	if err := tx.Model(&models.User{}).
		Where("id = ?", mitraID).
		Updates(mitraUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	// =========================
	// 🔴 INSERT TRANSACTIONS
	// =========================
	if err := tx.Create(&transactions).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Kirim push notification ke customer jika ada firebase_token
	if customer.FirebaseToken != nil && *customer.FirebaseToken != "" {
		msgPayload := map[string]interface{}{
			"data": map[string]interface{}{
				"notification_type": "ON_FINISH_RECEIVER",
				"title":             "Order udah selesai",
				"message":           "Order udah selesai dikerjain , semoga nyaman",
				"order_id":          id,
				"customer_id":       customerID,
				"mitra_id":          mitraID,
				"notif_type":        "order",
			},
		}
		_, _ = service.SendToDevice(s.DB, "customer", *order.Customer.FirebaseToken, msgPayload)
	}

	return nil
}

// ---------- 20. UpdateToFinish ----------
// func (s *OrderTransactionService) UpdateToFinish(id, customerID, mitraID string) error {
// 	order, err := s.OrderTransactionRepo.FindFullForFinish(id)
// 	if err != nil {
// 		return fmt.Errorf("order not found: %v", err)
// 	}
// 	if order.OrderStatus != "ON_PROGRESS" && order.OrderStatus != "OTW" {
// 		return fmt.Errorf("order cannot be finished from status %s", order.OrderStatus)
// 	}

// 	tx := s.DB.Begin()
// 	if tx.Error != nil {
// 		return tx.Error
// 	}

// 	now := time.Now()

// 	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", id).
// 		Updates(map[string]interface{}{
// 			"order_status": "FINISH",
// 			"updated_at":   now,
// 		}).Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	if err := tx.Model(&models.User{}).Where("id = ?", mitraID).
// 		Updates(map[string]interface{}{
// 			"user_status":  "stay",
// 			"is_busy":      "no",
// 			"today_order":  gorm.Expr("today_order + 1"),
// 			"total_order":  gorm.Expr("total_order + 1"),
// 			"today_income": gorm.Expr("today_income + ?", order.GrossAmountMitra),
// 			"updated_at":   now,
// 		}).Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	mitraTxID := uuid.New().String()
// 	mitraTx := models.Transaction{
// 		ID:                     mitraTxID,
// 		MitraID:                mitraID,
// 		CustomerID:             customerID,
// 		OrderID:                id,
// 		OrderIDTransaction: order.IDTransaction,
// 		BankName: "",
// 		LastAmount: 0,
// 		UserType:               "mitra",
// 		TransactionName:        "Pendapatan Order",
// 		TransactionAmount:      order.GrossAmountMitra,
// 		TransactionType:        "transaction_in",
// 		TransactionTypeFor:     "Uang Order",
// 		TransactionFor:         "order",
// 		TransactionStatus:      "success",
// 		TransactionDescription: "Pendapatan dari order yang telah selesai",
// 		TimezoneCode:           order.TimezoneCode,
// 		CreatedAt:              now,
// 		UpdatedAt:              now,
// 	}
// 	if err := tx.Create(&mitraTx).Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	if order.PaymentType == "balance" {
// 		if err := tx.Model(&models.User{}).Where("id = ?", customerID).
// 			Update("account_balance", gorm.Expr("account_balance - ?", order.GrossAmount)).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 		customerTxID := uuid.New().String()
// 		customerTx := models.Transaction{
// 			ID:                     customerTxID,
// 			MitraID:                mitraID,
// 			CustomerID:             customerID,
// 			OrderID:                id,
// 			UserType:               "customer",
// 			BankName: "",
// 			LastAmount: 0,
// 			TransactionName:        "Pembayaran Order",
// 			TransactionAmount:      order.GrossAmount,
// 			TransactionType:        "transaction_out",
// 			TransactionTypeFor:     "Uang Order",
// 			TransactionFor:         "order",
// 			TransactionStatus:      "success",
// 			TransactionDescription: "Pembayaran order yang telah selesai",
// 			TimezoneCode:           order.TimezoneCode,
// 			CreatedAt:              now,
// 			UpdatedAt:              now,
// 		}
// 		if err := tx.Create(&customerTx).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 	}

// 	// Delete the on_progress job if it exists
// 	if order.OnProgressJobID != "" {
// 		if err := queue.Inspector.DeleteTask("default", order.OnProgressJobID); err != nil {
// 		}
// 	}

// 	notifID := uuid.New().String()
// 	notif := models.Notification{
// 		ID:                  notifID,
// 		CustomerID:          customerID,
// 		MitraID:             mitraID,
// 		OrderID:             id,
// 		ServiceID:           order.ServiceID,
// 		SubServiceID:        order.SubServiceID,
// 		UserType:            "customer",
// 		NotificationType:    "ORDER_FINISH",
// 		NotificationTitle:   "Pesanan Selesai",
// 		NotificationMessage: "Pesanan Anda telah selesai dikerjakan",
// 		NotifType:           "order",
// 		IsRead:              "0",
// 	}
// 	_ = tx.Create(&notif).Error

// 	if err := tx.Commit().Error; err != nil {
// 		return err
// 	}

// 	// TODO: FCM notify customer
// 	// TODO: Socket.io emit admin
// 	return nil
// }

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
		MitraID:                helpers.StringPtr(mitraID),
		CustomerID:             helpers.StringPtr(customerID),
		OrderID:                helpers.StringPtr(id),
		SubOrderID:             &subID,
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
		LastAmount:             0,
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
			MitraID:                helpers.StringPtr(mitraID),
			CustomerID:             helpers.StringPtr(customerID),
			OrderID:                helpers.StringPtr(id),
			SubOrderID:             &subID,
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
			LastAmount:             0,
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
	return nil
}

// ---------- 22. CancelBlast ----------
func (s *OrderTransactionService) CancelBlast(ctx *gin.Context, orderID string) error {
	order, err := s.OrderTransactionRepo.FindById(orderID)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	allowedStatuses := []string{"FINDING_MITRA", "WAITING_FOR_SELECTED_MITRA"}
	isAllowed := false
	for _, st := range allowedStatuses {
		if order.OrderStatus == st {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		if order.OrderStatus != "" && (order.OrderStatus == "CANCELED" || order.OrderStatus == "CANCELED_VOID") {
			// Already canceled, treat as success
			return nil
		}
		return fmt.Errorf("order cannot be canceled from status %s", order.OrderStatus)
	}

	// Get payment data
	payment, err := s.PaymentRepo.FindById(order.PaymentID)
	if err != nil {
		return fmt.Errorf("payment not found")
	}

	// Ambil mitra IDs dari order_offers di DB (mitra IDs tidak disimpan di Redis)
	var playerMitraIDs []string
	sendPushNotifToMitra := false
	log.Printf("[CancelBlast] orderID=%s, order.TempID=%q", orderID, order.TempID)
	var orderOffers []models.OrderOffer
	if order.TempID != "" {
		s.DB.Where("temp_id = ? AND order_id = ?", order.TempID, orderID).Find(&orderOffers)
	} else {
		s.DB.Where("order_id = ?", orderID).Find(&orderOffers)
	}
	for _, offer := range orderOffers {
		playerMitraIDs = append(playerMitraIDs, offer.MitraID)
	}
	log.Printf("[CancelBlast] mitra IDs dari order_offers: %v", playerMitraIDs)
	if len(playerMitraIDs) > 0 {
		sendPushNotifToMitra = true
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()
	orderStatus := "CANCELED"
	voidStatus := ""

	if payment.Type == "ewallet" || payment.Type == "virtual account" {
		orderStatus = "CANCELED_VOID"
		// TODO: Call Xendit void API here and set voidStatus accordingly
		paymentClient := helpers.NewClient()
		resultVoid, _ := paymentClient.CreateVoidChargeXendit(ctx, order.PaymentIDPay)
		_ = resultVoid
		voidStatus = "VOID_SUCCEEDED" // Placeholder, should be set from API response
		if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"temp_id":      "",
				"order_status": orderStatus,
				"void_status":  voidStatus,
				"updated_at":   now,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else if payment.Type == "balance" {
		// Refund
		if err := tx.Model(&models.User{}).Where("id = ?", order.CustomerID).
			Update("account_balance", gorm.Expr("account_balance + ?", order.GrossAmount)).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"temp_id":      "",
				"order_status": "CANCELED_VOID",
				"updated_at":   now,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// Other payment types: delete order (hapus child records dulu karena ada FK constraint)
		if err := tx.Where("order_id = ?", order.ID).Delete(&models.Notification{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderOffer{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Where("id = ?", order.ID).Delete(&models.OrderTransaction{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Delete order_offers by temp_id if present, else by order_id
	if order.TempID != "" {
		if err := tx.Where("temp_id = ?", order.TempID).Delete(&models.OrderOffer{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderOffer{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Remove jobs
	if order.OfferExpiredJobID != "" {
		_ = queue.Inspector.DeleteTask("default", order.OfferExpiredJobID)
	}
	if order.OfferSelectedJobID != "" {
		_ = queue.Inspector.DeleteTask("default", order.OfferSelectedJobID)
	}
	if order.EwalletNotifyJobID != "" {
		_ = queue.Inspector.DeleteTask("default", order.EwalletNotifyJobID)
	}

	// TODO: Send FCM notification to customer and mitras using helpers (if available)
	// Example: helpers.SendPushNotificationToCustomer(...)
	// Example: helpers.SendPushNotificationToMitras(...)

	// --- Begin: Push Notification to Mitra (if needed) ---
	log.Printf("[CancelBlast] sendPushNotifToMitra=%v, playerMitraIDs=%v", sendPushNotifToMitra, playerMitraIDs)
	if sendPushNotifToMitra && len(playerMitraIDs) > 0 {
		// Get mitra firebase tokens
		var mitraTokens []string
		s.DB.Model(&models.User{}).Where("id IN ? AND user_type = ?", playerMitraIDs, "mitra").Pluck("firebase_token", &mitraTokens)
		log.Printf("[CancelBlast] mitraTokens (raw): %v", mitraTokens)
		// Remove empty tokens
		var filteredTokens []string
		for _, t := range mitraTokens {
			if t != "" {
				filteredTokens = append(filteredTokens, t)
			}
		}
		log.Printf("[CancelBlast] filteredTokens: %v", filteredTokens)
		if len(filteredTokens) > 0 {
			msg := map[string]interface{}{
				"data": map[string]interface{}{
					"notification_type": "CANCEL_BROADCAST",
					"title":             "Order dibatalin",
					"message":           "Customer gak jadi mesen",
					"order_id":          order.ID,
					"order_temp_id":     order.TempID,
					"customer_id":       order.CustomerID,
					"notification_id":   fmt.Sprintf("%d", order.NotificationID),
					"type":              "ORDER_CANCELED",
				},
				"tokens": filteredTokens,
			}
			_, err := service.SendMulticast(s.DB, "mitra", msg)
			log.Printf("[CancelBlast] SendMulticast err=%v", err)
		}
	}

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
			CustomerID:             helpers.StringPtr(order.CustomerID),
			MitraID:                order.MitraID,
			OrderID:                helpers.StringPtr(id),
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

	customer, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil {
		return fmt.Errorf("customer not found")
	}

	mitra, err := s.UserRepo.FindMitraById(mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found")
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
			"user_status":            "stay",
			"is_busy":                "no",
			"order_id_running":       nil,
			"sub_order_id_running":   nil,
			"service_id_running":     nil,
			"sub_service_id_running": nil,
			"updated_at":             now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	orderStatus := "CANCELED"
	updates := map[string]interface{}{
		"canceled_user":   canceledUser,
		"canceled_reason": canceledReason,
		"updated_at":      now,
	}

	if order.PaymentType == "ewallet" {
		orderStatus = "CANCELED_VOID"
		paymentClient := helpers.NewClient()
		resultVoid, _ := paymentClient.CreateVoidChargeXendit(context.Background(), order.PaymentIDPay)
		var responseVoid struct {
			VoidStatus string `json:"void_status"`
		}
		if err := json.Unmarshal(resultVoid, &responseVoid); err == nil && responseVoid.VoidStatus != "" {
			updates["void_status"] = fmt.Sprintf("VOID_%s", responseVoid.VoidStatus)
		}
	} else if order.PaymentType == "balance" {
		orderStatus = "CANCELED_VOID"
	}

	updates["order_status"] = orderStatus

	if err := tx.Model(&models.OrderTransaction{}).Where("id = ?", id).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	txRecord := models.Transaction{
		ID:                     uuid.New().String(),
		MitraID:                helpers.StringPtr(mitraID),
		CustomerID:             helpers.StringPtr(customerID),
		OrderID:                helpers.StringPtr(id),
		UserType:               "customer",
		TransactionName:        "Refund Biaya Order",
		TransactionAmount:      order.GrossAmount,
		TransactionType:        "transaction_out",
		TransactionTypeFor:     "Refund",
		TransactionFor:         "order",
		TransactionStatus:      "success",
		TransactionDescription: fmt.Sprintf("Pengembalian dana kepada customer %s karena order dibatalkan", customer.CompleteName),
		TimezoneCode:           order.TimezoneCode,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	if err := tx.Create(&txRecord).Error; err != nil {
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

	messageToCustomer := fmt.Sprintf("Yah, order kamu dibatalkan oleh mitra %s", mitra.CompleteName)
	if order.PaymentType == "virtual account" {
		messageToCustomer += fmt.Sprintf("\ndan uang mu sebesar %s dikembalikan ke rekening mu segera", helpers.FormatRupiah(order.GrossAmount))
	} else if order.PaymentType == "balance" {
		messageToCustomer += fmt.Sprintf("\ndan uang mu sebesar %s dikembalikan ke saldo", helpers.FormatRupiah(order.GrossAmount))
	} else if order.PaymentType == "ewallet" {
		messageToCustomer += fmt.Sprintf("\ndan uang mu sebesar %s dikembalikan ke ewallet mu segera", helpers.FormatRupiah(order.GrossAmount))
	}

	messageToMitra := fmt.Sprintf("Yah, order kamu dibatalkan oleh customer %s", customer.CompleteName)

	if canceledUser == "customer" {
		if mitra.FirebaseToken != nil && *mitra.FirebaseToken != "" {
			msgMitra := map[string]interface{}{
				"data": map[string]interface{}{
					"notification_type": "CANCEL_BROADCAST",
					"title":             "Order dibatalin",
					"message":           messageToMitra,
					"order_id":          id,
					"customer_id":       customerID,
					"mitra_id":          mitraID,
					"notif_type":        "order",
				},
			}
			_, _ = service.SendToDevice(s.DB, "mitra", *mitra.FirebaseToken, msgMitra)
		}
	} else if canceledUser == "mitra" {
		if customer.FirebaseToken != nil && *customer.FirebaseToken != "" {
			var cust models.User
			s.DB.Where("id = ?", customerID).First(&cust)
			msgCustomer := map[string]interface{}{
				"data": map[string]interface{}{
					"notification_type": "CANCEL_BROADCAST",
					"title":             "Order dibatalin",
					"message":           messageToCustomer,
					"order_id":          id,
					"customer_id":       customerID,
					"mitra_id":          mitraID,
					"account_balance":   strconv.FormatInt(cust.AccountBalance, 10),
					"notif_type":        "order",
				},
			}
			_, _ = service.SendToDevice(s.DB, "customer", *customer.FirebaseToken, msgCustomer)
		}
	}
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

// ---------- 19. GetOrderPendapatan ----------
func (s *OrderTransactionService) GetOrderPendapatan(mitraID string, paymentID int, orderTime string) ([]models.OrderTransaction, error) {
	startDate, endDate := helpers.GetStartEndDateFromString(orderTime)

	orders, err := s.OrderTransactionRepo.GetOrderPendapatan(mitraID, paymentID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		switch orders[i].OrderType {
		case "now":
			orders[i].OrderType = "Order Sekali"
		case "coming soon":
			orders[i].OrderType = "Order Sekali"
		case "repeat":
			orders[i].OrderType = "Order Berulang"
		default:
			orders[i].OrderType = "-"
		}
	}

	return orders, nil
}
