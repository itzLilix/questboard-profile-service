package infrastructure

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type localImageUploader struct {
	uploadDir     string
	publicBaseURL string
	maxSize       int64
}

func NewLocalImageUploader(uploadDir, publicBaseURL string, maxSize int64) *localImageUploader {
	return &localImageUploader{uploadDir: uploadDir, publicBaseURL: publicBaseURL, maxSize: maxSize}
}

var allowedExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

func (u *localImageUploader) Upload(file *multipart.FileHeader, subdir string) (string, error) {
	if file.Size > u.maxSize {
		return "", ErrFileTooLarge
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return "", ErrInvalidFileType
	}

	dstDir := filepath.Join(u.uploadDir, subdir)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	filename := uuid.New().String() + ext
	dstPath := filepath.Join(dstDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	return fmt.Sprintf("%s/uploads/%s/%s", u.publicBaseURL, subdir, filename), nil
}
