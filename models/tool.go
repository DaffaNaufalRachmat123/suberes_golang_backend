package models

type Tool struct {
	ID                     uint   `gorm:"primaryKey;column:id" json:"id"`
	ToolName               string `gorm:"column:tool_name" json:"tool_name"`
	ToolCount              int    `gorm:"column:tool_count" json:"tool_count"`
	ToolPrice              int    `gorm:"column:tool_price" json:"tool_price"`
	CompanyPriceAdditional int    `gorm:"column:company_price_additional" json:"company_price_additional"`
	ToolType               string `gorm:"column:tool_type" json:"tool_type"`
	DebtPerWeek            int    `gorm:"column:debt_per_week" json:"debt_per_week"`
	InstallmentPeriod      int    `gorm:"column:installment_period" json:"installment_period"`
	ToolImage              string `gorm:"type:text;column:tool_image" json:"tool_image"`
	IsOutOfStock           string `gorm:"type:enum('0','1');column:is_out_of_stock" json:"is_out_of_stock"`

	// Relations
	ToolsCredits []ToolsCredit `gorm:"foreignKey:ToolID;references:ID" json:"tools_credits"`
	UsersTools   []UsersTools  `gorm:"foreignKey:ToolID;references:ID" json:"users_tools"`
}

func (Tool) TableName() string {
	return "tools"
}
