package models

type UserTool struct {
	BaseModel
	UserID     string `gorm:"type:varchar(36)" json:"user_id"`
	ToolID     int    `gorm:"type:integer" json:"tool_id"`
	ToolName   string `gorm:"type:varchar(255)" json:"tool_name"`
	ToolCount  int    `gorm:"type:integer" json:"tool_count"`
	ToolStatus string `gorm:"type:varchar(20);check:tool_status IN ('Company','Mitra')" json:"tool_status"`

	// Associations
	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Tool *Tool `gorm:"foreignKey:ToolID;references:ID" json:"tool,omitempty"`
}

func (UserTool) TableName() string {
	return "users_tools"
}
