package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type SubServiceAdditionalRepository struct {
	DB *gorm.DB
}

func (r *SubServiceAdditionalRepository) CreateBulk(tx *gorm.DB, records []models.SubServiceAdditional) error {
	return tx.Create(&records).Error
}

func (r *SubServiceAdditionalRepository) DeleteBySubServiceID(tx *gorm.DB, subServiceID int) error {
	return tx.Where("sub_service_id = ?", subServiceID).Delete(&models.SubServiceAdditional{}).Error
}
