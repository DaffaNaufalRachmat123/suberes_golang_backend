package services

import (
	"errors"
	"fmt"
	"strings"
	"suberes_golang/models"
	"suberes_golang/repositories"

	"gorm.io/gorm"
)

type RatingService struct {
	RatingRepo           *repositories.RatingRepository
	OrderTransactionRepo *repositories.OrderTransactionRepository
	UserRepo             *repositories.UserRepository
	ServiceRepo          *repositories.ServiceRepository
	DB                   *gorm.DB
}

// GetMitraRatings returns ratings given to a mitra by customers.
func (s *RatingService) GetMitraRatings(mitraID string, limit, offset int) ([]models.UserRating, error) {
	return s.RatingRepo.FindMitraRatings(mitraID, limit, offset)
}

// GetCustomerRatings returns ratings given to customers by a mitra.
func (s *RatingService) GetCustomerRatings(mitraID string, limit, offset int) ([]models.UserRating, error) {
	return s.RatingRepo.FindCustomerRatings(mitraID, limit, offset)
}

type MitraRatingHomeResponse struct {
	ServerMessage string              `json:"server_message"`
	Status        string              `json:"status"`
	TotalRating   int64               `json:"total_rating"`
	AverageRating string              `json:"average_rating"`
	Rate5         int64               `json:"rate_5"`
	Rate4         int64               `json:"rate_4"`
	Rate3         int64               `json:"rate_3"`
	Rate2         int64               `json:"rate_2"`
	Rate1         int64               `json:"rate_1"`
	ReviewData    []models.UserRating `json:"review_data"`
}

// GetMitraRatingHome returns rating breakdown + last 5 reviews for the mitra home page.
func (s *RatingService) GetMitraRatingHome(mitraID string) (*MitraRatingHomeResponse, error) {
	totalRating, err := s.RatingRepo.CountMitraRatings(mitraID)
	if err != nil {
		return nil, err
	}

	mitra, err := s.UserRepo.FindMitraById(mitraID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var averageRating float64
	if mitra != nil {
		averageRating = mitra.UserRating
	}

	rate5, _ := s.RatingRepo.CountMitraRatingsByRange(mitraID, 5.0, 5.0)
	rate4, _ := s.RatingRepo.CountMitraRatingsByRange(mitraID, 4.0, 4.9)
	rate3, _ := s.RatingRepo.CountMitraRatingsByRange(mitraID, 3.0, 3.9)
	rate2, _ := s.RatingRepo.CountMitraRatingsByRange(mitraID, 2.0, 2.9)
	rate1, _ := s.RatingRepo.CountMitraRatingsByRange(mitraID, 1.0, 1.9)

	reviews, err := s.RatingRepo.FindMitraRatingsHome(mitraID)
	if err != nil {
		return nil, err
	}

	avg := fmt.Sprintf("%.1f", averageRating)

	return &MitraRatingHomeResponse{
		ServerMessage: "success",
		Status:        "ok",
		TotalRating:   totalRating,
		AverageRating: avg,
		Rate5:         rate5,
		Rate4:         rate4,
		Rate3:         rate3,
		Rate2:         rate2,
		Rate1:         rate1,
		ReviewData:    reviews,
	}, nil
}

// GetMitraRatingsPaginated returns paginated rating list.
func (s *RatingService) GetMitraRatingsPaginated(mitraID string, page, limit int) ([]models.UserRating, int64, error) {
	offset := (page - 1) * limit
	return s.RatingRepo.FindMitraRatingsPaginated(mitraID, limit, offset)
}

type CreateRatingRequest struct {
	Comment string `json:"comment"`
}

// CreateRatingToMitra handles rating submission from customer to mitra.
func (s *RatingService) CreateRatingToMitra(orderID, customerID, mitraID string, rating float64, req CreateRatingRequest) error {
	order, err := s.OrderTransactionRepo.FindById(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order not found")
		}
		return err
	}

	layananID, err := s.resolveLayananID(order.ServiceID)
	if err != nil {
		return err
	}

	customer, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil || customer == nil {
		return errors.New("customer or mitra data not found")
	}
	mitra, err := s.UserRepo.FindMitraById(mitraID)
	if err != nil || mitra == nil {
		return errors.New("customer or mitra data not found")
	}

	totalRatingSum, totalCount, err := s.RatingRepo.GetMitraRatingAggregate(mitraID)
	if err != nil {
		return err
	}

	var newAverage float64
	if totalCount > 0 {
		newAverage = (rating + totalRatingSum) / float64(totalCount+1)
	} else {
		newAverage = rating
	}

	comment := req.Comment
	if strings.ToLower(comment) == "empty" {
		comment = ""
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	newRating := &models.UserRating{
		OrderID:      orderID,
		CustomerID:   customerID,
		MitraID:      mitraID,
		LayananID:    layananID,
		ServiceID:    order.ServiceID,
		SubServiceID: order.SubServiceID,
		Rating:       rating,
		Comment:      comment,
		RatingType:   "customer to mitra",
	}
	if err := s.RatingRepo.CreateRating(tx, newRating); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.OrderTransaction{}).
		Where("id = ? AND order_status = ?", orderID, "FINISH").
		Updates(map[string]interface{}{"is_rated": "1", "rated": int(rating)}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.User{}).
		Where("id = ? AND user_type = ?", mitraID, "mitra").
		Update("user_rating", newAverage).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CreateRatingToCustomer handles rating submission from mitra to customer.
func (s *RatingService) CreateRatingToCustomer(orderID, customerID, mitraID string, rating float64, req CreateRatingRequest) error {
	order, err := s.OrderTransactionRepo.FindById(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order not found")
		}
		return err
	}

	layananID, err := s.resolveLayananID(order.ServiceID)
	if err != nil {
		return err
	}

	customer, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil || customer == nil {
		return errors.New("customer or mitra data not found")
	}
	mitra, err := s.UserRepo.FindMitraById(mitraID)
	if err != nil || mitra == nil {
		return errors.New("customer or mitra data not found")
	}

	totalRatingSum, totalCount, err := s.RatingRepo.GetCustomerRatingAggregate(customerID)
	if err != nil {
		return err
	}

	var newAverage float64
	if totalCount > 0 {
		newAverage = (rating + totalRatingSum) / float64(totalCount+1)
	} else {
		newAverage = rating
	}

	comment := req.Comment
	if strings.ToLower(comment) == "empty" {
		comment = ""
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	newRating := &models.UserRating{
		OrderID:      orderID,
		CustomerID:   customerID,
		MitraID:      mitraID,
		LayananID:    layananID,
		ServiceID:    order.ServiceID,
		SubServiceID: order.SubServiceID,
		Rating:       rating,
		Comment:      comment,
		RatingType:   "mitra to customer",
	}
	if err := s.RatingRepo.CreateRating(tx, newRating); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.OrderTransaction{}).
		Where("id = ? AND order_status = ?", orderID, "FINISH").
		Updates(map[string]interface{}{"is_rated_customer": "1", "rated_customer": int(rating)}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.User{}).
		Where("id = ? AND user_type = ?", customerID, "customer").
		Update("user_rating", newAverage).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// resolveLayananID walks service -> category_service to get layanan_id.
func (s *RatingService) resolveLayananID(serviceID int) (int, error) {
	category, err := s.ServiceRepo.FindServiceType(serviceID)
	if err != nil {
		return 0, errors.New("service not found")
	}
	return category.LayananID, nil
}
