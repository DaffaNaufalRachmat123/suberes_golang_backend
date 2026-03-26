package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type OrderChatRepository struct {
	DB *gorm.DB
}

func (r *OrderChatRepository) Create(tx *gorm.DB, orderChat models.OrderChat) error {
	return tx.Create(&orderChat).Error
}
