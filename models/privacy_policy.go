package models

type PrivacyPolicy struct {
	BaseModel

	AdminID           string `gorm:"column:admin_id" json:"admin_id"`
	PolicyTitle       string `gorm:"column:policy_title" json:"policy_title"`
	PolicyDescription string `gorm:"column:policy_description" json:"policy_description"`
	IsValid           string `gorm:"type:enum('0','1');column:is_valid" json:"is_valid"`
	UserType          string `gorm:"type:enum('customer','mitra');column:user_type" json:"user_type"`
}

func (PrivacyPolicy) TableName() string {
	return "privacy_policies"
}
