package models

import "time"

type NewsList struct {
	ID                 int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatorID          string    `gorm:"type:varchar(36)" json:"creator_id"`
	CreatorName        string    `gorm:"type:varchar(255)" json:"creator_name"`
	NewsTitle          string    `gorm:"type:varchar(255)" json:"news_title"`
	NewsBody           string    `gorm:"type:text" json:"news_body"`
	NewsType           string    `gorm:"type:varchar(20);check:news_type IN ('Suberes Update','News')" json:"news_type"`
	NewsImage          string    `gorm:"type:text" json:"news_image"`
	NewsImageSize      string    `gorm:"type:varchar(255)" json:"news_image_size"`
	NewsImageDimension string    `gorm:"type:varchar(255)" json:"news_image_dimension"`
	IsRevision         string    `gorm:"type:varchar(1);check:is_revision IN ('0','1')" json:"is_revision"`
	ReadCount          int       `gorm:"type:integer;default:0" json:"read_count"`
	LikeCount          int       `gorm:"type:integer;default:0" json:"like_count"`
	CommentCount       int       `gorm:"type:integer;default:0" json:"comment_count"`
	ShareCount         int       `gorm:"type:integer;default:0" json:"share_count"`
	Narasumber         string    `gorm:"type:varchar(255)" json:"narasumber"`
	IsBroadcast        string    `gorm:"type:varchar(1);check:is_broadcast IN ('0','1')" json:"is_broadcast"`
	TimezoneCode       string    `gorm:"type:varchar(50)" json:"timezone_code"`
	CreatedAt          time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt          time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`

	// Associations
	Creator       *User           `gorm:"foreignKey:CreatorID;references:ID" json:"creator,omitempty"`
	NewsImages    []NewsImageList `gorm:"foreignKey:NewsID" json:"news_images,omitempty"`
}

func (NewsList) TableName() string {
	return "news_lists"
}
