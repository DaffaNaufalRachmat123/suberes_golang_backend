package models

type BankList struct {
	BaseModel
	BankImage        string  `gorm:"type:text" json:"bank_image"`
	Name             string  `gorm:"type:varchar(255)" json:"name"`
	Code             string  `gorm:"type:varchar(255)" json:"code"`
	DisbursementCode string  `gorm:"type:varchar(255)" json:"disbursement_code"`
	MethodType       string  `gorm:"type:varchar(10);check:method_type IN ('bank','ewallet')" json:"method_type"`
	CanTopup         string  `gorm:"type:varchar(1);check:can_topup IN ('0','1')" json:"can_topup"`
	CanDisbursement  string  `gorm:"type:varchar(1);check:can_disbursement IN ('0','1')" json:"can_disbursement"`
	MinTopup         int     `gorm:"type:integer" json:"min_topup"`
	MinDisbursement  int     `gorm:"type:integer" json:"min_disbursement"`
	TopupFee         float64 `gorm:"type:float" json:"topup_fee"`
	DisbursementFee  float64 `gorm:"type:float" json:"disbursement_fee"`
	IsPercentage     string  `gorm:"type:varchar(1);check:is_percentage IN ('0','1')" json:"is_percentage"`
	disburse_type    string  `gorm:"type:varchar(255)" json:"type"`

	// Associations
	Transactions []Transaction `gorm:"foreignKey:BankID" json:"transactions,omitempty"`
}
