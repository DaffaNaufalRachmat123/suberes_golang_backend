package models

import "time"

type SubToolCredit struct {
	BaseModel
	ToolID              int       `gorm:"type:integer" json:"tool_id"`
	MitraID             string    `gorm:"type:varchar(36)" json:"mitra_id"`
	ToolsCreditsID      int       `gorm:"type:integer" json:"tools_credits_id"`
	AmountPaid          int       `gorm:"type:integer" json:"amount_paid"`
	PaidStatus          string    `gorm:"type:varchar(1);check:paid_status IN ('0','1')" json:"paid_status"`
	InstallmentDeadline time.Time `gorm:"type:timestamp" json:"installment_deadline"`

	// Associations
	Tool         *Tool          `gorm:"foreignKey:ToolID;references:ID" json:"tool,omitempty"`
	Mitra        *User          `gorm:"foreignKey:MitraID;references:ID" json:"mitra,omitempty"`
	ToolsCredit  *ToolCredit    `gorm:"foreignKey:ToolsCreditsID;references:ID" json:"tools_credit,omitempty"`
	Transactions []Transaction  `gorm:"foreignKey:SubToolsID" json:"transactions,omitempty"`
}

func (SubToolCredit) TableName() string {
	return "sub_tools_credits"
}
