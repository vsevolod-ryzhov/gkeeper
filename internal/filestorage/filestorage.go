package filestorage

import "context"

//go:generate mockery

// FileStorage defines the interface for binary file storage operations.
type FileStorage interface {
	Upload(ctx context.Context, objectKey string, data []byte) error
	Download(ctx context.Context, objectKey string) ([]byte, error)
	Delete(ctx context.Context, objectKey string) error
}
