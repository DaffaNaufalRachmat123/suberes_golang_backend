package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type SubServiceRepository struct {
	DB *gorm.DB
}

func (r *SubServiceRepository) FindByID(id int) (*models.SubService, error) {
	var subService models.SubService
	err := r.DB.Where("id = ?", id).First(&subService).Error
	return &subService, err
}
