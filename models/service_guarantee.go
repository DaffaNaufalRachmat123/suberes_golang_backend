package models

type ServiceGuarantee struct {
	BaseModel
	ServiceID            int    `gorm:"type:integer;unique" json:"service_id"`
	UserID               string `gorm:"type:varchar(36)" json:"user_id"`
	GuaranteeName        string `gorm:"type:varchar(255)" json:"guarantee_name"`
	GuaranteeDescription string `gorm:"type:varchar(255)" json:"guarantee_description"`
	IsGuaranteeEnabled   string `gorm:"type:varchar(1);check:is_guarantee_enabled IN ('0','1')" json:"is_guarantee_enabled"`

	// Associations
	Service *Service `gorm:"foreignKey:ServiceID;references:ID" json:"service,omitempty"`
	User    *User    `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (ServiceGuarantee) TableName() string {
	return "service_guarantees"
}
