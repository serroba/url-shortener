package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jaevor/go-nanoid"
	"github.com/serroba/web-demo-go/internal/store"
)

// URLHandler handles URL shortening operations.
type URLHandler struct {
	store      store.URLStore
	generateID func() string
	baseURL    string
}

// NewURLHandler creates a new URL handler with the given store.
func NewURLHandler(s store.URLStore, baseURL string, codeLength int) *URLHandler {
	gen, _ := nanoid.Standard(codeLength)

	return &URLHandler{
		store:      s,
		generateID: gen,
		baseURL:    baseURL,
	}
}

func (h *URLHandler) CreateShortURL(ctx context.Context, req *CreateShortURLRequest) (*CreateShortURLResponse, error) {
	if req.Body.URL == "" {
		return nil, huma.Error400BadRequest("url is required")
	}

	code := h.generateID()
	if err := h.store.Save(ctx, code, req.Body.URL); err != nil {
		return nil, huma.Error500InternalServerError("failed to save url")
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, code)

	resp := &CreateShortURLResponse{}
	resp.Headers.Location = shortURL
	resp.Body.Code = code
	resp.Body.ShortURL = shortURL
	resp.Body.OriginalURL = req.Body.URL

	return resp, nil
}

func (h *URLHandler) RedirectToURL(ctx context.Context, req *RedirectRequest) (*RedirectResponse, error) {
	url, err := h.store.Get(ctx, req.Code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, huma.Error404NotFound("short url not found")
		}

		return nil, huma.Error500InternalServerError("failed to get url")
	}

	resp := &RedirectResponse{
		Status: http.StatusMovedPermanently,
	}
	resp.Headers.Location = url

	return resp, nil
}
