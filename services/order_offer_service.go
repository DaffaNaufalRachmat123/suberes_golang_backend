package services

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"suberes_golang/models"
	"suberes_golang/repositories"

	"gorm.io/gorm"
)

type OrderOfferService struct {
	DB             *gorm.DB
	OrderOfferRepo *repositories.OrderOfferRepository
	OrderRepo      *repositories.OrderRepository
}

func NewOrderOfferService(db *gorm.DB) *OrderOfferService {
	return &OrderOfferService{
		DB:             db,
		OrderOfferRepo: &repositories.OrderOfferRepository{DB: db},
		OrderRepo:      &repositories.OrderRepository{DB: db},
	}
}

// GetIncomingOrderList returns a paginated list of order offers for the given mitra.
func (s *OrderOfferService) GetIncomingOrderList(mitraID string, page, limit int) ([]models.OrderOffer, int64, int, error) {
	offers, total, err := s.OrderOfferRepo.FindPaginatedByMitraID(mitraID, page, limit)
	if err != nil {
		return nil, 0, http.StatusInternalServerError, err
	}
	return offers, total, http.StatusOK, nil
}

// GetIncomingOrder returns a single order offer with full detail and a countdown timer.
// The countdown SQL is computed from the order status:
//   - WAITING_FOR_SELECTED_MITRA → based on order_blast_time
//   - FINDING_MITRA              → based on order_time
func (s *OrderOfferService) GetIncomingOrder(orderID, mitraID string) (*models.OrderOffer, int, error) {
	// First fetch order status for the countdown expression
	var orderData struct {
		OrderStatus  string
		TimezoneCode string
	}
	if err := s.DB.
		Model(&models.OrderTransaction{}).
		Select("order_status, timezone_code").
		Where("id = ?", orderID).
		Scan(&orderData).Error; err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("order not found: %w", err)
	}

	timeoutMinutesStr := os.Getenv("TIMEOUT_CAN_TAKE_ORDER")
	timeoutMinutes, _ := strconv.Atoi(timeoutMinutesStr)
	if timeoutMinutes == 0 {
		timeoutMinutes = 5
	}

	// Build the raw countdown expression (PostgreSQL syntax)
	var countdownSQL string

	if orderData.OrderStatus == "WAITING_FOR_SELECTED_MITRA" {
		countdownSQL = fmt.Sprintf(
			"GREATEST(0, EXTRACT(EPOCH FROM (order_transactions.order_blast_time + INTERVAL '%d minutes' - (NOW() AT TIME ZONE 'UTC')))::bigint)",
			timeoutMinutes,
		)
	} else {
		countdownSQL = fmt.Sprintf(
			"GREATEST(0, EXTRACT(EPOCH FROM (order_transactions.order_time + INTERVAL '%d minutes' - (NOW() AT TIME ZONE 'UTC')))::bigint)",
			timeoutMinutes,
		)
	}

	offer, err := s.OrderOfferRepo.FindDetailByOrderAndMitra(orderID, mitraID, countdownSQL)
	if err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("order offer not found: %w", err)
	}

	return offer, http.StatusOK, nil
}
