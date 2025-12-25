package handlers_test

import (
	"context"
	"errors"

	"github.com/serroba/web-demo-go/internal/shortener"
)

var errMock = errors.New("mock error")

const testURL = "https://example.com"

// mockStore is a test double for shortener.Repository that can be configured to return errors.
type mockStore struct {
	saveErr         error
	getByCodeErr    error
	getByHashErr    error
	saved           *shortener.ShortURL
	getByHashResult *shortener.ShortURL
}

func (m *mockStore) Save(_ context.Context, shortURL *shortener.ShortURL) error {
	m.saved = shortURL

	return m.saveErr
}

func (m *mockStore) GetByCode(_ context.Context, _ shortener.Code) (*shortener.ShortURL, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}

	return &shortener.ShortURL{
		Code:        "abc123",
		OriginalURL: testURL,
	}, nil
}

func (m *mockStore) GetByHash(_ context.Context, _ shortener.URLHash) (*shortener.ShortURL, error) {
	if m.getByHashErr != nil {
		return nil, m.getByHashErr
	}

	return m.getByHashResult, nil
}
