package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const SubPaymentImgAlias = "SUB_PAY_IMG_"

type SubPaymentService struct {
	DB *gorm.DB
}

func (s *SubPaymentService) GetAll(page, limit int) ([]models.SubPayment, int64, error) {
	var subPayments []models.SubPayment
	var total int64

	offset := (page - 1) * limit

	if err := s.DB.Model(&models.SubPayment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.DB.
		Preload("SubPaymentTutorial").
		Offset(offset).
		Limit(limit).
		Find(&subPayments).Error; err != nil {
		return nil, 0, err
	}

	return subPayments, total, nil
}

func (s *SubPaymentService) GetByID(id int) (*models.SubPayment, error) {
	var subPayment models.SubPayment
	if err := s.DB.Preload("SubPaymentTutorial").First(&subPayment, id).Error; err != nil {
		return nil, errors.New("sub payment not found")
	}
	return &subPayment, nil
}

func (s *SubPaymentService) Update(ctx *gin.Context, id int) (*models.SubPayment, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var subPayment models.SubPayment
	if err := tx.First(&subPayment, id).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("sub payment not found")
	}

	// Parse json_data from multipart form
	jsonData := ctx.PostForm("json_data")
	var req dtos.SubPaymentUpdateRequest
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			tx.Rollback()
			return nil, errors.New("invalid data format")
		}
	}

	updates := map[string]interface{}{}

	// Handle icon image upload
	fileHeader, fileErr := ctx.FormFile("file")
	if fileErr == nil && fileHeader != nil {
		iconPath, err := s.saveSubPaymentIcon(ctx, fileHeader.Filename)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		// Delete old icon if it's a local file
		if subPayment.Icon != "" {
			_ = helpers.DeleteImageIfExists(subPayment.Icon)
		}
		updates["icon"] = iconPath
		if err := ctx.SaveUploadedFile(fileHeader, s.fullIconPath(fileHeader.Filename, iconPath)); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.TitlePayment != "" {
		updates["title_payment"] = req.TitlePayment
	}
	if req.Enabled != "" {
		updates["enabled"] = req.Enabled
	}
	if req.Desc != "" {
		updates["desc"] = req.Desc
	}

	if len(updates) > 0 {
		if err := tx.Model(&subPayment).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Upsert tutorial
	if req.SubPaymentTutorial != nil {
		var tutorial models.SubPaymentTutorial
		err := tx.Where("sub_payment_id = ?", id).First(&tutorial).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Sync the sequence to avoid duplicate pkey from stale auto-increment
			// (occurs when rows were seeded with explicit IDs via raw SQL)
			tx.Exec("SELECT setval('sub_payment_tutorials_id_seq', (SELECT COALESCE(MAX(id), 0) FROM sub_payment_tutorials))")

			tutorial = models.SubPaymentTutorial{
				PaymentID:    subPayment.PaymentID,
				SubPaymentID: id,
				Title:        req.SubPaymentTutorial.Title,
				Description:  req.SubPaymentTutorial.Description,
			}
			if err := tx.Create(&tutorial).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else if err != nil {
			tx.Rollback()
			return nil, err
		} else {
			// Always update both fields (allow clearing title with empty string)
			tutorialUpdates := map[string]interface{}{
				"title":       req.SubPaymentTutorial.Title,
				"description": req.SubPaymentTutorial.Description,
			}
			if err := tx.Model(&tutorial).Updates(tutorialUpdates).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	var result models.SubPayment
	s.DB.Preload("SubPaymentTutorial").First(&result, id)
	return &result, nil
}

// saveSubPaymentIcon builds the stored icon path (relative, for DB).
func (s *SubPaymentService) saveSubPaymentIcon(ctx *gin.Context, originalFilename string) (string, error) {
	basePath := filepath.Join(helpers.RootPath(), os.Getenv("IMAGE_PATH_CONTROLLER"))
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", err
	}

	subPayPath := filepath.Join(basePath, os.Getenv("PAYMENTS_IMAGE_PATH"))
	if err := os.MkdirAll(subPayPath, 0755); err != nil {
		return "", err
	}

	now := time.Now()
	filename := fmt.Sprintf(
		"%s%d-%02d-%02d_%02d-%02d-%02d_%s",
		SubPaymentImgAlias,
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		originalFilename,
	)
	iconPath := os.Getenv("PAYMENTS_IMAGE_PATH") + filename
	return iconPath, nil
}

// fullIconPath returns the absolute path on disk for SaveUploadedFile.
func (s *SubPaymentService) fullIconPath(originalFilename, iconPath string) string {
	basePath := filepath.Join(helpers.RootPath(), os.Getenv("IMAGE_PATH_CONTROLLER"))
	dir := filepath.Join(basePath, os.Getenv("PAYMENTS_IMAGE_PATH"))
	// extract just the filename part from iconPath
	filename := filepath.Base(iconPath)
	_ = originalFilename
	return filepath.Join(dir, filename)
}
