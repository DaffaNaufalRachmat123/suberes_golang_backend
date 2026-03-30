package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type LayananServiceRepository struct {
	DB *gorm.DB
}

func (r *LayananServiceRepository) FindAllPagination(page, limit int) ([]models.LayananService, int64, error) {
	var layananServices []models.LayananService
	var total int64
	offset := (page - 1) * limit
	query := r.DB.Model(models.LayananService{}).Preload("CategoryServices").Order("created_at DESC")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Find(&layananServices).Error; err != nil {
		return nil, 0, err
	}
	return layananServices, total, nil
}

func (r *LayananServiceRepository) FindPopular(limit int) ([]models.LayananService, error) {
	var layananServices []models.LayananService
	err := r.DB.Limit(limit).Find(&layananServices).Error
	return layananServices, err
}
func (r *LayananServiceRepository) FindByID(id uint) (*models.LayananService, error) {
	var layanan models.LayananService
	if err := r.DB.First(&layanan, id).Error; err != nil {
		return nil, err
	}
	return &layanan, nil
}

func (r *LayananServiceRepository) Create(tx *gorm.DB, layanan *models.LayananService) error {
	return tx.Create(layanan).Error
}

func (r *LayananServiceRepository) Update(tx *gorm.DB, id uint, data map[string]interface{}) error {
	return tx.Model(&models.LayananService{}).Where("id = ?", id).Updates(data).Error
}

func (r *LayananServiceRepository) Delete(tx *gorm.DB, id uint) error {
	return tx.Delete(&models.LayananService{}, id).Error
}
