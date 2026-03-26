package models

import "time"

type Reward struct {
	BaseModel
	UserID            string    `gorm:"type:varchar(36)" json:"user_id"`
	ServiceID         int       `gorm:"type:integer" json:"service_id"`
	SubServiceID      int       `gorm:"type:integer" json:"sub_service_id"`
	RewardTitle       string    `gorm:"type:varchar(255)" json:"reward_title"`
	RewardDescription string    `gorm:"type:text" json:"reward_description"`
	RewardType        string    `gorm:"type:varchar(20);check:reward_type IN ('no level','silver','gold','platinum')" json:"reward_type"`
	RewardStartDate   time.Time `gorm:"type:timestamp" json:"reward_start_date"`
	RewardEndDate     time.Time `gorm:"type:timestamp" json:"reward_end_date"`

	// Associations
	User       *User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Service    *Service    `gorm:"foreignKey:ServiceID;references:ID" json:"service,omitempty"`
	SubService *SubService `gorm:"foreignKey:SubServiceID;references:ID" json:"sub_service,omitempty"`
}

func (Reward) TableName() string {
	return "rewards"
}
