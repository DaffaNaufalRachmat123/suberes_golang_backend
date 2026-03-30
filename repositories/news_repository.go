package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type NewsRepository struct {
	DB *gorm.DB
}

func (r *NewsRepository) FindAll(page, limit int) ([]models.NewsList, int64, error) {
	var news []models.NewsList
	var total int64
	offset := (page - 1) * limit
	query := r.DB.Model(&models.NewsList{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&news).Error
	return news, total, err
}

func (r *NewsRepository) FindById(id uint) (*models.NewsList, error) {
	var news models.NewsList
	err := r.DB.Preload("Users").Where("id = ?", id).First(&news).Error
	return &news, err
}
func (r *NewsRepository) FindPopular(limit int) ([]models.NewsList, error) {
	var news []models.NewsList
	err := r.DB.Limit(limit).Find(&news).Error
	return news, err
}
func (r *NewsRepository) Create(tx *gorm.DB, news *models.NewsList) error {
	return tx.Table("news_lists").Create(news).Error
}
func (r *NewsRepository) Update(tx *gorm.DB, news *models.NewsList) error {
	return tx.Table("news_lists").Save(news).Error
}
func (r *NewsRepository) Delete(tx *gorm.DB, news *models.NewsList) error {
	return tx.Table("news_lists").Delete(news).Error
}
