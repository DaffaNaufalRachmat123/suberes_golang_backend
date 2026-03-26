package models

type PaymentAccount struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	MitraID            string `gorm:"type:varchar(36)" json:"mitra_id"`
	PaymentID          int    `gorm:"type:integer" json:"payment_id"`
	BeneficiaryBank    string `gorm:"type:varchar(255)" json:"beneficiary_bank"`
	BeneficiaryAccount string `gorm:"type:varchar(255)" json:"beneficiary_account"`
	BeneficiaryName    string `gorm:"type:varchar(255)" json:"beneficiary_name"`
	BeneficiaryType    string `gorm:"type:varchar(20);check:beneficiary_type IN ('simulator','real')" json:"beneficiary_type"`
	BeneficiaryStatus  string `gorm:"type:varchar(20);check:beneficiary_status IN ('not active','active')" json:"beneficiary_status"`

	// Associations
	Mitra         *User         `gorm:"foreignKey:MitraID;references:ID" json:"mitra,omitempty"`
	PaymentMitra  *PaymentMitra `gorm:"foreignKey:PaymentID;references:ID" json:"payment_mitra,omitempty"`
}

func (PaymentAccount) TableName() string {
	return "payment_accounts"
}
