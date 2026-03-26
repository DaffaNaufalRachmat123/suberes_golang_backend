package models

type TermsCondition struct {
	BaseModel
	CreatorID   string `gorm:"type:varchar(36)" json:"creator_id"`
	Title       string `gorm:"type:varchar(255)" json:"title"`
	Body        string `gorm:"type:text" json:"body"`
	IsActive    string `gorm:"type:varchar(1);check:is_active IN ('0','1')" json:"is_active"`
	CanSelect   string `gorm:"type:varchar(1);check:can_select IN ('0','1')" json:"can_select"`
	TocType     string `gorm:"type:varchar(50);check:toc_type IN ('terms_of_condition','terms_of_service','privacy_policy')" json:"toc_type"`
	TocUserType string `gorm:"type:varchar(10);check:toc_user_type IN ('customer','mitra')" json:"toc_user_type"`

	// Associations
	Creator *User `gorm:"foreignKey:CreatorID;references:ID" json:"creator,omitempty"`
}

func (TermsCondition) TableName() string {
	return "terms_conditions"
}
