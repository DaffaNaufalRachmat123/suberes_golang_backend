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

type BantuanService struct {
	BantuanRepo *repositories.BantuanRepository
	DB          *gorm.DB
}

const (
	BantuanImgAlias = "BANTUAN_IMG_"
)

func (s *BantuanService) GetBantuans(page, limit int, helpType string) ([]models.Bantuan, int64, error) {
	return s.BantuanRepo.FindAll(page, limit, helpType)
}

func (s *BantuanService) GetBantuansAdmin(page, limit int) ([]models.Bantuan, int64, error) {
	return s.BantuanRepo.FindAllAdmin(page, limit)
}

func (s *BantuanService) GetBantuanByID(id uint) (*models.Bantuan, error) {
	bantuan, err := s.BantuanRepo.FindById(id)
	if err != nil {
		return nil, err
	}
	tx := s.DB.Begin()
	bantuan.WatchingCount += 1
	if err := s.BantuanRepo.Update(tx, bantuan); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return bantuan, nil
}

func (s *BantuanService) CreateBantuan(ctx *gin.Context) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.BantuanRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}

	form, _ := ctx.MultipartForm()
	contentFiles := form.File["content_images"]
	updatedBody := req.Description

	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	bantuanPath := filepath.Join(basePath, os.Getenv("BANTUAN_IMAGE_PATH"))
	if _, err := os.Stat(bantuanPath); os.IsNotExist(err) {
		_ = os.MkdirAll(bantuanPath, 0755)
	}

	if len(contentFiles) > 0 {
		re := regexp.MustCompile(`<img[^>]+src="blob:[^"]+"[^>]*>`)
		matches := re.FindAllString(updatedBody, -1)

		for i, match := range matches {
			if i < len(contentFiles) {
				cFile := contentFiles[i]

				now := time.Now()
				filename := fmt.Sprintf(
					"%s_%d-%02d-%02d_%02d-%02d-%02d_%s",
					BantuanImgAlias,
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
					cFile.Filename,
				)

				fullPath := filepath.Join(bantuanPath, filename)

				if err := ctx.SaveUploadedFile(cFile, fullPath); err != nil {
					return err
				}

				fileUrl := fmt.Sprintf(
					"%s/api/images/bantuan/%s",
					helpers.GetHostURL(ctx),
					filename,
				)

				srcRe := regexp.MustCompile(`src="[^"]+"`)
				updatedTag := srcRe.ReplaceAllString(
					match,
					fmt.Sprintf(`src="%s"`, fileUrl),
				)

				updatedBody = strings.Replace(updatedBody, match, updatedTag, 1)
			}
		}
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	bantuan := models.Bantuan{
		Title:       req.Title,
		Description: updatedBody,
		HelpType:    req.HelpType,
	}

	if err := s.BantuanRepo.Create(tx, &bantuan); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *BantuanService) UpdateBantuan(ctx *gin.Context, id uint) error {
	existBantuan, err := s.BantuanRepo.FindById(id)
	if err != nil {
		return err
	}
	if existBantuan == nil {
		return errors.New("Bantuan not found")
	}
	jsonData := ctx.PostForm("json_data")
	var req dtos.BantuanRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updatedBody := req.Description
	form, _ := ctx.MultipartForm()
	contentFiles := form.File["content_images"]
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	bantuanPath := filepath.Join(basePath, os.Getenv("BANTUAN_IMAGE_PATH"))
	if _, err := os.Stat(bantuanPath); os.IsNotExist(err) {
		_ = os.MkdirAll(bantuanPath, 0755)
	}

	if len(contentFiles) > 0 {
		re := regexp.MustCompile(`<img[^>]+src="blob:[^"]+"[^>]*>`)
		matches := re.FindAllString(updatedBody, -1)

		for i, match := range matches {
			if i < len(contentFiles) {
				cFile := contentFiles[i]

				now := time.Now()
				filename := fmt.Sprintf(
					"%s_%d-%02d-%02d_%02d-%02d-%02d_%s",
					BantuanImgAlias,
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
					cFile.Filename,
				)

				fullPath := filepath.Join(bantuanPath, filename)

				if err := ctx.SaveUploadedFile(cFile, fullPath); err != nil {
					tx.Rollback()
					return err
				}

				fileUrl := fmt.Sprintf(
					"%s/api/images/bantuan/%s",
					helpers.GetHostURL(ctx),
					filename,
				)

				srcRe := regexp.MustCompile(`src="[^"]+"`)
				updatedTag := srcRe.ReplaceAllString(match, fmt.Sprintf(`src="%s"`, fileUrl))
				updatedBody = strings.Replace(updatedBody, match, updatedTag, 1)
			}
		}
	}

	existBantuan.Title = req.Title
	existBantuan.Description = updatedBody
	existBantuan.HelpType = req.HelpType
	existBantuan.UpdatedAt = time.Now()

	err = s.BantuanRepo.Update(tx, existBantuan)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *BantuanService) DeleteBantuan(id uint) error {
	bantuan, err := s.BantuanRepo.FindById(id)
	if err != nil {
		return err
	}
	if bantuan == nil {
		return errors.New("Bantuan not found")
	}
	tx := s.BantuanRepo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	re := regexp.MustCompile(`<img[^>]+src="([^">]+)"`)
	matches := re.FindAllStringSubmatch(bantuan.Description, -1)

	for _, match := range matches {
		if len(match) > 1 {
			url := match[1]
			if strings.Contains(url, "/api/images/bantuan/") {
				parts := strings.Split(url, "/api/images/bantuan/")
				if len(parts) > 1 {
					filename := parts[1]
					filePath := filepath.Join(os.Getenv("BANTUAN_IMAGE_PATH"), filename)
					os.Remove(filePath)
				}
			}
		}
	}

	err = s.BantuanRepo.Delete(tx, bantuan)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
