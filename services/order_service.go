package services

import (
	"suberes_golang/models"
	"suberes_golang/repositories"
)

type OrderService struct {
	OrderTransactionRepo *repositories.OrderTransactionRepository
}

func NewOrderService(orderTransactionRepo *repositories.OrderTransactionRepository) *OrderService {
	return &OrderService{
		OrderTransactionRepo: orderTransactionRepo,
	}
}

func (s *OrderService) FindAllByStatusWithPagination(status string, page, limit int, search string) ([]models.OrderTransaction, int64, error) {
	return s.OrderTransactionRepo.FindAllByStatusWithPagination(status, page, limit, search)
}
