package models

type ToolsCredit struct {
	ID      uint   `gorm:"primaryKey;column:id" json:"id"`
	ToolID  int    `gorm:"column:tool_id" json:"tool_id"`
	MitraID string `gorm:"size:36;column:mitra_id" json:"mitra_id"`

	// Relations
	SubToolsCredits []SubToolsCredit `gorm:"foreignKey:ToolsCreditsID;references:ID" json:"sub_tools_credits"`
}

func (ToolsCredit) TableName() string {
	return "tools_credits"
}
