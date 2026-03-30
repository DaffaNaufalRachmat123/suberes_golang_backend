package models

import "time"

type LayananService struct {
	ID                    int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatorID             string    `gorm:"type:varchar(36)" json:"creator_id"`
	LayananTitle          string    `gorm:"type:varchar(255)" json:"layanan_title"`
	LayananDescription    string    `gorm:"type:text" json:"layanan_description"`
	LayananImage          string    `gorm:"type:text" json:"layanan_image"`
	LayananImageSize      string    `gorm:"type:varchar(255)" json:"layanan_image_size"`
	LayananImageDimension string    `gorm:"type:varchar(255)" json:"layanan_image_dimension"`
	IsActive              string    `gorm:"type:varchar(1);check:is_active IN ('0','1')" json:"is_active"`
	CreatedAt             time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt             time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`

	Creator          *User             `gorm:"foreignKey:CreatorID;references:ID" json:"creator,omitempty"`
	CategoryServices []CategoryService `gorm:"foreignKey:LayananID" json:"category_services,omitempty"`
	UserRatings      []UserRating      `gorm:"foreignKey:LayananID" json:"user_ratings,omitempty"`
}

func (LayananService) TableName() string {
	return "layanan_services"
}
