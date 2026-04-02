package services

import (
	"errors"
	"suberes_golang/dtos"
	"suberes_golang/models"
	"suberes_golang/repositories"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SubServiceService struct {
	SubServiceRepo           *repositories.SubServiceRepository
	SubServiceAdditionalRepo *repositories.SubServiceAdditionalRepository
	UserRepo                 *repositories.UserRepository
	DB                       *gorm.DB
}

func (s *SubServiceService) Create(req dtos.SubServiceCreateRequest) (*models.SubService, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	newSubService := models.SubService{
		ServiceID:             uint(req.ServiceID),
		SubPriceServiceTitle:  req.SubPriceServiceTitle,
		SubPriceService:       req.SubPriceService,
		SubServiceDescription: req.SubServiceDescription,
		CompanyPercentage:     req.CompanyPercentage,
		MinutesSubServices:    req.MinutesSubServices,
		Criteria:              req.Criteria,
		IsRecommended:         req.IsRecommended,
	}

	if err := s.SubServiceRepo.Create(tx, &newSubService); err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(req.SubServiceAdditionals) > 0 {
		items := make([]models.SubServiceAdditional, 0, len(req.SubServiceAdditionals))
		for _, add := range req.SubServiceAdditionals {
			items = append(items, models.SubServiceAdditional{
				SubServiceID:   int(newSubService.ID),
				Title:          add.Title,
				Amount:         add.Amount,
				BaseAmount:     add.Amount,
				AdditionalType: add.AdditionalType,
			})
		}

		if err := s.SubServiceAdditionalRepo.CreateBulk(tx, items); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &newSubService, nil
}

func (s *SubServiceService) Update(req dtos.SubServiceUpdateRequest) (*models.SubService, error) {
	found, err := s.SubServiceRepo.FindByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Sub service not found")
		}
		return nil, err
	}

	if found == nil {
		return nil, errors.New("Sub service not found")
	}

	updateData := map[string]interface{}{
		"service_id":              req.ServiceID,
		"sub_price_service_title": req.SubPriceServiceTitle,
		"sub_price_service":       req.SubPriceService,
		"sub_service_description": req.SubServiceDescription,
		"company_percentage":      req.CompanyPercentage,
		"minutes_sub_services":    req.MinutesSubServices,
		"criteria":                req.Criteria,
		"is_recommended":          req.IsRecommended,
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.SubServiceRepo.Update(tx, req.ID, updateData); err != nil {
		tx.Rollback()
		return nil, err
	}

	updated, err := s.SubServiceRepo.FindByID(req.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *SubServiceService) Delete(id int, userId string, password string) error {
	subService, err := s.SubServiceRepo.FindByID(id)
	if err != nil {
		return err
	}
	if subService == nil {
		return errors.New("Sub service not found")
	}

	admin, err := s.UserRepo.FindAdminById(userId)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("Admin not found")
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)) != nil {
		return errors.New("Password is wrong")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.SubServiceAdditionalRepo.DeleteBySubServiceID(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := s.SubServiceRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
