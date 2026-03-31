package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PanduanService struct {
	PanduanRepo *repositories.PanduanRepository
	DB          *gorm.DB
}

const (
	PanduanImgAlias = "GUIDE_IMG_"
)

func (s *PanduanService) GetPanduansCustomer(page, limit int) ([]models.GuideTable, int64, error) {
	return s.PanduanRepo.FindAllCustomer(page, limit)
}

func (s *PanduanService) GetPanduansMitra(page, limit int) ([]models.GuideTable, int64, error) {
	return s.PanduanRepo.FindAllMitra(page, limit)
}

func (s *PanduanService) GetPanduansAdmin(page, limit int) ([]models.GuideTable, int64, error) {
	return s.PanduanRepo.FindAllAdmin(page, limit)
}

func (s *PanduanService) GetPanduanByID(id uint) (*models.GuideTable, error) {
	return s.PanduanRepo.FindByID(id)
}

func (s *PanduanService) UpdateWatchingCount(id uint) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.PanduanRepo.UpdateWatchingCount(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *PanduanService) CreatePanduan(ctx *gin.Context) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.PanduanCreateRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}

	form, _ := ctx.MultipartForm()
	contentFiles := form.File["content_images"]
	updatedBody := req.GuideDescription

	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	panduanPath := filepath.Join(basePath, os.Getenv("GUIDE_IMAGE_PATH"))
	if _, err := os.Stat(panduanPath); os.IsNotExist(err) {
		_ = os.MkdirAll(panduanPath, 0755)
	}

	if len(contentFiles) > 0 {
		re := regexp.MustCompile(`<img[^>]+src="blob:[^"]+"[^>]*>`)
		matches := re.FindAllString(updatedBody, -1)

		for i, match := range matches {
			if i < len(contentFiles) {
				cFile := contentFiles[i]

				now := time.Now()
				filename := fmt.Sprintf(
					"%s%d-%02d-%02d_%02d-%02d-%02d_%s",
					PanduanImgAlias,
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
					cFile.Filename,
				)

				fullPath := filepath.Join(panduanPath, filename)

				if err := ctx.SaveUploadedFile(cFile, fullPath); err != nil {
					return err
				}

				fileUrl := fmt.Sprintf("%s://%s/api/images%s%s",
					ctx.Request.URL.Scheme,
					ctx.Request.Host,
					os.Getenv("GUIDE_IMAGE_PATH"),
					filename,
				)

				// Replace the blob src with fileUrl
				reSrc := regexp.MustCompile(`src="blob:[^"]*"`)
				updatedTag := reSrc.ReplaceAllString(match, `src="`+fileUrl+`"`)
				updatedBody = strings.Replace(updatedBody, match, updatedTag, 1)
			}
		}
	}

	panduan := &models.GuideTable{
		GuideTitle:       req.GuideTitle,
		GuideDescription: updatedBody,
		GuideType:        req.GuideType,
		WatchingCount:    0,
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.PanduanRepo.Create(tx, panduan); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *PanduanService) UpdatePanduan(ctx *gin.Context, id uint) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.PanduanUpdateRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}

	panduan, err := s.PanduanRepo.FindByID(id)
	if err != nil {
		return err
	}
	if panduan == nil {
		return errors.New("Panduan not found")
	}

	form, _ := ctx.MultipartForm()
	contentFiles := form.File["content_images"]
	updatedBody := req.GuideDescription

	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	panduanPath := filepath.Join(basePath, os.Getenv("GUIDE_IMAGE_PATH"))

	if len(contentFiles) > 0 {
		re := regexp.MustCompile(`<img[^>]+src="blob:[^"]+"[^>]*>`)
		matches := re.FindAllString(updatedBody, -1)

		for i, match := range matches {
			if i < len(contentFiles) {
				cFile := contentFiles[i]

				now := time.Now()
				filename := fmt.Sprintf(
					"%s%d-%02d-%02d_%02d-%02d-%02d_%s",
					PanduanImgAlias,
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
					cFile.Filename,
				)

				fullPath := filepath.Join(panduanPath, filename)

				if err := ctx.SaveUploadedFile(cFile, fullPath); err != nil {
					return err
				}

				fileUrl := fmt.Sprintf("%s://%s/api/images%s%s",
					ctx.Request.URL.Scheme,
					ctx.Request.Host,
					os.Getenv("GUIDE_IMAGE_PATH"),
					filename,
				)

				// Replace the blob src with fileUrl
				reSrc := regexp.MustCompile(`src="blob:[^"]*"`)
				updatedTag := reSrc.ReplaceAllString(match, `src="`+fileUrl+`"`)
				updatedBody = strings.Replace(updatedBody, match, updatedTag, 1)
			}
		}
	}

	panduan.GuideTitle = req.GuideTitle
	panduan.GuideDescription = updatedBody
	panduan.GuideType = req.GuideType

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.PanduanRepo.Update(tx, panduan); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *PanduanService) DeletePanduan(id uint) error {
	panduan, err := s.PanduanRepo.FindByID(id)
	if err != nil {
		return err
	}
	if panduan == nil {
		return errors.New("Panduan not found")
	}

	// Extract and remove images
	result := helpers.ExtractImagesFromText(panduan.GuideDescription)
	for _, guideImage := range result {
		// TODO: Implement removeImage function
		// removeImage(os.Getenv("GUIDE"), guideImage)
		_ = guideImage // avoid unused variable
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.PanduanRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
