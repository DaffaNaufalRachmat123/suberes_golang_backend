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

func (r *PaymentRepository) FindAllActive() ([]models.Payment, error) {
	var payments []models.Payment
	err := r.DB.
		Preload("SubPayments", "enabled = ?", "1").
		Where("is_active = ?", "1").
		Find(&payments).Error
	return payments, err
}

func (r *PaymentRepository) Create(tx *gorm.DB, payment *models.Payment) error {
	return tx.Create(payment).Error
}

func (r *PaymentRepository) Update(tx *gorm.DB, id int, updates map[string]interface{}) error {
	return tx.Model(&models.Payment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PaymentRepository) Delete(tx *gorm.DB, id int) error {
	return tx.Delete(&models.Payment{}, id).Error
}
