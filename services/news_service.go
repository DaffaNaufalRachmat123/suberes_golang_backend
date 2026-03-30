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

type NewsService struct {
	NewsRepo *repositories.NewsRepository
	DB       *gorm.DB
}

const (
	NewsImgAlias = "NEWS_IMG_"
)

func (s *NewsService) GetNews(page, limit int) ([]models.NewsList, int64, error) {
	return s.NewsRepo.FindAll(page, limit)
}

func (s *NewsService) GetNewsByID(id uint) (*models.NewsList, error) {
	return s.NewsRepo.FindById(id)
}

func (s *NewsService) GetPopularNews() ([]models.NewsList, error) {
	return s.NewsRepo.FindPopular(5)
}

func (s *NewsService) CreateNews(ctx *gin.Context) error {
	jsonData := ctx.PostForm("json_data")
	var req dtos.NewsRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return errors.New("Invalid data format")
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return errors.New("News image required")
	}
	width, height, err := helpers.GetImageDimension(fileHeader)
	if err != nil {
		return err
	}
	newsWidth, err := strconv.Atoi(os.Getenv("NEWS_IMAGE_WIDTH"))
	if err != nil {
		return fmt.Errorf("invalid NEWS_IMAGE_WIDTH")
	}

	newsHeight, err := strconv.Atoi(os.Getenv("NEWS_IMAGE_HEIGHT"))
	if err != nil {
		return fmt.Errorf("invalid NEWS_IMAGE_HEIGHT")
	}
	if width != newsWidth || height != newsHeight {
		return fmt.Errorf(
			"image dimension must be %dpx x %dpx",
			newsWidth,
			newsHeight,
		)
	}
	form, _ := ctx.MultipartForm()
	contentFiles := form.File["content_images"]
	updatedBody := req.NewsBody
	// === SETUP BASE PATH ===
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	newsPath := filepath.Join(basePath, os.Getenv("NEWS_IMAGE_PATH"))
	if _, err := os.Stat(newsPath); os.IsNotExist(err) {
		_ = os.MkdirAll(newsPath, 0755)
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
					NewsImgAlias,
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
					cFile.Filename,
				)

				fullPath := filepath.Join(newsPath, filename)

				if err := ctx.SaveUploadedFile(cFile, fullPath); err != nil {
					return err
				}

				fileUrl := fmt.Sprintf(
					"%s/api/images/news/%s",
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
		NewsImgAlias,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		fileHeader.Filename,
	)
	if err := ctx.SaveUploadedFile(fileHeader, os.Getenv("NEWS_IMAGE_PATH")+mainFilename); err != nil {
		return err
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	news := models.NewsList{
		CreatorID:          req.CreatorID,
		CreatorName:        req.CreatorName,
		NewsTitle:          req.NewsTitle,
		NewsBody:           req.NewsBody,
		NewsType:           req.NewsType,
		IsBroadcast:        req.IsBroadcast,
		NewsImage:          "/news/" + mainFilename,
		NewsImageSize:      strconv.FormatInt(fileHeader.Size/(1024*1024), 10),
		NewsImageDimension: fmt.Sprintf("%dpx and %dpx", width, height),
	}
	if err := s.NewsRepo.Create(tx, &news); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *NewsService) UpdateNews(ctx *gin.Context, id uint) error {
	existNews, err := s.NewsRepo.FindById(id)
	if err != nil {
		return err
	}
	if existNews == nil {
		return errors.New("News not found")
	}
	jsonData := ctx.PostForm("json_data")
	var req dtos.NewsRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return err
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updatedBody := req.NewsBody
	form, _ := ctx.MultipartForm()
	contentFiles := form.File["content_images"]
	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	newsPath := filepath.Join(basePath, os.Getenv("NEWS_IMAGE_PATH"))
	if _, err := os.Stat(newsPath); os.IsNotExist(err) {
		_ = os.MkdirAll(newsPath, 0755)
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
					NewsImgAlias,
					now.Year(),
					now.Month(),
					now.Day(),
					now.Hour(),
					now.Minute(),
					now.Second(),
					cFile.Filename,
				)

				fullPath := filepath.Join(newsPath, filename)

				if err := ctx.SaveUploadedFile(cFile, fullPath); err != nil {
					tx.Rollback()
					return err
				}

				fileUrl := fmt.Sprintf(
					"%s/api/images/news/%s",
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

		newsWidth, _ := strconv.Atoi(os.Getenv("NEWS_IMAGE_WIDTH"))
		newsHeight, _ := strconv.Atoi(os.Getenv("NEWS_IMAGE_HEIGHT"))

		if width != newsWidth || height != newsHeight {
			return fmt.Errorf(
				"image dimension must be %dpx x %dpx",
				newsWidth,
				newsHeight,
			)
		}

		// === FORMAT FILENAME ===
		now := time.Now()
		newFilename = fmt.Sprintf(
			"%s_%d-%02d-%02d_%02d-%02d-%02d_%s",
			NewsImgAlias,
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second(),
			fileHeader.Filename,
		)

		newFullPath := filepath.Join(newsPath, newFilename)

		if err := ctx.SaveUploadedFile(fileHeader, newFullPath); err != nil {
			tx.Rollback()
			return err
		}

		// === HAPUS IMAGE LAMA ===
		oldFilename := strings.TrimPrefix(existNews.NewsImage, "/news/")
		oldPath := filepath.Join(newsPath, oldFilename)
		_ = os.Remove(oldPath)

		newSize = strconv.FormatInt(fileHeader.Size/(1024*1024), 10)
		newDim = fmt.Sprintf("%dpx and %dpx", width, height)
	}

	existNews.CreatorID = req.CreatorID
	existNews.CreatorName = req.CreatorName
	existNews.NewsTitle = req.NewsTitle
	existNews.NewsBody = updatedBody 
	existNews.NewsType = req.NewsType
	existNews.IsBroadcast = req.IsBroadcast
	existNews.UpdatedAt = time.Now()

	if isNewImage {
		existNews.NewsImage = "/news/" + newFilename
		existNews.NewsImageSize = newSize
		existNews.NewsImageDimension = newDim
		existNews.IsRevision = "1"
	}
	err = s.NewsRepo.Update(tx, existNews)
	if err != nil {
		tx.Rollback()
		if isNewImage {
			os.Remove(filepath.Join(os.Getenv("NEWS_IMAGE_PATH"), newFilename))
		}
		return err
	}
	return tx.Commit().Error
}
func (s *NewsService) DeleteNews(id uint) error {
	news, err := s.NewsRepo.FindById(id)
	if err != nil {
		return err
	}
	if news == nil {
		return errors.New("News not found")
	}
	tx := s.NewsRepo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	re := regexp.MustCompile(`<img[^>]+src="([^">]+)"`)
	matches := re.FindAllStringSubmatch(news.NewsBody, -1)

	for _, match := range matches {
		if len(match) > 1 {
			url := match[1]
			if strings.Contains(url, "/api/images/news/") {
				parts := strings.Split(url, "/api/images/news/")
				if len(parts) > 1 {
					filename := parts[1]
					filePath := filepath.Join(os.Getenv("NEWS_IMAGE_PATH"), filename)
					os.Remove(filePath)
				}
			}
		}
	}
	if news.NewsImage != "" {
		filename := strings.TrimPrefix(news.NewsImage, "/news/")
		mainImgPath := filepath.Join(os.Getenv("NEWS_IMAGE_PATH"), filename)
		os.Remove(mainImgPath)
	}
	err = s.NewsRepo.Delete(tx, news)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
