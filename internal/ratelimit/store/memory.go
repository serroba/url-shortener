package store

import (
	"context"
	"sync"
	"time"
)

// Memory is an in-memory implementation of ratelimit.Store.
type Memory struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

// NewMemory creates a new in-memory rate limit store.
func NewMemory() *Memory {
	return &Memory{
		requests: make(map[string][]time.Time),
	}
}

func (s *Memory) Record(_ context.Context, key string, window time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	// Get existing timestamps and prune expired ones
	timestamps := s.requests[key]
	valid := make([]time.Time, 0, len(timestamps)+1)

	for _, ts := range timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}

	// Add current request
	valid = append(valid, now)
	s.requests[key] = valid

	return int64(len(valid)), nil
}
