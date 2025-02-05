package utils

import (
	"path/filepath"

	"github.com/google/uuid"
)

// GenerateUniqueFilename generates a unique filename using a UUID and preserves the original file extension.
func GenerateUniqueFilename(original string) string {
	ext := filepath.Ext(original)
	return uuid.New().String() + ext
}
