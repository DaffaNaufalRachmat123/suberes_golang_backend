package models

type BankList struct {
	BaseModel
	BankImage        string
	Name             string
	Code             string
	DisbursementCode string
	MethodType       string
	CanTopup         string
	CanDisbursement  string
	MinTopup         int
	MinDisbursement  int
	TopupFee         float64
	DisbursementFee  float64
	IsPercentage     string

	Transactions []Transaction `gorm:"foreignKey:BankID"`
}
