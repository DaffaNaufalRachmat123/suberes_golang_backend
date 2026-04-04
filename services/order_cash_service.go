package services

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
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
	layout := "2006-01-02 15:04:05"
	orderDateTime, err := time.ParseInLocation(layout, helpers.NormalizeDateTimeString(dto.OrderTime), loc)
	if err != nil {
		return "", 0, "", "", 400, err
	}

	if dto.OrderType == "coming soon" {

		// STRICT: orderDateTime.getDate() >= nowDateTime.getDate()
		if orderDateTime.Day() >= nowDateTime.Day() &&
			orderDateTime.Hour() >= 7 {

			// orderDateTime.setMinutes(orderDateTime.getMinutes() + minutesSubServices)
			orderDateTime = orderDateTime.Add(
				time.Duration(subService.MinutesSubServices) * time.Minute,
			)

			// if (hour >= 23 && minute > 0)
			if orderDateTime.Hour() >= 23 &&
				orderDateTime.Minute() > 0 {
				return "", 0, "", "", 400, errors.New("Batas maksimal order di jam 11 malam")
			}
		}
	} else if dto.OrderType == "repeat" {
		nowDate, err := helpers.GetTimezoneNowDateReturnDate(dto.TimezoneCode)
		if err != nil {
			return "", 0, "", "", 500, err
		}
		for i := 0; i < len(dto.OrderRepeatList); i++ {
			nowDateSecond := time.Date(
				nowDate.Year(),
				nowDate.Month(),
				nowDate.Day(),
				nowDate.Hour(),
				nowDate.Minute(),
				0,
				nowDate.Nanosecond(),
				nowDate.Location(),
			)
			orderDateTime := time.Now()
			orderTimeSplit := strings.Split(dto.OrderRepeatList[i].OrderTime, " ")
			orderTimeDate := strings.Split(orderTimeSplit[0], "-")
			orderTimeTime := strings.Split(orderTimeSplit[1], ":")
			year, _ := strconv.Atoi(orderTimeDate[0])

			monthStr := orderTimeDate[1]
			if len(monthStr) == 2 && monthStr[0] == '0' {
				monthStr = string(monthStr[1])
			}
			monthInt, _ := strconv.Atoi(monthStr)

			day, _ := strconv.Atoi(orderTimeDate[2])

			hour, _ := strconv.Atoi(orderTimeTime[0])
			minute, _ := strconv.Atoi(orderTimeTime[1])

			orderDateTime = time.Date(
				year,
				time.Month(monthInt),
				day,
				hour,
				minute,
				0,
				0,
				nowDate.Location(),
			)

			if orderDateTime.Day() >= nowDateSecond.Day() && orderDateTime.Hour() >= 7 {

				orderDateTime = orderDateTime.Add(
					time.Duration(subService.MinutesSubServices) * time.Minute,
				)

				fmt.Printf("After Add : %02d:%02d\n",
					orderDateTime.Hour(),
					orderDateTime.Minute(),
				)

				if orderDateTime.Hour() >= 23 && orderDateTime.Minute() > 0 {
					return "", 0, "", "", 400, errors.New("Batas maksimal jam order di jam 11 malam")
				}
			}
		}
	} else if dto.OrderType == "now" {
		nowTime := helpers.GetTimezoneNowDate(dto.TimezoneCode)
		splitNowTime := strings.Split(nowTime, " ")
		splitNowDateTime := strings.Split(splitNowTime[0], "-")
		splitNowTimeTime := strings.Split(splitNowTime[1], ":")

		nowYear, _ := strconv.Atoi(splitNowDateTime[0])
		nowMonth, _ := strconv.Atoi(splitNowDateTime[1])
		nowDay, _ := strconv.Atoi(splitNowDateTime[2])
		nowHours, _ := strconv.Atoi(splitNowTimeTime[0])
		nowMinutes, _ := strconv.Atoi(splitNowTimeTime[1])

		dateAdd := time.Date(
			nowYear,
			time.Month(nowMonth),
			nowDay,
			nowHours,
			nowMinutes,
			0,
			0,
			nowDateTime.Location(),
		)

		fmt.Printf("Now Time : %s\n", nowTime)

		if service.ServiceType == "Durasi" {

			dateAdd = dateAdd.Add(time.Duration(subService.MinutesSubServices) * time.Minute)

			addedDate := helpers.GetFormattedYearMonthDate(dateAdd)
			splitAddDate := strings.Split(addedDate, " ")
			splitTime := strings.Split(splitAddDate[1], ":")

			addedHours, _ := strconv.Atoi(splitTime[0])
			addedMinutes, _ := strconv.Atoi(splitTime[1])

			if nowHours >= 23 && nowMinutes >= 0 {
				return "", 0, "", "", 400, errors.New("Batas maksimal jam operasional sampai jam 11 malam untuk layanan ini")
			} else if addedHours >= 23 && addedMinutes >= 0 {
				return "", 0, "", "", 400, errors.New("Batas maksimal jam operasional sampai jam 11 malam untuk layanan ini")
			}
		}

		if nowHours >= 20 && nowMinutes >= 0 {
			return "", 0, "", "", 400, errors.New("Batas maksimal jam order sampai jam 8 malam")
		}
	}
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
	orderData, err := s.OrderTransactionRepo.FindDynamicOrderTransactionMap(
		nil,
		nil,
		"order_status IN ?",
		[]interface{}{
			[]string{"FINDING_MITRA", "WAITING_FOR_SELECTED_MITRA"},
		},
	)
	if err != nil {
		return 500, nil, err
	}
	if orderData == nil {
		return 404, nil, err
	}
	payment, err := s.PaymentRepo.FindById(orderData["payment_id"].(int))
	if err != nil {
		return 500, nil, err
	}
	if payment == nil {
		return 404, nil, err
	}
	customer, err := s.UserRepo.FindCustomerById(orderData["customer_id"].(string))
	if err != nil {
		return 500, nil, err
	}
	if customer == nil {
		return 404, nil, err
	}
	mitra, err := s.UserRepo.FindMitraById(orderData["mitra_id"].(string))
	if err != nil {
		return 500, nil, err
	}
	if mitra == nil {
		return 404, nil, err
	}
	playerMitraIDs, err := helpers.GetValue(data.TempID)
	if err != nil {
		return 500, nil, err
	}

	bookedOrderSign, _ := helpers.GetValue(fmt.Sprintf("BOOKED_ORDER_%s", orderData["id"]))

	log.Printf("PLAYER MITRA IDS : %s", playerMitraIDs)

	if bookedOrderSign != "" {
		return 409, nil, errors.New("this order was taken by another mitra")
	}

	err = helpers.SetValue(
		fmt.Sprintf("BOOKED_ORDER_%s", orderData["id"]),
		fmt.Sprintf("BOOKED_FOR_MITRA_%s", mitra.CompleteName),
	)
	if err != nil {
		return 500, nil, err
	}

	tx := s.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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
				return 500, nil, err
			}

			for _, u := range users {
				if u.FirebaseToken != nil {
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
					"order_id":          fmt.Sprintf("%v", orderData["id"]),
					"notification_id":   fmt.Sprintf("%v", orderData["notification_id"]),
					"customer_id":       fmt.Sprintf("%v", data.CustomerID),
					"notif_type":        "order",
				},
				"tokens": firebaseTokenArray,
			}

			_, err = service.SendMulticast(s.DB, "mitra", payloadMessage)
			if err != nil {
				log.Println("error:", err)
			}
		}
	}
	publicKey, privateKey, err := helpers.GenerateRsaKey()
	if err != nil {
		return 500, nil, err
	}
	publicKeyMitra := helpers.GeneratePublicKey(customer.SharedPrime, customer.SharedBase, customer.SharedSecret)
	publicKeyCustomer := helpers.GeneratePublicKey(customer.SharedPrime, customer.SharedBase, mitra.SharedSecret)
	mitraSecret := helpers.GenerateSharedSecret(publicKeyCustomer, customer.SharedSecret, customer.SharedPrime)
	customerSecret := helpers.GenerateSharedSecret(publicKeyMitra, mitra.SharedSecret, customer.SharedPrime)

	expiredJobId := orderData["offer_expired_job_id"].(string)
	offerSelectedJobId := orderData["offer_selected_job_id"].(string)

	var orderStatus string
	if orderData["order_type"] == "now" {
		orderStatus = "OTW"
	} else {
		orderStatus = "WAIT_SCHEDULE"
	}

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
	if orderData["order_type"] == "now" {
		payloadOrderUpdate["shared_prime"] = customer.SharedPrime
	}
	where := map[string]interface{}{
		"OR": map[string]interface{}{
			"order_status": []string{
				"FINDING_MITRA",
				"WAITING_FOR_SELECTED_MITRA",
			},
		},
	}
	if err := s.OrderTransactionRepo.UpdateWithConditions(tx, where, payloadOrderUpdate); err != nil {
		tx.Rollback()
		return 500, nil, err
	}
	if orderData["order_type"] == "repeat" {
		if err := s.OrderTransactionRepo.UpdateWithConditions(tx, map[string]interface{}{
			"order_id":     data.OrderID,
			"order_status": "FINDING_MITRA",
		}, map[string]interface{}{
			"order_status": "WAIT_SCHEDULE",
		}); err != nil {
			tx.Rollback()
			return 500, nil, err
		}
	}
	if orderData["order_type"] == "now" {
		if _, err := s.UserRepo.UpdateUserData(tx, map[string]interface{}{
			"id":        data.MitraID,
			"user_type": "mitra",
		}, map[string]interface{}{
			"is_busy":                "yes",
			"user_status":            "on progress",
			"order_id_running":       orderData["id"],
			"customer_id_running":    orderData["customer_id"],
			"service_id_running":     orderData["service_id"],
			"sub_service_id_running": orderData["sub_service_id_running"],
		}); err != nil {
			tx.Rollback()
			return 500, nil, err
		}
	} else if orderData["order_type"] == "coming soon" {
		dateTimeSplit := strings.Split(orderData["order_time"].(string), " ")
		dateSplit := strings.Split(dateTimeSplit[0], "-")
		timeSplit := strings.Split(dateTimeSplit[1], ":")

		year := dateSplit[0]
		monthInt, _ := strconv.Atoi(dateSplit[1])
		month := monthInt - 1
		day := dateSplit[2]
		hour := timeSplit[0]
		minute := timeSplit[1]

		yearInt, _ := strconv.Atoi(year)
		dayInt, _ := strconv.Atoi(day)
		hourInt, _ := strconv.Atoi(hour)
		minuteInt, _ := strconv.Atoi(minute)

		scheduleDate := time.Date(yearInt, time.Month(month+1), dayInt, hourInt, minuteInt, 0, 0, time.Local)
		scheduleWarningDate := scheduleDate.Add(3 * time.Minute)

		nodeScheduleDate := helpers.GetFormattedYearMonthDateTimeZone(scheduleDate, "UTC")
		nodeScheduleWarningDate := helpers.GetFormattedYearMonthDateTimeZone(scheduleWarningDate, "UTC")

		splitScheduledDate := strings.Split(nodeScheduleDate, " ")
		scheduleSplitDate := strings.Split(splitScheduledDate[0], "-")
		scheduleSplitTime := strings.Split(splitScheduledDate[1], ":")

		yearScheduled := scheduleSplitDate[0]
		monthScheduled := scheduleSplitDate[1]
		datesScheduled := scheduleSplitDate[2]
		hoursScheduled := scheduleSplitTime[0]
		minutesScheduled := scheduleSplitTime[1]

		if len(monthScheduled) == 2 {
			if strings.Split(monthScheduled, "")[0] == "0" {
				monthScheduled = strings.Split(monthScheduled, "")[1]
			}
		}
		if len(datesScheduled) == 2 {
			if strings.Split(datesScheduled, "")[0] == "0" {
				datesScheduled = strings.Split(datesScheduled, "")[1]
			}
		}
		if len(hoursScheduled) == 2 {
			if strings.Split(hoursScheduled, "")[0] == "0" {
				hoursScheduled = strings.Split(hoursScheduled, "")[1]
			}
		}
		if len(minutesScheduled) == 2 {
			if strings.Split(minutesScheduled, "")[0] == "0" {
				minutesScheduled = strings.Split(minutesScheduled, "")[1]
			}
		}

		splitWarningDate := strings.Split(nodeScheduleWarningDate, " ")
		warningSplitDate := strings.Split(splitWarningDate[0], "-")
		warningSplitTime := strings.Split(splitWarningDate[1], ":")

		yearWarning := warningSplitDate[0]
		monthWarning := warningSplitDate[1]
		dateWarning := warningSplitDate[2]
		hourWarning := warningSplitTime[0]
		minuteWarning := warningSplitTime[1]

		if len(monthWarning) == 2 {
			if strings.Split(monthWarning, "")[0] == "0" {
				monthWarning = strings.Split(monthWarning, "")[1]
			}
		}
		if len(dateWarning) == 2 {
			if strings.Split(dateWarning, "")[0] == "0" {
				dateWarning = strings.Split(dateWarning, "")[1]
			}
		}
		if len(hourWarning) == 2 {
			if strings.Split(hourWarning, "")[0] == "0" {
				hourWarning = strings.Split(hourWarning, "")[1]
			}
		}
		if len(minuteWarning) == 2 {
			if strings.Split(minuteWarning, "")[0] == "0" {
				minuteWarning = strings.Split(minuteWarning, "")[1]
			}
		}
		loc, err := time.LoadLocation(orderData["timezone_code"].(string))
		if err != nil {
			return 500, nil, err
		}
		minutes, _ := strconv.Atoi(minutesScheduled)
		hours, _ := strconv.Atoi(hoursScheduled)
		days, _ := strconv.Atoi(datesScheduled)
		months, _ := strconv.Atoi(monthScheduled)
		years, _ := strconv.Atoi(yearScheduled)

		scheduledTime := time.Date(
			years,
			time.Month(months), // No need subtract -1
			days,
			hours,
			minutes,
			0, // second
			0, // nanosecond
			loc,
		)
		queue.ScheduleOnceWithCallbackAt(scheduledTime, func() error {
			orderData, err := s.OrderTransactionRepo.FindDynamicOrderTransactionMap(
				nil,
				nil,
				"order_status IN ?",
				[]interface{}{
					[]string{"WAIT_SCHEDULE"},
				},
			)
			if err != nil {
				return err
			}
			if orderData == nil {
				return errors.New("Order data wait schedule not found")
			}
			customerLoad, err := s.UserRepo.FindCustomerById(orderData["customre_id"].(string))
			if err != nil {
				return err
			}
			if customerLoad == nil {
				return errors.New("Customer not found")
			}
			mitraLoad, err := s.UserRepo.FindMitraById(orderData["mitra_id"].(string))
			if err != nil {
				return err
			}
			if mitraLoad == nil {
				return errors.New("Mitra not found")
			}
			mitraPayload := map[string]interface{}{
				"data": map[string]interface{}{
					"notification_type": "COMING_SOON_ORDER_RUN_NOTIFICATION",
					"title":             "Order Terjadwal",
					"message":           fmt.Sprintf("Halo %s harap jalankan orderan mu sekarang", mitraLoad.CompleteName),
					"order_id":          orderData["id"].(string),
					"mitra_id":          orderData["mitra_id"].(string),
					"sub_order_id":      "-1",
					"customer_id":       orderData["customer_id"].(string),
					"notif_type":        "order",
				},
				"tokens": []string{*mitraLoad.FirebaseToken},
			}
			_, err = service.SendMulticast(s.DB, "mitra", mitraPayload)
			if err != nil {
				log.Println("error:", err)
			}

			customerPayload := map[string]interface{}{
				"data": map[string]interface{}{
					"notification_type": "COMING_SOON_ORDER_NOTIFICATION",
					"title":             "Order Terjadwal",
					"message": fmt.Sprintf(
						"Mitra %s seharusnya menjalankan orderan mu sekarang", mitraLoad.CompleteName,
					),
					"order_id":     orderData["id"].(string),
					"mitra_id":     orderData["mitra_id"].(string),
					"sub_order_id": "-1",
					"customer_id":  orderData["customer_id"].(string),
					"notif_type":   "order",
				},
				"tokens": []string{
					*customerLoad.FirebaseToken,
				},
			}
			_, err = service.SendMulticast(s.DB, "customer", customerPayload)
			if err != nil {
				log.Println("error:", err)
			}
			return nil
		})

		minutesWarning, _ := strconv.Atoi(minuteWarning)
		hoursWarning, _ := strconv.Atoi(hourWarning)
		daysWarning, _ := strconv.Atoi(dateWarning)
		monthsWarning, _ := strconv.Atoi(monthWarning)
		yearsWarning, _ := strconv.Atoi(yearWarning)

		scheduledWarningTime := time.Date(
			yearsWarning,
			time.Month(monthsWarning), // No need subtract -1
			daysWarning,
			hoursWarning,
			minutesWarning,
			0, // second
			0, // nanosecond
			loc,
		)
		queue.ScheduleOnceWithCallbackAt(scheduledWarningTime, func() error {
			orderWarningData, err := s.OrderTransactionRepo.FindById(orderData["id"].(string))
			if err != nil {
				return err
			}
			if orderWarningData == nil {
				return errors.New("Order data wait schedule not found")
			}
			paymentData, err := s.PaymentRepo.FindById(orderWarningData.PaymentID)
			if err != nil {
				return err
			}
			if paymentData == nil {
				return errors.New("Payment data not found")
			}
			customerData, err := s.UserRepo.FindCustomerById(orderWarningData.CustomerID)
			if err != nil {
				return err
			}
			if customerData == nil {
				return errors.New("Customer data not found")
			}
			mitraData, err := s.UserRepo.FindMitraById(helpers.DerefStr(orderWarningData.MitraID))
			if err != nil {
				return err
			}
			if mitraData == nil {
				return errors.New("Mitra data not found")
			}
			tx := s.DB.Begin()
			if mitraData.IsBusy == "no" && orderWarningData.OrderStatus == "WAIT_SCHEDULE" {
				err := s.OrderTransactionRepo.UpdateWithConditions(tx,
					map[string]interface{}{
						"id":          orderWarningData.ID,
						"mitra_id":    orderWarningData.MitraID,
						"customer_id": orderWarningData.CustomerID,
					}, map[string]interface{}{
						"order_status": "CANCELED",
					})
				if err != nil {
					log.Println("err : " + err.Error())
					return err
				}
				mitraPayload := map[string]interface{}{
					"data": map[string]interface{}{
						"notification_type": "CANCELED_ORDER_COMING_SOON_NOTIFICATION",
						"title":             "Perhatian !",
						"message":           fmt.Sprintf("Halo %s performa kamu diturunkan karena tidak menjalankan order", mitraData.CompleteName),
						"order_id":          orderWarningData.ID,
						"sub_order_id":      "-1",
						"mitra_id":          orderWarningData.MitraID,
						"customer_id":       orderWarningData.CustomerID,
						"notif_type":        "order",
					},
					"tokens": []string{*mitraData.FirebaseToken},
				}
				_, err = service.SendMulticast(s.DB, "mitra", mitraPayload)
				if err != nil {
					log.Println("err : " + err.Error())
					return err
				}
				msgCustomer := fmt.Sprintf("Order dibatalkan otomatis karena mitra %s tidak menjalankan order", mitraData.CompleteName)
				if paymentData.Type == "virtual account" || paymentData.Type == "ewallet" {
					msgCustomer += " dan uang kamu akan dikembalikan segera"
				}
				customerPayload := map[string]interface{}{
					"data": map[string]interface{}{
						"notification_type": "CANCELED_ORDER_LATE_NOTIFICATION",
						"title":             "Order terlambat & dibatalkan",
						"message":           msgCustomer,
						"order_id":          orderWarningData.ID,
						"sub_order_id":      "-1",
						"mitra_id":          orderWarningData.MitraID,
						"customer_id":       orderWarningData.CustomerID,
						"notif_type":        "order",
					},
					"tokens": []string{*customerData.FirebaseToken},
				}
				_, err = service.SendMulticast(s.DB, "customer", customerPayload)
				if err != nil {
					log.Println("err : " + err.Error())
					return err
				}
			}
			return nil
		})
	} else if orderData["order_type"] == "repeat" {
		orderRepeatData, err := s.OrderTransactionRepeatRepo.FindByWhereDynamic(tx, map[string]interface{}{
			"order_id":     orderData["id"],
			"order_status": "FINDING_MITRA",
		})
		if err != nil {
			return 500, nil, err
		}
		if orderRepeatData == nil {
			return 404, nil, errors.New("Order not found")
		}
		// if len(orderRepeatData) {
		// 	for _, elem := range orderRepeatData {
		// 		dateTimeSplitRepeat := strings.Split(elem.OrderTime , " ")
		// 		dateSplit := strings.Split(dateTimeSplitRepeat[0], "-")
		// 		timeSplit := strings.Split(dateTimeSplitRepeat[1], ":")
		// 		year := dateSplit[0]
		// 		monthInt, _ := strconv.Atoi(dateSplit[1])
		// 		day := dateSplit[2]
		// 		hour := timeSplit[0]
		// 		minute := timeSplit[1]

		// 	}
		// }
	}
	err = s.OrderOfferRepo.DeleteByWhere(tx, map[string]interface{}{
		"order_id": data.OrderID,
	})
	if err != nil {
		return 500, nil, err
	}
	payloadCreateChat := models.OrderChat{
		ID:           string(uuid.NewString()),
		OrderID:      data.OrderID,
		CustomerID:   orderData["customer_id"].(string),
		MitraID:      mitra.ID,
		ServiceID:    orderData["service_id"].(int),
		SubServiceID: orderData["sub_service_id"].(int),
	}
	err = s.OrderChatRepo.Create(tx, payloadCreateChat)
	if err != nil {
		return 500, nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return 500, nil, err
	}
	if expiredJobId != "" {
		err = queue.Inspector.DeleteTask(queue.TypeOrderOfferExpired, expiredJobId)
		if err != nil {
			return 500, nil, err
		}
	}
	if offerSelectedJobId != "" {
		err = queue.Inspector.DeleteTask(queue.TypeOrderSelectedExpired, offerSelectedJobId)
		if err != nil {
			return 500, nil, err
		}
	}
	if customer.FirebaseToken != nil {
		payloadCustomer := map[string]interface{}{
			"data": map[string]interface{}{
				"notification_type": "GOT_ORDER",
				"title":             "Kamu dapet mitra",
				"message":           "Halo, kamu sudah dapat mitra untuk order mu",
				"order_id":          data.OrderID,
				"mitra_id":          data.MitraID,
				"customer_id":       data.CustomerID,
				"notif_type":        "order",
			},
			"tokens": []string{*customer.FirebaseToken},
		}
		_, err = service.SendMulticast(s.DB, "customer", payloadCustomer)
		if err != nil {
			log.Println("error:", err)
		}
	}
	err = helpers.DeleteValue(data.TempID)
	if err != nil {
		return 500, nil, err
	}
	err = helpers.DeleteValue(fmt.Sprintf("SELECT_MITRA_%s", orderData["id"]))
	if err != nil {
		return 500, nil, err
	}
	onlineAdminList, err := s.UserRepo.FindOnlineAdmins()
	if err != nil {
		return 500, nil, err
	}
	runningCount, err := s.OrderTransactionRepo.CountRunningOrders()
	if err != nil {
		log.Println(err)
		return 500, nil, err
	}

	waitingCount, err := s.OrderTransactionRepo.CountWaitingOrders()
	if err != nil {
		log.Println(err)
		return 500, nil, err
	}
	if len(onlineAdminList) > 0 {
		for _, socketID := range onlineAdminList {
			realtime.Server.BroadcastToRoom(
				"/",
				socketID,
				constants.MESSAGE_SOCKET_ADMIN,
				map[string]interface{}{
					"notification_type":   constants.NOTIFICATION_ORDER_RUNNING,
					"order_id":            orderData["id"],
					"order_running_count": runningCount,
					"order_waiting_count": waitingCount,
				},
			)
		}
	}
	payloadResponse := map[string]interface{}{
		"order_id":       data.OrderID,
		"sub_id":         data.SubID,
		"temp_id":        data.TempID,
		"customer_id":    data.CustomerID,
		"mitra_id":       data.MitraID,
		"order_type":     data.OrderType,
		"payment_type":   payment.Type,
		"server_message": "Sucessfully took the order",
		"status":         "success",
	}
	if orderData["order_type"] == "now" {
		payloadResponse["shared_prime"] = customer.SharedPrime
	}
	return 200, payloadResponse, nil
}
