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

func (r *CategoryServiceRepository) FindByID(id uint) (*models.CategoryService, error) {
	var cs models.CategoryService
	if err := r.DB.First(&cs, id).Error; err != nil {
		return nil, err
	}
	return &cs, nil
}

func (r *CategoryServiceRepository) Create(tx *gorm.DB, cs *models.CategoryService) error {
	return tx.Create(cs).Error
}

func (r *CategoryServiceRepository) Update(tx *gorm.DB, id uint, data map[string]interface{}) error {
	return tx.Model(&models.CategoryService{}).Where("id = ?", id).Updates(data).Error
}

func (r *CategoryServiceRepository) Delete(tx *gorm.DB, id uint) error {
	return tx.Delete(&models.CategoryService{}, id).Error
}
