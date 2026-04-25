package services

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"suberes_golang/constants"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/queue"
	"suberes_golang/realtime"
	"suberes_golang/repositories"
	"suberes_golang/service"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type OrderCashService struct {
	DB                         *gorm.DB
	UserRepo                   *repositories.UserRepository
	ServiceRepo                *repositories.ServiceRepository
	SubServiceRepo             *repositories.SubServiceRepository
	LayananServiceRepo         *repositories.LayananServiceRepository
	SubPaymentRepo             *repositories.SubPaymentRepository
	SubServiceAddedRepo        *repositories.SubServiceAddedRepository
	PaymentRepo                *repositories.PaymentRepository
	OrderRepo                  *repositories.OrderRepository
	OrderChatRepo              *repositories.OrderChatRepository
	OrderOfferRepo             *repositories.OrderOfferRepository
	OrderTransactionRepo       *repositories.OrderTransactionRepository
	OrderTransactionRepeatRepo *repositories.OrderTransactionRepeatsRepository
}

func (s *OrderCashService) CreateOrderCash(customerId string, dto dtos.CreateOrderDTO) (string, int, string, string, int, error) {
	service, err := s.ServiceRepo.FindByID(dto.ServiceID)
	if err != nil {
		return "", 0, "", "", 500, err
	}
	if service == nil {
		return "", 0, "", "", 404, err
	}

	subService, err := s.SubServiceRepo.FindByID(dto.SubServiceID)
	if err != nil {
		return "", 0, "", "", 500, err
	}
	if subService == nil {
		return "", 0, "", "", 404, err
	}
	customerData, err := s.UserRepo.FindCustomerById(customerId)
	if err != nil {
		return "", 0, "", "", 500, err
	}
	if customerData == nil {
		return "", 0, "", "", 404, err
	}
	paymentData, err := s.PaymentRepo.FindById(dto.PaymentID)
	if err != nil {
		return "", 0, "", "", 500, err
	}
	if paymentData == nil {
		return "", 0, "", "", 404, err
	}
	subPaymentData, err := s.SubPaymentRepo.FindById(dto.SubPaymentID)
	if err != nil {
		return "", 0, "", "", 500, err
	}
	if subPaymentData == nil {
		return "", 0, "", "", 404, err
	}
	if subPaymentData.TitlePayment == "BALANCE" {
		if customerData.AccountBalance < dto.GrossAmount {
			return "", 0, "", "", 402, err
		}

	}
	grossAmount := dto.GrossAmount
	grossAmountMitra := grossAmount - ((grossAmount * int64(subService.CompanyPercentage)) / 100)
	grossAmountCompany := (grossAmount * int64(subService.CompanyPercentage)) / 100

	nowDateTime, err := helpers.GetTimezoneNowDateReturnDate(dto.TimezoneCode)
	if err != nil {
		return "", 0, "", "", 500, err
	}

	loc, err := time.LoadLocation(dto.TimezoneCode)
	if err != nil {
		return "", 0, "", "", 500, err
	}

	nowDateTime = nowDateTime.AddDate(0, 0, 1)
	if err != nil {
		return "", 0, "", "", 400, err
	}
	// if dto.OrderType == "now" {
	// 	nowTime := helpers.GetTimezoneNowDate(dto.TimezoneCode)
	// 	splitNowTime := strings.Split(nowTime, " ")
	// 	splitNowDateTime := strings.Split(splitNowTime[0], "-")
	// 	splitNowTimeTime := strings.Split(splitNowTime[1], ":")

	// 	nowYear, _ := strconv.Atoi(splitNowDateTime[0])
	// 	nowMonth, _ := strconv.Atoi(splitNowDateTime[1])
	// 	nowDay, _ := strconv.Atoi(splitNowDateTime[2])
	// 	nowHours, _ := strconv.Atoi(splitNowTimeTime[0])
	// 	nowMinutes, _ := strconv.Atoi(splitNowTimeTime[1])

	// 	dateAdd := time.Date(
	// 		nowYear,
	// 		time.Month(nowMonth),
	// 		nowDay,
	// 		nowHours,
	// 		nowMinutes,
	// 		0,
	// 		0,
	// 		nowDateTime.Location(),
	// 	)

	// 	fmt.Printf("Now Time : %s\n", nowTime)

	// 	if service.ServiceType == "Durasi" {

	// 		dateAdd = dateAdd.Add(time.Duration(subService.MinutesSubServices) * time.Minute)

	// 		addedDate := helpers.GetFormattedYearMonthDate(dateAdd)
	// 		splitAddDate := strings.Split(addedDate, " ")
	// 		splitTime := strings.Split(splitAddDate[1], ":")

	// 		addedHours, _ := strconv.Atoi(splitTime[0])
	// 		addedMinutes, _ := strconv.Atoi(splitTime[1])

	// 		if nowHours >= 23 && nowMinutes >= 0 {
	// 			return "", 0, "", "", 400, errors.New("Batas maksimal jam operasional sampai jam 11 malam untuk layanan ini")
	// 		} else if addedHours >= 23 && addedMinutes >= 0 {
	// 			return "", 0, "", "", 400, errors.New("Batas maksimal jam operasional sampai jam 11 malam untuk layanan ini")
	// 		}
	// 	}

	// 	if nowHours >= 23 && nowMinutes >= 0 {
	// 		return "", 0, "", "", http.StatusForbidden, errors.New("Batas maksimal jam order sampai jam 8 malam")
	// 	}
	// }
	createdAtString := helpers.GetTimezoneNowDate(dto.TimezoneCode)
	orderSoonTime := dto.OrderTime

	var orderTimeCreate time.Time

	if dto.OrderType == "coming soon" {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", helpers.NormalizeDateTimeString(dto.OrderTime), loc)
		orderTimeCreate = t.UTC()
	} else {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", createdAtString, loc)
		orderTimeCreate = t.UTC()
	}

	orderTimestampNowDate := strings.Split(createdAtString, " ")[0]
	dateParts := strings.Split(orderTimestampNowDate, "-")

	day := dateParts[2]
	monthRaw := dateParts[1]
	year := dateParts[0]

	var monthName string
	if strings.HasPrefix(monthRaw, "0") {
		monthName = helpers.ConvertNumberToMonthString(strings.TrimPrefix(monthRaw, "0"))
	} else {
		monthName = helpers.ConvertNumberToMonthString(monthRaw)
	}

	timePart := strings.Split(createdAtString, " ")[1]
	timeSplit := strings.Split(timePart, ":")

	orderTimestampNow := fmt.Sprintf(
		"%s %s %s %s:%s",
		day,
		monthName,
		year,
		timeSplit[0],
		timeSplit[1],
	)

	var orderTimestamp string
	if dto.OrderType == "coming soon" {
		orderTimestamp = dto.OrderTimestamp
	} else {
		orderTimestamp = orderTimestampNow
	}

	var idTransaction string
	if dto.OrderType == "now" {
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_NOW"), helpers.RandomString(6))
	} else if dto.OrderType == "coming soon" {
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_COMING_SOON"), helpers.RandomString(6))
	} else if dto.OrderType == "repeat" {
		idTransaction = fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_REPEAT"), helpers.RandomString(6))
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return "", 0, "", "", 500, tx.Error
	}
	if paymentData.Type == "balance" {
		if customerData.AccountBalance >= grossAmount {
			_, err := s.UserRepo.UpdateUserData(tx, map[string]interface{}{
				"account_balance": customerData.AccountBalance - grossAmount,
			}, map[string]interface{}{
				"id":        customerId,
				"user_type": "customer",
			})
			if err != nil {
				tx.Rollback()
				return "", 0, "", "", 500, err
			}
		}
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randomNotificationID := int(n.Int64())
	orderData := &models.OrderTransaction{
		CustomerID:            customerId,
		ServiceID:             dto.ServiceID,
		SubServiceID:          dto.SubServiceID,
		CustomerName:          customerData.CompleteName,
		OrderType:             dto.OrderType,
		MitraGender:           strings.ToLower(dto.MitraGender),
		OrderTime:             orderTimeCreate,
		OrderTimestamp:        orderTimestamp,
		Address:               dto.Address,
		OrderNote:             dto.OrderNote,
		PaymentID:             subPaymentData.PaymentID,
		SubPaymentID:          dto.SubPaymentID,
		PaymentType:           paymentData.Type,
		OrderStatus:           "FINDING_MITRA",
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
		IsRated:               "0",
		IsRatedCustomer:       "0",
		IsMitraOnline:         "0",
		IsCustomerOnline:      "0",
		IsPaidCustomer:        "0",
		IsLive:                "false",
		IsClosed:              "0",
		IsSingleUse:           "0",
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}
	if dto.OrderType == "coming soon" {
		orderData.OrderOriginSoonTime = orderSoonTime
	}
	order, err := s.OrderTransactionRepo.CreateOrderData(tx, *orderData)
	if err != nil {
		tx.Rollback()
		return "", 0, "", "", 500, err
	}
	var subAddPayload []map[string]interface{}

	if len(dto.OrderAdditionalList) > 0 {
		for i := 0; i < len(dto.OrderAdditionalList); i++ {
			payload := map[string]interface{}{
				"order_id":           order.ID,
				"customer_id":        customerId,
				"sub_service_add_id": dto.OrderAdditionalList[i].ID,
			}

			subAddPayload = append(subAddPayload, payload)
		}
	}
	err = s.SubServiceAddedRepo.CreateBulk(tx, subAddPayload)
	if err != nil {
		tx.Rollback()
		return "", 0, "", "", 500, err
	}
	if dto.OrderType == "repeat" {
		var payloadRepeat []map[string]interface{}

		for _, elem := range dto.OrderRepeatList {

			dateTimeSplit := strings.Split(elem.OrderTime, " ")
			if len(dateTimeSplit) == 2 {

				subPriceService := subService.SubPriceService
				grossAddAdditional := dto.GrossAddAdditional
				companyPercentage := subService.CompanyPercentage

				grossAmountRepeat := subPriceService + grossAddAdditional
				grossAmountCompanyRepeat := (float64(grossAmountRepeat) * companyPercentage) / 100
				grossAmountMitraRepeat := float64(grossAmountRepeat) - grossAmountCompanyRepeat

				payload := map[string]interface{}{
					"order_id":             order.ID,
					"customer_id":          customerId,
					"service_id":           dto.ServiceID,
					"sub_service_id":       dto.SubServiceID,
					"customer_name":        customerData.CompleteName,
					"address":              dto.Address,
					"order_time":           elem.OrderTime,
					"order_timestamp":      elem.OrderTimestamp,
					"payment_id":           dto.PaymentID,
					"sub_payment_id":       dto.SubPaymentID,
					"order_status":         "FINDING_MITRA",
					"id_transaction":       fmt.Sprintf("%s-%s", os.Getenv("PREFIX_ORDER_REPEAT"), helpers.RandomString(4)),
					"gross_amount":         grossAmountRepeat,
					"gross_amount_company": grossAmountCompanyRepeat,
					"gross_amount_mitra":   grossAmountMitraRepeat,
					"customer_latitude":    dto.CustomerLatitude,
					"customer_longitude":   dto.CustomerLongitude,
				}

				payloadRepeat = append(payloadRepeat, payload)
			}
		}

		if len(payloadRepeat) > 0 {
			if err := tx.Model(&models.OrderTransactionRepeat{}).Create(&payloadRepeat).Error; err != nil {
				return "", 0, "", "", 500, err
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		return "", 0, "", "", 500, err
	}
	payload, _ := json.Marshal(queue.OrderQueueCashPayload{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
	})
	task := asynq.NewTask(queue.TypeOrderQueueCash, payload)
	_, err = queue.AsynqClient.Enqueue(task)
	if err != nil {
		return "", 0, "", "", 500, err
	}
	return order.ID, -1, order.CustomerID, helpers.DerefStr(order.MitraID), 200, nil
}

func (s *OrderCashService) AcceptOrder(data dtos.AcceptOrderDTO) (int, map[string]interface{}, error) {
	// Local helpers for safe type conversion from map[string]interface{} DB scan results.
	asString := func(v interface{}) string {
		switch s := v.(type) {
		case string:
			return s
		case []byte:
			return string(s)
		case nil:
			return ""
		default:
			return fmt.Sprintf("%v", s)
		}
	}
	asInt := func(v interface{}) int {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		case uint64:
			return int(n)
		default:
			return 0
		}
	}

	// 1. Find the order filtered by order_id, customer_id, and allowed statuses.
	orderData, err := s.OrderTransactionRepo.FindDynamicOrderTransactionMap(
		nil,
		map[string]interface{}{
			"id":          data.OrderID,
			"customer_id": data.CustomerID,
		},
		"order_status IN ?",
		[]interface{}{
			[]string{"FINDING_MITRA", "WAITING_FOR_SELECTED_MITRA"},
		},
	)
	if err != nil {
		fmt.Printf("[AcceptOrder] ERROR FindDynamicOrderTransactionMap: %v\n", err)
		return 500, nil, err
	}
	if orderData == nil {
		fmt.Printf("[AcceptOrder] ERROR orderData nil\n")
		return 409, nil, errors.New("order not found")
	}

	// 2. Extract payment_id with safe type assertion.
	var paymentID int
	switch v := orderData["payment_id"].(type) {
	case int:
		paymentID = v
	case int64:
		paymentID = int(v)
	case float64:
		paymentID = int(v)
	}
	payment, err := s.PaymentRepo.FindById(paymentID)
	if err != nil {
		fmt.Printf("[AcceptOrder] ERROR FindById Payment: %v\n", err)
		return 500, nil, err
	}
	if payment == nil {
		fmt.Printf("[AcceptOrder] ERROR payment nil\n")
		return 404, nil, errors.New("payment not found")
	}

	// 3. Extract customer_id and fetch customer data.
	var customerID string
	switch v := orderData["customer_id"].(type) {
	case string:
		customerID = v
	case []byte:
		customerID = string(v)
	default:
		fmt.Printf("[AcceptOrder] ERROR customer_id tipe tidak dikenali\n")
		return 500, nil, fmt.Errorf("customer_id tipe tidak dikenali")
	}

	customer, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil {
		fmt.Printf("[AcceptOrder] ERROR FindCustomerById: %v\n", err)
		return 500, nil, err
	}
	if customer == nil {
		fmt.Printf("[AcceptOrder] ERROR customer nil\n")
		return 404, nil, errors.New("customer not found")
	}

	mitra, err := s.UserRepo.FindMitraById(data.MitraID)
	if err != nil {
		fmt.Printf("[AcceptOrder] ERROR FindMitraById: %v\n", err)
		return 500, nil, err
	}
	if mitra == nil {
		fmt.Printf("[AcceptOrder] ERROR mitra nil\n")
		return 404, nil, errors.New("mitra not found")
	}

	// 4. Check Redis for booking state.
	playerMitraIDs, err := helpers.GetValue(data.TempID)
	if err != nil && err != redis.Nil {
		fmt.Printf("[AcceptOrder] ERROR GetValue TempID: %v\n", err)
		return 500, nil, err
	}

	bookedOrderSign, err := helpers.GetValue(fmt.Sprintf("BOOKED_ORDER_%s", data.OrderID))
	if err != nil && err != redis.Nil {
		fmt.Printf("[AcceptOrder] ERROR GetValue BOOKED_ORDER: %v\n", err)
		return 500, nil, err
	}

	log.Printf("PLAYER MITRA IDS : %s", playerMitraIDs)

	if bookedOrderSign != "" {
		fmt.Printf("[AcceptOrder] order was taken by another mitra\n")
		return 409, nil, errors.New("this order was taken by another mitra")
	}

	// Lock the order for this mitra before doing any DB writes.
	if err = helpers.SetValue(
		fmt.Sprintf("BOOKED_ORDER_%s", data.OrderID),
		fmt.Sprintf("BOOKED_FOR_MITRA_%s", mitra.CompleteName),
	); err != nil {
		fmt.Printf("[AcceptOrder] ERROR SetValue BOOKED_ORDER: %v\n", err)
		return 500, nil, err
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 5. Notify other mitras in the broadcast group that they lost the order.
	var playerMitraIDArray []string
	if playerMitraIDs != "" {
		playerMitraIDArray = strings.Split(playerMitraIDs, ",")
	}

	if len(playerMitraIDArray) >= 1 {
		var filteredIDs []string
		for _, id := range playerMitraIDArray {
			if id != mitra.ID {
				filteredIDs = append(filteredIDs, id)
			}
		}

		var firebaseTokenArray []string
		if len(filteredIDs) > 0 {
			var users []models.User
			if err := tx.Select("firebase_token").Where("id IN ?", filteredIDs).
				Find(&users).Error; err != nil {
				tx.Rollback()
				fmt.Printf("[AcceptOrder] ERROR Find users for LOST_BROADCAST: %v\n", err)
				return 500, nil, err
			}
			for _, u := range users {
				if u.FirebaseToken != nil && *u.FirebaseToken != "" && *u.FirebaseToken != "null" {
					firebaseTokenArray = append(firebaseTokenArray, *u.FirebaseToken)
				}
			}
		}

		if len(firebaseTokenArray) > 0 {
			payloadMessage := map[string]interface{}{
				"data": map[string]string{
					"notification_type": "LOST_BROADCAST",
					"title":             "Yah...kamu kehilangan order",
					"message":           fmt.Sprintf("Tawaran order dari customer %s telah diambil mitra lain", customer.CompleteName),
					"order_id":          data.OrderID,
					"notification_id":   fmt.Sprintf("%v", orderData["notification_id"]),
					"customer_id":       data.CustomerID,
					"notif_type":        "order",
				},
				"tokens": firebaseTokenArray,
			}
			if _, err = service.SendMulticast(s.DB, "mitra", payloadMessage); err != nil {
				log.Printf("[AcceptOrder] WARNING SendMulticast LOST_BROADCAST: %v", err)
			}
		}
	}

	// 6. Generate RSA and Diffie-Hellman keys.
	publicKey, privateKey, err := helpers.GenerateRsaKey()
	if err != nil {
		tx.Rollback()
		fmt.Printf("[AcceptOrder] ERROR GenerateRsaKey: %v\n", err)
		return 500, nil, err
	}
	publicKeyMitra := helpers.GeneratePublicKey(customer.SharedPrime, customer.SharedBase, customer.SharedSecret)
	publicKeyCustomer := helpers.GeneratePublicKey(customer.SharedPrime, customer.SharedBase, mitra.SharedSecret)
	mitraSecret := helpers.GenerateSharedSecret(publicKeyCustomer, customer.SharedSecret, customer.SharedPrime)
	customerSecret := helpers.GenerateSharedSecret(publicKeyMitra, mitra.SharedSecret, customer.SharedPrime)

	// 7. Safely extract async job IDs for later cleanup.
	expiredJobId := asString(orderData["offer_expired_job_id"])
	offerSelectedJobId := asString(orderData["offer_selected_job_id"])
	orderType := asString(orderData["order_type"])

	var orderStatus string
	if orderType == "now" {
		orderStatus = "OTW"
	} else {
		orderStatus = "WAIT_SCHEDULE"
	}

	// 8. Update the order transaction with mitra details, keys, and new status.
	payloadOrderUpdate := map[string]interface{}{
		"mitra_id":              data.MitraID,
		"private_key_rsa":       privateKey,
		"public_key_rsa":        publicKey,
		"public_key_mitra":      publicKeyMitra,
		"public_key_customer":   publicKeyCustomer,
		"mitra_secret":          mitraSecret,
		"customer_secret":       customerSecret,
		"notification_id":       0,
		"offer_expired_job_id":  nil,
		"offer_selected_job_id": nil,
		"order_status":          orderStatus,
	}
	if orderType == "now" {
		payloadOrderUpdate["shared_prime"] = customer.SharedPrime
	}
	where := map[string]interface{}{
		"AND": map[string]interface{}{
			"id": data.OrderID,
			"order_status": []string{
				"FINDING_MITRA",
				"WAITING_FOR_SELECTED_MITRA",
			},
		},
	}
	if err := s.OrderTransactionRepo.UpdateWithConditions(tx, where, payloadOrderUpdate); err != nil {
		tx.Rollback()
		fmt.Printf("[AcceptOrder] ERROR UpdateWithConditions: %v\n", err)
		return 500, nil, err
	}

	// 9. Handle order-type-specific updates.
	if orderType == "repeat" {
		// Move all FINDING_MITRA repeat sub-orders to WAIT_SCHEDULE.
		if err := tx.Model(&models.OrderTransactionRepeat{}).
			Where("order_id = ? AND order_status = ?", data.OrderID, "FINDING_MITRA").
			Update("order_status", "WAIT_SCHEDULE").Error; err != nil {
			tx.Rollback()
			fmt.Printf("[AcceptOrder] ERROR update OrderTransactionRepeat: %v\n", err)
			return 500, nil, err
		}
	}

	if orderType == "now" {
		// Mark mitra as busy and running this order.
		if _, err := s.UserRepo.UpdateUserData(tx,
			map[string]interface{}{
				"id":        data.MitraID,
				"user_type": "mitra",
			},
			map[string]interface{}{
				"is_busy":                "yes",
				"user_status":            "on progress",
				"order_id_running":       data.OrderID,
				"customer_id_running":    customerID,
				"service_id_running":     orderData["service_id"],
				"sub_service_id_running": orderData["sub_service_id"],
			},
		); err != nil {
			tx.Rollback()
			fmt.Printf("[AcceptOrder] ERROR UpdateUserData Mitra: %v\n", err)
			return 500, nil, err
		}
	}

	// 10. Remove all competing offers for this order.
	if err := s.OrderOfferRepo.DeleteByWhere(tx, map[string]interface{}{
		"order_id": data.OrderID,
	}); err != nil {
		tx.Rollback()
		fmt.Printf("[AcceptOrder] ERROR DeleteByWhere OrderOffer: %v\n", err)
		return 500, nil, err
	}

	// 11. Create the order chat room between customer and mitra.
	payloadCreateChat := models.OrderChat{
		ID:           uuid.NewString(),
		OrderID:      data.OrderID,
		CustomerID:   customerID,
		MitraID:      mitra.ID,
		ServiceID:    asInt(orderData["service_id"]),
		SubServiceID: asInt(orderData["sub_service_id"]),
	}
	if err := s.OrderChatRepo.Create(tx, payloadCreateChat); err != nil {
		tx.Rollback()
		fmt.Printf("[AcceptOrder] ERROR Create OrderChat: %v\n", err)
		return 500, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		fmt.Printf("[AcceptOrder] ERROR Commit: %v\n", err)
		return 500, nil, err
	}

	// 12. Schedule coming-soon run and warning tasks via asynq.
	if orderType == "coming soon" {
		var orderTime time.Time
		switch v := orderData["order_time"].(type) {
		case time.Time:
			orderTime = v
		case []byte:
			orderTime, _ = time.Parse("2006-01-02 15:04:05", string(v))
		case string:
			orderTime, _ = time.Parse("2006-01-02 15:04:05", v)
		}
		if !orderTime.IsZero() {
			runPayload, _ := queue.NewOrderComingSoonRunTask(data.OrderID)
			runTask := asynq.NewTask(queue.TypeOrderComingSoonRun, runPayload)
			if _, err := queue.AsynqClient.Enqueue(runTask, asynq.ProcessAt(orderTime)); err != nil {
				log.Printf("[AcceptOrder] WARNING enqueue ComingSoonRun: %v", err)
			}
			warningPayload, _ := queue.NewOrderComingSoonWarningTask(data.OrderID)
			warningTask := asynq.NewTask(queue.TypeOrderComingSoonWarning, warningPayload)
			if _, err := queue.AsynqClient.Enqueue(warningTask, asynq.ProcessAt(orderTime.Add(3*time.Minute))); err != nil {
				log.Printf("[AcceptOrder] WARNING enqueue ComingSoonWarning: %v", err)
			}
		}
	}

	// 13. Remove the offer-expired and offer-selected queue jobs.
	if expiredJobId != "" {
		if err := queue.Inspector.DeleteTask(queue.TypeOrderOfferExpired, expiredJobId); err != nil {
			log.Printf("[AcceptOrder] WARNING DeleteTask OfferExpired: %v", err)
		}
	}
	if offerSelectedJobId != "" {
		if err := queue.Inspector.DeleteTask(queue.TypeOrderSelectedExpired, offerSelectedJobId); err != nil {
			log.Printf("[AcceptOrder] WARNING DeleteTask OfferSelectedExpired: %v", err)
		}
	}

	// 14. Notify customer that a mitra has been found.
	if customer.FirebaseToken != nil && *customer.FirebaseToken != "" {
		payloadCustomer := map[string]interface{}{
			"data": map[string]interface{}{
				"notification_type": "GOT_ORDER",
				"title":             "Kamu dapat mitra",
				"message":           "Halo, kamu sudah dapat mitra untuk order mu",
				"order_id":          data.OrderID,
				"mitra_id":          data.MitraID,
				"customer_id":       data.CustomerID,
				"notif_type":        "order",
			},
			"tokens": []string{*customer.FirebaseToken},
		}
		if _, err = service.SendMulticast(s.DB, "customer", payloadCustomer); err != nil {
			log.Printf("[AcceptOrder] WARNING SendMulticast Customer: %v", err)
		}
	}

	// 15. Clean up Redis broadcast keys.
	if err := helpers.DeleteValue(data.TempID); err != nil {
		log.Printf("[AcceptOrder] WARNING DeleteValue TempID: %v", err)
	}
	if err := helpers.DeleteValue(fmt.Sprintf("SELECT_MITRA_%s", data.OrderID)); err != nil {
		log.Printf("[AcceptOrder] WARNING DeleteValue SELECT_MITRA: %v", err)
	}

	// 16. Broadcast live order counts to all online admin sockets.
	onlineAdminList, err := s.UserRepo.FindOnlineAdmins()
	if err != nil {
		log.Printf("[AcceptOrder] WARNING FindOnlineAdmins: %v", err)
	}
	runningCount, _ := s.OrderTransactionRepo.CountRunningOrders()
	waitingCount, _ := s.OrderTransactionRepo.CountWaitingOrders()
	for _, socketID := range onlineAdminList {
		realtime.Server.BroadcastToRoom(
			"/",
			socketID,
			constants.MESSAGE_SOCKET_ADMIN,
			map[string]interface{}{
				"notification_type":   constants.NOTIFICATION_ORDER_RUNNING,
				"order_id":            data.OrderID,
				"order_running_count": runningCount,
				"order_waiting_count": waitingCount,
			},
		)
	}

	payloadResponse := map[string]interface{}{
		"order_id":       data.OrderID,
		"sub_id":         data.SubID,
		"temp_id":        data.TempID,
		"customer_id":    data.CustomerID,
		"mitra_id":       data.MitraID,
		"order_type":     data.OrderType,
		"payment_type":   payment.Type,
		"server_message": "successfully took the order",
		"status":         "success",
	}
	if orderType == "now" {
		payloadResponse["shared_prime"] = customer.SharedPrime
	}
	return 200, payloadResponse, nil
}
