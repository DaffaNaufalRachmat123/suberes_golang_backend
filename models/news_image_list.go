package models

type NewsImageList struct {
	ID                 int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	NewsID             int    `gorm:"column:news_id" json:"news_id"`
	NewsImage          string `gorm:"type:text;column:news_image" json:"news_image"`
	NewsImageSize      string `gorm:"column:news_image_size" json:"news_image_size"`
	NewsImageDimension string `gorm:"column:news_image_dimension" json:"news_image_dimension"`
	NewsImageSource    string `gorm:"column:news_image_source" json:"news_image_source"`
}

func (NewsImageList) TableName() string {
	return "news_image_lists"
}
