package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ServiceService struct {
	ServiceRepo *repositories.ServiceRepository
	UserRepo    *repositories.UserRepository
	DB          *gorm.DB
}

const (
	ServiceImgAlias = "SERVICE_IMG_"
)

func (s *ServiceService) GetServices(parent_id, page, limit int) ([]models.Service, int64, error) {
	return s.ServiceRepo.FindAllPagination(parent_id, page, limit)
}
func (s *ServiceService) SearchServices(layananID int, serviceName string) ([]models.CategoryService, error) {
	return s.ServiceRepo.Search(layananID, serviceName)
}
func (s *ServiceService) GetServiceByID(id int) (*models.Service, error) {
	return s.ServiceRepo.FindByID(id)
}
func (s *ServiceService) GetLayananServices(id int) ([]models.LayananService, error) {
	return s.ServiceRepo.FindLayananServices(id)
}
func (s *ServiceService) GetServiceType(serviceID int) (*models.CategoryService, error) {

	data, err := s.ServiceRepo.FindServiceType(serviceID)
	if err != nil {
		return nil, err
	}

	for i := range data.Services {
		if data.Services[i].ServiceType == "Durasi" {
			sort.Slice(data.Services[i].SubServices, func(a, b int) bool {
				return data.Services[i].SubServices[a].MinutesSubServices <
					data.Services[i].SubServices[b].MinutesSubServices
			})
		}
	}

	return data, nil
}

func (s *ServiceService) GetPopular() ([]models.Service, error) {
	return s.ServiceRepo.FindPopular(5)
}
func (s *ServiceService) Create(ctx *gin.Context) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.ServiceRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data request")
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return errors.New("Service image required")
	}
	width, height, err := helpers.GetImageDimension(fileHeader)
	if err != nil {
		return err
	}
	serviceWidth, err := strconv.Atoi(os.Getenv("SERVICE_IMAGE_WIDTH"))
	if err != nil {
		return fmt.Errorf("Invalid Service Image Width")
	}
	serviceHeight, err := strconv.Atoi(os.Getenv("SERVICE_IMAGE_HEIGHT"))
	if err != nil {
		return fmt.Errorf("Invalid Service Image Height")
	}
	if width != serviceWidth || height != serviceHeight {
		return fmt.Errorf(
			"Image dimension must be %dpx x %dpx",
			serviceWidth,
			serviceHeight,
		)
	}
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}
	servicePath := filepath.Join(basePath, os.Getenv("SERVICE_IMAGE_PATH"))
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		_ = os.MkdirAll(servicePath, 0755)
	}
	now := time.Now()
	filename := fmt.Sprintf(
		"%s%d-%02d-%02d_%02d-%02d-%02d_%s",
		ServiceImgAlias,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		fileHeader.Filename,
	)
	fullPath := filepath.Join(servicePath, filename)
	if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
		return err
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	payload := models.Service{
		ServiceName:           req.ServiceName,
		ServiceDescription:    req.ServiceDescription,
		ServiceImageThumbnail: os.Getenv("SERVICE_IMAGE_PATH") + filename,
		ServiceCount:          0,
		ServiceType:           req.ServiceType,
		ServiceCategory:       req.ServiceCategory,
		IsActive:              "1",
	}
	if err := tx.Create(&payload).Error; err != nil {
		tx.Rollback()
		_ = os.Remove(fullPath)
		return err
	}
	tx.Commit()
	return nil
}

func (s *ServiceService) Update(ctx *gin.Context) error {

	var req dtos.ServiceUpdateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return errors.New("Invalid data request")
	}
	var existing models.Service
	if err := s.DB.Where("id = ?", req.ID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Service not found")
		}
		return err
	}

	tx := s.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	payload := models.Service{
		ServiceName:        req.ServiceName,
		ServiceDescription: req.ServiceDescription,
		ServiceCategory:    req.ServiceCategory,
		ServiceType:        req.ServiceType,
	}

	if err := tx.Model(&models.Service{}).
		Where("id = ?", req.ID).
		Updates(payload).Error; err != nil {
		tx.Rollback()
		return err
	}

	if existing.ServiceType != req.ServiceType {
		if err := tx.Where("service_id = ?", req.ID).
			Delete(&models.SubService{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
func (s *ServiceService) UpdateImage(ctx *gin.Context) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.ServiceUpdateRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid request format")
	}
	service, err := s.ServiceRepo.FindByID(req.ID)
	if err != nil {
		return err
	}
	if service == nil {
		return errors.New("Service not found")
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return err
	}
	width, height, err := helpers.GetImageDimension(fileHeader)
	if err != nil {
		return err
	}
	serviceWidth, err := strconv.Atoi(os.Getenv("SERVICE_IMAGE_WIDTH"))
	if err != nil {
		return fmt.Errorf("Invalid SERVICE_IMAGE_WIDTH")
	}
	serviceHeight, err := strconv.Atoi(os.Getenv("SERVICE_IMAGE_HEIGHT"))
	if err != nil {
		return fmt.Errorf("Invalid SERVICE_IMAGE_HEIGHT")
	}
	if width != serviceWidth || height != serviceHeight {
		return fmt.Errorf(
			"Image dimension must be %dpx x %dpx", serviceWidth, serviceHeight,
		)
	}
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}
	servicePath := filepath.Join(basePath, os.Getenv("SERVICE_IMAGE_PATH"))
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		_ = os.MkdirAll(servicePath, 0755)
	}
	now := time.Now()
	filename := fmt.Sprintf(
		"%s%d-%02d-%02d_%02d-%02d-%02d_%s",
		ServiceImgAlias,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		fileHeader.Filename,
	)
	fullPath := filepath.Join(servicePath, filename)
	if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
		return err
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	oldImgPath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
		service.ServiceImageThumbnail,
	)
	_ = os.Remove(oldImgPath)
	payload := models.Service{
		ServiceName:           req.ServiceName,
		ServiceDescription:    req.ServiceDescription,
		ServiceImageThumbnail: os.Getenv("SERVICE_IMAGE_PATH") + filename,
		ServiceType:           req.ServiceType,
		ServiceCategory:       req.ServiceCategory,
	}
	if err := tx.Model(&service).Updates(payload).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (s *ServiceService) GetServiceDetail(id int) (*models.Service, error) {
	return s.ServiceRepo.FindServiceWithSubServicesByID(id)
}

func (s *ServiceService) Delete(id int, userId string, password string) error {
	service, err := s.ServiceRepo.FindByID(id)
	if err != nil {
		return err
	}
	if service == nil {
		return errors.New("Service not found")
	}
	admin, err := s.UserRepo.FindAdminById(userId)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("Admin not found")
	}
	if bcrypt.CompareHashAndPassword(
		[]byte(admin.Password),
		[]byte(password),
	) != nil {
		return errors.New("Password is wrong")
	}
	tx := s.DB.Begin()
	if err := tx.Delete(&service).Error; err != nil {
		tx.Rollback()
		return err
	}
	_ = helpers.DeleteImageIfExists(service.ServiceImageThumbnail)
	return tx.Commit().Error
}
