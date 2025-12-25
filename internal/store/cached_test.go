package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/serroba/web-demo-go/internal/cache"
	"github.com/serroba/web-demo-go/internal/shortener"
	"github.com/serroba/web-demo-go/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	saveFunc      func(ctx context.Context, shortURL *shortener.ShortURL) error
	getByCodeFunc func(ctx context.Context, code shortener.Code) (*shortener.ShortURL, error)
	getByHashFunc func(ctx context.Context, hash shortener.URLHash) (*shortener.ShortURL, error)
	callCount     int
}

func (m *mockStore) Save(ctx context.Context, shortURL *shortener.ShortURL) error {
	m.callCount++

	if m.saveFunc != nil {
		return m.saveFunc(ctx, shortURL)
	}

	return nil
}

func (m *mockStore) GetByCode(ctx context.Context, code shortener.Code) (*shortener.ShortURL, error) {
	m.callCount++

	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, code)
	}

	return nil, shortener.ErrNotFound
}

func (m *mockStore) GetByHash(ctx context.Context, hash shortener.URLHash) (*shortener.ShortURL, error) {
	m.callCount++

	if m.getByHashFunc != nil {
		return m.getByHashFunc(ctx, hash)
	}

	return nil, shortener.ErrNotFound
}

func TestCachedRepository_GetByCode(t *testing.T) {
	t.Run("cache miss fetches from store and caches", func(t *testing.T) {
		url := &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
		}
		mock := &mockStore{
			getByCodeFunc: func(_ context.Context, _ shortener.Code) (*shortener.ShortURL, error) {
				return url, nil
			},
		}
		lru := cache.New(10)
		cached := store.NewCachedRepository(mock, lru)

		// First call - cache miss
		result, err := cached.GetByCode(context.Background(), "abc123")

		require.NoError(t, err)
		assert.Equal(t, url, result)
		assert.Equal(t, 1, mock.callCount, "store should be called on cache miss")

		// Second call - cache hit
		result, err = cached.GetByCode(context.Background(), "abc123")

		require.NoError(t, err)
		assert.Equal(t, url, result)
		assert.Equal(t, 1, mock.callCount, "store should NOT be called on cache hit")
	})

	t.Run("cache miss with error does not cache", func(t *testing.T) {
		storeErr := errors.New("store error")
		mock := &mockStore{
			getByCodeFunc: func(_ context.Context, _ shortener.Code) (*shortener.ShortURL, error) {
				return nil, storeErr
			},
		}
		lru := cache.New(10)
		cached := store.NewCachedRepository(mock, lru)

		_, err := cached.GetByCode(context.Background(), "abc123")

		require.ErrorIs(t, err, storeErr)
		assert.Equal(t, 0, lru.Len(), "error should not be cached")
	})

	t.Run("ErrNotFound is not cached", func(t *testing.T) {
		callCount := 0
		mock := &mockStore{
			getByCodeFunc: func(_ context.Context, _ shortener.Code) (*shortener.ShortURL, error) {
				callCount++

				return nil, shortener.ErrNotFound
			},
		}
		lru := cache.New(10)
		cached := store.NewCachedRepository(mock, lru)

		// First call
		_, err := cached.GetByCode(context.Background(), "missing")
		require.ErrorIs(t, err, shortener.ErrNotFound)

		// Second call - should still hit store (not cached)
		_, err = cached.GetByCode(context.Background(), "missing")
		require.ErrorIs(t, err, shortener.ErrNotFound)

		assert.Equal(t, 2, callCount, "store should be called each time for not found")
	})
}

func TestCachedRepository_Save(t *testing.T) {
	t.Run("save updates cache", func(t *testing.T) {
		url := &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
		}
		mock := &mockStore{}
		lru := cache.New(10)
		cached := store.NewCachedRepository(mock, lru)

		err := cached.Save(context.Background(), url)

		require.NoError(t, err)
		assert.Equal(t, 1, lru.Len(), "cache should have the saved item")

		// GetByCode should return from cache without hitting store
		mock.callCount = 0
		result, err := cached.GetByCode(context.Background(), "abc123")

		require.NoError(t, err)
		assert.Equal(t, url, result)
		assert.Equal(t, 0, mock.callCount, "store should not be called (cache hit)")
	})

	t.Run("save error does not update cache", func(t *testing.T) {
		saveErr := errors.New("save failed")
		mock := &mockStore{
			saveFunc: func(_ context.Context, _ *shortener.ShortURL) error {
				return saveErr
			},
		}
		lru := cache.New(10)
		cached := store.NewCachedRepository(mock, lru)

		url := &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
		}
		err := cached.Save(context.Background(), url)

		require.ErrorIs(t, err, saveErr)
		assert.Equal(t, 0, lru.Len(), "cache should not be updated on error")
	})
}

func TestCachedRepository_GetByHash(t *testing.T) {
	t.Run("passes through to store without caching", func(t *testing.T) {
		url := &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
			URLHash:     "hash123",
		}
		mock := &mockStore{
			getByHashFunc: func(_ context.Context, _ shortener.URLHash) (*shortener.ShortURL, error) {
				return url, nil
			},
		}
		lru := cache.New(10)
		cached := store.NewCachedRepository(mock, lru)

		// First call
		result, err := cached.GetByHash(context.Background(), "hash123")

		require.NoError(t, err)
		assert.Equal(t, url, result)
		assert.Equal(t, 1, mock.callCount)

		// Second call - should still hit store (not cached)
		result, err = cached.GetByHash(context.Background(), "hash123")

		require.NoError(t, err)
		assert.Equal(t, url, result)
		assert.Equal(t, 2, mock.callCount, "store should be called each time (no caching)")
	})
}
