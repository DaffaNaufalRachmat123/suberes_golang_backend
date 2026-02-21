package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type SubPaymentRepository struct {
	DB *gorm.DB
}

func (r *SubPaymentRepository) FindById(id int) (*models.SubPayment, error) {
	var subPayment models.SubPayment
	err := r.DB.Where("id = ?", id).First(&subPayment)
	return &subPayment, err.Error
}
