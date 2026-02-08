package models

type PaymentMitra struct {
	BaseModel

	Title string `gorm:"column:title" json:"title"`
	Desc  string `gorm:"type:text;column:desc" json:"desc"`
	Icon  string `gorm:"type:text;column:icon" json:"icon"`

	// Relations
	PaymentAccounts []PaymentAccount `gorm:"foreignKey:PaymentID;references:ID" json:"payment_accounts"`
}

func (PaymentMitra) TableName() string {
	return "payment_mitras"
}
