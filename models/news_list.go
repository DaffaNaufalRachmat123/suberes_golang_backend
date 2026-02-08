package models

import "time"

type NewsList struct {
	ID                 int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatorID          int    `gorm:"column:creator_id" json:"creator_id"`
	CreatorName        string `gorm:"column:creator_name" json:"creator_name"`
	NewsTitle          string `gorm:"column:news_title" json:"news_title"`
	NewsBody           string `gorm:"type:text;column:news_body" json:"news_body"`
	NewsType           string `gorm:"type:enum('Suberes Update','News');column:news_type" json:"news_type"`
	NewsImage          string `gorm:"type:text;column:news_image" json:"news_image"`
	NewsImageSize      string `gorm:"column:news_image_size" json:"news_image_size"`
	NewsImageDimension string `gorm:"column:news_image_dimension" json:"news_image_dimension"`
	IsRevision         string `gorm:"type:enum('0','1');column:is_revision" json:"is_revision"`
	ReadCount          int    `gorm:"column:read_count" json:"read_count"`
	LikeCount          int    `gorm:"column:like_count" json:"like_count"`
	CommentCount       int    `gorm:"column:comment_count" json:"comment_count"`
	ShareCount         int    `gorm:"column:share_count" json:"share_count"`
	Narasumber         string `gorm:"column:narasumber" json:"narasumber"`
	IsBroadcast        string `gorm:"type:enum('0','1');column:is_broadcast" json:"is_broadcast"`
	TimezoneCode       string `gorm:"size:50;column:timezone_code" json:"timezone_code"`

	CreatedAt time.Time `gorm:"column:createdAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updatedAt" json:"updated_at"`

	// Relations
	NewsImages []NewsImageList `gorm:"foreignKey:NewsID;references:ID" json:"news_images"`
}

func (NewsList) TableName() string {
	return "news_lists"
}
