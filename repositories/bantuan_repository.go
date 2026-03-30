package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type BantuanRepository struct {
	DB *gorm.DB
}

func (r *BantuanRepository) FindAll(page, limit int, helpType string) ([]models.Bantuan, int64, error) {
	var bantuans []models.Bantuan
	var total int64
	offset := (page - 1) * limit
	query := r.DB.Model(&models.Bantuan{})

	if helpType != "" {
		query = query.Where("help_type = ?", helpType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("watching_count DESC").Offset(offset).Limit(limit).Find(&bantuans).Error
	return bantuans, total, err
}

func (r *BantuanRepository) FindAllAdmin(page, limit int) ([]models.Bantuan, int64, error) {
	var bantuans []models.Bantuan
	var total int64
	offset := (page - 1) * limit
	query := r.DB.Model(&models.Bantuan{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&bantuans).Error
	return bantuans, total, err
}

func (r *BantuanRepository) FindById(id uint) (*models.Bantuan, error) {
	var bantuan models.Bantuan
	err := r.DB.Where("id = ?", id).First(&bantuan).Error
	return &bantuan, err
}

func (r *BantuanRepository) Create(tx *gorm.DB, bantuan *models.Bantuan) error {
	return tx.Table("help_table").Create(bantuan).Error
}

func (r *BantuanRepository) Update(tx *gorm.DB, bantuan *models.Bantuan) error {
	return tx.Table("help_table").Save(bantuan).Error
}

func (r *BantuanRepository) Delete(tx *gorm.DB, bantuan *models.Bantuan) error {
	return tx.Table("help_table").Delete(bantuan).Error
}
