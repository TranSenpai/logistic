package storage

import (
	"context"
	"mime/multipart"
)

type FileStorage interface {
	Upload(ctx context.Context, file *multipart.FileHeader, folder string, prefix string) (string, string, string, error)
	Delete(ctx context.Context, publicID string) error
}
