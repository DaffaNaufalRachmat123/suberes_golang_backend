package models

type NewsImageList struct {
	ID                 int    `gorm:"primaryKey;autoIncrement" json:"id"`
	NewsID             int    `gorm:"type:integer" json:"news_id"`
	NewsImage          string `gorm:"type:text" json:"news_image"`
	NewsImageSize      string `gorm:"type:varchar(255)" json:"news_image_size"`
	NewsImageDimension string `gorm:"type:varchar(255)" json:"news_image_dimension"`
	NewsImageSource    string `gorm:"type:varchar(255)" json:"news_image_source"`

	// Associations
	News *NewsList `gorm:"foreignKey:NewsID;references:ID" json:"news,omitempty"`
}

func (NewsImageList) TableName() string {
	return "news_image_lists"
}
