package models

import "time"

type Bantuan struct {
	ID            uint      `gorm:"primary_key" json:"id"`
	Title         string    `gorm:"size:255" json:"title"`
	Description   string    `gorm:"type:text" json:"description"`
	HelpType      string    `gorm:"size:50" json:"help_type"`
	WatchingCount int       `gorm:"default:0" json:"watching_count"`
	CreatedAt     time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Bantuan) TableName() string {
	return "help_tables"
}
