package store

import (
	"context"
	"sync"

	"github.com/serroba/web-demo-go/internal/shortener"
)

// MemoryStore is an in-memory implementation of shortener.Repository.
type MemoryStore struct {
	mu     sync.RWMutex
	urls   map[shortener.Code]*shortener.ShortURL // code -> entity
	hashes map[shortener.URLHash]shortener.Code   // urlHash -> code (index for hash lookups)
}

// NewMemoryStore creates a new in-memory URL store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urls:   make(map[shortener.Code]*shortener.ShortURL),
		hashes: make(map[shortener.URLHash]shortener.Code),
	}
}

func (m *MemoryStore) Save(_ context.Context, shortURL *shortener.ShortURL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortURL.Code] = shortURL

	// Index by hash if present (for hash strategy)
	if shortURL.URLHash != "" {
		m.hashes[shortURL.URLHash] = shortURL.Code
	}

	return nil
}

func (m *MemoryStore) GetByCode(_ context.Context, code shortener.Code) (*shortener.ShortURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shortURL, ok := m.urls[code]
	if !ok {
		return nil, shortener.ErrNotFound
	}

	return shortURL, nil
}

func (m *MemoryStore) GetByHash(_ context.Context, hash shortener.URLHash) (*shortener.ShortURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	code, ok := m.hashes[hash]
	if !ok {
		return nil, shortener.ErrNotFound
	}

	shortURL, ok := m.urls[code]
	if !ok {
		return nil, shortener.ErrNotFound
	}

	return shortURL, nil
}
