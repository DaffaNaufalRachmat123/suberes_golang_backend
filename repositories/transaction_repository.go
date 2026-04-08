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

func (r *TransactionRepository) FindAllWithPagination(page, limit int, search, transactionType string) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.DB.Model(&models.Transaction{})

	if search != "" {
		query = query.Where("disbursement_id LIKE ? OR topup_id LIKE ? OR order_id_transaction LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if transactionType != "" {
		query = query.Where("transaction_type = ?", transactionType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Preload("MitraTransactionData").Preload("CustomerTransactionData").Preload("OrderTransaction").Limit(limit).Offset(offset).Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *TransactionRepository) GetTransactionTypesByMitraIDAndDate(mitraID string, startDate, endDate string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := r.DB.Model(&models.Transaction{}).
		Select("transaction_for, CASE transaction_for WHEN 'order' THEN 'Order' WHEN 'cicilan' THEN 'Cicilan' WHEN 'Other' THEN 'Lainnya' ELSE '-' END as transaction_for_show").
		Where("mitra_id = ? AND createdAt BETWEEN ? AND ?", mitraID, startDate, endDate).
		Group("transaction_for").
		Find(&result).Error

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *TransactionRepository) FindAllByMitraIDWithPagination(mitraID, transactionFor, startDate, endDate string, page, limit int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.DB.Model(&models.Transaction{}).
		Where("mitra_id = ? AND transaction_for = ? AND createdAt BETWEEN ? AND ?", mitraID, transactionFor, startDate, endDate)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Preload("OrderTransaction.OrderTransactionRepeats").
		Preload("Tool").
		Preload("ToolsCredit").
		Preload("SubToolsCredit").
		Preload("MitraTransactionData").
		Limit(limit).
		Offset(offset).
		Order("createdAt DESC").
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *TransactionRepository) FindDisbursementsByMitraID(mitraID string, page, limit int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.DB.Model(&models.Transaction{}).
		Where("mitra_id = ? AND transaction_for = ?", mitraID, "disbursement")

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Limit(limit).Offset(offset).Order("createdAt DESC").Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *TransactionRepository) FindTopupTransactionByExternalIDForCallback(externalID string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.DB.Where("external_id = ?", externalID).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) FindPendingDisbursementByExternalID(externalID string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.DB.Where("external_id = ? AND transaction_status = ? AND transaction_for = ?", externalID, "pending", "disbursement").First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) FindMitraDisburseTransactionsPaginated(mitraID string, page, limit int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.DB.Model(&models.Transaction{}).
		Where("mitra_id = ? AND (transaction_for = ? OR transaction_for = ?)", mitraID, "disbursement", "topup")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("MitraTransactionData").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}
	return transactions, total, nil
}

func (r *TransactionRepository) FindCustomerDisburseTransactionsPaginated(customerID string, page, limit int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.DB.Model(&models.Transaction{}).
		Where("customer_id = ? AND (transaction_for = ? OR transaction_for = ?)", customerID, "disbursement", "topup")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("CustomerTransactionData").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}
	return transactions, total, nil
}

func (r *TransactionRepository) FindMitraTransactionDetail(id, mitraID, idempotencyKey string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.DB.Where("id = ? AND mitra_id = ? AND idempotency_key = ?", id, mitraID, idempotencyKey).
		Preload("MitraTransactionData").
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) FindCustomerTransactionDetail(id, customerID string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.DB.Where("id = ? AND customer_id = ?", id, customerID).
		Preload("CustomerTransactionData").
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) UpdateTopupSuccess(tx *gorm.DB, externalID string, lastAmount int64) error {
	return tx.Model(&models.Transaction{}).
		Where("external_id = ? AND transaction_status = ?", externalID, "pending").
		Updates(map[string]interface{}{
			"last_amount":        lastAmount,
			"transaction_status": "success",
		}).Error
}

func (r *TransactionRepository) UpdateDisbursementFailure(tx *gorm.DB, transactionID, failureCode string, refundLastAmount int64) error {
	return tx.Model(&models.Transaction{}).
		Where("id = ?", transactionID).
		Updates(map[string]interface{}{
			"transaction_status": "failed",
			"failure_code":       failureCode,
			"last_amount":        gorm.Expr("last_amount + ?", refundLastAmount),
		}).Error
}

func (r *TransactionRepository) UpdateDisbursementStatus(tx *gorm.DB, transactionID, status string) error {
	return tx.Model(&models.Transaction{}).
		Where("id = ?", transactionID).
		Update("transaction_status", status).Error
}
