package models

import "time"

type LayananService struct {
	BaseModel // Asumsi ada ID auto increment

	CreatorID             string `gorm:"column:creator_id" json:"creator_id"`
	LayananTitle          string `gorm:"column:layanan_title" json:"layanan_title"`
	LayananDescription    string `gorm:"type:text;column:layanan_description" json:"layanan_description"`
	LayananImage          string `gorm:"type:text;column:layanan_image" json:"layanan_image"`
	LayananImageSize      string `gorm:"column:layanan_image_size" json:"layanan_image_size"`
	LayananImageDimension string `gorm:"column:layanan_image_dimension" json:"layanan_image_dimension"`
	IsActive              string `gorm:"type:enum('0','1');column:is_active" json:"is_active"`

	CreatedAt time.Time `gorm:"column:createdAt" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updatedAt" json:"updated_at"`
}

func (LayananService) TableName() string {
	return "layanan_services"
}
