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

func (r *TransactionRepository) FindTransactionByExternalID(tx *gorm.DB, externalID string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := tx.Where("external_id = ? AND transaction_for = ? AND transaction_status = ?", externalID, "topup", "pending").First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) UpdateTransactionStatus(tx *gorm.DB, transactionID string, status string) error {
	return tx.Model(&models.Transaction{}).Where("id = ?", transactionID).Update("transaction_status", status).Error
}
