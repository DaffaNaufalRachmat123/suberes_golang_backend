package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BannerService struct {
	BannerRepo *repositories.BannerRepository
	DB         *gorm.DB
}

const (
	BannerImgAlias = "BNR_IMG_"
)

func (s *BannerService) GetBanners(page, limit int) ([]models.BannerList, int64, error) {
	return s.BannerRepo.FindAll(page, limit)
}
func (s *BannerService) GetBannerByID(id uint) (*models.BannerList, error) {
	return s.BannerRepo.FindById(id)
}
func (s *BannerService) GetPopularBanners() ([]models.BannerList, error) {
	return s.BannerRepo.FindPopular(5)
}
func (s *BannerService) CreateBanner(ctx *gin.Context) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.BannerRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}

	adminRepo := &repositories.AdminRepository{DB: s.DB}
	adminData, err := adminRepo.FindAdminByID(req.CreatorID)
	if err != nil || adminData == nil {
		return errors.New("admin not found")
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return errors.New("Banner image required")
	}
	width, height, err := helpers.GetImageDimension(fileHeader)
	if err != nil {
		return err
	}
	bannerWidth, err := strconv.Atoi(os.Getenv("BANNER_IMAGE_WIDTH"))
	if err != nil {
		return fmt.Errorf("invalid BANNER_IMAGE_WIDTH")
	}

	bannerHeight, err := strconv.Atoi(os.Getenv("BANNER_IMAGE_HEIGHT"))
	if err != nil {
		return fmt.Errorf("invalid BANNER_IMAGE_HEIGHT")
	}
	if width != bannerWidth || height != bannerHeight {
		return fmt.Errorf(
			"image dimension must be %dpx x %dpx",
			bannerWidth,
			bannerHeight,
		)
	}
	form, _ := ctx.MultipartForm()
	contentFiles := form.File["content_images"]
	updatedBody := req.BannerBody
	// === SETUP BASE PATH ===
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	bannerPath := filepath.Join(basePath, os.Getenv("BANNER_IMAGE_PATH"))
	if _, err := os.Stat(bannerPath); os.IsNotExist(err) {
		_ = os.MkdirAll(bannerPath, 0755)
	}

	if len(contentFiles) > 0 {
		re := regexp.MustCompile(`<img[^>]+src="blob:[^"]+"[^>]*>`)
		matches := re.FindAllString(updatedBody, -1)

		for i, match := range matches {
			if i < len(contentFiles) {
				cFile := contentFiles[i]

				// === FORMAT FILENAME ===
				now := time.Now()
				filename := fmt.Sprintf(
					"%s_%d-%02d-%02d_%02d-%02d-%02d_%s",
					BannerImgAlias,
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
					cFile.Filename,
				)

				fullPath := filepath.Join(bannerPath, filename)

				if err := ctx.SaveUploadedFile(cFile, fullPath); err != nil {
					return err
				}

				fileUrl := fmt.Sprintf(
					"%s/api/images/banner/%s",
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
	now := time.Now()
	mainFilename := fmt.Sprintf(
		"%s_%d-%02d-%02d_%02d-%02d-%02d_%s",
		BannerImgAlias,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		fileHeader.Filename,
	)
	if err := ctx.SaveUploadedFile(fileHeader, os.Getenv("BANNER_IMAGE_PATH")+mainFilename); err != nil {
		return err
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	banner := models.BannerList{
		CreatorID:            req.CreatorID,
		CreatorName:          req.CreatorName,
		BannerTitle:          req.BannerTitle,
		BannerBody:           req.BannerBody,
		BannerType:           req.BannerType,
		IsBroadcast:          req.IsBroadcast,
		BannerImage:          "/banner/" + mainFilename,
		BannerImageSize:      strconv.FormatInt(fileHeader.Size, 10),
		BannerImageDimension: fmt.Sprintf("%dpx and %dpx", width, height),
	}
	if err := s.BannerRepo.Create(tx, &banner); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *BannerService) UpdateBanner(ctx *gin.Context, id uint) error {
	existBanner, err := s.BannerRepo.FindById(id)
	if err != nil {
		return err
	}
	if existBanner == nil {
		return errors.New("Banner not found")
	}
	jsonData := ctx.PostForm("json_data")
	var req dtos.BannerRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return err
	}

	adminRepo := &repositories.AdminRepository{DB: s.DB}
	adminData, err := adminRepo.FindAdminByID(req.CreatorID)
	if err != nil || adminData == nil {
		return errors.New("admin not found")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updatedBody := req.BannerBody
	form, _ := ctx.MultipartForm()
	contentFiles := form.File["content_images"]
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	bannerPath := filepath.Join(basePath, os.Getenv("BANNER_IMAGE_PATH"))
	if _, err := os.Stat(bannerPath); os.IsNotExist(err) {
		_ = os.MkdirAll(bannerPath, 0755)
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
					BannerImgAlias,
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
					cFile.Filename,
				)

				fullPath := filepath.Join(bannerPath, filename)

				if err := ctx.SaveUploadedFile(cFile, fullPath); err != nil {
					tx.Rollback()
					return err
				}

				fileUrl := fmt.Sprintf(
					"%s/api/images/banners/%s",
					helpers.GetHostURL(ctx),
					filename,
				)

				srcRe := regexp.MustCompile(`src="[^"]+"`)
				updatedTag := srcRe.ReplaceAllString(match, fmt.Sprintf(`src="%s"`, fileUrl))
				updatedBody = strings.Replace(updatedBody, match, updatedTag, 1)
			}
		}
	}
	fileHeader, err := ctx.FormFile("file")

	var isNewImage bool = false
	var newFilename string
	var newSize string
	var newDim string

	if err == nil {
		isNewImage = true

		// === VALIDASI DIMENSI ===
		width, height, errDim := helpers.GetImageDimension(fileHeader)
		if errDim != nil {
			tx.Rollback()
			return errors.New("failed to decode image config")
		}

		bannerWidth, _ := strconv.Atoi(os.Getenv("BANNER_IMAGE_WIDTH"))
		bannerHeight, _ := strconv.Atoi(os.Getenv("BANNER_IMAGE_HEIGHT"))

		if width != bannerWidth || height != bannerHeight {
			return fmt.Errorf(
				"image dimension must be %dpx x %dpx",
				bannerWidth,
				bannerHeight,
			)
		}

		// === FORMAT FILENAME ===
		now := time.Now()
		newFilename = fmt.Sprintf(
			"%s_%d-%02d-%02d_%02d-%02d-%02d_%s",
			BannerImgAlias,
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second(),
			fileHeader.Filename,
		)

		newFullPath := filepath.Join(bannerPath, newFilename)

		if err := ctx.SaveUploadedFile(fileHeader, newFullPath); err != nil {
			tx.Rollback()
			return err
		}

		// === HAPUS IMAGE LAMA ===
		oldFilename := strings.TrimPrefix(existBanner.BannerImage, "/banner/")
		oldPath := filepath.Join(bannerPath, oldFilename)
		_ = os.Remove(oldPath)

		newSize = strconv.FormatInt(fileHeader.Size, 10)
		newDim = fmt.Sprintf("%dpx and %dpx", width, height)
	}

	// 6. Update Fields
	existBanner.CreatorID = req.CreatorID
	existBanner.CreatorName = req.CreatorName
	existBanner.BannerTitle = req.BannerTitle
	existBanner.BannerBody = updatedBody // Body hasil proses regex
	existBanner.BannerType = req.BannerType
	existBanner.IsBroadcast = req.IsBroadcast
	existBanner.UpdatedAt = time.Now()

	// Jika gambar diganti, update field gambar & flag revisi
	if isNewImage {
		existBanner.BannerImage = "/banner/" + newFilename
		existBanner.BannerImageSize = newSize
		existBanner.BannerImageDimension = newDim
		existBanner.IsRevision = "1"
	}
	err = s.BannerRepo.Update(tx, existBanner)
	if err != nil {
		tx.Rollback()
		if isNewImage {
			os.Remove(filepath.Join(os.Getenv("BANNER_IMAGE_PATH"), newFilename))
		}
		return err
	}
	return tx.Commit().Error
}
func (s *BannerService) DeleteBanner(id uint) error {
	banner, err := s.BannerRepo.FindById(id)
	if err != nil {
		return err
	}
	if banner == nil {
		return errors.New("Banner not found")
	}
	tx := s.BannerRepo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	re := regexp.MustCompile(`<img[^>]+src="([^">]+)"`)
	matches := re.FindAllStringSubmatch(banner.BannerBody, -1)

	for _, match := range matches {
		if len(match) > 1 {
			// match[1] berisi URL, misal: http://localhost:8080/api/images/banner/BNR_IMG_...jpg
			url := match[1]

			// Kita harus ekstrak nama filenya saja
			if strings.Contains(url, "/api/images/banner/") {
				parts := strings.Split(url, "/api/images/banner/")
				if len(parts) > 1 {
					filename := parts[1]
					filePath := filepath.Join(os.Getenv("BANNER_IMAGE_PATH"), filename)

					// Hapus File
					os.Remove(filePath)
				}
			}
		}
	}
	if banner.BannerImage != "" {
		filename := strings.TrimPrefix(banner.BannerImage, "/banner/")
		mainImgPath := filepath.Join(os.Getenv("BANNER_IMAGE_PATH"), filename)
		os.Remove(mainImgPath)
	}
	err = s.BannerRepo.Delete(tx, banner)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
