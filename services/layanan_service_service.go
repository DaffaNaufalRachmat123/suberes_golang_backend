package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LayananServiceService struct {
	LayananServiceRepo *repositories.LayananServiceRepository
	CategoryServiceRepo *repositories.CategoryServiceRepository
	DB                 *gorm.DB
}

const (
	LayananImgAlias = "LAYANAN_IMG_"
)

func (s *LayananServiceService) GetLayananService(page, limit int) ([]models.LayananService, int64, error) {
	return s.LayananServiceRepo.FindAllPagination(page, limit)
}

func (s *LayananServiceService) GetCategoryServiceByLayananID(layananID, page, limit int) ([]models.CategoryService, int64, error) {
	return s.CategoryServiceRepo.FindAllByLayananIDPagination(layananID, page, limit)
}

func (s *LayananServiceService) GetLayananByID(id uint) (*models.LayananService, error) {
	return s.LayananServiceRepo.FindByID(id)
}
func (s *LayananServiceService) GetLayananPopular() ([]models.LayananService, error) {
	return s.LayananServiceRepo.FindPopular(5)
}
func (s *LayananServiceService) Create(ctx *gin.Context) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.LayananServiceRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return errors.New("Layanan service image required")
	}
	width, height, err := helpers.GetImageDimension(fileHeader)
	if err != nil {
		return err
	}
	layananWidth, err := strconv.Atoi(os.Getenv("LAYANAN_IMAGE_WIDTH"))
	if err != nil {
		return fmt.Errorf("Invalid LAYANAN_IMAGE_WIDTH")
	}
	layananHeight, err := strconv.Atoi(os.Getenv("LAYANAN_IMAGE_HEIGHT"))
	if err != nil {
		return fmt.Errorf("Invalid LAYANAN_IMAGE_HEIGHT")
	}
	if width != layananWidth || height != layananHeight {
		return fmt.Errorf(
			"Image dimension must be %dpx x %dpx", layananWidth, layananHeight,
		)
	}
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	layananPath := filepath.Join(basePath, os.Getenv("LAYANAN_IMAGE_PATH"))
	if _, err := os.Stat(layananPath); os.IsNotExist(err) {
		_ = os.MkdirAll(layananPath, 0755)
	}

	// === SETUP FILENAME (setara filename multer) ===
	now := time.Now()
	filename := fmt.Sprintf(
		"%s%d-%02d-%02d_%02d-%02d-%02d_%s",
		LayananImgAlias,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		fileHeader.Filename,
	)

	fullPath := filepath.Join(layananPath, filename)

	// === SAVE FILE ===
	if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
		return err
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	payload := models.LayananService{
		LayananTitle:          req.LayananTitle,
		LayananDescription:    req.LayananDescription,
		LayananImage:          os.Getenv("LAYANAN_IMAGE_PATH") + filename,
		LayananImageSize:      strconv.FormatInt((fileHeader.Size / (1024 * 1024)), 10),
		LayananImageDimension: fmt.Sprintf("%dpx x %dpx", width, height),
		IsActive:              req.IsActive,
	}

	if err := tx.Create(&payload).Error; err != nil {
		tx.Rollback()
		_ = os.Remove(fullPath)
		return err
	}

	tx.Commit()
	return nil
}
func (s *LayananServiceService) Update(ctx *gin.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return err
	}

	layanan, err := s.LayananServiceRepo.FindByID(uint(id))
	if err != nil {
		return errors.New("Could not load layanan service data")
	}
	if layanan == nil {
		return errors.New("Layanan service not found")
	}
	jsonData := ctx.PostForm("json_data")
	var req dtos.LayananServiceRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid request format")
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return errors.New("Layanan service image required")
	}
	width, height, err := helpers.GetImageDimension(fileHeader)
	if err != nil {
		return err
	}
	layananWidth, err := strconv.Atoi(os.Getenv("LAYANAN_IMAGE_WIDTH"))
	if err != nil {
		return fmt.Errorf("Invalid LAYANAN_IMAGE_WIDTH")
	}
	layananHeight, err := strconv.Atoi(os.Getenv("LAYANAN_IMAGE_HEIGHT"))
	if err != nil {
		return fmt.Errorf("Invalid LAYANAN_IMAGE_HEIGHT")
	}
	if width != layananWidth || height != layananHeight {
		return fmt.Errorf(
			"Image dimension must be %dpx x %dpx", layananWidth, layananHeight,
		)
	}
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	layananPath := filepath.Join(basePath, os.Getenv("LAYANAN_IMAGE_PATH"))
	if _, err := os.Stat(layananPath); os.IsNotExist(err) {
		_ = os.MkdirAll(layananPath, 0755)
	}

	// === SETUP FILENAME (setara filename multer) ===
	now := time.Now()
	filename := fmt.Sprintf(
		"%s%d-%02d-%02d_%02d-%02d-%02d_%s",
		LayananImgAlias,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		fileHeader.Filename,
	)

	fullPath := filepath.Join(layananPath, filename)

	// === SAVE FILE ===
	if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
		return err
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	oldImagePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
		layanan.LayananImage,
	)

	// ignore error kalau file tidak ada
	_ = os.Remove(oldImagePath)

	// === UPDATE DATA ===
	payload := models.LayananService{
		LayananTitle:          req.LayananTitle,
		LayananDescription:    req.LayananDescription,
		LayananImage:          filepath.Join(os.Getenv("LAYANAN_IMAGE_PATH"), filename),
		LayananImageSize:      strconv.FormatInt(fileHeader.Size/(1024*1024), 10),
		LayananImageDimension: fmt.Sprintf("%dpx x %dpx", width, height),
		IsActive:              req.IsActive,
	}

	if err := tx.Model(&layanan).Updates(payload).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
func (s *LayananServiceService) Delete(id uint) error {
	layanan, err := s.LayananServiceRepo.FindByID(id)
	if err != nil {
		return err
	}
	if layanan == nil {
		return errors.New("layanan service not found")
	}

	tx := s.DB.Begin()

	// hapus image
	_ = helpers.DeleteImageIfExists(layanan.LayananImage)

	if err := tx.Delete(&layanan).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
