package models

type Tool struct {
	ID                     uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ToolName               string `gorm:"type:varchar(255)" json:"tool_name"`
	ToolCount              int    `gorm:"type:integer" json:"tool_count"`
	ToolPrice              int    `gorm:"type:integer" json:"tool_price"`
	CompanyPriceAdditional int    `gorm:"type:integer" json:"company_price_additional"`
	ToolType               string `gorm:"type:varchar(255)" json:"tool_type"`
	DebtPerWeek            int    `gorm:"type:integer" json:"debt_per_week"`
	InstallmentPeriod      int    `gorm:"type:integer" json:"installment_period"`
	ToolImage              string `gorm:"type:text" json:"tool_image"`
	IsOutOfStock           string `gorm:"type:varchar(1);check:is_out_of_stock IN ('0','1')" json:"is_out_of_stock"`

	// Associations
	ToolCredits  []ToolCredit  `gorm:"foreignKey:ToolID" json:"tool_credits,omitempty"`
	UsersTools   []UserTool    `gorm:"foreignKey:ToolID" json:"users_tools,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:ToolID" json:"transactions,omitempty"`
}

func (Tool) TableName() string {
	return "tools"
}
