package backends

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// StorageBackend defines the interface all storage backends must implement.
// This allows swapping MinIO for S3, GCS, Azure Blob, or local disk 
// without changing the rest of the storage service.
type StorageBackend interface {
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, int64, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	CreateBucket(ctx context.Context, bucket string) error
	DeleteBucket(ctx context.Context, bucket string) error
	BucketExists(ctx context.Context, bucket string) (bool, error)
	PresignedGetURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
	PresignedPutURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
}

// MinIOBackend implements StorageBackend using MinIO
type MinIOBackend struct {
	client *minio.Client
	log    *zap.Logger
}

// NewMinIOBackend creates a MinIO storage backend
func NewMinIOBackend(endpoint, accessKey, secretKey string, useSSL bool, log *zap.Logger) (*MinIOBackend, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	log.Info("MinIO backend initialized", zap.String("endpoint", endpoint))
	return &MinIOBackend{client: client, log: log}, nil
}

func (m *MinIOBackend) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (m *MinIOBackend) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, int64, error) {
	obj, err := m.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, err
	}
	stat, err := obj.Stat()
	if err != nil {
		return nil, 0, err
	}
	return obj, stat.Size, nil
}

func (m *MinIOBackend) DeleteObject(ctx context.Context, bucket, key string) error {
	return m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (m *MinIOBackend) CreateBucket(ctx context.Context, bucket string) error {
	return m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}

func (m *MinIOBackend) DeleteBucket(ctx context.Context, bucket string) error {
	return m.client.RemoveBucket(ctx, bucket)
}

func (m *MinIOBackend) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return m.client.BucketExists(ctx, bucket)
}

func (m *MinIOBackend) PresignedGetURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (m *MinIOBackend) PresignedPutURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedPutObject(ctx, bucket, key, expiry)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
