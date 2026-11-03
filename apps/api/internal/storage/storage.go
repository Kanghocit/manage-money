package storage

import (
	"fmt"
	"io"

	"github.com/manage-money/api/internal/config"
)

type BlobStore interface {
	Put(key string, r io.Reader, size int64, contentType string) error
	Get(key string) (io.ReadCloser, string, error)
}

func New(cfg config.Config) (BlobStore, error) {
	switch cfg.StorageDriver {
	case "s3":
		return NewS3(cfg)
	default:
		return NewLocal(cfg.LocalStoragePath)
	}
}

func UserObjectKey(userID, filename string) string {
	return fmt.Sprintf("users/%s/%s", userID, filename)
}
