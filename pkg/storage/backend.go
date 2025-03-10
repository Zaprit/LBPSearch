package storage

import (
	"context"
	"io"
)

type ResourceStorageBackend interface {
	GetResource(ctx context.Context, path string) (io.ReadCloser, error)
	HasResource(ctx context.Context, path string) (bool, error)
}

type IconCacheBackend interface {
	HasIcon(ctx context.Context, hash string) (bool, error)
	PutIcon(ctx context.Context, hash string, r io.Reader) error
	GetIconURL(ctx context.Context, hash string) (string, error)
}

type LevelCacheBackend interface {
	HasLevel(ctx context.Context, id string) (bool, error)
	PutLevel(ctx context.Context, id string, r io.Reader) error
	GetLevelURL(ctx context.Context, id string) (string, error)
}
