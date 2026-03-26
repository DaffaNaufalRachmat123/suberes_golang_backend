package payloads

type GetNearestMitraProductionPayload struct {
	CustomerID           int     `json:"customer_id"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	InitialRange         int     `json:"initial_range"`
	MaxRange             int     `json:"max_range"`
	UserGender           string  `json:"user_gender"`
	OrderType            string  `json:"order_type"`
	SubPaymentID         int     `json:"sub_payment_id"`
	IsAutoBid            string  `json:"is_auto_bid"`
	ServiceDuration      int     `json:"service_duration"`
	CustomerTimezoneCode string  `json:"customer_timezone_code"`
	CustomerTimeOrder    string  `json:"customer_time_order"`
	JsonOrderTimes       []struct {
		OrderTime string `json:"order_time"`
	} `json:"json_order_times"`
	GrossAmountCompany float64 `json:"gross_amount_company"`
	IsWithTime           bool    `json:"is_with_time"`
	Limit                int     `json:"limit"`
	Page                 int     `json:"page"`
}
