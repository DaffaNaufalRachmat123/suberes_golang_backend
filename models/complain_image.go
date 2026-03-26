package models

type ComplainImage struct {
	BaseModel
	ComplainID         string `gorm:"type:varchar(36)" json:"complain_id"`
	ImageName          string `gorm:"type:text" json:"image_name"`
	ImageSize          string `gorm:"type:varchar(255)" json:"image_size"`
	ImageSizeDimension string `gorm:"type:varchar(255)" json:"image_size_dimension"`

	// Associations
	Complain *Complain `gorm:"foreignKey:ComplainID;references:ID" json:"complain,omitempty"`
}

func (ComplainImage) TableName() string {
	return "complain_images"
}
