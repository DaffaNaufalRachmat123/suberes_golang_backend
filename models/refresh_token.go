package models

import "time"

type RefreshToken struct {
	BaseModel
	UsersID   string    `gorm:"type:varchar(36);index;not null" json:"users_id"`
	TokenHash string    `gorm:"type:text;not null" json:"token_hash"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null" json:"expires_at"`
	Revoked   string    `gorm:"type:varchar(1);default:'0';check:revoked IN ('0','1')" json:"revoked"`

	// Associations
	User *User `gorm:"foreignKey:UsersID;references:ID" json:"user,omitempty"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
