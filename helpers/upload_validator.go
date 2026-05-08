package helpers

import (
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const MaxUploadSize = 10 << 20 // 10MB

var AllowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

var AllowedImageMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func ValidateUploadedFile(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > MaxUploadSize {
		return errors.New("file size exceeds 10MB limit")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !AllowedImageExtensions[ext] {
		return errors.New("file extension not allowed, only jpg/jpeg/png/gif/webp")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errors.New("unable to open file")
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return errors.New("unable to read file content")
	}

	mimeType := http.DetectContentType(buffer[:n])
	if !AllowedImageMIME[mimeType] {
		return errors.New("file content type not allowed, only image files accepted")
	}

	return nil
}

func SanitizeFilename(filename string) string {
	base := filepath.Base(filename)

	base = strings.ReplaceAll(base, "..", "")
	base = strings.ReplaceAll(base, "/", "")
	base = strings.ReplaceAll(base, "\\", "")

	if base == "" || base == "." {
		base = "upload"
	}

	return base
}
