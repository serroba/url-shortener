package store

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// RedisStore is a Redis implementation of URLStore.
type RedisStore struct {
	client *redis.Client
	prefix string
}

// NewRedisStore creates a new Redis-backed URL store.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: "url:",
	}
}

func (r *RedisStore) Save(ctx context.Context, code, url string) error {
	return r.client.Set(ctx, r.prefix+code, url, 0).Err()
}

func (r *RedisStore) Get(ctx context.Context, code string) (string, error) {
	url, err := r.client.Get(ctx, r.prefix+code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}

		return "", err
	}

	return url, nil
}
