package handlers_test

import (
	"context"
	"errors"

	"github.com/serroba/web-demo-go/internal/domain"
)

var errMock = errors.New("mock error")

const testURL = "https://example.com"

// mockStore is a test double for ShortURLRepository that can be configured to return errors.
type mockStore struct {
	saveErr         error
	getByCodeErr    error
	getByHashErr    error
	saved           *domain.ShortURL
	getByHashResult *domain.ShortURL
}

func (m *mockStore) Save(_ context.Context, shortURL *domain.ShortURL) error {
	m.saved = shortURL

	return m.saveErr
}

func (m *mockStore) GetByCode(_ context.Context, _ domain.Code) (*domain.ShortURL, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}

	return &domain.ShortURL{
		Code:        "abc123",
		OriginalURL: testURL,
	}, nil
}

func (m *mockStore) GetByHash(_ context.Context, _ domain.URLHash) (*domain.ShortURL, error) {
	if m.getByHashErr != nil {
		return nil, m.getByHashErr
	}

	return m.getByHashResult, nil
}
