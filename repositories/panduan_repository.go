package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type PanduanRepository struct {
	DB *gorm.DB
}

// FindAllCustomer retrieves all panduan for customer with pagination, ordered by watching_count DESC
func (r *PanduanRepository) FindAllCustomer(page, limit int) ([]models.GuideTable, int64, error) {
	var panduans []models.GuideTable
	var total int64
	offset := (page - 1) * limit

	query := r.DB.Model(&models.GuideTable{}).Where("guide_type = ?", "customer")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("watching_count DESC").Offset(offset).Limit(limit).Find(&panduans).Error
	return panduans, total, err
}

// FindAllMitra retrieves all panduan for mitra with pagination, ordered by watching_count DESC
func (r *PanduanRepository) FindAllMitra(page, limit int) ([]models.GuideTable, int64, error) {
	var panduans []models.GuideTable
	var total int64
	offset := (page - 1) * limit

	query := r.DB.Model(&models.GuideTable{}).Where("guide_type = ?", "mitra")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("watching_count DESC").Offset(offset).Limit(limit).Find(&panduans).Error
	return panduans, total, err
}

// FindAllAdmin retrieves all panduan for admin with pagination, ordered by createdAt DESC
func (r *PanduanRepository) FindAllAdmin(page, limit int) ([]models.GuideTable, int64, error) {
	var panduans []models.GuideTable
	var total int64
	offset := (page - 1) * limit

	query := r.DB.Model(&models.GuideTable{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&panduans).Error
	return panduans, total, err
}

// FindByID retrieves a single panduan by ID
func (r *PanduanRepository) FindByID(id uint) (*models.GuideTable, error) {
	var panduan models.GuideTable
	err := r.DB.Where("id = ?", id).First(&panduan).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &panduan, err
}

// Create creates a new panduan
func (r *PanduanRepository) Create(tx *gorm.DB, panduan *models.GuideTable) error {
	return tx.Create(panduan).Error
}

// Update updates a panduan
func (r *PanduanRepository) Update(tx *gorm.DB, panduan *models.GuideTable) error {
	return tx.Save(panduan).Error
}

// UpdateWatchingCount increments watching_count
func (r *PanduanRepository) UpdateWatchingCount(tx *gorm.DB, id uint) error {
	return tx.Model(&models.GuideTable{}).Where("id = ?", id).Update("watching_count", gorm.Expr("watching_count + 1")).Error
}

// Delete deletes a panduan
func (r *PanduanRepository) Delete(tx *gorm.DB, id uint) error {
	return tx.Where("id = ?", id).Delete(&models.GuideTable{}).Error
}
