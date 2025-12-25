package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/serroba/web-demo-go/internal/handlers"
	"github.com/serroba/web-demo-go/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortURL(t *testing.T) {
	t.Run("creates short url successfully", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req := &handlers.CreateShortURLRequest{}
		req.Body.URL = "https://example.com/very/long/path"

		resp, err := handler.CreateShortURL(context.Background(), req)

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Body.Code)
		assert.Equal(t, "https://example.com/very/long/path", resp.Body.OriginalURL)
		assert.Contains(t, resp.Body.ShortURL, resp.Body.Code)
		assert.Equal(t, resp.Body.ShortURL, resp.Headers.Location)
	})

	t.Run("returns error when url is empty", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req := &handlers.CreateShortURLRequest{}
		req.Body.URL = ""

		resp, err := handler.CreateShortURL(context.Background(), req)

		assert.Nil(t, resp)
		assert.Error(t, err)
	})

	t.Run("returns error for invalid strategy", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req := &handlers.CreateShortURLRequest{}
		req.Body.URL = "https://example.com"
		req.Body.Strategy = "invalid"

		resp, err := handler.CreateShortURL(context.Background(), req)

		assert.Nil(t, resp)
		assert.Error(t, err)
	})

	t.Run("token strategy creates new code for same URL", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req := &handlers.CreateShortURLRequest{}
		req.Body.URL = "https://example.com"
		req.Body.Strategy = handlers.StrategyToken

		resp1, err1 := handler.CreateShortURL(context.Background(), req)
		resp2, err2 := handler.CreateShortURL(context.Background(), req)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, resp1.Body.Code, resp2.Body.Code)
	})

	t.Run("hash strategy returns same code for same URL", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req := &handlers.CreateShortURLRequest{}
		req.Body.URL = "https://example.com"
		req.Body.Strategy = handlers.StrategyHash

		resp1, err1 := handler.CreateShortURL(context.Background(), req)
		resp2, err2 := handler.CreateShortURL(context.Background(), req)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, resp1.Body.Code, resp2.Body.Code)
	})

	t.Run("hash strategy returns same code for equivalent URLs", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req1 := &handlers.CreateShortURLRequest{}
		req1.Body.URL = "https://example.com/path"
		req1.Body.Strategy = handlers.StrategyHash

		req2 := &handlers.CreateShortURLRequest{}
		req2.Body.URL = "HTTPS://EXAMPLE.COM/path/"
		req2.Body.Strategy = handlers.StrategyHash

		resp1, err1 := handler.CreateShortURL(context.Background(), req1)
		resp2, err2 := handler.CreateShortURL(context.Background(), req2)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, resp1.Body.Code, resp2.Body.Code)
	})

	t.Run("hash strategy returns different codes for different URLs", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req1 := &handlers.CreateShortURLRequest{}
		req1.Body.URL = "https://example.com/path1"
		req1.Body.Strategy = handlers.StrategyHash

		req2 := &handlers.CreateShortURLRequest{}
		req2.Body.URL = "https://example.com/path2"
		req2.Body.Strategy = handlers.StrategyHash

		resp1, err1 := handler.CreateShortURL(context.Background(), req1)
		resp2, err2 := handler.CreateShortURL(context.Background(), req2)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, resp1.Body.Code, resp2.Body.Code)
	})

	t.Run("defaults to token strategy when not specified", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req := &handlers.CreateShortURLRequest{}
		req.Body.URL = "https://example.com"
		// Strategy not set - should default to token

		resp1, err1 := handler.CreateShortURL(context.Background(), req)
		resp2, err2 := handler.CreateShortURL(context.Background(), req)

		require.NoError(t, err1)
		require.NoError(t, err2)
		// Token strategy: different codes for same URL
		assert.NotEqual(t, resp1.Body.Code, resp2.Body.Code)
	})
}

func TestRedirectToURL(t *testing.T) {
	t.Run("redirects to original url", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		_ = memStore.Save(context.Background(), "abc123", "https://example.com")
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req := &handlers.RedirectRequest{Code: "abc123"}

		resp, err := handler.RedirectToURL(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusMovedPermanently, resp.Status)
		assert.Equal(t, "https://example.com", resp.Headers.Location)
	})

	t.Run("returns 404 when code not found", func(t *testing.T) {
		memStore := store.NewMemoryStore()
		handler := handlers.NewURLHandler(memStore, "http://localhost:8888", 8)

		req := &handlers.RedirectRequest{Code: "notfound"}

		resp, err := handler.RedirectToURL(context.Background(), req)

		assert.Nil(t, resp)
		assert.Error(t, err)
	})
}
