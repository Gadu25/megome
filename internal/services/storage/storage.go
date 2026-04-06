package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type R2Client struct {
	Client *minio.Client
	Bucket string
}

type Config struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Bucket    string
	Secure    bool
}

// NewR2Client initializes the client
func NewR2Client(cfg Config) (*R2Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &R2Client{
		Client: minioClient,
		Bucket: cfg.Bucket,
	}, nil
}

// GenerateKey creates a safe object key
func GenerateKey(prefix, filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}
	return fmt.Sprintf("%s/%s", prefix, filename), nil
}

// UploadFromReader uploads a file directly (server-side)
func (r *R2Client) UploadFromReader(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := r.Client.PutObject(ctx, r.Bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", key, err)
	}
	return nil
}

// DeleteObject deletes a single file by key
func (r *R2Client) DeleteObject(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	err := r.Client.RemoveObject(ctx, r.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}

	return nil
}
