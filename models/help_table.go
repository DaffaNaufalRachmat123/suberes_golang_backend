package models

type HelpTable struct {
	BaseModel

	Title         string `gorm:"column:title" json:"title"`
	Description   string `gorm:"type:text;column:description" json:"description"`
	WatchingCount int    `gorm:"column:watching_count" json:"watching_count"`
	HelpType      string `gorm:"type:enum('customer','mitra');column:help_type" json:"help_type"`
}

func (HelpTable) TableName() string {
	return "help_tables"
}
