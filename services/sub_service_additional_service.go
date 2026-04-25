package services

import (
	"errors"
	"suberes_golang/dtos"
	"suberes_golang/models"
	"suberes_golang/repositories"

	"gorm.io/gorm"
)

type SubServiceAdditionalService struct {
	SubServiceAdditionalRepo *repositories.SubServiceAdditionalRepository
	DB                       *gorm.DB
}

func (s *SubServiceAdditionalService) Create(req dtos.CreateSubServiceAdditionalRequest) (*models.SubServiceAdditional, error) {
	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	additional := models.SubServiceAdditional{
		SubServiceID:   req.SubServiceID,
		Title:          req.Title,
		BaseAmount:     req.BaseAmount,
		Amount:         req.Amount,
		AdditionalType: req.AdditionalType,
	}

	if err := s.SubServiceAdditionalRepo.Create(tx, &additional); err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create sub service additional")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &additional, nil
}
