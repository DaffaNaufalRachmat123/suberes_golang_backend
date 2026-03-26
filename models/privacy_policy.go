package models

type PrivacyPolicy struct {
	BaseModel
	AdminID           string `gorm:"type:varchar(36)" json:"admin_id"`
	PolicyTitle       string `gorm:"type:varchar(255)" json:"policy_title"`
	PolicyDescription string `gorm:"type:text" json:"policy_description"`
	IsValid           string `gorm:"type:varchar(1);check:is_valid IN ('0','1')" json:"is_valid"`
	UserType          string `gorm:"type:varchar(10);check:user_type IN ('customer','mitra')" json:"user_type"`

	// Associations
	Admin *User `gorm:"foreignKey:AdminID;references:ID" json:"admin,omitempty"`
}

func (PrivacyPolicy) TableName() string {
	return "privacy_policies"
}
