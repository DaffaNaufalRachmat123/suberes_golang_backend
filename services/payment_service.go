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
	"suberes_golang/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentService struct {
	PaymentRepo *repositories.PaymentRepository
	DB          *gorm.DB
}

const (
	PaymentImgAlias = "PAY_IMG_"
)

// validCreateTypes mirrors the JS create_model / update_model_image joi validation:
// valid('tunai','virtual account','ewallet','balance')
var validCreateTypes = map[string]bool{
	"tunai":           true,
	"virtual account": true,
	"ewallet":         true,
	"balance":         true,
}

// validUpdateTypes mirrors the JS update_model joi validation:
// valid('tunai','virtual account','transfer','balance')
var validUpdateTypes = map[string]bool{
	"tunai":           true,
	"virtual account": true,
	"transfer":        true,
	"balance":         true,
}

// GetAllActive returns all active payments with their enabled sub_payments.
// Mirrors: GET /index (no pagination – payment methods are a small, finite list).
func (s *PaymentService) GetAllActive() ([]models.Payment, error) {
	return s.PaymentRepo.FindAllActive()
}

// Create handles multipart upload (json_data field + file).
// Mirrors: POST /create
func (s *PaymentService) Create(ctx *gin.Context) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.PaymentCreateRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}
	if err := validatePaymentCreate(req); err != nil {
		return err
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return errors.New("Payment image required")
	}

	paymentPath, filename, fullPath, err := buildPaymentImagePath(ctx, fileHeader.Filename)
	if err != nil {
		return err
	}
	_ = paymentPath

	if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
		return err
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	payment := models.Payment{
		Title:    req.Title,
		Type:     req.Type,
		Desc:     req.Desc,
		Icon:     os.Getenv("PAYMENTS_IMAGE_PATH") + filename,
		IsActive: "1",
	}
	if err := s.PaymentRepo.Create(tx, &payment); err != nil {
		tx.Rollback()
		_ = os.Remove(fullPath)
		return err
	}

	return tx.Commit().Error
}

// UpdateWithImage handles multipart upload (json_data field + file), replacing the icon.
// Mirrors: PUT /update/image/:id
func (s *PaymentService) UpdateWithImage(ctx *gin.Context, id int) error {
	existing, err := s.PaymentRepo.FindById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Payment not found")
		}
		return err
	}

	jsonData := ctx.PostForm("json_data")
	var req dtos.PaymentCreateRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}
	if err := validatePaymentCreate(req); err != nil {
		return err
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return errors.New("Payment image required")
	}

	_, filename, fullPath, err := buildPaymentImagePath(ctx, fileHeader.Filename)
	if err != nil {
		return err
	}

	if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
		return err
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updates := map[string]interface{}{
		"title": req.Title,
		"type":  req.Type,
		"desc":  req.Desc,
		"icon":  os.Getenv("PAYMENTS_IMAGE_PATH") + filename,
	}
	if err := s.PaymentRepo.Update(tx, id, updates); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Remove old image after successful commit (mirrors JS: await removeImage after commit).
	_ = helpers.DeleteImageIfExists(existing.Icon)
	return nil
}

// Update handles a plain JSON body update (no image change).
// Mirrors: PUT /update/:id
func (s *PaymentService) Update(ctx *gin.Context, id int) error {
	_, err := s.PaymentRepo.FindById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Payment not found")
		}
		return err
	}

	var req dtos.PaymentUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return errors.New("title, type, and desc are required")
	}
	if !validUpdateTypes[req.Type] {
		return fmt.Errorf("type must be one of: tunai, virtual account, transfer, balance")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updates := map[string]interface{}{
		"title": req.Title,
		"type":  req.Type,
		"desc":  req.Desc,
	}
	if err := s.PaymentRepo.Update(tx, id, updates); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Delete removes a payment record and its icon from disk.
// Mirrors: DELETE /remove/:id
func (s *PaymentService) Delete(id int) error {
	existing, err := s.PaymentRepo.FindById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Payment not found")
		}
		return err
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.PaymentRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Remove icon after successful commit (mirrors JS: await removeImage after commit).
	_ = helpers.DeleteImageIfExists(existing.Icon)
	return nil
}

// --- helpers ---

func validatePaymentCreate(req dtos.PaymentCreateRequest) error {
	if req.Title == "" || req.Type == "" || req.Desc == "" {
		return errors.New("title, type, and desc are required")
	}
	if !validCreateTypes[req.Type] {
		return fmt.Errorf("type must be one of: tunai, virtual account, ewallet, balance")
	}
	return nil
}

// buildPaymentImagePath resolves the payment image directory and returns
// (paymentPath, filename, fullPath, error).
func buildPaymentImagePath(ctx *gin.Context, originalFilename string) (string, string, string, error) {
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return "", "", "", err
		}
	}

	paymentPath := filepath.Join(basePath, os.Getenv("PAYMENTS_IMAGE_PATH"))
	if _, err := os.Stat(paymentPath); os.IsNotExist(err) {
		if err := os.MkdirAll(paymentPath, 0755); err != nil {
			return "", "", "", err
		}
	}

	now := time.Now()
	filename := fmt.Sprintf(
		"%s%d-%02d-%02d_%02d-%02d-%02d_%s",
		PaymentImgAlias,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		originalFilename,
	)

	fullPath := filepath.Join(paymentPath, filename)
	return paymentPath, filename, fullPath, nil
}
