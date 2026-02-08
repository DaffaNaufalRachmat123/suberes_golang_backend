package models

type Payment struct {
	ID       uint `gorm:"primaryKey"`
	Icon     string
	IsActive string
	Title    string
	Type     string
	Desc     string

	SubPayments []SubPayment `gorm:"foreignKey:PaymentID"`
}
