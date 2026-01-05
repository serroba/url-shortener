package store

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// slidingWindow tracks requests in a sliding time window using the same
// algorithm as github.com/serroba/rate/window.SlidingLimiter.
type slidingWindow struct {
	window time.Duration
	q      []time.Time
	head   int
}

func newSlidingWindow(dur time.Duration) *slidingWindow {
	return &slidingWindow{
		window: dur,
		q:      make([]time.Time, 0),
	}
}

// record adds a new request timestamp and returns the current count.
// It prunes expired entries before counting.
func (sw *slidingWindow) record(now time.Time) int64 {
	cutoff := now.Add(-sw.window)

	// Prune expired entries (same algorithm as serroba/rate)
	for sw.head < len(sw.q) && sw.q[sw.head].Before(cutoff) {
		sw.head++
	}

	// Compact the slice if needed (same as serroba/rate)
	if sw.head > 0 && sw.head*2 >= len(sw.q) {
		sw.q = append([]time.Time(nil), sw.q[sw.head:]...)
		sw.head = 0
	}

	// Add current request
	sw.q = append(sw.q, now)

	return int64(len(sw.q) - sw.head)
}

// Memory is an in-memory implementation of ratelimit.Store.
// It uses a sliding window algorithm matching github.com/serroba/rate.
type Memory struct {
	mu      sync.Mutex
	windows map[string]*slidingWindow
}

// NewMemory creates a new in-memory rate limit store.
func NewMemory() *Memory {
	return &Memory{
		windows: make(map[string]*slidingWindow),
	}
}

// Record records a request and returns the current count within the window.
// The sliding window algorithm matches github.com/serroba/rate/window.SlidingLimiter.
func (s *Memory) Record(_ context.Context, key string, dur time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create composite key including window duration for unique windows
	compositeKey := fmt.Sprintf("%s:%d", key, dur.Milliseconds())

	sw, ok := s.windows[compositeKey]
	if !ok {
		sw = newSlidingWindow(dur)
		s.windows[compositeKey] = sw
	}

	return sw.record(time.Now()), nil
}
