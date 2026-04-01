package storage

import (
	"errors"
	"strings"
)

const MaxObjectSize = 5 * 1024 * 1024 * 1024 // 5GB for example

// ValidateKey ensures the key is safe and non-empty.
func ValidateKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("object key cannot be empty")
	}
	if strings.Contains(key, "..") {
		return errors.New("object key cannot contain '..'")
	}
	return nil
}

// ValidateContentSize ensures the content is within limits.
func ValidateContentSize(size int64) error {
	if size < 0 || size > MaxObjectSize {
		return errors.New("content size out of allowed range")
	}
	return nil
}
