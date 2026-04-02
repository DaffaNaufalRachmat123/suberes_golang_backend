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
	if err := r.DB.First(&subService, id).Error; err != nil {
		return nil, err
	}
	return &subService, nil
}

func (r *SubServiceRepository) Create(tx *gorm.DB, subService *models.SubService) error {
	return tx.Create(subService).Error
}

func (r *SubServiceRepository) Update(tx *gorm.DB, id int, data map[string]interface{}) error {
	return tx.Model(&models.SubService{}).Where("id = ?", id).Updates(data).Error
}

func (r *SubServiceRepository) Delete(tx *gorm.DB, id int) error {
	return tx.Delete(&models.SubService{}, id).Error
}
