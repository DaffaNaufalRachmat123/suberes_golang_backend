package models

import "time"

type BaseModel struct {
	ID        int `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
