package repositories

import (
"suberes_golang/models"

"gorm.io/gorm"
)

type TermsConditionRepository struct {
	DB *gorm.DB
}

// FindAll retrieves all terms conditions with pagination
func (r *TermsConditionRepository) FindAll(page, limit int) ([]models.TermsCondition, int64, error) {
	var termsConditions []models.TermsCondition
	var total int64
	offset := (page - 1) * limit

	query := r.DB.Model(&models.TermsCondition{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&termsConditions).Error
	return termsConditions, total, err
}

// FindByID retrieves a single terms condition by ID
func (r *TermsConditionRepository) FindByID(id uint) (*models.TermsCondition, error) {
	var termsCondition models.TermsCondition
	err := r.DB.Where("id = ?", id).First(&termsCondition).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &termsCondition, err
}

// FindByTypeAndUserType retrieves active terms condition by type and user type
func (r *TermsConditionRepository) FindByTypeAndUserType(tocType, tocUserType, isActive string) (*models.TermsCondition, error) {
	var termsCondition models.TermsCondition
	query := r.DB.Where("toc_type = ? AND toc_user_type = ?", tocType, tocUserType)
	
	if isActive != "" {
		query = query.Where("is_active = ?", isActive)
	}

	err := query.First(&termsCondition).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &termsCondition, err
}

// FindActiveExcept retrieves active terms condition excluding a specific ID
func (r *TermsConditionRepository) FindActiveExcept(tocType, tocUserType string, excludeID uint) (*models.TermsCondition, error) {
	var termsCondition models.TermsCondition
	err := r.DB.Where(
"toc_type = ? AND toc_user_type = ? AND is_active = '1' AND id != ?",
tocType, tocUserType, excludeID,
).First(&termsCondition).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &termsCondition, err
}

// Create inserts a new terms condition record
func (r *TermsConditionRepository) Create(tx *gorm.DB, termsCondition *models.TermsCondition) error {
	return tx.Table("terms_conditions").Create(termsCondition).Error
}

// Update updates an existing terms condition record
func (r *TermsConditionRepository) Update(tx *gorm.DB, termsCondition *models.TermsCondition) error {
	return tx.Table("terms_conditions").Save(termsCondition).Error
}

// UpdateStatus updates only the status fields of a terms condition
func (r *TermsConditionRepository) UpdateStatus(tx *gorm.DB, id uint, isActive, canSelect string) error {
	return tx.Table("terms_conditions").
		Where("id = ?", id).
		Updates(map[string]interface{}{
"is_active": isActive,
"can_select": canSelect,
}).Error
}

// UpdateMultipleByTypeAndUserType updates multiple records by type and user type
func (r *TermsConditionRepository) UpdateMultipleByTypeAndUserType(tx *gorm.DB, tocType, tocUserType string, updates map[string]interface{}, excludeID uint) error {
	return tx.Table("terms_conditions").
		Where("toc_type = ? AND toc_user_type = ? AND id != ?", tocType, tocUserType, excludeID).
		Updates(updates).Error
}

// Delete removes a terms condition record
func (r *TermsConditionRepository) Delete(tx *gorm.DB, termsCondition *models.TermsCondition) error {
	return tx.Table("terms_conditions").Delete(termsCondition).Error
}

// DeleteByID removes a terms condition by ID
func (r *TermsConditionRepository) DeleteByID(tx *gorm.DB, id uint) error {
	return tx.Table("terms_conditions").Where("id = ?", id).Delete(&models.TermsCondition{}).Error
}
