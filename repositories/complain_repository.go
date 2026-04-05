package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type ComplainRepository struct {
	DB *gorm.DB
}

// FindAllAdmin returns paginated complains with user (customer/mitra) preloaded, filtered by complain_code search.
func (r *ComplainRepository) FindAllAdmin(page, limit int, search string) ([]models.Complain, int64, error) {
	var complains []models.Complain
	var total int64
	offset := (page - 1) * limit

	query := r.DB.Model(&models.Complain{}).
		Joins("JOIN users ON users.id = complains.customer_id AND users.user_type IN ('customer','mitra')").
		Where("complains.complain_code LIKE ?", "%"+search+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Preload("Customer").
		Order("complains.id DESC").
		Offset(offset).
		Limit(limit).
		Find(&complains).Error
	return complains, total, err
}

// FindAllCustomer returns paginated complains (no filter by user, as per JS source).
func (r *ComplainRepository) FindAllCustomer(page, limit int) ([]models.Complain, int64, error) {
	var complains []models.Complain
	var total int64
	offset := (page - 1) * limit

	query := r.DB.Model(&models.Complain{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&complains).Error
	return complains, total, err
}

// FindAllMitra returns paginated complains (same as customer index per JS source).
func (r *ComplainRepository) FindAllMitra(page, limit int) ([]models.Complain, int64, error) {
	return r.FindAllCustomer(page, limit)
}

// FindByID returns a single complain with its images preloaded.
func (r *ComplainRepository) FindByID(id int) (*models.Complain, error) {
	var complain models.Complain
	err := r.DB.
		Preload("ComplainImages").
		Where("id = ?", id).
		First(&complain).Error
	if err != nil {
		return nil, err
	}
	return &complain, nil
}

// Create inserts a new complain record inside a transaction.
func (r *ComplainRepository) Create(tx *gorm.DB, complain *models.Complain) error {
	return tx.Create(complain).Error
}

// BulkCreateImages inserts multiple ComplainImage records inside a transaction.
func (r *ComplainRepository) BulkCreateImages(tx *gorm.DB, images []models.ComplainImage) error {
	return tx.Create(&images).Error
}

// UpdateStatus updates the status field for the given complain id.
func (r *ComplainRepository) UpdateStatus(tx *gorm.DB, id int, status string) error {
	return tx.Model(&models.Complain{}).Where("id = ?", id).Update("status", status).Error
}

// Delete removes a complain by id.
func (r *ComplainRepository) Delete(tx *gorm.DB, id int) error {
	return tx.Where("id = ?", id).Delete(&models.Complain{}).Error
}
