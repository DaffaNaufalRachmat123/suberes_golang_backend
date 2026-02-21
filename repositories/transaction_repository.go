package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type TransactionRepository struct {
	DB *gorm.DB
}

func (r *TransactionRepository) FindById(id int) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.DB.Where("id = ?", id).First(&transaction)
	return &transaction, err.Error
}
func (r *TransactionRepository) CreateTransaction(tx *gorm.DB, transaction *models.Transaction) error {
	return tx.Create(transaction).Error
}
