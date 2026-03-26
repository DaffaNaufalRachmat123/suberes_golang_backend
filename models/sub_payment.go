package models

import "time"

type SubPayment struct {
	ID           int    `gorm:"primaryKey;autoIncrement" json:"id"`
	PaymentID    int    `gorm:"type:integer" json:"payment_id"`
	Icon         string `gorm:"type:varchar(255)" json:"icon"`
	Title        string `gorm:"type:varchar(255)" json:"title"`
	TitlePayment string `gorm:"type:varchar(255)" json:"title_payment"`
	Enabled      string `gorm:"type:varchar(1);check:enabled IN ('0','1')" json:"enabled"`
	Desc         string `gorm:"type:varchar(255)" json:"desc"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// Associations
	Payment            *Payment            `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
	OrderTransactions  []OrderTransaction  `gorm:"foreignKey:SubPaymentID" json:"order_transactions,omitempty"`
	SubPaymentTutorial *SubPaymentTutorial `gorm:"foreignKey:SubPaymentID" json:"sub_payment_tutorial,omitempty"`
}

func (SubPayment) TableName() string {
	return "sub_payments"
}
