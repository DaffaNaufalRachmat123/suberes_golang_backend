package models

type PaymentMitra struct {
	BaseModel
	Title string `gorm:"type:varchar(255)" json:"title"`
	Desc  string `gorm:"type:text" json:"desc"`
	Icon  string `gorm:"type:text" json:"icon"`

	// Associations
	PaymentAccounts []PaymentAccount `gorm:"foreignKey:PaymentID" json:"payment_accounts,omitempty"`
}

func (PaymentMitra) TableName() string {
	return "payment_mitras"
}
