package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type CategoryServiceRepository struct {
	DB *gorm.DB
}

func (r *CategoryServiceRepository) FindAllByLayananIDPagination(layananID, page, limit int) ([]models.CategoryService, int64, error) {
	var categoryServices []models.CategoryService
	var total int64

	offset := (page - 1) * limit

	err := r.DB.Model(&models.CategoryService{}).Where("layanan_id = ?", layananID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.DB.Where("layanan_id = ?", layananID).Order("id desc").Limit(limit).Offset(offset).Find(&categoryServices).Error
	if err != nil {
		return nil, 0, err
	}

	return categoryServices, total, nil
}
