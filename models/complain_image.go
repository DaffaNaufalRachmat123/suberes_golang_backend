package models

type ComplainImage struct {
	BaseModel

	ComplainID         string `gorm:"size:36;column:complain_id" json:"complain_id"`
	ImageName          string `gorm:"type:text;column:image_name" json:"image_name"`
	ImageSize          string `gorm:"column:image_size" json:"image_size"`
	ImageSizeDimension string `gorm:"column:image_size_dimension" json:"image_size_dimension"`
}

func (ComplainImage) TableName() string {
	return "complain_images"
}
