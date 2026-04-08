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

// CountMitraRatings returns the total count of "customer to mitra" ratings for a mitra.
func (r *RatingRepository) CountMitraRatings(mitraID string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.UserRating{}).
		Where("mitra_id = ? AND rating_type = ?", mitraID, "customer to mitra").
		Count(&count).Error
	return count, err
}

// CountMitraRatingsByRange counts ratings in a given score range for a mitra.
func (r *RatingRepository) CountMitraRatingsByRange(mitraID string, min, max float64) (int64, error) {
	var count int64
	err := r.DB.Model(&models.UserRating{}).
		Where("mitra_id = ? AND rating >= ? AND rating <= ?", mitraID, min, max).
		Count(&count).Error
	return count, err
}

// FindMitraRatingsPaginated returns paginated "customer to mitra" ratings with associations.
func (r *RatingRepository) FindMitraRatingsPaginated(mitraID string, limit, offset int) ([]models.UserRating, int64, error) {
	var ratings []models.UserRating
	var total int64

	query := r.DB.Model(&models.UserRating{}).
		Where("mitra_id = ? AND rating_type = ?", mitraID, "customer to mitra")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Customer").
		Preload("OrderTransaction").
		Preload("LayananService").
		Preload("Service").
		Preload("SubService").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&ratings).Error

	return ratings, total, err
}

// FindMitraRatingsHome returns the latest 5 "customer to mitra" ratings with associations (for home page).
func (r *RatingRepository) FindMitraRatingsHome(mitraID string) ([]models.UserRating, error) {
	var ratings []models.UserRating
	err := r.DB.
		Where("mitra_id = ? AND rating_type = ?", mitraID, "customer to mitra").
		Preload("Customer").
		Preload("LayananService").
		Preload("Service").
		Preload("SubService").
		Order("created_at DESC").
		Limit(5).
		Find(&ratings).Error
	return ratings, err
}

// GetMitraRatingAggregate returns the sum and count of ratings for a mitra.
func (r *RatingRepository) GetMitraRatingAggregate(mitraID string) (totalRating float64, totalCount int64, err error) {
	type result struct {
		TotalRating float64
		TotalCount  int64
	}
	var res result
	err = r.DB.Model(&models.UserRating{}).
		Select("COALESCE(SUM(rating), 0) as total_rating, COUNT(id) as total_count").
		Where("mitra_id = ?", mitraID).
		Scan(&res).Error
	return res.TotalRating, res.TotalCount, err
}

// GetCustomerRatingAggregate returns the sum and count of ratings for a customer.
func (r *RatingRepository) GetCustomerRatingAggregate(customerID string) (totalRating float64, totalCount int64, err error) {
	type result struct {
		TotalRating float64
		TotalCount  int64
	}
	var res result
	err = r.DB.Model(&models.UserRating{}).
		Select("COALESCE(SUM(rating), 0) as total_rating, COUNT(id) as total_count").
		Where("customer_id = ?", customerID).
		Scan(&res).Error
	return res.TotalRating, res.TotalCount, err
}

// CreateRating inserts a new rating record inside a transaction.
func (r *RatingRepository) CreateRating(tx *gorm.DB, rating *models.UserRating) error {
	return tx.Create(rating).Error
}
