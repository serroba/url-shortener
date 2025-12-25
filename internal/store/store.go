package store

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("url not found")

// URLStore defines the interface for URL storage operations.
type URLStore interface {
	Save(ctx context.Context, code, url string) error
	Get(ctx context.Context, code string) (string, error)
}
