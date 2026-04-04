package dtos

type OrderSummary struct {
	OrderCount int64
	Pendapatan int64
}

type CreateOrderDTO struct {
	ServiceID             int                  `json:"service_id" form:"service_id" binding:"required"`
	SubServiceID          int                  `json:"sub_service_id" form:"sub_service_id" binding:"required"`
	GrossAmount           int64                `json:"gross_amount" form:"gross_amount" binding:"required"`
	PaymentID             int                  `json:"payment_id" form:"payment_id" binding:"required"`
	SubPaymentID          int                  `json:"sub_payment_id" form:"sub_payment_id" binding:"required"`
	Description           string               `json:"description" form:"description"`
	OrderType             string               `json:"order_type" form:"order_type" binding:"required"`
	MitraGender           string               `json:"mitra_gender" form:"mitra_gender" binding:"required"`
	OrderTime             string               `json:"order_time" form:"order_time" binding:"required"`
	OrderTimestamp        string               `json:"order_timestamp" form:"order_timestamp" binding:"required"`
	Address               string               `json:"address" form:"address" binding:"required"`
	Locality              string               `json:"locality" form:"locality"`
	City                  string               `json:"city" form:"city"`
	Region                string               `json:"region" form:"region"`
	Country               string               `json:"country" form:"country"`
	PostalCode            string               `json:"postal_code" form:"postal_code"`
	Landmark              string               `json:"landmark" form:"landmark"`
	GrossAddAdditional    int64                `json:"gross_add_additional" form:"gross_add_additional"`
	GrossAmountAdditional int64                `json:"gross_amount_additional" form:"gross_amount_additional"`
	OrderCountAdditional  int                  `json:"order_count_additional" form:"order_count_additional"`
	BankCode              string               `json:"bank_code" form:"bank_code"`
	IsAdditional          string               `json:"is_additional" form:"is_additional" binding:"required"`
	OrderAdditionalList   []OrderAdditionalDTO `json:"order_additional_list" form:"order_additional_list"`
	OrderRepeatList       []OrderRepeatDTO     `json:"order_repeat_list" form:"order_repeat_list"`
	OrderServiceCount     int                  `json:"order_service_count" form:"order_service_count"`
	TimezoneCode          string               `json:"timezone_code" form:"timezone_code" binding:"required"`
	OrderNote             string               `json:"order_note" form:"order_note"`
	CustomerLatitude      float64              `json:"customer_latitude" form:"customer_latitude" binding:"required"`
	CustomerLongitude     float64              `json:"customer_longitude" form:"customer_longitude" binding:"required"`
}

type OrderAdditionalDTO struct {
	ID             int     `json:"id" form:"id" binding:"required"`
	SubServiceID   int     `json:"sub_service_id" form:"sub_service_id" binding:"required"`
	Title          string  `json:"title" form:"title" binding:"required"`
	Amount         float64 `json:"amount" form:"amount" binding:"required"`
	AdditionalType string  `json:"additional_type" form:"additional_type" binding:"required"`
	IsChoice       string  `json:"is_choice" form:"is_choice" binding:"required"`
}

type OrderRepeatDTO struct {
	OrderTime      string `json:"order_time" form:"order_time" binding:"required"`
	OrderTimestamp string `json:"order_timestamp" form:"order_timestamp" binding:"required"`
}

type AcceptOrderDTO struct {
	OrderID    string `json:"order_id" form:"order_id" binding:"required"`
	SubID      int    `json:"sub_id" form:"sub_id" binding:"required"`
	TempID     string `json:"temp_id" form:"temp_id" binding:"required"`
	CustomerID string `json:"customer_id" form:"customer_id" binding:"required"`
	MitraID    string `json:"mitra_id" form:"mitra_id" binding:"required"`
	OrderType  string `json:"order_type" form:"order_type" binding:"required"`
}
