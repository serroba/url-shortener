package store

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis is a Redis implementation of ratelimit.Store using sorted sets.
type Redis struct {
	client *redis.Client
	prefix string
}

// NewRedis creates a new Redis-backed rate limit store.
func NewRedis(client *redis.Client) *Redis {
	return &Redis{
		client: client,
		prefix: "ratelimit:",
	}
}

// Record records a request and returns the count of requests in the current window.
// Uses Redis sorted sets with timestamps as scores for sliding window implementation.
func (s *Redis) Record(ctx context.Context, key string, window time.Duration) (int64, error) {
	now := time.Now()
	nowUnix := float64(now.UnixNano())
	cutoff := float64(now.Add(-window).UnixNano())
	redisKey := s.prefix + key

	// Use a pipeline for atomic operations
	pipe := s.client.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatFloat(cutoff, 'f', -1, 64))

	// Add current request with unique member (timestamp + counter)
	// Using UnixNano as both score and member ensures uniqueness
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  nowUnix,
		Member: strconv.FormatInt(now.UnixNano(), 10),
	})

	// Count entries in the window
	countCmd := pipe.ZCard(ctx, redisKey)

	// Set TTL to auto-expire the key after the window
	pipe.Expire(ctx, redisKey, window+time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return countCmd.Val(), nil
}
