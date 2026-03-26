package models

type OrderSelectedMitra struct {
	BaseModel
	OrderID     string `gorm:"type:varchar(36)" json:"order_id"`
	MitraID     string `gorm:"type:varchar(36)" json:"mitra_id"`
	OfferStatus string `gorm:"type:varchar(20);check:offer_status IN ('SELECTED','CANCELED','RECEIVED','ACCEPTED','REJECTED')" json:"offer_status"`

	// Associations
	OrderTransaction *OrderTransaction `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
	Mitra            *User             `gorm:"foreignKey:MitraID;references:ID" json:"mitra,omitempty"`
}

func (OrderSelectedMitra) TableName() string {
	return "order_selected_mitras"
}
