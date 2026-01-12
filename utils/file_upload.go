package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	UploadDir          = "uploads/images"
	MaxFileSize        = 5 << 20 // 5 MB
	AllowedExtensions  = ".jpg,.jpeg,.png,.webp"
)

// ValidateImageFile - Validasi file gambar
func ValidateImageFile(fileHeader *multipart.FileHeader) error {
	// Cek ukuran file
	if fileHeader.Size > MaxFileSize {
		return fmt.Errorf("ukuran file terlalu besar, maksimal 5MB")
	}

	// Cek extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := strings.Split(AllowedExtensions, ",")
	
	isValid := false
	for _, allowedExt := range allowedExts {
		if ext == strings.TrimSpace(allowedExt) {
			isValid = true
			break
		}
	}
	
	if !isValid {
		return fmt.Errorf("format file tidak didukung, gunakan: jpg, jpeg, png, webp")
	}

	return nil
}

// SaveUploadedFile - Simpan file yang di-upload
func SaveUploadedFile(c *gin.Context, formKey string) (string, error) {
	// Ambil file dari form
	file, err := c.FormFile(formKey)
	if err != nil {
		return "", fmt.Errorf("file tidak ditemukan")
	}

	// Validasi file
	if err := ValidateImageFile(file); err != nil {
		return "", err
	}

	// Buat folder jika belum ada
	if err := os.MkdirAll(UploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("gagal membuat folder upload")
	}

	// Generate nama file unik
	ext := filepath.Ext(file.Filename)
	timestamp := time.Now().UnixNano() // Gunakan Nanosecond untuk lebih unik
	filename := fmt.Sprintf("%d%s", timestamp, ext)
	filepath := filepath.Join(UploadDir, filename)

	// Simpan file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		return "", fmt.Errorf("gagal menyimpan file")
	}

	// Return path relative (untuk disimpan di database)
	return filepath, nil
}
