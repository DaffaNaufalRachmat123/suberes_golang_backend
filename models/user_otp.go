package models

import "time"

type UserOTP struct {
	BaseModel
	UsersID     string    `gorm:"type:varchar(36);primaryKey" json:"users_id"`
	OTPCode     string    `gorm:"type:text" json:"otp_code"`
	OTPType     string    `gorm:"type:varchar(50);check:otp_type IN ('login_code','email_verification_code','change_data','change_pin','change_phone_number','forgot_password')" json:"otp_type"`
	SessionTime time.Time `gorm:"type:time" json:"session_time"`

	// Associations
	User *User `gorm:"foreignKey:UsersID;references:ID" json:"user,omitempty"`
}

func (UserOTP) TableName() string {
	return "users_otps"
}
