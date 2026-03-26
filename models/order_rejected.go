package models

type OrderRejected struct {
	BaseModel
	OrderID      string `gorm:"type:varchar(36)" json:"order_id"`
	CustomerID   string `gorm:"type:varchar(36)" json:"customer_id"`
	MitraID      string `gorm:"type:varchar(36)" json:"mitra_id"`
	ServiceID    int    `gorm:"type:integer" json:"service_id"`
	SubServiceID int    `gorm:"type:integer" json:"sub_service_id"`

	// Associations
	OrderTransaction *OrderTransaction `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
}

func (OrderRejected) TableName() string {
	return "order_rejecteds"
}
