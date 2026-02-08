package models

type TermsConditions struct {
	BaseModel

	CreatorID   string `gorm:"column:creator_id" json:"creator_id"`
	Title       string `gorm:"column:title" json:"title"`
	Body        string `gorm:"type:text;column:body" json:"body"`
	IsActive    string `gorm:"type:enum('0','1');column:is_active" json:"is_active"`
	CanSelect   string `gorm:"type:enum('0','1');column:can_select" json:"can_select"`
	TocType     string `gorm:"type:enum('terms_of_condition','terms_of_service','privacy_policy');column:toc_type" json:"toc_type"`
	TocUserType string `gorm:"type:enum('customer','mitra');column:toc_user_type" json:"toc_user_type"`
}

func (TermsConditions) TableName() string {
	return "terms_conditions"
}
