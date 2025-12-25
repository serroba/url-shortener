package store_test

import (
	"context"
	"testing"

	"github.com/serroba/web-demo-go/internal/shortener"
	"github.com/serroba/web-demo-go/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_Save(t *testing.T) {
	t.Run("saves short url successfully", func(t *testing.T) {
		s := store.NewMemoryStore()

		err := s.Save(context.Background(), &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
		})

		require.NoError(t, err)

		// Verify it can be retrieved
		shortURL, err := s.GetByCode(context.Background(), "abc123")

		require.NoError(t, err)
		assert.Equal(t, "https://example.com", shortURL.OriginalURL)
	})

	t.Run("indexes by hash when present", func(t *testing.T) {
		s := store.NewMemoryStore()

		err := s.Save(context.Background(), &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
			URLHash:     "somehash",
		})

		require.NoError(t, err)

		// Verify it can be retrieved by hash
		shortURL, err := s.GetByHash(context.Background(), "somehash")

		require.NoError(t, err)
		assert.Equal(t, shortener.Code("abc123"), shortURL.Code)
	})

	t.Run("overwrites existing url", func(t *testing.T) {
		s := store.NewMemoryStore()
		_ = s.Save(context.Background(), &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
		})

		err := s.Save(context.Background(), &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://other.com",
		})

		require.NoError(t, err)

		shortURL, _ := s.GetByCode(context.Background(), "abc123")

		assert.Equal(t, "https://other.com", shortURL.OriginalURL)
	})
}

func TestMemoryStore_GetByCode(t *testing.T) {
	t.Run("returns short url when found", func(t *testing.T) {
		s := store.NewMemoryStore()
		_ = s.Save(context.Background(), &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
		})

		shortURL, err := s.GetByCode(context.Background(), "abc123")

		require.NoError(t, err)
		assert.Equal(t, "https://example.com", shortURL.OriginalURL)
	})

	t.Run("returns ErrNotFound when code does not exist", func(t *testing.T) {
		s := store.NewMemoryStore()

		shortURL, err := s.GetByCode(context.Background(), "notfound")

		assert.Nil(t, shortURL)
		assert.ErrorIs(t, err, shortener.ErrNotFound)
	})
}

func TestMemoryStore_GetByHash(t *testing.T) {
	t.Run("returns short url when hash exists", func(t *testing.T) {
		s := store.NewMemoryStore()
		_ = s.Save(context.Background(), &shortener.ShortURL{
			Code:        "abc123",
			OriginalURL: "https://example.com",
			URLHash:     "somehash",
		})

		shortURL, err := s.GetByHash(context.Background(), "somehash")

		require.NoError(t, err)
		assert.Equal(t, shortener.Code("abc123"), shortURL.Code)
		assert.Equal(t, "https://example.com", shortURL.OriginalURL)
	})

	t.Run("returns ErrNotFound when hash does not exist", func(t *testing.T) {
		s := store.NewMemoryStore()

		shortURL, err := s.GetByHash(context.Background(), "nonexistent")

		assert.Nil(t, shortURL)
		assert.ErrorIs(t, err, shortener.ErrNotFound)
	})
}
