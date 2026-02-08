package models

import "time"

type UserOTP struct {
	UsersID     string    `gorm:"primaryKey;size:36;column:users_id" json:"users_id"`
	OTPCode     string    `gorm:"column:otp_code" json:"otp_code"`
	OTPType     string    `gorm:"column:otp_type" json:"otp_type"`
	SessionTime time.Time `gorm:"column:session_time" json:"session_time"`

	// Perbaikan: column:createdAt menjadi column:created_at
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime:true" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updatedAt;autoUpdateTime:true" json:"updated_at"`
}

func (UserOTP) TableName() string {
	return "users_otps"
}
