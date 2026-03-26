package models

import "time"

type OrderTransactionRepeat struct {
	ID                 int       `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID            string    `gorm:"type:varchar(36)" json:"order_id"`
	CustomerID         string    `gorm:"type:varchar(36)" json:"customer_id"`
	MitraID            string    `gorm:"type:varchar(36)" json:"mitra_id"`
	ServiceID          int       `gorm:"type:integer" json:"service_id"`
	SubServiceID       int       `gorm:"type:integer" json:"sub_service_id"`
	CustomerName       string    `gorm:"type:varchar(255)" json:"customer_name"`
	Address            string    `gorm:"type:varchar(255)" json:"address"`
	OrderTime          time.Time `gorm:"type:timestamp" json:"order_time"`
	OrderTimestamp     string    `gorm:"type:varchar(255)" json:"order_timestamp"`
	CanceledReason     string    `gorm:"type:text" json:"canceled_reason"`
	CanceledUser       string    `gorm:"type:varchar(10);check:canceled_user IN ('customer','mitra')" json:"canceled_user"`
	OrderNote          string    `gorm:"type:text" json:"order_note"`
	PaymentID          int       `gorm:"type:integer" json:"payment_id"`
	SubPaymentID       int       `gorm:"type:integer" json:"sub_payment_id"`
	OrderStatus        string    `gorm:"type:varchar(20);check:order_status IN ('WAITING_PAYMENT','WAIT_SCHEDULE','OTW','ON_PROGRESS','FINISH','CANCELED','FINDING_MITRA')" json:"order_status"`
	IDTransaction      string    `gorm:"type:varchar(255)" json:"id_transaction"`
	GrossAmount        int64     `gorm:"type:bigint" json:"gross_amount"`
	GrossAmountMitra   int64     `gorm:"type:bigint" json:"gross_amount_mitra"`
	GrossAmountCompany int64     `gorm:"type:bigint" json:"gross_amount_company"`
	CustomerLatitude   float64   `gorm:"type:float" json:"customer_latitude"`
	CustomerLongitude  float64   `gorm:"type:float" json:"customer_longitude"`
	MitraLatitude      float64   `gorm:"type:float" json:"mitra_latitude"`
	MitraLongitude     float64   `gorm:"type:float" json:"mitra_longitude"`
	CreatedAt          time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt          time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`

	// Associations
	OrderTransaction *OrderTransaction `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
	Customer         *User             `gorm:"foreignKey:CustomerID;references:ID" json:"customer,omitempty"`
	Mitra            *User             `gorm:"foreignKey:MitraID;references:ID" json:"mitra,omitempty"`
	Service          *Service          `gorm:"foreignKey:ServiceID;references:ID" json:"service,omitempty"`
	SubService       *SubService       `gorm:"foreignKey:SubServiceID;references:ID" json:"sub_service,omitempty"`
	Transactions     []Transaction     `gorm:"foreignKey:SubOrderID" json:"transactions,omitempty"`
	Notifications    []Notification    `gorm:"foreignKey:SubOrderID" json:"notifications,omitempty"`
}

func (OrderTransactionRepeat) TableName() string {
	return "order_transaction_repeats"
}
