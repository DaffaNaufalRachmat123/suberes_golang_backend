package models

type CategoryService struct {
	ID              uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	LayananID       int    `gorm:"column:layanan_id" json:"layanan_id"`
	CreatorID       int    `gorm:"column:creator_id" json:"creator_id"`
	CategoryService string `gorm:"type:text;column:category_service" json:"category_service"` // Nama field sama dengan nama tabel di JS, disesuaikan

	// Relations
	Services []Service `gorm:"foreignKey:ParentID;references:ID" json:"services"` // Asumsi ada model Service
}

func (CategoryService) TableName() string {
	return "category_services"
}
