package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"trbb/internal/config"
)

// ImageFolder — 圖片子目錄常數
const (
	FolderEvents   = "events"
	FolderProducts = "products"
	FolderAvatars  = "avatars"
	FolderGeneral  = "general"
	FolderTraining = "training"
)

type Storage struct {
	client *minio.Client
	cfg    config.MinioConfig
}

func New(cfg config.MinioConfig) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio: %w", err)
	}

	s := &Storage{client: client, cfg: cfg}

	// Ensure all required buckets exist with correct policy
	for _, bucket := range []string{cfg.BucketPublic, cfg.BucketPrivate, cfg.BucketImages} {
		if err := s.ensureBucket(context.Background(), bucket); err != nil {
			return nil, err
		}
	}

	// Set images bucket to public read
	if err := s.setBucketPublicRead(context.Background(), cfg.BucketImages); err != nil {
		// Non-fatal: log but continue (may already be set)
		fmt.Printf("[storage] warn: set public policy on %s: %v\n", cfg.BucketImages, err)
	}

	return s, nil
}

func (s *Storage) ensureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket %s: %w", bucket, err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("make bucket %s: %w", bucket, err)
		}
	}
	return nil
}

// setBucketPublicRead sets a bucket to allow public anonymous GET
func (s *Storage) setBucketPublicRead(ctx context.Context, bucket string) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": ["*"]},
			"Action": ["s3:GetObject"],
			"Resource": ["arn:aws:s3:::%s/*"]
		}]
	}`, bucket)
	return s.client.SetBucketPolicy(ctx, bucket, policy)
}

// UploadImage uploads an image to the images bucket under the given folder.
// Returns the object path (folder/filename) — NOT the full URL.
// The caller combines this with MINIO_EXTERNAL_URL to form the full URL.
func (s *Storage) UploadImage(ctx context.Context, folder, filename, contentType string, reader io.Reader, size int64) (string, error) {
	objectName := path.Join(folder, filename)
	_, err := s.client.PutObject(ctx, s.cfg.BucketImages, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}
	return objectName, nil
}

// ImageURL returns the full public URL for an image object path.
// e.g. "events/abc123.jpg" → "https://images.trbbtw.com/images/events/abc123.jpg"
func (s *Storage) ImageURL(objectPath string) string {
	return fmt.Sprintf("%s/%s/%s", s.cfg.ExternalURL, s.cfg.BucketImages, objectPath)
}

// Upload puts an object into the specified bucket and returns the public URL.
func (s *Storage) Upload(ctx context.Context, bucket, objectName, contentType string, reader io.Reader, size int64) (string, error) {
	_, err := s.client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload object: %w", err)
	}
	return fmt.Sprintf("%s/%s/%s", s.cfg.ExternalURL, bucket, objectName), nil
}

// Delete removes an object from a bucket.
func (s *Storage) Delete(ctx context.Context, bucket, objectName string) error {
	return s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

// PresignedURL generates a temporary presigned GET URL for private objects.
func (s *Storage) PresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, bucket, objectName, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}
	return u.String(), nil
}

func (s *Storage) PublicBucket() string  { return s.cfg.BucketPublic }
func (s *Storage) PrivateBucket() string { return s.cfg.BucketPrivate }
func (s *Storage) ImagesBucket() string  { return s.cfg.BucketImages }
func (s *Storage) ExternalURL() string   { return s.cfg.ExternalURL }
