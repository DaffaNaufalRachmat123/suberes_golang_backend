package models

type BannerList struct {
	BaseModel

	CreatorID            string `gorm:"column:creator_id" json:"creator_id"`
	CreatorName          string `gorm:"column:creator_name" json:"creator_name"`
	BannerTitle          string `gorm:"column:banner_title" json:"banner_title"`
	BannerBody           string `gorm:"type:text;column:banner_body" json:"banner_body"`
	BannerImage          string `gorm:"type:text;column:banner_image" json:"banner_image"`
	BannerImageSize      string `gorm:"column:banner_image_size" json:"banner_image_size"`
	BannerImageDimension string `gorm:"column:banner_image_dimension" json:"banner_image_dimension"`
	BannerType           string `gorm:"type:enum('promo','coupon','visi misi','other');column:banner_type" json:"banner_type"`
	IsRevision           string `gorm:"type:enum('0','1');column:is_revision" json:"is_revision"`
	IsBroadcast          string `gorm:"type:enum('0','1');column:is_broadcast" json:"is_broadcast"`
}

func (BannerList) TableName() string {
	return "banner_lists"
}
