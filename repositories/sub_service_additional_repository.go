package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type SubServiceAdditionalRepository struct {
	DB *gorm.DB
}

func (r *SubServiceAdditionalRepository) FindByID(id int) (*models.SubServiceAdditional, error) {
	var record models.SubServiceAdditional
	if err := r.DB.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *SubServiceAdditionalRepository) Create(tx *gorm.DB, record *models.SubServiceAdditional) error {
	return tx.Create(record).Error
}

func (r *SubServiceAdditionalRepository) CreateBulk(tx *gorm.DB, records []models.SubServiceAdditional) error {
	return tx.Create(&records).Error
}

func (r *SubServiceAdditionalRepository) Update(tx *gorm.DB, id int, data map[string]interface{}) error {
	return tx.Model(&models.SubServiceAdditional{}).Where("id = ?", id).Updates(data).Error
}

func (r *SubServiceAdditionalRepository) Delete(tx *gorm.DB, id int) error {
	return tx.Delete(&models.SubServiceAdditional{}, id).Error
}

func (r *SubServiceAdditionalRepository) DeleteBySubServiceID(tx *gorm.DB, subServiceID int) error {
	return tx.Where("sub_service_id = ?", subServiceID).Delete(&models.SubServiceAdditional{}).Error
}
