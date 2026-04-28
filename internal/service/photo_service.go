package service

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PhotoService struct {
	basePath       string
	quality        int
	maxSizeBytes   int64
	retentionDays  int
	allowedFormats []string
}

func NewPhotoService(basePath string) *PhotoService {
	return &PhotoService{
		basePath:       basePath,
		quality:        85,
		maxSizeBytes:   5 * 1024 * 1024, // 5MB
		retentionDays:  365,              // 1 year
		allowedFormats: []string{"jpeg", "jpg", "png"},
	}
}

// SavePhoto saves a photo from base64 data and returns the file path
func (s *PhotoService) SavePhoto(userID int64, photoType string, base64Data string) (string, error) {
	// Validate photo type
	if photoType != "masuk" && photoType != "pulang" && photoType != "enrollment" {
		return "", fmt.Errorf("invalid photo type: %s", photoType)
	}

	// Remove data URL prefix if present (e.g., "data:image/jpeg;base64,")
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	// Decode base64
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Validate size
	if int64(len(imageData)) > s.maxSizeBytes {
		return "", fmt.Errorf("photo size exceeds maximum allowed size")
	}

	// Decode image to validate format
	img, format, err := image.Decode(strings.NewReader(string(imageData)))
	if err != nil {
		return "", fmt.Errorf("invalid image format: %w", err)
	}

	// Validate format
	if !s.isFormatAllowed(format) {
		return "", fmt.Errorf("unsupported image format: %s", format)
	}

	// Generate file path
	now := time.Now()
	dirPath := filepath.Join(s.basePath, fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()))
	
	// Create directory if not exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate filename with hash to prevent duplicates
	hash := sha256.Sum256(imageData)
	filename := fmt.Sprintf("%d_%s_%s_%x.jpg", userID, photoType, now.Format("150405"), hash[:8])
	filePath := filepath.Join(dirPath, filename)

	// Save as JPEG with compression
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Encode as JPEG with quality setting
	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: s.quality}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// Return relative path from base
	relativePath := strings.TrimPrefix(filePath, s.basePath)
	relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
	
	return relativePath, nil
}

// GetPhotoPath returns the full path to a photo
func (s *PhotoService) GetPhotoPath(relativePath string) string {
	return filepath.Join(s.basePath, relativePath)
}

// PhotoExists checks if a photo file exists
func (s *PhotoService) PhotoExists(relativePath string) bool {
	fullPath := s.GetPhotoPath(relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// DeletePhoto deletes a photo file
func (s *PhotoService) DeletePhoto(relativePath string) error {
	fullPath := s.GetPhotoPath(relativePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete photo: %w", err)
	}
	return nil
}

// CleanupOldPhotos removes photos older than retention period
func (s *PhotoService) CleanupOldPhotos() (int, error) {
	cutoffDate := time.Now().AddDate(0, 0, -s.retentionDays)
	deletedCount := 0

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is older than retention period
		if info.ModTime().Before(cutoffDate) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to delete old photo %s: %w", path, err)
			}
			deletedCount++
		}

		return nil
	})

	if err != nil {
		return deletedCount, fmt.Errorf("cleanup failed: %w", err)
	}

	return deletedCount, nil
}

// GetStorageStats returns storage statistics
func (s *PhotoService) GetStorageStats() (map[string]interface{}, error) {
	var totalSize int64
	var fileCount int

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to calculate storage stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_files":      fileCount,
		"total_size_bytes": totalSize,
		"total_size_mb":    float64(totalSize) / (1024 * 1024),
		"base_path":        s.basePath,
	}

	return stats, nil
}

// ReadPhoto reads a photo file and returns it as a reader
func (s *PhotoService) ReadPhoto(relativePath string) (io.ReadCloser, error) {
	fullPath := s.GetPhotoPath(relativePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open photo: %w", err)
	}
	return file, nil
}

// isFormatAllowed checks if image format is allowed
func (s *PhotoService) isFormatAllowed(format string) bool {
	format = strings.ToLower(format)
	for _, allowed := range s.allowedFormats {
		if format == allowed {
			return true
		}
	}
	return false
}

// EnsureBasePathExists creates the base photo directory if it doesn't exist
func (s *PhotoService) EnsureBasePathExists() error {
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create base photo directory: %w", err)
	}
	return nil
}

