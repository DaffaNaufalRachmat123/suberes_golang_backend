package models

type BeneficiaryTransaction struct {
	ID                uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID            int    `gorm:"column:user_id" json:"user_id"`
	BeneficiaryID     int    `gorm:"column:beneficiary_id" json:"beneficiary_id"`
	ExternalID        string `gorm:"column:external_id" json:"external_id"`
	TransactionAmount int    `gorm:"column:transaction_amount" json:"transaction_amount"`
	TransactionStatus string `gorm:"column:transaction_status" json:"transaction_status"`
}

func (BeneficiaryTransaction) TableName() string {
	return "beneficiary_transactions"
}
