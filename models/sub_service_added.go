package models

type SubServiceAdded struct {
	BaseModel
	OrderID         string `gorm:"type:varchar(36)" json:"order_id"`
	SubServiceAddID int    `gorm:"type:integer" json:"sub_service_add_id"`
	Title           string `gorm:"type:varchar(255);default:''" json:"title"`
	BaseAmount      int    `gorm:"type:integer;default:0" json:"base_amount"`
	Amount          int    `gorm:"type:integer;default:0" json:"amount"`
	AdditionalType  string `gorm:"type:varchar(20);check:additional_type IN ('choice','cashback','discount','free')" json:"additional_type"`

	// Associations
	OrderTransaction      *OrderTransaction      `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
	SubServiceAdditional *SubServiceAdditional `gorm:"foreignKey:SubServiceAddID;references:ID" json:"sub_service_additional,omitempty"`
}

func (SubServiceAdded) TableName() string {
	return "sub_service_added"
}
