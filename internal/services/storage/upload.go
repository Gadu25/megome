package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Upload uploads content to the specified key in the bucket.
func (r *R2Client) Upload(ctx context.Context, key string, content []byte, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &r.bucket,
		Key:         &key,
		Body:        bytes.NewReader(content),
		ContentType: &contentType,
		ACL:         types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", key, err)
	}
	return nil
}

// UploadFromReader uploads content from an io.Reader.
func (r *R2Client) UploadFromReader(ctx context.Context, key string, reader io.Reader, contentType string) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return fmt.Errorf("failed to read from reader: %w", err)
	}
	return r.Upload(ctx, key, buf.Bytes(), contentType)
}
