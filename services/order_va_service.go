package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	"suberes_golang/repositories"
	"suberes_golang/service"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type OrderVAService struct {
	DB                         *gorm.DB
	UserRepo                   *repositories.UserRepository
	ServiceRepo                *repositories.ServiceRepository
	SubServiceRepo             *repositories.SubServiceRepository
	SubPaymentRepo             *repositories.SubPaymentRepository
	PaymentRepo                *repositories.PaymentRepository
	SubServiceAddedRepo        *repositories.SubServiceAddedRepository
	OrderTransactionRepo       *repositories.OrderTransactionRepository
	OrderTransactionRepeatRepo *repositories.OrderTransactionRepeatsRepository
}

func NewOrderVAService(db *gorm.DB) *OrderVAService {
	return &OrderVAService{
		DB:                         db,
		UserRepo:                   &repositories.UserRepository{DB: db},
		ServiceRepo:                &repositories.ServiceRepository{DB: db},
		SubServiceRepo:             &repositories.SubServiceRepository{DB: db},
		SubPaymentRepo:             &repositories.SubPaymentRepository{DB: db},
		PaymentRepo:                &repositories.PaymentRepository{DB: db},
		SubServiceAddedRepo:        &repositories.SubServiceAddedRepository{DB: db},
		OrderTransactionRepo:       &repositories.OrderTransactionRepository{DB: db},
		OrderTransactionRepeatRepo: &repositories.OrderTransactionRepeatsRepository{DB: db},
	}
}

// CreateOrderVA creates a new VA (virtual account) order.
func (s *OrderVAService) CreateOrderVA(customerID string, dto dtos.CreateOrderDTO) (string, int, string, string, int, error) {
	log.Printf("[CreateOrderVA] START customer_id=%s order_type=%s sub_payment_id=%d gross_amount=%d", customerID, dto.OrderType, dto.SubPaymentID, dto.GrossAmount)

	serviceData, err := s.ServiceRepo.FindByID(dto.ServiceID)
	if err != nil || serviceData == nil {
		log.Printf("[CreateOrderVA] ERROR service not found service_id=%d err=%v", dto.ServiceID, err)
		return "", 0, "", "", http.StatusNotFound, errors.New("service not found")
	}

	subService, err := s.SubServiceRepo.FindByID(dto.SubServiceID)
	if err != nil || subService == nil {
		log.Printf("[CreateOrderVA] ERROR sub service not found sub_service_id=%d err=%v", dto.SubServiceID, err)
		return "", 0, "", "", http.StatusNotFound, errors.New("sub service not found")
	}

	customerData, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil || customerData == nil {
		log.Printf("[CreateOrderVA] ERROR customer not found customer_id=%s err=%v", customerID, err)
		return "", 0, "", "", http.StatusNotFound, errors.New("customer not found")
	}

	subPayment, err := s.SubPaymentRepo.FindById(dto.SubPaymentID)
	if err != nil || subPayment == nil {
		log.Printf("[CreateOrderVA] ERROR sub payment not found sub_payment_id=%d err=%v", dto.SubPaymentID, err)
		return "", 0, "", "", http.StatusNotFound, errors.New("sub payment not found")
	}

	grossAmount := dto.GrossAmount
	grossAmountMitra := grossAmount - ((grossAmount * int64(subService.CompanyPercentage)) / 100)
	grossAmountCompany := grossAmount - grossAmountMitra

	// Time validation
	loc, err := time.LoadLocation(dto.TimezoneCode)
	if err != nil {
		log.Printf("[CreateOrderVA] ERROR invalid timezone_code=%s err=%v", dto.TimezoneCode, err)
		return "", 0, "", "", http.StatusInternalServerError, err
	}

	layout := "2006-01-02 15:04:05"
	nowDateTime, _ := helpers.GetTimezoneNowDateReturnDate(dto.TimezoneCode)

	if dto.OrderType == "coming soon" {
		orderDateTime, err := time.ParseInLocation(layout, helpers.NormalizeDateTimeString(dto.OrderTime), loc)
		if err != nil {
			log.Printf("[CreateOrderVA] ERROR parse order_time=%s err=%v", dto.OrderTime, err)
			return "", 0, "", "", http.StatusBadRequest, err
		}
		if orderDateTime.Day() >= nowDateTime.Day() && orderDateTime.Hour() >= 7 {
			orderDateTime = orderDateTime.Add(time.Duration(subService.MinutesSubServices) * time.Minute)
			if orderDateTime.Hour() >= 23 && orderDateTime.Minute() > 0 {
				log.Printf("[CreateOrderVA] ERROR order_time exceeds max working hours order_time=%s", dto.OrderTime)
				return "", 0, "", "", http.StatusBadRequest, errors.New("Batas maksimal jam order di jam 11 malam")
			}
		}
	} else if dto.OrderType == "repeat" {
		for _, rep := range dto.OrderRepeatList {
			orderDateTime, err := time.ParseInLocation(layout, helpers.NormalizeDateTimeString(rep.OrderTime), loc)
			if err != nil {
				log.Printf("[CreateOrderVA] ERROR parse repeat order_time=%s err=%v", rep.OrderTime, err)
				return "", 0, "", "", http.StatusBadRequest, err
			}
			if orderDateTime.Day() >= nowDateTime.Day() && orderDateTime.Hour() >= 7 {
				orderDateTime = orderDateTime.Add(time.Duration(subService.MinutesSubServices) * time.Minute)
				if orderDateTime.Hour() >= 23 && orderDateTime.Minute() > 0 {
					log.Printf("[CreateOrderVA] ERROR repeat order_time exceeds max working hours order_time=%s", rep.OrderTime)
					return "", 0, "", "", http.StatusBadRequest, errors.New("Batas maksimal jam order di jam 11 malam")
				}
			}
		}
	} else if dto.OrderType == "now" {
		// nowHours := nowDateTime.Hour()
		// nowMinutes := nowDateTime.Minute()

		// if serviceData.ServiceType == "Durasi" {
		// 	dateAdd := nowDateTime.Add(time.Duration(subService.MinutesSubServices) * time.Minute)
		// 	if nowHours >= 23 && nowMinutes >= 0 {
		// 		return "", 0, "", "", http.StatusForbidden, errors.New("Batas maksimal jam operasional sampai jam 11 malam untuk layanan ini")
		// 	}
		// 	if dateAdd.Hour() >= 23 && dateAdd.Minute() >= 0 {
		// 		return "", 0, "", "", http.StatusForbidden, errors.New("Batas maksimal jam operasional sampai jam 11 malam untuk layanan ini")
		// 	}
		// }
		// if nowHours >= 20 && nowMinutes >= 0 {
		// 	return "", 0, "", "", http.StatusForbidden, errors.New("Batas maksimal jam order sampai jam 8 malam")
		// }
	}

	var idTransaction string
	switch dto.OrderType {
	case "now":
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_NOW"), helpers.RandomString(6))
	case "coming soon":
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_COMING_SOON"), helpers.RandomString(6))
	case "repeat":
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_REPEAT"), helpers.RandomString(6))
	}

	createdAtString := helpers.GetTimezoneNowDate(dto.TimezoneCode)
	var orderTimeCreate time.Time
	if dto.OrderType == "coming soon" {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", helpers.NormalizeDateTimeString(dto.OrderTime), loc)
		orderTimeCreate = t.UTC()
	} else {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", createdAtString, loc)
		orderTimeCreate = t.UTC()
	}

	// Compute order_timestamp server-side for non-coming-soon orders (matches Node.js createOrderVa behavior)
	orderTimestampNowDate := strings.Split(createdAtString, " ")[0]
	dateParts := strings.Split(orderTimestampNowDate, "-")
	tsDay := dateParts[2]
	tsMonthRaw := dateParts[1]
	tsYear := dateParts[0]
	var tsMonthName string
	if strings.HasPrefix(tsMonthRaw, "0") {
		tsMonthName = helpers.ConvertNumberToMonthString(strings.TrimPrefix(tsMonthRaw, "0"))
	} else {
		tsMonthName = helpers.ConvertNumberToMonthString(tsMonthRaw)
	}
	tsTimePart := strings.Split(createdAtString, " ")[1]
	tsTimeSplit := strings.Split(tsTimePart, ":")
	orderTimestampNow := fmt.Sprintf("%s %s %s %s:%s", tsDay, tsMonthName, tsYear, tsTimeSplit[0], tsTimeSplit[1])
	var orderTimestamp string
	if dto.OrderType == "coming soon" {
		orderTimestamp = dto.OrderTimestamp
	} else {
		orderTimestamp = orderTimestampNow
	}

	timeoutMinutesStr := os.Getenv("TIMEOUT_COMING_SOON_VA_PAYMENT")
	timeoutMinutes, _ := strconv.Atoi(timeoutMinutesStr)
	if timeoutMinutes == 0 {
		timeoutMinutes = 30
	}

	// Call Xendit VA creation API
	// Payload mirrors JS: { external_id, bank_code, name, expected_amount, is_single_use }
	vaRequestPayload := map[string]interface{}{
		"external_id":     fmt.Sprintf("VA_REQUEST_%s", idTransaction),
		"bank_code":       dto.BankCode,
		"name":            customerData.CompleteName,
		"expected_amount": grossAmount,
		"is_single_use":   false,
		"is_closed":       true,
	}
	paymentClient := helpers.NewClient()
	log.Printf("[CreateOrderVA] Xendit VA request payload: %+v", vaRequestPayload)
	vaRespBytes, err := paymentClient.CreateVirtualAccount(context.Background(), vaRequestPayload)
	if err != nil {
		log.Printf("[CreateOrderVA] ERROR xendit VA creation failed id_transaction=%s err=%v", idTransaction, err)
		return "", 0, "", "", http.StatusInternalServerError, fmt.Errorf("xendit VA creation failed: %w", err)
	}
	log.Printf("[CreateOrderVA] Xendit VA raw response: %s", string(vaRespBytes))
	var vaResp map[string]interface{}
	if err := json.Unmarshal(vaRespBytes, &vaResp); err != nil {
		log.Printf("[CreateOrderVA] ERROR parse xendit VA response id_transaction=%s err=%v", idTransaction, err)
		return "", 0, "", "", http.StatusInternalServerError, fmt.Errorf("failed to parse xendit VA response: %w", err)
	}
	log.Printf("[CreateOrderVA] Xendit VA parsed: va_id=%v external_id=%v account_number=%v bank_code=%v expiration_date=%v status=%v",
		vaResp["id"], vaResp["external_id"], vaResp["account_number"], vaResp["bank_code"], vaResp["expiration_date"], vaResp["status"])
	// Map Xendit response fields — mirrors JS payloadVaResponse
	vaID := fmt.Sprintf("%v", vaResp["id"])
	vaOwnerID := fmt.Sprintf("%v", vaResp["owner_id"])
	vaExternalID := fmt.Sprintf("%v", vaResp["external_id"])
	vaAccountNumber := fmt.Sprintf("%v", vaResp["account_number"])
	vaBankCode := fmt.Sprintf("%v", vaResp["bank_code"])
	vaMerchantCode := fmt.Sprintf("%v", vaResp["merchant_code"])
	vaName := fmt.Sprintf("%v", vaResp["name"])
	vaExpirationDate := fmt.Sprintf("%v", vaResp["expiration_date"])
	if vaExpirationDate == "" || vaExpirationDate == "<nil>" || vaExpirationDate == "%!v(MISSING)" {
		vaExpirationDate = time.Now().UTC().Add(time.Duration(timeoutMinutes) * time.Minute).Format(time.RFC3339)
	}
	vaExpectedAmount := int(grossAmount)
	if v, ok := vaResp["expected_amount"]; ok {
		switch val := v.(type) {
		case float64:
			vaExpectedAmount = int(val)
		}
	}
	vaIsClosed := "0"
	if v, ok := vaResp["is_closed"]; ok && v == true {
		vaIsClosed = "1"
	}
	vaIsSingleUse := "0"
	if v, ok := vaResp["is_single_use"]; ok && v == true {
		vaIsSingleUse = "1"
	}
	vaCurrency := fmt.Sprintf("%v", vaResp["currency"])
	vaXenditStatus := fmt.Sprintf("%v", vaResp["status"])

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randomNotificationID := int(n.Int64()) + 1000

	orderData := &models.OrderTransaction{
		CustomerID:            customerID,
		ServiceID:             dto.ServiceID,
		SubServiceID:          dto.SubServiceID,
		CustomerName:          customerData.CompleteName,
		OrderType:             dto.OrderType,
		MitraGender:           strings.ToLower(dto.MitraGender),
		OrderTime:             orderTimeCreate,
		OrderTimeTemp:         orderTimeCreate,
		OrderTimestamp:        orderTimestamp,
		Address:               dto.Address,
		OrderNote:             dto.OrderNote,
		PaymentID:             subPayment.PaymentID,
		SubPaymentID:          dto.SubPaymentID,
		PaymentType:           "virtual account",
		OrderStatus:           "PROCESSING_PAYMENT",
		IDTransaction:         idTransaction,
		IsAdditional:          helpers.NormalizeIsAdditional(dto.IsAdditional),
		OrderCountAdditional:  dto.OrderCountAdditional,
		GrossAmount:           grossAmount,
		GrossAmountMitra:      grossAmountMitra,
		GrossAmountCompany:    grossAmountCompany,
		GrossAmountAdditional: dto.GrossAmountAdditional,
		TimezoneCode:          dto.TimezoneCode,
		CustomerLatitude:      dto.CustomerLatitude,
		CustomerLongitude:     dto.CustomerLongitude,
		NotificationID:        randomNotificationID,
		OfferExpiredJobID:     uuid.New().String(),
		OfferSelectedJobID:    uuid.New().String(),
		VAID:                  vaID,
		OwnerID:               vaOwnerID,
		ExternalID:            vaExternalID,
		AccountNumber:         vaAccountNumber,
		BankCode:              vaBankCode,
		MerchantCode:          vaMerchantCode,
		Name:                  vaName,
		IsClosed:              vaIsClosed,
		ExpectedAmount:        vaExpectedAmount,
		ExpirationDate:        vaExpirationDate,
		IsSingleUse:           vaIsSingleUse,
		Currency:              vaCurrency,
		XenditStatus:          vaXenditStatus,
		IsRated:               "0",
		IsRatedCustomer:       "0",
		IsMitraOnline:         "0",
		IsCustomerOnline:      "0",
		IsPaidCustomer:        "0",
		IsLive:                "false",
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}
	if dto.OrderType == "coming soon" {
		orderData.OrderOriginSoonTime = dto.OrderTime
	}

	order, err := s.OrderTransactionRepo.CreateOrderData(tx, *orderData)
	if err != nil {
		log.Printf("[CreateOrderVA] ERROR create order id_transaction=%s err=%v", idTransaction, err)
		tx.Rollback()
		return "", 0, "", "", http.StatusInternalServerError, err
	}
	log.Printf("[CreateOrderVA] Order created order_id=%s id_transaction=%s customer_id=%s", order.ID, idTransaction, customerID)

	// Sub service added
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

	// Balance deduction
	if subPayment.TitlePayment == "BALANCE" {
		if customerData.AccountBalance >= grossAmount {
			tx.Model(&models.User{}).
				Where("id = ? AND user_type = ?", customerID, "customer").
				Update("account_balance", gorm.Expr("account_balance - ?", grossAmount))
		}
	}

	// Repeat records
	if dto.OrderType == "repeat" && len(dto.OrderRepeatList) > 0 {
		var payloadRepeat []map[string]interface{}
		for _, elem := range dto.OrderRepeatList {
			parts := strings.Split(elem.OrderTime, " ")
			if len(parts) == 2 {
				grossRepeat := subService.SubPriceService + dto.GrossAddAdditional
				grossCompanyRepeat := (float64(grossRepeat) * subService.CompanyPercentage) / 100
				grossMitraRepeat := float64(grossRepeat) - grossCompanyRepeat
				payloadRepeat = append(payloadRepeat, map[string]interface{}{
					"order_id":             order.ID,
					"customer_id":          customerID,
					"service_id":           dto.ServiceID,
					"sub_service_id":       dto.SubServiceID,
					"customer_name":        customerData.CompleteName,
					"address":              dto.Address,
					"order_time":           elem.OrderTime,
					"order_timestamp":      elem.OrderTimestamp,
					"payment_id":           subPayment.PaymentID,
					"sub_payment_id":       dto.SubPaymentID,
					"order_status":         "WAITING_PAYMENT",
					"id_transaction":       fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_REPEAT"), helpers.RandomString(4)),
					"gross_amount":         grossRepeat,
					"gross_amount_company": grossCompanyRepeat,
					"gross_amount_mitra":   grossMitraRepeat,
					"customer_latitude":    dto.CustomerLatitude,
					"customer_longitude":   dto.CustomerLongitude,
				})
			}
		}
		if len(payloadRepeat) > 0 {
			tx.Model(&models.OrderTransactionRepeat{}).Create(&payloadRepeat)
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[CreateOrderVA] ERROR commit tx order_id=%s err=%v", order.ID, err)
		return "", 0, "", "", http.StatusInternalServerError, err
	}

	// Enqueue payment expiry job: cancel order to CANCELED_LATE_PAYMENT if not paid in time
	notifyPayload, _ := queue.NewOrderEwalletNotifyExpiredTask(order.ID, customerID)
	notifyTask := asynq.NewTask(queue.TypeOrderVAEwalletNotifyExpired, notifyPayload)
	if _, enqErr := queue.AsynqClient.Enqueue(notifyTask, asynq.ProcessIn(time.Duration(timeoutMinutes)*time.Minute)); enqErr != nil {
		log.Printf("[CreateOrderVA] ERROR enqueue payment expiry job order_id=%s err=%v", order.ID, enqErr)
	} else {
		log.Printf("[CreateOrderVA] Payment expiry job enqueued order_id=%s timeout=%d minutes", order.ID, timeoutMinutes)
	}

	log.Printf("[CreateOrderVA] SUCCESS order_id=%s id_transaction=%s customer_id=%s mitra_id=%s", order.ID, idTransaction, order.CustomerID, helpers.DerefStr(order.MitraID))
	return order.ID, -1, order.CustomerID, helpers.DerefStr(order.MitraID), http.StatusOK, nil
}

// AcceptOrderVA handles mitra accepting a VA order in FINDING_MITRA status.
func (s *OrderVAService) AcceptOrderVA(dto dtos.AcceptOrderDTO) (int, map[string]interface{}, error) {
	orderData, err := s.OrderTransactionRepo.FindDynamicOrderTransactionMap(
		nil,
		nil,
		"id = ? AND customer_id = ? AND order_status = ?",
		[]interface{}{dto.OrderID, dto.CustomerID, "FINDING_MITRA"},
	)
	if err != nil || orderData == nil {
		return http.StatusNotFound, nil, errors.New("order not found")
	}

	payment, err := s.PaymentRepo.FindById(int(getInt64(orderData, "payment_id")))
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

	// Race condition guard via Redis (temp_id)
	playerMitraIDs, _ := helpers.GetValue(dto.TempID)
	helpers.DeleteValue(dto.TempID)

	if playerMitraIDs == "" {
		return http.StatusConflict, nil, errors.New("this order was taken by another mitra")
	}

	// Notify other mitras
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

	if err := tx.Model(&models.OrderTransaction{}).
		Where("id = ? AND order_status = ?", dto.OrderID, "FINDING_MITRA").
		Updates(map[string]interface{}{
			"mitra_id":            dto.MitraID,
			"private_key_rsa":     privateKey,
			"public_key_rsa":      publicKey,
			"public_key_mitra":    publicKeyMitra,
			"public_key_customer": publicKeyCustomer,
			"order_status":        orderStatus,
		}).Error; err != nil {
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
		rp, _ := queue.NewOrderComingSoonRunTask(dto.OrderID)
		queue.ScheduleOnceAt(queue.TypeOrderComingSoonRun, rp, scheduleAt)
		wp, _ := queue.NewOrderComingSoonWarningTask(dto.OrderID)
		queue.ScheduleOnceAt(queue.TypeOrderComingSoonWarning, wp, warningAt)
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
		ServiceID:    int(getInt64(orderData, "service_id")),
		SubServiceID: int(getInt64(orderData, "sub_service_id")),
	})

	if err := tx.Commit().Error; err != nil {
		return http.StatusInternalServerError, nil, err
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

	// Admin FCM + socket.io
	var adminTokens []string
	var admins []models.User
	s.DB.Select("firebase_token").
		Where("user_type = ? AND is_logged_in = ?", "admin", "1").
		Find(&admins)
	for _, a := range admins {
		if a.FirebaseToken != nil && *a.FirebaseToken != "" {
			adminTokens = append(adminTokens, *a.FirebaseToken)
		}
	}
	if len(adminTokens) > 0 {
		service.SendMulticast(s.DB, "admin", map[string]interface{}{
			"data": map[string]string{
				"notification_type": "MITRA_ORDER_NOTIFICATION",
				"title":             "Mitra Sedang Menjalankan Order",
				"message":           fmt.Sprintf("Mitra %s sedang menjalankan order customer : %s", mitra.CompleteName, customer.CompleteName),
				"mitra_id":          mitra.ID,
				"customer_id":       customer.ID,
			},
			"tokens": adminTokens,
		})
	}

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

	return http.StatusOK, response, nil
}

// getInt64 safely extracts an int64 value from a map[string]interface{}.
func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		}
	}
	log.Printf("getInt64: key %q not found or wrong type in map", key)
	return 0
}
