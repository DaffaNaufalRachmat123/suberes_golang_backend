package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type BannerRepository struct {
	DB *gorm.DB
}

func (r *BannerRepository) FindAll(page, limit int) ([]models.BannerList, int64, error) {
	var banners []models.BannerList
	var total int64
	offset := (page - 1) * limit
	query := r.DB.Model(&models.BannerList{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&banners).Error
	return banners, total, err
}

func (r *BannerRepository) FindById(id uint) (*models.BannerList, error) {
	var banner models.BannerList
	err := r.DB.Where("id = ?", id).First(&banner).Error
	return &banner, err
}
func (r *BannerRepository) FindPopular(limit int) ([]models.BannerList, error) {
	var banners []models.BannerList
	err := r.DB.Limit(limit).Find(&banners).Error
	return banners, err
}
func (r *BannerRepository) Create(tx *gorm.DB, banner *models.BannerList) error {
	return tx.Table("banner_lists").Create(banner).Error
}
func (r *BannerRepository) Update(tx *gorm.DB, banner *models.BannerList) error {
	return tx.Table("banner_lists").Save(banner).Error
}
func (r *BannerRepository) Delete(tx *gorm.DB, banner *models.BannerList) error {
	return tx.Table("banner_lists").Delete(banner).Error
}
