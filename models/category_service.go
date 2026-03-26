package models

type CategoryService struct {
	ID              uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	LayananID       int    `gorm:"type:integer" json:"layanan_id"`
	CreatorID       string `gorm:"type:varchar(36)" json:"creator_id"`
	CategoryService string `gorm:"type:text" json:"category_service"`

	// Associations
	LayananService *LayananService `gorm:"foreignKey:LayananID;references:ID" json:"layanan_service,omitempty"`
	Services       []Service       `gorm:"foreignKey:parent_id;references:ID" json:"services,omitempty"`
}

func (CategoryService) TableName() string {
	return "category_services"
}
