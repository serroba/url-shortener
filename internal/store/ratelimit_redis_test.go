package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/serroba/web-demo-go/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Check if Redis is available
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available at %s: %v", addr, err)
	}

	return client
}

func TestRateLimitRedisStore(t *testing.T) {
	client := getRedisClient(t)

	t.Run("records and counts requests", func(t *testing.T) {
		s := store.NewRateLimitRedisStore(client)
		key := "test:ratelimit:records:" + t.Name()

		// Clean up before test
		client.Del(context.Background(), "ratelimit:"+key)

		count1, err := s.Record(context.Background(), key, time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count1)

		count2, err := s.Record(context.Background(), key, time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count2)

		count3, err := s.Record(context.Background(), key, time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count3)
	})

	t.Run("tracks keys independently", func(t *testing.T) {
		s := store.NewRateLimitRedisStore(client)
		key1 := "test:ratelimit:independent:key1:" + t.Name()
		key2 := "test:ratelimit:independent:key2:" + t.Name()

		// Clean up before test
		client.Del(context.Background(), "ratelimit:"+key1, "ratelimit:"+key2)

		_, _ = s.Record(context.Background(), key1, time.Minute)
		_, _ = s.Record(context.Background(), key1, time.Minute)

		count, err := s.Record(context.Background(), key2, time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "key2 should have its own counter")
	})

	t.Run("prunes expired entries", func(t *testing.T) {
		s := store.NewRateLimitRedisStore(client)
		key := "test:ratelimit:prune:" + t.Name()

		// Clean up before test
		client.Del(context.Background(), "ratelimit:"+key)

		// Record some requests with a short window
		_, _ = s.Record(context.Background(), key, 50*time.Millisecond)
		_, _ = s.Record(context.Background(), key, 50*time.Millisecond)

		// Wait for them to expire
		time.Sleep(60 * time.Millisecond)

		// New request should only count itself
		count, err := s.Record(context.Background(), key, 50*time.Millisecond)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "expired entries should be pruned")
	})
}
