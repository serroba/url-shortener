package shortener

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a short URL is not found.
var ErrNotFound = errors.New("short url not found")

// Repository defines the interface for short URL storage operations.
type Repository interface {
	Save(ctx context.Context, shortURL *ShortURL) error
	GetByCode(ctx context.Context, code Code) (*ShortURL, error)
	GetByHash(ctx context.Context, hash URLHash) (*ShortURL, error)
}
