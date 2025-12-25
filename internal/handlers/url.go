package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jaevor/go-nanoid"
	"github.com/serroba/web-demo-go/internal/domain"
)

// URLHandler handles URL shortening operations.
type URLHandler struct {
	store      domain.ShortURLRepository
	generateID func() string
	baseURL    string
}

// NewURLHandler creates a new URL handler with the given store.
func NewURLHandler(s domain.ShortURLRepository, baseURL string, codeLength int) *URLHandler {
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

	// Determine strategy (default to token)
	strategy := req.Body.Strategy
	if strategy == "" {
		strategy = StrategyToken
	}

	var shortURL *domain.ShortURL

	var err error

	switch strategy {
	case StrategyHash:
		shortURL, err = h.createWithHashStrategy(ctx, req.Body.URL)
	case StrategyToken:
		shortURL, err = h.createWithTokenStrategy(ctx, req.Body.URL)
	default:
		return nil, huma.Error400BadRequest("invalid strategy: must be 'token' or 'hash'")
	}

	if err != nil {
		return nil, huma.Error500InternalServerError("failed to save url")
	}

	fullShortURL := fmt.Sprintf("%s/%s", h.baseURL, shortURL.Code)

	resp := &CreateShortURLResponse{}
	resp.Headers.Location = fullShortURL
	resp.Body.Code = string(shortURL.Code)
	resp.Body.ShortURL = fullShortURL
	resp.Body.OriginalURL = shortURL.OriginalURL

	return resp, nil
}

// createWithTokenStrategy always generates a new code (current behavior).
func (h *URLHandler) createWithTokenStrategy(ctx context.Context, url string) (*domain.ShortURL, error) {
	shortURL := &domain.ShortURL{
		Code:        domain.Code(h.generateID()),
		OriginalURL: url,
		URLHash:     "", // empty for token strategy
	}

	if err := h.store.Save(ctx, shortURL); err != nil {
		return nil, err
	}

	return shortURL, nil
}

// createWithHashStrategy checks for existing hash mapping first (deduplication).
func (h *URLHandler) createWithHashStrategy(ctx context.Context, rawURL string) (*domain.ShortURL, error) {
	// Normalize URL for consistent hashing
	normalizedURL, err := NormalizeURL(rawURL)
	if err != nil {
		return nil, err
	}

	// Compute hash of normalized URL
	urlHash := domain.URLHash(HashURL(normalizedURL))

	// Check if we already have a short URL for this hash
	existing, err := h.store.GetByHash(ctx, urlHash)
	if err == nil {
		// Found existing - return it (deduplication)
		return existing, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		// Unexpected error
		return nil, err
	}

	// No existing mapping - create new short URL
	shortURL := &domain.ShortURL{
		Code:        domain.Code(h.generateID()),
		OriginalURL: rawURL,
		URLHash:     urlHash,
	}

	if err = h.store.Save(ctx, shortURL); err != nil {
		return nil, err
	}

	return shortURL, nil
}

func (h *URLHandler) RedirectToURL(ctx context.Context, req *RedirectRequest) (*RedirectResponse, error) {
	shortURL, err := h.store.GetByCode(ctx, domain.Code(req.Code))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, huma.Error404NotFound("short url not found")
		}

		return nil, huma.Error500InternalServerError("failed to get url")
	}

	resp := &RedirectResponse{
		Status: http.StatusMovedPermanently,
	}
	resp.Headers.Location = shortURL.OriginalURL

	return resp, nil
}
