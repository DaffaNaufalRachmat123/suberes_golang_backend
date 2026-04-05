package services

import (
	"suberes_golang/models"
	"suberes_golang/repositories"
)

type RatingService struct {
	RatingRepo *repositories.RatingRepository
}

// GetMitraRatings returns ratings given to a mitra by customers.
func (s *RatingService) GetMitraRatings(mitraID string, limit, offset int) ([]models.UserRating, error) {
	return s.RatingRepo.FindMitraRatings(mitraID, limit, offset)
}

// GetCustomerRatings returns ratings given to customers by a mitra.
func (s *RatingService) GetCustomerRatings(mitraID string, limit, offset int) ([]models.UserRating, error) {
	return s.RatingRepo.FindCustomerRatings(mitraID, limit, offset)
}
