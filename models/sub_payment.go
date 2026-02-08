package models

type SubPayment struct {
	ID           uint   `gorm:"primaryKey;column:id" json:"id"`
	PaymentID    uint   `gorm:"column:payment_id" json:"payment_id"`
	Icon         string `gorm:"column:icon" json:"icon"`
	Title        string `gorm:"column:title" json:"title"`
	TitlePayment string `gorm:"column:title_payment" json:"title_payment"`
	Enabled      string `gorm:"column:enabled" json:"enabled"`
	Desc         string `gorm:"column:desc" json:"desc"`
}
