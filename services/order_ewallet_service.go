package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"crypto/rand"

	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/queue"
	"suberes_golang/realtime"
	"suberes_golang/repositories"
	"suberes_golang/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type OrderEwalletService struct {
	DB                         *gorm.DB
	UserRepo                   *repositories.UserRepository
	ServiceRepo                *repositories.ServiceRepository
	SubServiceRepo             *repositories.SubServiceRepository
	SubPaymentRepo             *repositories.SubPaymentRepository
	PaymentRepo                *repositories.PaymentRepository
	SubServiceAddedRepo        *repositories.SubServiceAddedRepository
	OrderRepo                  *repositories.OrderRepository
	OrderChatRepo              *repositories.OrderChatRepository
	OrderOfferRepo             *repositories.OrderOfferRepository
	OrderTransactionRepo       *repositories.OrderTransactionRepository
	OrderTransactionRepeatRepo *repositories.OrderTransactionRepeatsRepository
}

func NewOrderEwalletService(db *gorm.DB) *OrderEwalletService {
	return &OrderEwalletService{
		DB:                         db,
		UserRepo:                   &repositories.UserRepository{DB: db},
		ServiceRepo:                &repositories.ServiceRepository{DB: db},
		SubServiceRepo:             &repositories.SubServiceRepository{DB: db},
		SubPaymentRepo:             &repositories.SubPaymentRepository{DB: db},
		PaymentRepo:                &repositories.PaymentRepository{DB: db},
		SubServiceAddedRepo:        &repositories.SubServiceAddedRepository{DB: db},
		OrderRepo:                  &repositories.OrderRepository{DB: db},
		OrderChatRepo:              &repositories.OrderChatRepository{DB: db},
		OrderOfferRepo:             &repositories.OrderOfferRepository{DB: db},
		OrderTransactionRepo:       &repositories.OrderTransactionRepository{DB: db},
		OrderTransactionRepeatRepo: &repositories.OrderTransactionRepeatsRepository{DB: db},
	}
}

// CallbackNotification is a no-op endpoint that returns 200 OK.
func (s *OrderEwalletService) CallbackNotification() (int, error) {
	return http.StatusOK, nil
}

// CreateOrderEwallet creates a new ewallet order and triggers a Xendit ewallet charge.
func (s *OrderEwalletService) CreateOrderEwallet(ctx *gin.Context, customerID string, dto dtos.CreateOrderDTO) (string, int, string, string, int, error) {
	serviceData, err := s.ServiceRepo.FindByID(dto.ServiceID)
	if err != nil || serviceData == nil {
		return "", 0, "", "", http.StatusNotFound, errors.New("service not found")
	}

	subService, err := s.SubServiceRepo.FindByID(dto.SubServiceID)
	if err != nil || subService == nil {
		return "", 0, "", "", http.StatusNotFound, errors.New("sub service not found")
	}

	customerData, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil || customerData == nil {
		return "", 0, "", "", http.StatusNotFound, errors.New("customer not found")
	}

	subPayment, err := s.SubPaymentRepo.FindById(dto.SubPaymentID)
	if err != nil || subPayment == nil {
		return "", 0, "", "", http.StatusNotFound, errors.New("sub payment not found")
	}

	grossAmount := dto.GrossAmount

	if subPayment.TitlePayment == "BALANCE" {
		if float64(customerData.AccountBalance) < float64(grossAmount) {
			return "", 0, "", "", http.StatusPaymentRequired, errors.New("insufficient account balance")
		}
	}

	grossAmountMitra := grossAmount - ((grossAmount * int64(subService.CompanyPercentage)) / 100)
	grossAmountCompany := grossAmount - grossAmountMitra
	grossAmountCompanyAfterDeduction := grossAmountCompany

	// Apply sub_service_additional discounts/cashbacks
	if len(dto.OrderAdditionalList) > 0 {
		subServiceIDs := make([]int, 0, len(dto.OrderAdditionalList))
		for _, item := range dto.OrderAdditionalList {
			subServiceIDs = append(subServiceIDs, item.ID)
		}

		var additionals []models.SubServiceAdditional
		s.DB.Where("id IN ?", subServiceIDs).Find(&additionals)

		var sumDeduction int64
		for _, a := range additionals {
			switch a.AdditionalType {
			case "discount":
				sumDeduction += int64((float64(a.Amount) * float64(grossAmount)) / 100)
			case "cashback":
				sumDeduction += int64(a.Amount)
			case "choice":
				diff := a.Amount - a.BaseAmount
				if diff < 0 {
					sumDeduction += int64(diff)
				}
			}
		}

		if sumDeduction < 0 {
			grossAmountCompanyAfterDeduction = grossAmountCompany + sumDeduction
		} else {
			grossAmountCompanyAfterDeduction = grossAmountCompany - sumDeduction
		}
	}

	// Time validation
	loc, err := time.LoadLocation(dto.TimezoneCode)
	if err != nil {
		return "", 0, "", "", http.StatusInternalServerError, err
	}

	layout := "2006-01-02 15:04:05"
	nowDateTime, _ := helpers.GetTimezoneNowDateReturnDate(dto.TimezoneCode)

	if dto.OrderType == "coming soon" {
		orderDateTime, err := time.ParseInLocation(layout, helpers.NormalizeDateTimeString(dto.OrderTime), loc)
		if err != nil {
			return "", 0, "", "", http.StatusBadRequest, err
		}
		if orderDateTime.Day() >= nowDateTime.Day() && orderDateTime.Hour() >= 7 {
			orderDateTime = orderDateTime.Add(time.Duration(subService.MinutesSubServices) * time.Minute)
			if orderDateTime.Hour() >= 23 && orderDateTime.Minute() > 0 {
				return "", 0, "", "", http.StatusForbidden, errors.New("Batas maksimal jam order di jam 11 malam")
			}
		}
	} else if dto.OrderType == "repeat" {
		for _, rep := range dto.OrderRepeatList {
			orderDateTime, err := time.ParseInLocation(layout, helpers.NormalizeDateTimeString(rep.OrderTime), loc)
			if err != nil {
				return "", 0, "", "", http.StatusBadRequest, err
			}
			if orderDateTime.Day() >= nowDateTime.Day() && orderDateTime.Hour() >= 7 {
				orderDateTime = orderDateTime.Add(time.Duration(subService.MinutesSubServices) * time.Minute)
				if orderDateTime.Hour() >= 23 && orderDateTime.Minute() > 0 {
					return "", 0, "", "", http.StatusForbidden, errors.New("Batas maksimal jam order di jam 11 malam")
				}
			}
		}
	}

	// Calculate order_time_create
	createdAtString := helpers.GetTimezoneNowDate(dto.TimezoneCode)
	var orderTimeCreate time.Time
	if dto.OrderType == "coming soon" {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", helpers.NormalizeDateTimeString(dto.OrderTime), loc)
		orderTimeCreate = t.UTC()
	} else {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", createdAtString, loc)
		orderTimeCreate = t.UTC()
	}

	timeoutMinutesStr := os.Getenv("TIMEOUT_COMING_SOON_VA_PAYMENT")
	timeoutMinutes, _ := strconv.Atoi(timeoutMinutesStr)
	if timeoutMinutes == 0 {
		timeoutMinutes = 30
	}

	// Format order timestamp
	parts := strings.Split(createdAtString, " ")
	orderTimestampNow := createdAtString
	if len(parts) == 2 {
		dateParts := strings.Split(parts[0], "-")
		timeParts := strings.Split(parts[1], ":")
		if len(dateParts) == 3 && len(timeParts) >= 2 {
			mStr := dateParts[1]
			if strings.HasPrefix(mStr, "0") && len(mStr) > 1 {
				mStr = mStr[1:]
			}
			monthName := helpers.ConvertNumberToMonthString(mStr)
			orderTimestampNow = fmt.Sprintf("%s %s %s %s:%s", dateParts[2], monthName, dateParts[0], timeParts[0], timeParts[1])
		}
	}

	orderTimestamp := dto.OrderTimestamp
	if dto.OrderType != "coming soon" {
		orderTimestamp = orderTimestampNow
	}

	// Generate id_transaction
	var idTransaction string
	switch dto.OrderType {
	case "now":
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_NOW"), helpers.RandomString(6))
	case "coming soon":
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_COMING_SOON"), helpers.RandomString(6))
	case "repeat":
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_REPEAT"), helpers.RandomString(6))
	}

	orderID := uuid.New().String()

	// Call Xendit ewallet API
	paymentClient := helpers.NewClient()
	ewalletPayload := make(map[string]interface{})

	if subPayment.TitlePayment == "ID_DANA" {
		referenceID := fmt.Sprintf("ORDER_ID_%s#%s", orderID, customerID)
		// Ambil baseURL dari Gin context agar dinamis sesuai request
		baseURL := helpers.GetHostURL(ctx)
		successURL := os.Getenv("EWALLET_SUCCESS_REDIRECT_URL")
		if successURL == "" {
			successURL = fmt.Sprintf("%s/api/orders/order_payment_status/%s", strings.TrimRight(baseURL, "/"), idTransaction)
		} else {
			if strings.Contains(successURL, "%s") {
				successURL = fmt.Sprintf(successURL, idTransaction)
			} else {
				if !strings.HasSuffix(successURL, "/") {
					successURL += "/"
				}
				successURL += idTransaction
			}
		}

		ewalletPayload = map[string]interface{}{
			"reference_id":    referenceID,
			"currency":        "IDR",
			"amount":          grossAmount,
			"checkout_method": "ONE_TIME_PAYMENT",
			"channel_code":    "ID_DANA",
			"channel_properties": map[string]interface{}{
				"success_redirect_url": successURL,
			},
			"metadata": map[string]interface{}{
				"branch_area": "PLUIT",
				"branch_city": "JAKARTA",
			},
		}
	} else {
		referenceID := fmt.Sprintf("ORDER_ID_%s#%s", orderID, customerID)
		ewalletPayload = map[string]interface{}{
			"reference_id": referenceID,
			"currency":     "IDR",
			"amount":       grossAmount,
			"channel_code": subPayment.TitlePayment,
			"channel_properties": map[string]interface{}{
				"success_redirect_url": os.Getenv("EWALLET_SUCCESS_REDIRECT_URL"),
			},
			"metadata": map[string]interface{}{
				"order_id":    idTransaction,
				"customer_id": customerID,
			},
		}
	}

	ewalletRespBytes, err := paymentClient.CreateEwalletChargeXendit(context.Background(), ewalletPayload)
	if err != nil {
		return "", 0, "", "", http.StatusInternalServerError, fmt.Errorf("xendit ewallet error: %w", err)
	}

	var ewalletResp map[string]interface{}
	json.Unmarshal(ewalletRespBytes, &ewalletResp)

	paymentIDPay := ""
	mobileEwallet := ""
	checkoutURL := ""
	if ewalletResp != nil {
		if id, ok := ewalletResp["id"].(string); ok {
			paymentIDPay = id
		}
		if actions, ok := ewalletResp["actions"].(map[string]interface{}); ok {
			if mobile, ok := actions["mobile_web_checkout_url"].(string); ok {
				mobileEwallet = mobile
			} else if mobile, ok := actions["mobile_deeplink_checkout_url"].(string); ok {
				mobileEwallet = mobile
			}
			if web, ok := actions["desktop_web_checkout_url"].(string); ok {
				checkoutURL = web
			}
		}
	}

	ewalletNotifyJobID := uuid.New().String()

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randomNotificationID := int(n.Int64()) + 1000

	orderData := &models.OrderTransaction{
		ID:                               orderID,
		CustomerID:                       customerID,
		ServiceID:                        dto.ServiceID,
		SubServiceID:                     dto.SubServiceID,
		CustomerName:                     customerData.CompleteName,
		OrderType:                        dto.OrderType,
		MitraGender:                      strings.ToLower(dto.MitraGender),
		OrderTime:                        orderTimeCreate,
		OrderTimestamp:                   orderTimestamp,
		Address:                          dto.Address,
		OrderNote:                        dto.OrderNote,
		PaymentID:                        subPayment.PaymentID,
		SubPaymentID:                     dto.SubPaymentID,
		PaymentType:                      "ewallet",
		OrderStatus:                      "WAITING_PAYMENT",
		IDTransaction:                    idTransaction,
		IsAdditional:                     helpers.NormalizeIsAdditional(dto.IsAdditional),
		OrderCountAdditional:             dto.OrderCountAdditional,
		GrossAmount:                      grossAmount,
		GrossAmountMitra:                 grossAmountMitra,
		GrossAmountCompany:               grossAmountCompany,
		GrossAmountCompanyAfterDeduction: grossAmountCompanyAfterDeduction,
		GrossAmountAdditional:            dto.GrossAmountAdditional,
		TimezoneCode:                     dto.TimezoneCode,
		CustomerLatitude:                 dto.CustomerLatitude,
		CustomerLongitude:                dto.CustomerLongitude,
		NotificationID:                   randomNotificationID,
		OfferExpiredJobID:                uuid.New().String(),
		OfferSelectedJobID:               uuid.New().String(),
		EwalletNotifyJobID:               ewalletNotifyJobID,
		PaymentIDPay:                     paymentIDPay,
		MobileEwallet:                    mobileEwallet,
		CheckoutURLEwallet:               checkoutURL,
		IsRated:                          "0",
		IsRatedCustomer:                  "0",
		IsMitraOnline:                    "0",
		IsCustomerOnline:                 "0",
		IsPaidCustomer:                   "0",
		IsLive:                           "false",
		IsClosed:                         "0",
		IsSingleUse:                      "0",
		CreatedAt:                        time.Now().UTC(),
		UpdatedAt:                        time.Now().UTC(),
	}
	if dto.OrderType == "coming soon" {
		orderData.OrderOriginSoonTime = dto.OrderTime
	}

	order, err := s.OrderTransactionRepo.CreateOrderData(tx, *orderData)
	if err != nil {
		fmt.Println("Error creating order data:", err.Error())
		tx.Rollback()
		return "", 0, "", "", http.StatusInternalServerError, err
	}

	// Create sub_service_added records
	if len(dto.OrderAdditionalList) > 0 {
		var subAddPayload []map[string]interface{}
		for _, item := range dto.OrderAdditionalList {
			subAddPayload = append(subAddPayload, map[string]interface{}{
				"order_id":           order.ID,
				"customer_id":        customerID,
				"sub_service_add_id": item.ID,
			})
		}
		s.SubServiceAddedRepo.CreateBulk(tx, subAddPayload)
	}

	if err := tx.Commit().Error; err != nil {
		return "", 0, "", "", http.StatusInternalServerError, err
	}

	// Enqueue ewallet notify expired task
	notifyPayload, _ := queue.NewOrderEwalletNotifyExpiredTask(order.ID, customerID)
	notifyTask := asynq.NewTask(queue.TypeOrderEwalletNotifyExpired, notifyPayload)
	queue.AsynqClient.Enqueue(notifyTask, asynq.ProcessIn(time.Duration(timeoutMinutes)*time.Minute))

	return order.ID, -1, order.CustomerID, helpers.DerefStr(order.MitraID), http.StatusOK, nil
}

// AcceptOrderEwallet handles mitra accepting an ewallet order in FINDING_MITRA status.
func (s *OrderEwalletService) AcceptOrderEwallet(dto dtos.AcceptOrderDTO) (int, map[string]interface{}, error) {
	orderData, err := s.OrderTransactionRepo.FindDynamicOrderTransactionMap(
		nil,
		nil,
		"id = ? AND customer_id = ? AND order_status = ?",
		[]interface{}{dto.OrderID, dto.CustomerID, "FINDING_MITRA"},
	)
	if err != nil || orderData == nil {
		return http.StatusNotFound, nil, errors.New("order not found")
	}

	payment, err := s.PaymentRepo.FindById(int(orderData["payment_id"].(int64)))
	if err != nil || payment == nil {
		return http.StatusNotFound, nil, errors.New("payment not found")
	}

	customer, err := s.UserRepo.FindCustomerById(fmt.Sprintf("%v", orderData["customer_id"]))
	if err != nil || customer == nil {
		return http.StatusNotFound, nil, errors.New("customer not found")
	}

	mitra, err := s.UserRepo.FindMitraById(dto.MitraID)
	if err != nil || mitra == nil {
		return http.StatusNotFound, nil, errors.New("mitra not found")
	}

	// Race condition guard via Redis
	playerMitraIDs, _ := helpers.GetValue(dto.TempID)
	helpers.DeleteValue(dto.TempID)

	if playerMitraIDs != "" {
		return http.StatusConflict, nil, errors.New("this order was taken by another mitra")
	}

	// Notify other mitras that they lost the order
	ids := strings.Split(playerMitraIDs, ",")
	var filteredIDs []string
	for _, id := range ids {
		if id != mitra.ID {
			filteredIDs = append(filteredIDs, id)
		}
	}
	if len(filteredIDs) > 0 {
		var users []models.User
		s.DB.Select("firebase_token").Where("id IN ?", filteredIDs).Find(&users)
		var tokens []string
		for _, u := range users {
			if u.FirebaseToken != nil && *u.FirebaseToken != "" {
				tokens = append(tokens, *u.FirebaseToken)
			}
		}
		if len(tokens) > 0 {
			service.SendMulticast(s.DB, "mitra", map[string]interface{}{
				"data": map[string]string{
					"notification_type": "LOST_BROADCAST",
					"title":             "Yah...kamu kehilangan order",
					"message":           fmt.Sprintf("Tawaran order dari customer %s telah diambil mitra lain", customer.CompleteName),
					"order_id":          dto.OrderID,
					"customer_id":       dto.CustomerID,
					"notif_type":        "order",
				},
				"tokens": tokens,
			})
		}
	}

	publicKey, privateKey, err := helpers.GenerateRsaKey()
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	publicKeyMitra := helpers.GeneratePublicKey(customer.SharedPrime, customer.SharedBase, customer.SharedSecret)
	publicKeyCustomer := helpers.GeneratePublicKey(customer.SharedPrime, customer.SharedBase, mitra.SharedSecret)

	expiredJobID := fmt.Sprintf("%v", orderData["offer_expired_job_id"])
	selectedJobID := fmt.Sprintf("%v", orderData["offer_selected_job_id"])

	orderStatus := "WAIT_SCHEDULE"
	if dto.OrderType == "now" {
		orderStatus = "OTW"
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updatePayload := map[string]interface{}{
		"mitra_id":              dto.MitraID,
		"private_key_rsa":       privateKey,
		"public_key_rsa":        publicKey,
		"public_key_mitra":      publicKeyMitra,
		"public_key_customer":   publicKeyCustomer,
		"notification_id":       0,
		"offer_expired_job_id":  nil,
		"offer_selected_job_id": nil,
		"order_status":          orderStatus,
	}

	if err := tx.Model(&models.OrderTransaction{}).
		Where("id = ? AND order_status = ?", dto.OrderID, "FINDING_MITRA").
		Updates(updatePayload).Error; err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, nil, err
	}

	if dto.OrderType == "repeat" {
		tx.Model(&models.OrderTransactionRepeat{}).
			Where("order_id = ? AND order_status = ?", dto.OrderID, "FINDING_MITRA").
			Update("order_status", "WAIT_SCHEDULE")
	}

	if dto.OrderType == "now" {
		tx.Model(&models.User{}).
			Where("id = ? AND user_type = ?", dto.MitraID, "mitra").
			Updates(map[string]interface{}{
				"is_busy":                "yes",
				"user_status":            "on progress",
				"order_id_running":       dto.OrderID,
				"customer_id_running":    dto.CustomerID,
				"service_id_running":     orderData["service_id"],
				"sub_service_id_running": orderData["sub_service_id"],
			})
	} else if dto.OrderType == "coming soon" {
		var orderFull models.OrderTransaction
		s.DB.Where("id = ?", dto.OrderID).First(&orderFull)
		scheduleAt := orderFull.OrderTime
		warningAt := scheduleAt.Add(3 * time.Minute)
		runPayload, _ := queue.NewOrderComingSoonRunTask(dto.OrderID)
		queue.ScheduleOnceAt(queue.TypeOrderComingSoonRun, runPayload, scheduleAt)
		warnPayload, _ := queue.NewOrderComingSoonWarningTask(dto.OrderID)
		queue.ScheduleOnceAt(queue.TypeOrderComingSoonWarning, warnPayload, warningAt)
	} else if dto.OrderType == "repeat" {
		var repeats []models.OrderTransactionRepeat
		s.DB.Where("order_id = ? AND order_status = ?", dto.OrderID, "WAIT_SCHEDULE").Find(&repeats)
		for _, rep := range repeats {
			runAt := rep.OrderTime
			warningAt := runAt.Add(3 * time.Minute)
			rp, _ := queue.NewOrderComingSoonRunTask(dto.OrderID)
			queue.ScheduleOnceAt(queue.TypeOrderComingSoonRun, rp, runAt)
			wp, _ := queue.NewOrderComingSoonWarningTask(dto.OrderID)
			queue.ScheduleOnceAt(queue.TypeOrderComingSoonWarning, wp, warningAt)
		}
	}

	// Delete order_offers
	tx.Where("order_id = ?", dto.OrderID).Delete(&models.OrderOffer{})

	// Create order_chat
	tx.Create(&models.OrderChat{
		ID:           uuid.New().String(),
		OrderID:      dto.OrderID,
		CustomerID:   dto.CustomerID,
		MitraID:      dto.MitraID,
		ServiceID:    int(orderData["service_id"].(int64)),
		SubServiceID: int(orderData["sub_service_id"].(int64)),
	})

	if err := tx.Commit().Error; err != nil {
		return http.StatusInternalServerError, nil, err
	}

	// Delete asynq offer jobs
	if expiredJobID != "" && expiredJobID != "<nil>" {
		queue.Inspector.DeleteTask("default", expiredJobID)
	}
	if selectedJobID != "" && selectedJobID != "<nil>" {
		queue.Inspector.DeleteTask("default", selectedJobID)
	}

	// FCM to customer
	if customer.FirebaseToken != nil && *customer.FirebaseToken != "" {
		service.SendMulticast(s.DB, "customer", map[string]interface{}{
			"data": map[string]string{
				"notification_type": "GOT_ORDER",
				"title":             "Kamu dapat mitra",
				"message":           "Halo, kamu sudah dapat mitra untuk order mu",
				"order_id":          dto.OrderID,
				"mitra_id":          dto.MitraID,
				"customer_id":       dto.CustomerID,
				"notif_type":        "order",
			},
			"tokens": []string{*customer.FirebaseToken},
		})
	}

	// Socket.io to admins
	emitAdminOrderCount(s.DB)

	response := map[string]interface{}{
		"order_id":     dto.OrderID,
		"sub_id":       dto.SubID,
		"temp_id":      dto.TempID,
		"customer_id":  dto.CustomerID,
		"mitra_id":     dto.MitraID,
		"order_type":   dto.OrderType,
		"payment_type": payment.Type,
		"status":       "success",
	}
	if dto.OrderType == "now" {
		response["shared_prime"] = customer.SharedPrime
	}

	return http.StatusOK, response, nil
}

// emitAdminOrderCount sends the current running/waiting order counts to all online admins via socket.io.
func emitAdminOrderCount(db *gorm.DB) {
	var admins []models.User
	db.Select("socket_id").
		Where("user_type IN ? AND is_logged_in = ?", []string{"admin", "superadmin"}, "1").
		Find(&admins)

	var runningCount int64
	db.Model(&models.OrderTransaction{}).
		Where("order_status IN ?", []string{"OTW", "ON_PROGRESS"}).
		Count(&runningCount)

	var waitingCount int64
	db.Model(&models.OrderTransaction{}).
		Where("order_status = ?", "WAITING_FOR_SELECTED_MITRA").
		Count(&waitingCount)

	for _, admin := range admins {
		if admin.SocketID == "" {
			continue
		}
		realtime.Server.BroadcastToRoom("/", admin.SocketID, "admin_message", map[string]interface{}{
			"notification_type":   "NOTIFICATION_ORDER_RUNNING",
			"order_running_count": runningCount,
			"order_waiting_count": waitingCount,
		})
	}
}
