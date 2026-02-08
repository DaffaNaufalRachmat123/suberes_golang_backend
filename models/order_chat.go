package models

type OrderChats struct {
	ID string `gorm:"primaryKey;type:uuid;column:id" json:"id"`

	CustomerID     string `gorm:"size:36;column:customer_id" json:"customer_id"`
	MitraID        string `gorm:"size:36;column:mitra_id" json:"mitra_id"`
	OrderID        string `gorm:"type:uuid;column:order_id" json:"order_id"`
	ServiceID      int    `gorm:"column:service_id" json:"service_id"`
	SubServiceID   int    `gorm:"column:sub_service_id" json:"sub_service_id"`
	OrderChatCount int    `gorm:"column:order_chat_count" json:"order_chat_count"`

	// Relations
	OrderChatMessages []OrderChatMessage `gorm:"foreignKey:OrderChatID;references:ID" json:"order_chat_messages"`
}

func (OrderChats) TableName() string {
	return "order_chats"
}
