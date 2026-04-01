package storage

import (
	"fmt"
	"path/filepath"
	"strings"
)

// GenerateKey creates a safe key from a file name and optional prefix.
func GenerateKey(prefix, filename string) (string, error) {
	clean := filepath.Clean(filename)
	if clean == "." || clean == ".." || strings.Contains(clean, "..") {
		return "", fmt.Errorf("invalid filename: %s", filename)
	}

	key := strings.TrimPrefix(clean, "/")
	if prefix != "" {
		key = fmt.Sprintf("%s/%s", strings.TrimSuffix(prefix, "/"), key)
	}

	return key, nil
}
