package models

type SubServiceAdditional struct {
	BaseModel
	SubServiceID   int     `gorm:"type:integer" json:"sub_service_id"`
	Title          string  `gorm:"type:varchar(255)" json:"title"`
	BaseAmount     float64 `gorm:"type:float" json:"base_amount"`
	Amount         float64 `gorm:"type:float" json:"amount"`
	AdditionalType string  `gorm:"type:varchar(255)" json:"additional_type"`

	// Associations
	SubService       *SubService       `gorm:"foreignKey:SubServiceID;references:ID" json:"sub_service,omitempty"`
	SubServiceAddeds []SubServiceAdded `gorm:"foreignKey:SubServiceAddID" json:"sub_service_addeds,omitempty"`
}

func (SubServiceAdditional) TableName() string {
	return "sub_service_additionals"
}
