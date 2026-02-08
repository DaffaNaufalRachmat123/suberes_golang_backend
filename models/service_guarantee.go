package models

type ServiceGuarantee struct {
	BaseModel

	ServiceID            int    `gorm:"unique;column:service_id" json:"service_id"`
	UserID               int    `gorm:"column:user_id" json:"user_id"`
	GuaranteeName        string `gorm:"column:guarantee_name" json:"guarantee_name"`
	GuaranteeDescription string `gorm:"column:guarantee_description" json:"guarantee_description"`
	IsGuaranteeEnabled   string `gorm:"type:enum('0','1');column:is_guarantee_enabled" json:"is_guarantee_enabled"`
}

func (ServiceGuarantee) TableName() string {
	return "service_guarantees"
}
