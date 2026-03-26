package models

type SubPaymentTutorial struct {
	BaseModel
	PaymentID    int    `gorm:"type:integer" json:"payment_id"`
	SubPaymentID int    `gorm:"type:integer" json:"sub_payment_id"`
	Title        string `gorm:"type:varchar(255)" json:"title"`
	Description  string `gorm:"type:text" json:"description"`

	// Associations
	Payment    *Payment    `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
	SubPayment *SubPayment `gorm:"foreignKey:SubPaymentID" json:"sub_payment,omitempty"`
}

func (SubPaymentTutorial) TableName() string {
	return "sub_payment_tutorials"
}
