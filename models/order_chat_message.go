package models

type OrderChatMessage struct {
	ID                  string `gorm:"primaryKey;column:id" json:"id"` // STRING PK di JS
	OrderChatID         string `gorm:"type:text;column:order_chat_id" json:"order_chat_id"`
	Message             string `gorm:"type:text;column:message" json:"message"`
	TypeMessage         string `gorm:"type:enum('text','image','video','audio');column:type_message" json:"type_message"`
	MessageFilePath     string `gorm:"type:text;column:message_file_path" json:"message_file_path"`
	BlurMessageFilePath string `gorm:"type:text;column:blur_message_file_path" json:"blur_message_file_path"`
	MessageFileSize     string `gorm:"column:message_file_size" json:"message_file_size"`
	FromMessage         string `gorm:"type:enum('customer','mitra');column:from_message" json:"from_message"`
	IsMessageSent       string `gorm:"type:enum('0','1');column:is_message_sent" json:"is_message_sent"`
	IsMessageRead       string `gorm:"type:enum('0','1');column:is_message_read" json:"is_message_read"`
	IsMessageListened   string `gorm:"type:enum('0','1');column:is_message_listened" json:"is_message_listened"`

	// Menggunakan string sesuai definisi JS: DataTypes.STRING(100)
	CreatedAt string `gorm:"size:100;column:createdAt" json:"created_at"`
	UpdatedAt string `gorm:"size:100;column:updatedAt" json:"updated_at"`
}

func (OrderChatMessage) TableName() string {
	return "order_chat_messages"
}
