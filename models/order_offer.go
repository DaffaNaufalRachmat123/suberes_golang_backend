package models

type OrderOffer struct {
	BaseModel
	TempID     string `gorm:"type:varchar(255)" json:"temp_id"`
	OrderID    string `gorm:"type:varchar(36)" json:"order_id"`
	CustomerID string `gorm:"type:varchar(36)" json:"customer_id"`
	MitraID    string `gorm:"type:varchar(36)" json:"mitra_id"`

	// Associations
	OrderTransaction *OrderTransaction `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
	Customer         *User             `gorm:"foreignKey:CustomerID;references:ID" json:"customer,omitempty"`
	Mitra            *User             `gorm:"foreignKey:MitraID;references:ID" json:"mitra,omitempty"`
}

func (OrderOffer) TableName() string {
	return "order_offers"
}
