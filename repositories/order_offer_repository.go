package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type OrderOfferRepository struct {
	DB *gorm.DB
}

func (r *OrderOfferRepository) FindOneByWhere(tx *gorm.DB, where map[string]interface{}) (*models.OrderOffer, error) {
	var order models.OrderOffer

	db := r.DB
	if tx != nil {
		db = tx
	}

	err := db.
		Model(&models.OrderOffer{}).
		Where(where).
		First(&order).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderOfferRepository) DeleteByWhere(tx *gorm.DB, where map[string]interface{}) error {

	db := r.DB
	if tx != nil {
		db = tx
	}

	err := db.
		Where(where).
		Delete(&models.OrderOffer{}).Error

	if err != nil {
		return err
	}

	return nil
}
