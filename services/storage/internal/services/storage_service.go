package services

import (
	"context"
	"errors"
	"fmt"
	"io"


	"github.com/jackc/pgx/v5"
	"github.com/machinelearningprodigy/OmniBase/storage/internal/backends"
	"go.uber.org/zap"
)

// Interfaces for DB interactions (simplified for Phase 1)

type StorageService struct {
	db      *pgx.Conn
	backend backends.StorageBackend
	log     *zap.Logger
}

func NewStorageService(db *pgx.Conn, backend backends.StorageBackend, log *zap.Logger) *StorageService {
	return &StorageService{
		db:      db,
		backend: backend,
		log:     log,
	}
}

// CheckPermission checks if the user has access via RLS (simulated)
// Full RLS is enforced via PostgREST/Postgres, but the storage Go service needs
// to check permissions against Postgres directly often doing a "SET autho.uid" query first.
func (s *StorageService) checkAccess(ctx context.Context, userID, sql string, args ...any) (bool, error) {
	// For Phase 1 we use basic backend-enforced checks instead of full Postgres RLS emulation in the Go app,
	// but the architecture is here
	return true, nil
}

type Bucket struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

func (s *StorageService) ListBuckets(ctx context.Context, userID string) ([]Bucket, error) {
	rows, err := s.db.Query(ctx, "SELECT id, name, public FROM storage.buckets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []Bucket
	for rows.Next() {
		var b Bucket
		if err := rows.Scan(&b.ID, &b.Name, &b.Public); err != nil {
			return nil, err
		}
		buckets = append(buckets, b)
	}
	return buckets, nil
}

func (s *StorageService) CreateBucket(ctx context.Context, userID, id, name string, public bool) error {
	// 1. Create in metadata DB
	_, err := s.db.Exec(ctx,
		"INSERT INTO storage.buckets (id, name, owner, public) VALUES ($1, $2, $3, $4)",
		id, name, userID, public,
	)
	if err != nil {
		return fmt.Errorf("metadata error: %w", err)
	}

	// 2. Create in backend (e.g., MinIO)
	if err := s.backend.CreateBucket(ctx, id); err != nil {
		// Rollback DB
		s.db.Exec(ctx, "DELETE FROM storage.buckets WHERE id = $1", id)
		return fmt.Errorf("backend error: %w", err)
	}
	return nil
}

func (s *StorageService) DeleteBucket(ctx context.Context, userID, id string) error {
	// 1. Delete from backend
	if err := s.backend.DeleteBucket(ctx, id); err != nil {
		s.log.Warn("failed to delete backend bucket", zap.Error(err), zap.String("bucket", id))
	}
	// 2. Delete metadata
	_, err := s.db.Exec(ctx, "DELETE FROM storage.buckets WHERE id = $1", id)
	return err
}

func (s *StorageService) UploadObject(ctx context.Context, userID, bucketID, path string, reader io.Reader, size int64, contentType string) error {
	// Upsert object metadata
	_, err := s.db.Exec(ctx, `
		INSERT INTO storage.objects (bucket_id, name, owner, content_type, size)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (bucket_id, name)
		DO UPDATE SET size = EXCLUDED.size, content_type = EXCLUDED.content_type, updated_at = NOW()
	`, bucketID, path, userID, contentType, size)
	if err != nil {
		return fmt.Errorf("metadata error: %w", err)
	}

	// Upload to backend
	objKey := fmt.Sprintf("%s/%s", bucketID, path)
	if err := s.backend.PutObject(ctx, bucketID, objKey, reader, size, contentType); err != nil {
		return fmt.Errorf("backend error: %w", err)
	}

	return nil
}

func (s *StorageService) DeleteObjects(ctx context.Context, userID, bucketID string, paths []string) error {
	for _, path := range paths {
		objKey := fmt.Sprintf("%s/%s", bucketID, path)
		if err := s.backend.DeleteObject(ctx, bucketID, objKey); err != nil {
			s.log.Warn("backend delete failed", zap.Error(err))
		}
		s.db.Exec(ctx, "DELETE FROM storage.objects WHERE bucket_id = $1 AND name = $2", bucketID, path)
	}
	return nil
}

func (s *StorageService) GetPublicObject(ctx context.Context, bucketID, path string) (io.ReadCloser, int64, string, error) {
	// Check if bucket is public
	var public bool
	err := s.db.QueryRow(ctx, "SELECT public FROM storage.buckets WHERE id = $1", bucketID).Scan(&public)
	if err != nil || !public {
		return nil, 0, "", errors.New("bucket not found or not public")
	}

	var contentType string
	s.db.QueryRow(ctx, "SELECT content_type FROM storage.objects WHERE bucket_id = $1 AND name = $2", bucketID, path).Scan(&contentType)

	objKey := fmt.Sprintf("%s/%s", bucketID, path)
	reader, size, err := s.backend.GetObject(ctx, bucketID, objKey)
	return reader, size, contentType, err
}

func (s *StorageService) CreateSignedURL(ctx context.Context, bucketID, path string, expiresIn int) (string, error) {
	var count int
	err := s.db.QueryRow(ctx, "SELECT count(*) FROM storage.objects WHERE bucket_id = $1 AND name = $2", bucketID, path).Scan(&count)
	if err != nil {
		return "", fmt.Errorf("metadata error: %w", err)
	}
	if count == 0 {
		return "", errors.New("object not found")
	}
	// The actual signed URL token generation is handled by the handler
	return "", nil 
}

func (s *StorageService) GetObjectDirect(ctx context.Context, bucketID, path string) (io.ReadCloser, int64, string, error) {
	var contentType string
	s.db.QueryRow(ctx, "SELECT content_type FROM storage.objects WHERE bucket_id = $1 AND name = $2", bucketID, path).Scan(&contentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objKey := fmt.Sprintf("%s/%s", bucketID, path)
	reader, size, err := s.backend.GetObject(ctx, bucketID, objKey)
	return reader, size, contentType, err
}
