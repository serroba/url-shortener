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
