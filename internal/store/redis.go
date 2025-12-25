package store

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/serroba/web-demo-go/internal/domain"
)

// RedisStore is a Redis implementation of ShortURLRepository.
type RedisStore struct {
	client  *redis.Client
	prefix  string // "url:" prefix for code->entity (stored as Redis hash)
	hashKey string // "url_hashes" for urlHash->code lookup
}

// NewRedisStore creates a new Redis-backed URL store.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client:  client,
		prefix:  "url:",
		hashKey: "url_hashes",
	}
}

func (r *RedisStore) Save(ctx context.Context, shortURL *domain.ShortURL) error {
	pipe := r.client.Pipeline()

	// Store entity as Redis hash
	pipe.HSet(ctx, r.prefix+string(shortURL.Code), map[string]interface{}{
		"code":         string(shortURL.Code),
		"original_url": shortURL.OriginalURL,
		"url_hash":     string(shortURL.URLHash),
	})

	// Index by hash if present (for hash strategy)
	if shortURL.URLHash != "" {
		pipe.HSet(ctx, r.hashKey, string(shortURL.URLHash), string(shortURL.Code))
	}

	_, err := pipe.Exec(ctx)

	return err
}

func (r *RedisStore) GetByCode(ctx context.Context, code domain.Code) (*domain.ShortURL, error) {
	result, err := r.client.HGetAll(ctx, r.prefix+string(code)).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, domain.ErrNotFound
	}

	return &domain.ShortURL{
		Code:        domain.Code(result["code"]),
		OriginalURL: result["original_url"],
		URLHash:     domain.URLHash(result["url_hash"]),
	}, nil
}

func (r *RedisStore) GetByHash(ctx context.Context, hash domain.URLHash) (*domain.ShortURL, error) {
	code, err := r.client.HGet(ctx, r.hashKey, string(hash)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrNotFound
		}

		return nil, err
	}

	return r.GetByCode(ctx, domain.Code(code))
}
