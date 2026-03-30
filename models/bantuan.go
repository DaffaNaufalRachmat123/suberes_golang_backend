package models

import "time"

type Bantuan struct {
	ID            uint      `gorm:"primary_key"`
	Title         string    `gorm:"size:255"`
	Description   string    `gorm:"type:text"`
	HelpType      string    `gorm:"size:50"`
	WatchingCount int       `gorm:"default:0"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (Bantuan) TableName() string {
	return "help_table"
}
