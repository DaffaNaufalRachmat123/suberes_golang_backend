package services

import (
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
)

type TransactionService struct {
	TransactionRepo *repositories.TransactionRepository
}

func NewTransactionService(transactionRepo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{
		TransactionRepo: transactionRepo,
	}
}

func (s *TransactionService) FindAllWithPagination(page, limit int, search, transactionType string) ([]models.Transaction, int64, error) {
	return s.TransactionRepo.FindAllWithPagination(page, limit, search, transactionType)
}

func (s *TransactionService) GetTransactionTypesByMitraIDAndDate(mitraID string, date string) ([]map[string]interface{}, error) {
	startDate, endDate := helpers.GetStartEndDateFromString(date)
	return s.TransactionRepo.GetTransactionTypesByMitraIDAndDate(mitraID, startDate, endDate)
}

func (s *TransactionService) FindAllByMitraIDWithPagination(mitraID, transactionFor, transactionTime string, page, limit int) ([]models.Transaction, int64, error) {
	startDate, endDate := helpers.GetStartEndDateFromString(transactionTime)
	return s.TransactionRepo.FindAllByMitraIDWithPagination(mitraID, transactionFor, startDate, endDate, page, limit)
}

func (s *TransactionService) FindDisbursementsByMitraID(mitraID string, page, limit int) ([]models.Transaction, int64, error) {
	return s.TransactionRepo.FindDisbursementsByMitraID(mitraID, page, limit)
}
