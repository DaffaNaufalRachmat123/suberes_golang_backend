package models

type PaymentAccount struct {
	ID                 uint `gorm:"primaryKey"`
	MitraID            string
	PaymentID          uint
	BeneficiaryBank    string
	BeneficiaryAccount string
	BeneficiaryName    string
	BeneficiaryType    string
	BeneficiaryStatus  string
}
