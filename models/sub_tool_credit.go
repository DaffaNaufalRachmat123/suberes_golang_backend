package models

import "time"

type SubToolsCredit struct {
	ID                  uint      `gorm:"primaryKey;column:id" json:"id"`
	ToolID              int       `gorm:"column:tool_id" json:"tool_id"`
	MitraID             string    `gorm:"size:36;column:mitra_id" json:"mitra_id"`
	ToolsCreditsID      int       `gorm:"column:tools_credits_id" json:"tools_credits_id"`
	AmountPaid          int       `gorm:"column:amount_paid" json:"amount_paid"`
	PaidStatus          string    `gorm:"type:enum('0','1');column:paid_status" json:"paid_status"`
	InstallmentDeadline time.Time `gorm:"column:installment_deadline" json:"installment_deadline"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:true" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:true" json:"updated_at"`
}

func (SubToolsCredit) TableName() string {
	return "sub_tools_credits"
}
