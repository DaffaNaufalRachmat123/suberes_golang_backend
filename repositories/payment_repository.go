package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	DB *gorm.DB
}

func (r *PaymentRepository) FindById(id int) (*models.Payment, error) {
	var payment models.Payment
	err := r.DB.Where("id = ?", id).First(&payment)
	return &payment, err.Error
}
