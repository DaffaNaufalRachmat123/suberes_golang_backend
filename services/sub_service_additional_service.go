package services

import (
	"errors"
	"suberes_golang/dtos"
	"suberes_golang/models"
	"suberes_golang/repositories"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SubServiceAdditionalService struct {
	SubServiceAdditionalRepo *repositories.SubServiceAdditionalRepository
	UserRepo                 *repositories.UserRepository
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

func (s *SubServiceAdditionalService) Update(req dtos.UpdateSubServiceAdditionalRequest) (*models.SubServiceAdditional, error) {
	found, err := s.SubServiceAdditionalRepo.FindByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("sub service additional not found")
		}
		return nil, err
	}
	if found == nil {
		return nil, errors.New("sub service additional not found")
	}

	updateData := map[string]interface{}{
		"sub_service_id":  req.SubServiceID,
		"title":           req.Title,
		"base_amount":     req.BaseAmount,
		"amount":          req.Amount,
		"additional_type": req.AdditionalType,
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.SubServiceAdditionalRepo.Update(tx, req.ID, updateData); err != nil {
		tx.Rollback()
		return nil, err
	}

	updated, err := s.SubServiceAdditionalRepo.FindByID(req.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *SubServiceAdditionalService) Delete(id int, userId string, password string) error {
	record, err := s.SubServiceAdditionalRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("sub service additional not found")
		}
		return err
	}
	if record == nil {
		return errors.New("sub service additional not found")
	}

	admin, err := s.UserRepo.FindAdminById(userId)
	if err != nil {
		return errors.New("admin not found")
	}
	if admin == nil {
		return errors.New("admin not found")
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)) != nil {
		return errors.New("password is wrong")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.SubServiceAdditionalRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
