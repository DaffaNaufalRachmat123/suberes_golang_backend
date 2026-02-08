package models

type SubPaymentTutorials struct {
	BaseModel

	PaymentID    int    `gorm:"column:payment_id" json:"payment_id"`
	SubPaymentID int    `gorm:"column:sub_payment_id" json:"sub_payment_id"`
	Title        string `gorm:"column:title" json:"title"`
	Description  string `gorm:"type:text;column:description" json:"description"`
}

func (SubPaymentTutorials) TableName() string {
	return "sub_payment_tutorials"
}
