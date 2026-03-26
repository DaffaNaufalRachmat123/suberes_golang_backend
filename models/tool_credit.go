package models

type ToolCredit struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ToolID  int    `gorm:"type:integer" json:"tool_id"`
	MitraID string `gorm:"type:varchar(36)" json:"mitra_id"`

	// Associations
	Tool            *Tool            `gorm:"foreignKey:ToolID;references:ID" json:"tool,omitempty"`
	Mitra           *User            `gorm:"foreignKey:MitraID;references:ID" json:"mitra,omitempty"`
	SubToolCredits  []SubToolCredit  `gorm:"foreignKey:ToolsCreditsID" json:"sub_tool_credits,omitempty"`
	Transactions    []Transaction    `gorm:"foreignKey:ToolsCreditsID" json:"transactions,omitempty"`
}

func (ToolCredit) TableName() string {
	return "tools_credits"
}
