package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type SubServiceAddedRepository struct {
	DB *gorm.DB
}

func (r *SubServiceAddedRepository) CreateBulk(tx *gorm.DB, data []map[string]interface{}) error {
	return tx.Model(&models.SubServiceAdded{}).Create(&data).Error
}
