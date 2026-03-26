package models

type HelpTable struct {
	BaseModel
	Title         string `gorm:"type:varchar(255)" json:"title"`
	Description   string `gorm:"type:text" json:"description"`
	WatchingCount int    `gorm:"type:integer;default:0" json:"watching_count"`
	HelpType      string `gorm:"type:varchar(10);check:help_type IN ('customer','mitra')" json:"help_type"`
}

func (HelpTable) TableName() string {
	return "help_tables"
}
