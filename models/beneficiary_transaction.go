package models

type BeneficiaryTransaction struct {
	ID                uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            string `gorm:"type:varchar(36)" json:"user_id"`
	BeneficiaryID     int    `gorm:"type:integer" json:"beneficiary_id"`
	ExternalID        string `gorm:"type:varchar(255)" json:"external_id"`
	TransactionAmount int    `gorm:"type:integer" json:"transaction_amount"`
	TransactionStatus string `gorm:"type:varchar(255)" json:"transaction_status"`

	// Associations
	User        *User           `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Beneficiary *PaymentAccount `gorm:"foreignKey:BeneficiaryID;references:ID" json:"beneficiary,omitempty"`
}

func (BeneficiaryTransaction) TableName() string {
	return "beneficiary_transactions"
}
