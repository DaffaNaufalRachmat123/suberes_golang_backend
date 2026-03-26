package models

type OrderChat struct {
	ID             string `gorm:"type:varchar(36);primaryKey" json:"id"`
	CustomerID     string `gorm:"type:varchar(36)" json:"customer_id"`
	MitraID        string `gorm:"type:varchar(36)" json:"mitra_id"`
	OrderID        string `gorm:"type:varchar(36)" json:"order_id"`
	ServiceID      int    `gorm:"type:integer" json:"service_id"`
	SubServiceID   int    `gorm:"type:integer" json:"sub_service_id"`
	OrderChatCount int    `gorm:"type:integer" json:"order_chat_count"`

	// Associations
	Customer          *User              `gorm:"foreignKey:CustomerID;references:ID" json:"customer,omitempty"`
	Mitra             *User              `gorm:"foreignKey:MitraID;references:ID" json:"mitra,omitempty"`
	OrderTransaction  *OrderTransaction  `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
	Service           *Service           `gorm:"foreignKey:ServiceID;references:ID" json:"service,omitempty"`
	SubService        *SubService        `gorm:"foreignKey:SubServiceID;references:ID" json:"sub_service,omitempty"`
	OrderChatMessages []OrderChatMessage `gorm:"foreignKey:OrderChatID" json:"order_chat_messages,omitempty"`
}

func (OrderChat) TableName() string {
	return "order_chats"
}
