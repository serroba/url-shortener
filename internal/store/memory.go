package store

import (
	"context"
	"sync"

	"github.com/serroba/web-demo-go/internal/domain"
)

// MemoryStore is an in-memory implementation of ShortURLRepository.
type MemoryStore struct {
	mu     sync.RWMutex
	urls   map[domain.Code]*domain.ShortURL // code -> entity
	hashes map[domain.URLHash]domain.Code   // urlHash -> code (index for hash lookups)
}

// NewMemoryStore creates a new in-memory URL store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urls:   make(map[domain.Code]*domain.ShortURL),
		hashes: make(map[domain.URLHash]domain.Code),
	}
}

func (m *MemoryStore) Save(_ context.Context, shortURL *domain.ShortURL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortURL.Code] = shortURL

	// Index by hash if present (for hash strategy)
	if shortURL.URLHash != "" {
		m.hashes[shortURL.URLHash] = shortURL.Code
	}

	return nil
}

func (m *MemoryStore) GetByCode(_ context.Context, code domain.Code) (*domain.ShortURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shortURL, ok := m.urls[code]
	if !ok {
		return nil, domain.ErrNotFound
	}

	return shortURL, nil
}

func (m *MemoryStore) GetByHash(_ context.Context, hash domain.URLHash) (*domain.ShortURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	code, ok := m.hashes[hash]
	if !ok {
		return nil, domain.ErrNotFound
	}

	shortURL, ok := m.urls[code]
	if !ok {
		return nil, domain.ErrNotFound
	}

	return shortURL, nil
}
