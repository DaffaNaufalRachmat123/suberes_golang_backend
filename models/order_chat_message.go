package models

import "time"

type OrderChatMessage struct {
	ID                  string    `gorm:"type:varchar(255);primaryKey" json:"id"`
	OrderChatID         string    `gorm:"type:text" json:"order_chat_id"`
	Message             string    `gorm:"type:text" json:"message"`
	TypeMessage         string    `gorm:"type:varchar(10);check:type_message IN ('text','image','video' , 'audio')" json:"type_message"`
	MessageFilePath     string    `gorm:"type:text" json:"message_file_path"`
	BlurMessageFilePath string    `gorm:"type:text" json:"blur_message_file_path"`
	MessageFileSize     string    `gorm:"type:varchar(255)" json:"message_file_size"`
	FromMessage         string    `gorm:"type:varchar(10);check:from_message IN ('customer','mitra')" json:"from_message"`
	IsMessageSent       string    `gorm:"type:varchar(1);check:is_message_sent IN ('0','1')" json:"is_message_sent"`
	IsMessageRead       string    `gorm:"type:varchar(1);check:is_message_read IN ('0','1')" json:"is_message_read"`
	IsMessageListened   string    `gorm:"type:varchar(1);check:is_message_listened IN ('0','1')" json:"is_message_listened"`
	CreatedAt           time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt           time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`

	// Associations
	OrderChat *OrderChat `gorm:"foreignKey:OrderChatID;references:ID" json:"order_chat,omitempty"`
}

func (OrderChatMessage) TableName() string {
	return "order_chat_messages"
}
