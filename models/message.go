package models

type Message struct {
	ID                int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SenderID          int    `gorm:"column:sender_id" json:"sender_id"`
	RecipientIDs      string `gorm:"column:recipient_ids" json:"recipient_ids"`
	SenderName        string `gorm:"column:sender_name" json:"sender_name"`
	IsBlastMessage    string `gorm:"type:enum('yes','no');column:is_blast_message" json:"is_blast_message"`
	IsMultipleMessage string `gorm:"type:enum('yes','no');column:is_multiple_message" json:"is_multiple_message"`
	ImageMessage      string `gorm:"column:image_message" json:"image_message"`
	Title             string `gorm:"column:title" json:"title"`
	CaptionText       string `gorm:"column:caption_text" json:"caption_text"`
	Body              string `gorm:"type:text;column:body" json:"body"`
	TypeMessage       string `gorm:"type:enum('ORDER','CAMPAIGN','EXPIRED_PAYMENT','FAILURE_PAYMENT','SUCCESS_PAYMENT','PENDING_PAYMENT','ORDER_CANCELED_BY_ADMIN','NOTIFICATION','MESSAGE');column:type_message" json:"type_message"`
	IsRead            string `gorm:"type:enum('0','1');column:is_read" json:"is_read"`
}

func (Message) TableName() string {
	return "messages"
}
