package models

import "time"

type OrderTransactionRepeat struct {
	ID                 int       `gorm:"primaryKey;column:id" json:"id"`
	OrderID            string    `gorm:"type:uuid;column:order_id" json:"order_id"`
	CustomerID         string    `gorm:"size:36;column:customer_id" json:"customer_id"`
	MitraID            string    `gorm:"size:36;column:mitra_id" json:"mitra_id"`
	ServiceID          int       `gorm:"column:service_id" json:"service_id"`
	SubServiceID       int       `gorm:"column:sub_service_id" json:"sub_service_id"`
	CustomerName       string    `gorm:"column:customer_name" json:"customer_name"`
	Address            string    `gorm:"column:address" json:"address"`
	OrderTime          time.Time `gorm:"column:order_time" json:"order_time"`
	OrderTimestamp     string    `gorm:"column:order_timestamp" json:"order_timestamp"`
	CanceledReason     string    `gorm:"type:text;column:canceled_reason" json:"canceled_reason"`
	CanceledUser       string    `gorm:"type:enum('customer','mitra');column:canceled_user" json:"canceled_user"`
	OrderNote          string    `gorm:"type:text;column:order_note" json:"order_note"`
	PaymentID          int       `gorm:"column:payment_id" json:"payment_id"`
	SubPaymentID       int       `gorm:"column:sub_payment_id" json:"sub_payment_id"`
	OrderStatus        string    `gorm:"column:order_status" json:"order_status"`
	IDTransaction      string    `gorm:"column:id_transaction" json:"id_transaction"`
	GrossAmount        int       `gorm:"column:gross_amount" json:"gross_amount"`
	GrossAmountMitra   int       `gorm:"column:gross_amount_mitra" json:"gross_amount_mitra"`
	GrossAmountCompany int       `gorm:"column:gross_amount_company" json:"gross_amount_company"`
	CustomerLatitude   float64   `gorm:"column:customer_latitude" json:"customer_latitude"`
	CustomerLongitude  float64   `gorm:"column:customer_longitude" json:"customer_longitude"`
	MitraLatitude      float64   `gorm:"column:mitra_latitude" json:"mitra_latitude"`
	MitraLongitude     float64   `gorm:"column:mitra_longitude" json:"mitra_longitude"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:true" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:true" json:"updated_at"`
}

func (OrderTransactionRepeat) TableName() string {
	return "order_transaction_repeats"
}
