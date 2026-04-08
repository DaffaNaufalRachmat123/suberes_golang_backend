package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type BankListRepository struct {
	DB *gorm.DB
}

func (r *BankListRepository) FindTopupBanks(page, limit int) ([]models.BankList, int64, error) {
	var banks []models.BankList
	var total int64

	query := r.DB.Model(&models.BankList{}).Where("can_topup = ?", "1")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("id ASC").Limit(limit).Offset(offset).Find(&banks).Error
	if err != nil {
		return nil, 0, err
	}
	return banks, total, nil
}

func (r *BankListRepository) FindDisbursementBanks(page, limit int) ([]models.BankList, int64, error) {
	var banks []models.BankList
	var total int64

	query := r.DB.Model(&models.BankList{}).Where("can_disbursement = ?", "1")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("id ASC").Limit(limit).Offset(offset).Find(&banks).Error
	if err != nil {
		return nil, 0, err
	}
	return banks, total, nil
}

func (r *BankListRepository) BulkCreate(tx *gorm.DB, banks []models.BankList) error {
	return tx.Create(&banks).Error
}

func (r *BankListRepository) FindByID(id int) (*models.BankList, error) {
	var bank models.BankList
	err := r.DB.Where("id = ?", id).First(&bank).Error
	if err != nil {
		return nil, err
	}
	return &bank, nil
}

func (r *BankListRepository) Update(tx *gorm.DB, id int, updates map[string]interface{}) error {
	return tx.Model(&models.BankList{}).Where("id = ?", id).Updates(updates).Error
}
