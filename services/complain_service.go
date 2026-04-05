package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ComplainService struct {
	ComplainRepo *repositories.ComplainRepository
	DB           *gorm.DB
}

const complainImgAlias = "CMP_IMG_"

// GetAllAdmin returns paginated complains for admin with search by complain_code.
func (s *ComplainService) GetAllAdmin(page, limit int, search string) ([]models.Complain, int64, error) {
	return s.ComplainRepo.FindAllAdmin(page, limit, search)
}

// GetAllCustomer returns paginated complains for customer index.
func (s *ComplainService) GetAllCustomer(page, limit int) ([]models.Complain, int64, error) {
	return s.ComplainRepo.FindAllCustomer(page, limit)
}

// GetAllMitra returns paginated complains for mitra index.
func (s *ComplainService) GetAllMitra(page, limit int) ([]models.Complain, int64, error) {
	return s.ComplainRepo.FindAllMitra(page, limit)
}

// GetDetail returns a single complain with images.
func (s *ComplainService) GetDetail(id int) (*models.Complain, error) {
	return s.ComplainRepo.FindByID(id)
}

// Create creates a complain and saves uploaded images.
func (s *ComplainService) Create(ctx *gin.Context) error {
	customerID := ctx.PostForm("customer_id")
	titleProblem := ctx.PostForm("title_problem")
	problem := ctx.PostForm("problem")

	if strings.TrimSpace(customerID) == "" || strings.TrimSpace(titleProblem) == "" || strings.TrimSpace(problem) == "" {
		return errors.New("customer_id, title_problem, and problem are required")
	}

	form, err := ctx.MultipartForm()
	if err != nil && !errors.Is(err, multipart.ErrMessageTooLarge) {
		form = &multipart.Form{}
	}

	// Prepare image save directory
	basePath := filepath.Join(helpers.RootPath(), os.Getenv("IMAGE_PATH_CONTROLLER"))
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}
	complainPath := filepath.Join(basePath, os.Getenv("COMPLAIN_IMAGE_PATH"))
	if _, err := os.Stat(complainPath); os.IsNotExist(err) {
		_ = os.MkdirAll(complainPath, 0755)
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	complainCode := fmt.Sprintf("#%s", strings.ToUpper(helpers.GenerateRandomAlphaNum(7)))

	complain := &models.Complain{
		ComplainCode: complainCode,
		CustomerID:   customerID,
		TitleProblem: titleProblem,
		Problem:      problem,
		Status:       "SENT",
	}

	if err := s.ComplainRepo.Create(tx, complain); err != nil {
		tx.Rollback()
		return err
	}

	// Save uploaded image files and build bulk insert
	var complainImages []models.ComplainImage
	if form != nil {
		files := form.File["file"]
		for _, fileHeader := range files {
			filename := helpers.GenerateFilename(complainImgAlias, fileHeader.Filename)
			fullPath := filepath.Join(complainPath, filename)

			if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
				tx.Rollback()
				return err
			}

			width, height, dimErr := helpers.GetImageDimension(fileHeader)
			sizeDimension := ""
			if dimErr == nil {
				sizeDimension = fmt.Sprintf("Width : %d Height : %d", width, height)
			}

			complainImages = append(complainImages, models.ComplainImage{
				ComplainID:         strconv.Itoa(complain.ID),
				ImageName:          filename,
				ImageSize:          formatComplainFileSize(fileHeader.Size),
				ImageSizeDimension: sizeDimension,
			})
		}
	}

	if len(complainImages) > 0 {
		if err := s.ComplainRepo.BulkCreateImages(tx, complainImages); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// UpdateStatus updates the status of a complain. Returns error if not found.
func (s *ComplainService) UpdateStatus(id int, status string) error {
	_, err := s.ComplainRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("complain not found")
		}
		return err
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.ComplainRepo.UpdateStatus(tx, id, status); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Remove deletes a complain by id. Returns error if not found.
func (s *ComplainService) Remove(id int) error {
	_, err := s.ComplainRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("complain not found")
		}
		return err
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.ComplainRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// formatComplainFileSize formats file size in bytes into a human-readable string (mirrors JS formatBytes).
func formatComplainFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
