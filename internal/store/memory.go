package store

import (
	"context"
	"sync"
)

// MemoryStore is an in-memory implementation of URLStore.
type MemoryStore struct {
	mu   sync.RWMutex
	urls map[string]string
}

// NewMemoryStore creates a new in-memory URL store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urls: make(map[string]string),
	}
}

func (m *MemoryStore) Save(_ context.Context, code, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[code] = url

	return nil
}

func (m *MemoryStore) Get(_ context.Context, code string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, ok := m.urls[code]
	if !ok {
		return "", ErrNotFound
	}

	return url, nil
}
