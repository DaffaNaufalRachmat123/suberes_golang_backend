package services

import (
	"errors"
	"suberes_golang/models"
	"suberes_golang/repositories"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CategoryServiceService struct {
	CategoryServiceRepo *repositories.CategoryServiceRepository
	UserRepo            *repositories.UserRepository
	DB                  *gorm.DB
}

func (s *CategoryServiceService) GetDetail(id uint) (*models.CategoryService, error) {
	cs, err := s.CategoryServiceRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("Category service not found")
	}
	return cs, nil
}

func (s *CategoryServiceService) Create(layananID int, categoryService string, creatorID string) error {
	payload := &models.CategoryService{
		LayananID:       layananID,
		CategoryService: categoryService,
		CreatorID:       creatorID,
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.CategoryServiceRepo.Create(tx, payload); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *CategoryServiceService) Update(id uint, layananID int, categoryService string, creatorID string) error {
	cs, err := s.CategoryServiceRepo.FindByID(id)
	if err != nil {
		return errors.New("Category service not found")
	}
	if cs == nil {
		return errors.New("Category service not found")
	}

	data := map[string]interface{}{
		"layanan_id":       layananID,
		"category_service": categoryService,
		"creator_id":       creatorID,
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.CategoryServiceRepo.Update(tx, id, data); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *CategoryServiceService) Delete(id uint, adminID string, password string) error {
	cs, err := s.CategoryServiceRepo.FindByID(id)
	if err != nil {
		return errors.New("Category service not found")
	}
	if cs == nil {
		return errors.New("Category service not found")
	}

	admin, err := s.UserRepo.FindAdminById(adminID)
	if err != nil {
		return errors.New("Admin not found")
	}
	if admin == nil {
		return errors.New("Admin not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return errors.New("Unauthorized")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.CategoryServiceRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
