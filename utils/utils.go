package utils

import (
	"path/filepath"
	"strings"
)

// IsImage checks if the given key represents an image file
func IsImage(key string) bool {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(key))
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".avif"}
	for _, imgExt := range imageExtensions {
		if ext == imgExt {
			return true
		}
	}

	return false
}
