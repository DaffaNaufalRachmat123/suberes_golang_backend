package models

type UsersTools struct {
	BaseModel

	UserID     int    `gorm:"column:user_id" json:"user_id"`
	ToolID     int    `gorm:"column:tool_id" json:"tool_id"`
	ToolName   string `gorm:"column:tool_name" json:"tool_name"`
	ToolCount  int    `gorm:"column:tool_count" json:"tool_count"`
	ToolStatus string `gorm:"type:enum('Company','Mitra');column:tool_status" json:"tool_status"`
}

func (UsersTools) TableName() string {
	return "users_tools"
}
