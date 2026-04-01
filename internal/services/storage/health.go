package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// HealthCheck verifies if the bucket is accessible.
func (r *R2Client) HealthCheck(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &r.bucket,
	})
	if err != nil {
		return fmt.Errorf("bucket health check failed: %w", err)
	}
	return nil
}
