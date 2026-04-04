package models

import "time"

type Bantuan struct {
	ID            uint      `gorm:"primary_key"`
	Title         string    `gorm:"size:255"`
	Description   string    `gorm:"type:text"`
	HelpType      string    `gorm:"size:50"`
	WatchingCount int       `gorm:"default:0"`
	CreatedAt     time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Bantuan) TableName() string {
	return "help_tables"
}
