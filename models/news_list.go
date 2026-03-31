package models

import "time"

type NewsList struct {
	ID                 int64     `gorm:"primaryKey" json:"id"`
	CreatorID          string    `gorm:"column:creator_id" json:"creator_id"`
	CreatorName        string    `gorm:"column:creator_name" json:"creator_name"`
	NewsTitle          string    `gorm:"column:news_title" json:"news_title"`
	NewsBody           string    `gorm:"column:news_body" json:"news_body"`
	NewsType           string    `gorm:"column:news_type" json:"news_type"`
	NewsImage          string    `gorm:"column:news_image" json:"news_image"`
	NewsImageSize      string    `gorm:"column:news_image_size" json:"news_image_size"`
	NewsImageDimension string    `gorm:"column:news_image_dimension" json:"news_image_dimension"`
	IsRevision         string    `gorm:"column:is_revision;type:char(1);check:is_revision IN ('0','1')" json:"is_revision"`
	IsBroadcast        string    `gorm:"column:is_broadcast;type:char(1);check:is_broadcast IN ('0','1')" json:"is_broadcast"`
	ReadCount          int       `gorm:"column:read_count" json:"read_count"`
	LikeCount          int       `gorm:"column:like_count" json:"like_count"`
	CommentCount       int       `gorm:"column:comment_count" json:"comment_count"`
	ShareCount         int       `gorm:"column:share_count" json:"share_count"`
	Narasumber         string    `gorm:"column:narasumber" json:"narasumber"`
	TimezoneCode       string    `gorm:"column:timezone_code;size:50" json:"timezone_code"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Users              User      `gorm:"foreignKey:CreatorID" json:"user"`
}

func (NewsList) TableName() string {
	return "news_lists"
}
