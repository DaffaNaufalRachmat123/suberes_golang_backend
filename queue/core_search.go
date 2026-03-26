package queue

import (
	"fmt"
	"os"
	"strconv"
	"suberes_golang/config"
	"suberes_golang/models"
)

func searchQueryNowCash(latitude float64, longitude float64, isActive string, isBusy string, distance float64, isAutoBid string, minutesSubServices int, userGender string, grossAmountCompany float64, page int, limit int) string {
	timeZone := "Asia/Jakarta"
	searchStringWithoutRating := fmt.Sprintf(`SELECT
		COALESCE((SELECT SUM(CASE WHEN TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), '%s', b.timezone_code), b.timezone_code, 'UTC'), b.order_time) > 0 AND TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), '%s', b.timezone_code), b.timezone_code, 'UTC'), b.order_time) < %d THEN 1 ELSE 0 END) FROM order_transactions b WHERE b.mitra_id = a.id AND b.order_status = "WAIT_SCHEDULE"),0) AS count_order,
		COALESCE((SELECT SUM(CASE WHEN TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), '%s', b.timezone_code), b.timezone_code, 'UTC'), c.order_time) > 0 AND TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), '%s', b.timezone_code), b.timezone_code, 'UTC'), c.order_time) < %d THEN 1 ELSE 0 END) FROM order_transactions b LEFT JOIN order_transaction_repeats c ON c.order_id = b.id WHERE c.order_id = b.id AND c.mitra_id = a.id),0) AS count_order_repeat,
		a.id, a.firebase_token, a.is_auto_bid, a.account_balance,
		(SELECT COALESCE(SUM(c.debt_per_week),0) FROM tools_credits b LEFT JOIN tools c ON b.tool_id = c.id WHERE b.mitra_id = a.id) AS total_hutang,
		(6371 * acos(cos(radians(%f)) * cos(radians(latitude)) * cos(radians(longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(latitude)))) AS distance,
		(a.account_balance - COALESCE((SELECT SUM(x.gross_amount_company) FROM order_transactions x WHERE x.order_status = 'WAIT_SCHEDULE'),0)) AS money_left
		FROM users a
		WHERE a.user_gender = '%s' AND a.is_logged_in = '1' AND a.is_active = '%s' AND a.is_busy = '%s' AND a.is_auto_bid = '%s' AND a.is_suspended = '0' AND a.user_type = 'mitra'
		HAVING total_hutang <= a.account_balance AND money_left >= %f AND distance <= %f AND a.account_balance >= %f AND count_order = 0 AND count_order_repeat = 0;`,
		timeZone, timeZone, minutesSubServices, timeZone, timeZone, minutesSubServices, latitude, longitude, latitude, userGender, isActive, isBusy, isAutoBid, grossAmountCompany, distance, grossAmountCompany)
	return searchStringWithoutRating
}

func searchQueryNow(latitude float64, longitude float64, isActive string, isBusy string, distance float64, isAutoBid string, minutesSubServices int, userGender string) string {
	timeZone := "Asia/Jakarta"
	searchString := fmt.Sprintf(`SELECT
		COALESCE((SELECT SUM(CASE WHEN TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), '%s', b.timezone_code), b.timezone_code, 'UTC'), b.order_time) > 0 AND TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), '%s', b.timezone_code), b.timezone_code, 'UTC'), b.order_time) < %d THEN 1 ELSE 0 END) FROM order_transactions b WHERE b.mitra_id = a.id AND b.order_status = "WAIT_SCHEDULE"),0) AS count_order,
		COALESCE((SELECT SUM(CASE WHEN TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), '%s', b.timezone_code), b.timezone_code, 'UTC'), c.order_time) > 0 AND TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), '%s', b.timezone_code), b.timezone_code, 'UTC'), c.order_time) < %d THEN 1 ELSE 0 END) FROM order_transactions b LEFT JOIN order_transaction_repeats c ON c.order_id = b.id WHERE c.order_id = b.id AND c.mitra_id = a.id),0) AS count_order_repeat,
		a.id, a.firebase_token, a.is_auto_bid, a.account_balance,
		(SELECT COALESCE(SUM(c.debt_per_week),0) FROM tools_credits b LEFT JOIN tools c ON b.tool_id = c.id WHERE b.mitra_id = a.id) AS total_hutang,
		(6371 * acos(cos(radians(%f)) * cos(radians(latitude)) * cos(radians(longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(latitude)))) AS distance
		FROM users a
		WHERE a.user_gender = '%s' AND a.is_logged_in = '1' AND a.is_active = '%s' AND a.is_busy = '%s' AND a.is_auto_bid = '%s' AND a.is_suspended = '0' AND a.user_type = 'mitra'
		HAVING total_hutang <= a.account_balance AND distance <= %f AND count_order = 0 AND count_order_repeat = 0
		LIMIT 1;`,
		timeZone, timeZone, minutesSubServices, timeZone, timeZone, minutesSubServices, latitude, longitude, latitude, userGender, isActive, isBusy, isAutoBid, distance)
	return searchString
}

func searchQueryNowCashWithoutTime(latitude float64, longitude float64, isActive string, isBusy string, distance float64, isAutoBid string, userGender string, grossAmountCompany float64) string {
	searchStringWithoutRating := fmt.Sprintf(`SELECT
		a.id, a.firebase_token, a.is_auto_bid, a.account_balance,
		(SELECT COALESCE(SUM(c.debt_per_week), 0) FROM tools_credits b LEFT JOIN tools c ON b.tool_id = c.id WHERE b.mitra_id = a.id) AS total_hutang,
		(6371 * acos(cos(radians(%f)) * cos(radians(latitude)) * cos(radians(longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(latitude)))) AS distance,
		(a.account_balance - (CASE WHEN(SELECT SUM(x.gross_amount_company) FROM order_transactions x WHERE x.order_status = 'WAIT_SCHEDULE') IS NULL THEN 0 ELSE (SELECT SUM(x.gross_amount_company) FROM order_transactions x WHERE x.order_status = 'WAIT_SCHEDULE') END)) as money_left
		FROM users a
		WHERE a.user_gender = '%s' AND a.is_logged_in = '1' AND a.is_active = '%s' AND a.is_busy = '%s' AND a.is_auto_bid = '%s' AND a.is_suspended = '0' AND a.user_type = 'mitra'
		HAVING total_hutang <= a.account_balance AND distance <= %f AND money_left >= %f AND a.account_balance >= %f
		LIMIT 1;`,
		latitude, longitude, latitude, userGender, isActive, isBusy, isAutoBid, distance, grossAmountCompany, grossAmountCompany)
	return searchStringWithoutRating
}

func searchQueryNowWithoutTime(latitude float64, longitude float64, isActive string, isBusy string, distance float64, isAutoBid string, userGender string) string {
	searchString := fmt.Sprintf(`SELECT
		a.id, a.firebase_token, a.is_auto_bid, a.account_balance,
		(SELECT COALESCE(SUM(c.debt_per_week), 0) FROM tools_credits b LEFT JOIN tools c ON b.tool_id = c.id WHERE b.mitra_id = a.id) AS total_hutang,
		(6371 * acos(cos(radians(%f)) * cos(radians(latitude)) * cos(radians(longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(latitude)))) AS distance
		FROM users a
		WHERE a.user_gender = '%s' AND a.is_logged_in = '1' AND a.is_active = '%s' AND a.is_busy = '%s' AND a.is_auto_bid = '%s' AND a.is_suspended = '0' AND a.user_type = 'mitra'
		HAVING total_hutang <= a.account_balance AND distance <= %f
		LIMIT 1;`,
		latitude, longitude, latitude, userGender, isActive, isBusy, isAutoBid, distance)
	return searchString
}

type GetNearestMitraProductionParams struct {
	CustomerID           string
	Latitude             float64
	Longitude            float64
	InitialRange         float64
	MaxRange             float64
	UserGender           string
	OrderType            string
	SubPaymentID         int
	IsAutoBid            string
	ServiceDuration      int
	CustomerTimezoneCode string
	CustomerTimeOrder    string
	JsonOrderTimes       []models.OrderTransactionRepeat
	GrossAmountCompany   float64
	IsWithTime           bool
	Limit                int
	Page                 int
}

type MitraResult struct {
	IsAvailableNextTime bool
	PayloadMitra        []models.User
	InitRange           float64
	TriedRange          float64
}

func GetNearestMitraProduction(params GetNearestMitraProductionParams) (*MitraResult, error) {
	var customerRating models.User
	if err := config.DB.Where("id = ? AND user_type = ?", params.CustomerID, "customer").First(&customerRating).Error; err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	var subPaymentData models.SubPayment
	if err := config.DB.Where("id = ?", params.SubPaymentID).First(&subPaymentData).Error; err != nil {
		return nil, fmt.Errorf("sub payment not found: %w", err)
	}

	var paymentData models.Payment
	if err := config.DB.Where("id = ?", subPaymentData.PaymentID).First(&paymentData).Error; err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	initRange := params.InitialRange
	isAutoBid := params.IsAutoBid

	for initRange <= params.MaxRange {
		var resultQuery []models.User
		var query string

		if params.OrderType == "now" {
			if paymentData.Type == "tunai" {
				if params.IsWithTime {
					query = searchQueryNowCash(params.Latitude, params.Longitude, "yes", "no", initRange, isAutoBid, params.ServiceDuration+15, params.UserGender, params.GrossAmountCompany, params.Page, params.Limit)
				} else {
					query = searchQueryNowCashWithoutTime(params.Latitude, params.Longitude, "yes", "no", initRange, isAutoBid, params.UserGender, params.GrossAmountCompany)
				}
			} else {
				if params.IsWithTime {
					query = searchQueryNow(params.Latitude, params.Longitude, "yes", "no", initRange, isAutoBid, params.ServiceDuration+15, params.UserGender)
				} else {
					query = searchQueryNowWithoutTime(params.Latitude, params.Longitude, "yes", "no", initRange, isAutoBid, params.UserGender)
				}
			}

			if err := config.DB.Raw(query).Scan(&resultQuery).Error; err != nil {
				return nil, fmt.Errorf("error executing raw query: %w", err)
			}

			if len(resultQuery) > 0 {
				return &MitraResult{
					IsAvailableNextTime: false,
					PayloadMitra:        resultQuery,
					InitRange:           initRange,
				}, nil
			}
		}

		if isAutoBid == "no" {
			initRange++
		}
		isAutoBid = toggleAutoBid(isAutoBid)
	}

	return &MitraResult{
		TriedRange:   initRange,
		PayloadMitra: []models.User{},
	}, nil
}

func toggleAutoBid(isAutoBid string) string {
	if isAutoBid == "yes" {
		return "no"
	}
	return "yes"
}

func convertToInt(s string, defaultVal int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return i
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}
