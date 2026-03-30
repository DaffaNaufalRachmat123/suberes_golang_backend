package models

import "time"

type NewsList struct {
	ID                 uint   `gorm:"primaryKey"`
	CreatorID          int    `gorm:"column:creator_id"`
	CreatorName        string `gorm:"column:creator_name"`
	NewsTitle          string `gorm:"column:news_title"`
	NewsBody           string `gorm:"column:news_body"`
	NewsType           string `gorm:"column:news_type"`
	NewsImage          string `gorm:"column:news_image"`
	NewsImageSize      string `gorm:"column:news_image_size"`
	NewsImageDimension string `gorm:"column:news_image_dimension"`
	IsRevision         string `gorm:"column:is_revision;type:char(1);check:is_revision IN ('0','1')"`
	IsBroadcast        string `gorm:"column:is_broadcast;type:char(1);check:is_broadcast IN ('0','1')"`
	ReadCount          int    `gorm:"column:read_count"`
	LikeCount          int    `gorm:"column:like_count"`
	CommentCount       int    `gorm:"column:comment_count"`
	ShareCount         int    `gorm:"column:share_count"`
	Narasumber         string `gorm:"column:narasumber"`
	TimezoneCode       string `gorm:"column:timezone_code;size:50"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Users              User `gorm:"foreignKey:CreatorID"`
}

func (NewsList) TableName() string {
	return "news_lists"
}
