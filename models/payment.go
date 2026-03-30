package models

import "time"

type Payment struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Icon      string `gorm:"type:varchar(255)" json:"icon"`
	IsActive  string `gorm:"type:varchar(1);check:is_active IN ('0','1')" json:"is_active"`
	Title     string `gorm:"type:varchar(255)" json:"title"`
	Type      string `gorm:"type:varchar(20);check:type IN ('tunai','virtual account','transfer' , 'ewallet','balance')" json:"type"`
	Desc      string `gorm:"type:varchar(255)" json:"desc"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// Associations
	SubPayments         []SubPayment         `gorm:"foreignKey:PaymentID" json:"sub_payments,omitempty"`
	OrderTransactions   []OrderTransaction   `gorm:"foreignKey:PaymentID" json:"order_transactions,omitempty"`
	SubPaymentTutorials []SubPaymentTutorial `gorm:"foreignKey:PaymentID" json:"sub_payment_tutorials,omitempty"`
}
