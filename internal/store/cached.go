package store

import (
	"context"

	"github.com/serroba/web-demo-go/internal/shortener"
)

// Cache defines the interface for URL caching.
// This allows swapping cache implementations (LRU, SLRU, Clock, FIFO).
type Cache interface {
	Get(key string) (*shortener.ShortURL, bool)
	Set(key string, value *shortener.ShortURL)
	Len() int
}

// CachedRepository wraps a Repository with a cache for GetByCode lookups.
type CachedRepository struct {
	store shortener.Repository
	cache Cache
}

// NewCachedRepository creates a new cached repository decorator.
func NewCachedRepository(store shortener.Repository, c Cache) *CachedRepository {
	return &CachedRepository{
		store: store,
		cache: c,
	}
}

// Save stores a short URL and updates the cache.
func (c *CachedRepository) Save(ctx context.Context, shortURL *shortener.ShortURL) error {
	if err := c.store.Save(ctx, shortURL); err != nil {
		return err
	}

	// Write-through: update cache after successful save
	c.cache.Set(string(shortURL.Code), shortURL)

	return nil
}

// GetByCode retrieves a short URL by its code, using cache-aside pattern.
func (c *CachedRepository) GetByCode(ctx context.Context, code shortener.Code) (*shortener.ShortURL, error) {
	// Check cache first
	if url, ok := c.cache.Get(string(code)); ok {
		return url, nil
	}

	// Cache miss - fetch from store
	url, err := c.store.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Populate cache
	c.cache.Set(string(code), url)

	return url, nil
}

// GetByHash retrieves a short URL by its hash (pass-through, not cached).
func (c *CachedRepository) GetByHash(ctx context.Context, hash shortener.URLHash) (*shortener.ShortURL, error) {
	return c.store.GetByHash(ctx, hash)
}
