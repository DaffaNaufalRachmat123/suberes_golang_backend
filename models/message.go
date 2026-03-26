package models

type Message struct {
	ID                int    `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID          string `gorm:"type:varchar(36)" json:"sender_id"`
	RecipientIDs      string `gorm:"type:varchar(255)" json:"recipient_ids"`
	SenderName        string `gorm:"type:varchar(255)" json:"sender_name"`
	IsBlastMessage    string `gorm:"type:varchar(3);check:is_blast_message IN ('yes','no')" json:"is_blast_message"`
	IsMultipleMessage string `gorm:"type:varchar(3);check:is_multiple_message IN ('yes','no')" json:"is_multiple_message"`
	ImageMessage      string `gorm:"type:varchar(255)" json:"image_message"`
	Title             string `gorm:"type:varchar(255)" json:"title"`
	CaptionText       string `gorm:"type:varchar(255)" json:"caption_text"`
	Body              string `gorm:"type:text" json:"body"`
	TypeMessage       string `gorm:"type:varchar(50);check:type_message IN ('ORDER','CAMPAIGN','EXPIRED_PAYMENT','FAILURE_PAYMENT','SUCCESS_PAYMENT','PENDING_PAYMENT','ORDER_CANCELED_BY_ADMIN' , 'NOTIFICATION' , 'MESSAGE')" json:"type_message"`
	IsRead            string `gorm:"type:varchar(1);check:is_read IN ('0','1')" json:"is_read"`

	// Associations
	Sender *User `gorm:"foreignKey:SenderID;references:ID" json:"sender,omitempty"`
}

func (Message) TableName() string {
	return "messages"
}
