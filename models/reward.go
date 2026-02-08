package models

import "time"

type Reward struct {
	BaseModel

	UserID            int       `gorm:"column:user_id" json:"user_id"`
	ServiceID         int       `gorm:"column:service_id" json:"service_id"`
	SubServiceID      int       `gorm:"column:sub_service_id" json:"sub_service_id"`
	RewardTitle       string    `gorm:"column:reward_title" json:"reward_title"`
	RewardDescription string    `gorm:"type:text;column:reward_description" json:"reward_description"`
	RewardType        string    `gorm:"type:enum('no level','silver','gold','platinum');column:reward_type" json:"reward_type"`
	RewardStartDate   time.Time `gorm:"column:reward_start_date" json:"reward_start_date"`
	RewardEndDate     time.Time `gorm:"column:reward_end_date" json:"reward_end_date"`
}

func (Reward) TableName() string {
	return "rewards"
}
