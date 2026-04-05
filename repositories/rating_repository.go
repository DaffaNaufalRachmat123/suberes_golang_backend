package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type RatingRepository struct {
	DB *gorm.DB
}

// FindMitraRatings returns ratings of type "customer to mitra" for a given mitra, with preloaded order and service.
func (r *RatingRepository) FindMitraRatings(mitraID string, limit, offset int) ([]models.UserRating, error) {
	var ratings []models.UserRating
	err := r.DB.
		Where("mitra_id = ? AND rating_type = ?", mitraID, "customer to mitra").
		Preload("OrderTransaction.Service").
		Limit(limit).
		Offset(offset).
		Find(&ratings).Error
	if err != nil {
		return nil, err
	}
	return ratings, nil
}

// FindCustomerRatings returns ratings of type "mitra to customer" for a given mitra, with pagination.
func (r *RatingRepository) FindCustomerRatings(mitraID string, limit, offset int) ([]models.UserRating, error) {
	var ratings []models.UserRating
	err := r.DB.
		Where("mitra_id = ? AND rating_type = ?", mitraID, "mitra to customer").
		Limit(limit).
		Offset(offset).
		Find(&ratings).Error
	if err != nil {
		return nil, err
	}
	return ratings, nil
}
