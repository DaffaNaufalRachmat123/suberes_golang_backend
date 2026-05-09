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

const BankListImgAlias = "BANK_IMG_"

type BankListService struct {
	BankListRepo *repositories.BankListRepository
	UserRepo     *repositories.UserRepository
	DB           *gorm.DB
}

func NewBankListService(db *gorm.DB) *BankListService {
	return &BankListService{
		BankListRepo: &repositories.BankListRepository{DB: db},
		UserRepo:     &repositories.UserRepository{DB: db},
		DB:           db,
	}
}

// GetAllBankLists returns paginated bank lists without filter.
func (s *BankListService) GetAllBankLists(page, limit int) ([]models.BankList, int64, error) {
	return s.BankListRepo.FindAllBankLists(page, limit)
}

// GetTopupBanks returns paginated banks where can_topup = '1'.
func (s *BankListService) GetTopupBanks(page, limit int) ([]models.BankList, int64, error) {
	return s.BankListRepo.FindTopupBanks(page, limit)
}

// GetDisbursementBanks returns paginated banks where can_disbursement = '1'.
func (s *BankListService) GetDisbursementBanks(page, limit int) ([]models.BankList, int64, error) {
	return s.BankListRepo.FindDisbursementBanks(page, limit)
}

// BulkCreateBanks bulk-inserts items with method_type = 'bank'.
func (s *BankListService) BulkCreateBanks(adminID string, items []dtos.BankListCreateItem) error {
	return s.bulkCreate(adminID, items, "bank")
}

// BulkCreateEwallets bulk-inserts items with method_type = 'ewallet'.
func (s *BankListService) BulkCreateEwallets(adminID string, items []dtos.BankListCreateItem) error {
	return s.bulkCreate(adminID, items, "ewallet")
}

func (s *BankListService) bulkCreate(adminID string, items []dtos.BankListCreateItem, methodType string) error {
	admin, err := s.UserRepo.FindAdminById(adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("admin not found")
		}
		return err
	}
	_ = admin

	banks := make([]models.BankList, 0, len(items))
	for _, item := range items {
		banks = append(banks, models.BankList{
			Name:             item.Name,
			Code:             item.Code,
			DisbursementCode: item.DisbursementCode,
			BankImage:        item.BankImage,
			CanTopup:         item.CanTopup,
			CanDisbursement:  item.CanDisbursement,
			MinTopup:         item.MinTopup,
			MinDisbursement:  item.MinDisbursement,
			TopupFee:         item.TopupFee,
			DisbursementFee:  item.DisbursementFee,
			IsPercentage:     item.IsPercentage,
			MethodType:       methodType,
		})
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	if err := s.BankListRepo.BulkCreate(tx, banks); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateWithContext handles multipart/form-data (json_data + optional file) for bank_list update.
func (s *BankListService) UpdateWithContext(ctx *gin.Context, id int) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.BankListUpdateRequest
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			return errors.New("invalid data format")
		}
	}

	bank, err := s.BankListRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bank not found")
		}
		return err
	}

	// Handle image file upload
	fileHeader, fileErr := ctx.FormFile("file")
	if fileErr == nil && fileHeader != nil {
		basePath := filepath.Join(helpers.RootPath(), os.Getenv("IMAGE_PATH_CONTROLLER"))
		_ = os.MkdirAll(basePath, 0755)
		bankImgPath := filepath.Join(basePath, os.Getenv("PAYMENTS_IMAGE_PATH"))
		_ = os.MkdirAll(bankImgPath, 0755)

		now := time.Now()
		filename := fmt.Sprintf("%s%d-%02d-%02d_%02d-%02d-%02d_%s",
			BankListImgAlias,
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second(),
			fileHeader.Filename,
		)
		fullPath := filepath.Join(bankImgPath, filename)
		if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
			return err
		}
		// Delete old image if local
		if bank.BankImage != "" {
			_ = helpers.DeleteImageIfExists(bank.BankImage)
		}
		req.BankImage = os.Getenv("PAYMENTS_IMAGE_PATH") + filename
	}

	return s.Update(id, &req)
}

// Update merges the non-zero fields from the request into the bank record.
func (s *BankListService) Update(id int, req *dtos.BankListUpdateRequest) error {
	_, err := s.BankListRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bank not found")
		}
		return err
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.DisbursementCode != "" {
		updates["disbursement_code"] = req.DisbursementCode
	}
	if req.BankImage != "" {
		updates["bank_image"] = req.BankImage
	}
	if req.CanTopup != "" {
		updates["can_topup"] = req.CanTopup
	}
	if req.CanDisbursement != "" {
		updates["can_disbursement"] = req.CanDisbursement
	}
	if req.MinTopup != 0 {
		updates["min_topup"] = req.MinTopup
	}
	if req.MinDisbursement != 0 {
		updates["min_disbursement"] = req.MinDisbursement
	}
	if req.TopupFee != 0 {
		updates["topup_fee"] = req.TopupFee
	}
	if req.DisbursementFee != 0 {
		updates["disbursement_fee"] = req.DisbursementFee
	}
	if req.IsPercentage != "" {
		updates["is_percentage"] = req.IsPercentage
	}
	if req.MethodType != "" {
		updates["method_type"] = req.MethodType
	}
	if req.Tutorial != "" {
		updates["tutorial"] = req.Tutorial
	}

	if len(updates) == 0 {
		return nil
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	if err := s.BankListRepo.Update(tx, id, updates); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
