package models

import "time"

type SubServiceAdded struct {
	ID              uint   `gorm:"primaryKey;column:id" json:"id"`
	OrderID         string `gorm:"size:36;column:order_id" json:"order_id"` // UUID
	SubServiceAddID int    `gorm:"column:sub_service_add_id" json:"sub_service_add_id"`
	Title           string `gorm:"column:title;default:''" json:"title"`
	BaseAmount      int    `gorm:"column:base_amount;default:0" json:"base_amount"`
	Amount          int    `gorm:"column:amount;default:0" json:"amount"`
	AdditionalType  string `gorm:"type:enum('choice','cashback','discount','free');column:additional_type" json:"additional_type"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:true" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:true" json:"updated_at"`
}

func (SubServiceAdded) TableName() string {
	return "sub_service_added"
}
