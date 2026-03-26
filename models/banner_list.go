package models

type BannerList struct {
	BaseModel
	CreatorID            string `gorm:"type:varchar(36)" json:"creator_id"`
	CreatorName          string `gorm:"type:varchar(255)" json:"creator_name"`
	BannerTitle          string `gorm:"type:varchar(255)" json:"banner_title"`
	BannerBody           string `gorm:"type:text" json:"banner_body"`
	BannerImage          string `gorm:"type:text" json:"banner_image"`
	BannerImageSize      string `gorm:"type:varchar(255)" json:"banner_image_size"`
	BannerImageDimension string `gorm:"type:varchar(255)" json:"banner_image_dimension"`
	BannerType           string `gorm:"type:varchar(20);check:banner_type IN ('promo','coupon','visi misi','other')" json:"banner_type"`
	IsRevision           string `gorm:"type:varchar(1);check:is_revision IN ('0','1')" json:"is_revision"`
	IsBroadcast          string `gorm:"type:varchar(1);check:is_broadcast IN ('0','1')" json:"is_broadcast"`

	// Associations
	Creator *User `gorm:"foreignKey:CreatorID;references:ID" json:"creator,omitempty"`
}

func (BannerList) TableName() string {
	return "banner_lists"
}
