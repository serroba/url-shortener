package shortener_test

import (
	"context"
	"errors"
	"testing"

	"github.com/serroba/web-demo-go/internal/shortener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	saveFunc      func(ctx context.Context, shortURL *shortener.ShortURL) error
	getByCodeFunc func(ctx context.Context, code shortener.Code) (*shortener.ShortURL, error)
	getByHashFunc func(ctx context.Context, hash shortener.URLHash) (*shortener.ShortURL, error)
}

func (m *mockRepository) Save(ctx context.Context, shortURL *shortener.ShortURL) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, shortURL)
	}
	return nil
}

func (m *mockRepository) GetByCode(ctx context.Context, code shortener.Code) (*shortener.ShortURL, error) {
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, code)
	}
	return nil, shortener.ErrNotFound
}

func (m *mockRepository) GetByHash(ctx context.Context, hash shortener.URLHash) (*shortener.ShortURL, error) {
	if m.getByHashFunc != nil {
		return m.getByHashFunc(ctx, hash)
	}
	return nil, shortener.ErrNotFound
}

func TestTokenStrategy_Shorten(t *testing.T) {
	t.Run("generates new code and saves", func(t *testing.T) {
		var savedURL *shortener.ShortURL
		repo := &mockRepository{
			saveFunc: func(_ context.Context, s *shortener.ShortURL) error {
				savedURL = s
				return nil
			},
		}
		generator := func() string { return "abc123" }

		strategy := shortener.NewTokenStrategy(repo, generator)
		result, err := strategy.Shorten(context.Background(), "https://example.com")

		require.NoError(t, err)
		assert.Equal(t, shortener.Code("abc123"), result.Code)
		assert.Equal(t, "https://example.com", result.OriginalURL)
		assert.Empty(t, result.URLHash)
		assert.Equal(t, savedURL, result)
	})

	t.Run("returns error when save fails", func(t *testing.T) {
		saveErr := errors.New("save failed")
		repo := &mockRepository{
			saveFunc: func(_ context.Context, _ *shortener.ShortURL) error {
				return saveErr
			},
		}
		generator := func() string { return "abc123" }

		strategy := shortener.NewTokenStrategy(repo, generator)
		result, err := strategy.Shorten(context.Background(), "https://example.com")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, saveErr)
	})
}

func TestHashStrategy_Shorten(t *testing.T) {
	t.Run("returns existing short URL when hash exists", func(t *testing.T) {
		existing := &shortener.ShortURL{
			Code:        "existing",
			OriginalURL: "https://example.com",
			URLHash:     "somehash",
		}
		repo := &mockRepository{
			getByHashFunc: func(_ context.Context, _ shortener.URLHash) (*shortener.ShortURL, error) {
				return existing, nil
			},
		}
		generator := func() string { return "newcode" }

		strategy := shortener.NewHashStrategy(repo, generator)
		result, err := strategy.Shorten(context.Background(), "https://example.com")

		require.NoError(t, err)
		assert.Equal(t, existing, result)
	})

	t.Run("creates new short URL when hash not found", func(t *testing.T) {
		var savedURL *shortener.ShortURL
		repo := &mockRepository{
			getByHashFunc: func(_ context.Context, _ shortener.URLHash) (*shortener.ShortURL, error) {
				return nil, shortener.ErrNotFound
			},
			saveFunc: func(_ context.Context, s *shortener.ShortURL) error {
				savedURL = s
				return nil
			},
		}
		generator := func() string { return "newcode" }

		strategy := shortener.NewHashStrategy(repo, generator)
		result, err := strategy.Shorten(context.Background(), "https://example.com")

		require.NoError(t, err)
		assert.Equal(t, shortener.Code("newcode"), result.Code)
		assert.Equal(t, "https://example.com", result.OriginalURL)
		assert.NotEmpty(t, result.URLHash)
		assert.Equal(t, savedURL, result)
	})

	t.Run("returns error when GetByHash fails with non-ErrNotFound", func(t *testing.T) {
		repoErr := errors.New("repository error")
		repo := &mockRepository{
			getByHashFunc: func(_ context.Context, _ shortener.URLHash) (*shortener.ShortURL, error) {
				return nil, repoErr
			},
		}
		generator := func() string { return "newcode" }

		strategy := shortener.NewHashStrategy(repo, generator)
		result, err := strategy.Shorten(context.Background(), "https://example.com")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("returns error when Save fails", func(t *testing.T) {
		saveErr := errors.New("save failed")
		repo := &mockRepository{
			getByHashFunc: func(_ context.Context, _ shortener.URLHash) (*shortener.ShortURL, error) {
				return nil, shortener.ErrNotFound
			},
			saveFunc: func(_ context.Context, _ *shortener.ShortURL) error {
				return saveErr
			},
		}
		generator := func() string { return "newcode" }

		strategy := shortener.NewHashStrategy(repo, generator)
		result, err := strategy.Shorten(context.Background(), "https://example.com")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, saveErr)
	})

	t.Run("returns error for invalid URL", func(t *testing.T) {
		repo := &mockRepository{}
		generator := func() string { return "newcode" }

		strategy := shortener.NewHashStrategy(repo, generator)
		result, err := strategy.Shorten(context.Background(), "://invalid")

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
