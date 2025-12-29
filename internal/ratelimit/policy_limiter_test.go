package ratelimit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/serroba/web-demo-go/internal/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	counts map[string]int64
	err    error
}

func newMockStore() *mockStore {
	return &mockStore{counts: make(map[string]int64)}
}

func (m *mockStore) Record(_ context.Context, key string, _ time.Duration) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}

	m.counts[key]++

	return m.counts[key], nil
}

func TestPolicyLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeGlobal, 10, time.Minute).
		Build()

	limiter := ratelimit.NewPolicyLimiter(store, policy)

	for range 10 {
		allowed, exceeded, err := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{ratelimit.ScopeGlobal})
		require.NoError(t, err)
		assert.True(t, allowed)
		assert.Nil(t, exceeded)
	}
}

func TestPolicyLimiter_DeniesRequestsOverLimit(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeGlobal, 5, time.Minute).
		Build()

	limiter := ratelimit.NewPolicyLimiter(store, policy)

	// First 5 requests should be allowed
	for range 5 {
		allowed, _, err := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{ratelimit.ScopeGlobal})
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// 6th request should be denied
	allowed, exceeded, err := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{ratelimit.ScopeGlobal})
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.NotNil(t, exceeded)
	assert.Equal(t, ratelimit.ScopeGlobal, exceeded.Scope)
	assert.Equal(t, int64(6), exceeded.Count)
	assert.Equal(t, int64(5), exceeded.Config.Max)
}

func TestPolicyLimiter_ChecksMultipleScopes(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeGlobal, 100, time.Minute).
		AddLimit(ratelimit.ScopeWrite, 2, time.Minute).
		Build()

	limiter := ratelimit.NewPolicyLimiter(store, policy)

	scopes := []ratelimit.Scope{ratelimit.ScopeGlobal, ratelimit.ScopeWrite}

	// First 2 write requests should be allowed
	for range 2 {
		allowed, _, err := limiter.Allow(context.Background(), "client1", scopes)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// 3rd write request should be denied (write limit exceeded, not global)
	allowed, exceeded, err := limiter.Allow(context.Background(), "client1", scopes)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, ratelimit.ScopeWrite, exceeded.Scope)
}

func TestPolicyLimiter_IndependentClientTracking(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeGlobal, 2, time.Minute).
		Build()

	limiter := ratelimit.NewPolicyLimiter(store, policy)

	// Client 1 uses their limit
	for range 2 {
		allowed, _, _ := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{ratelimit.ScopeGlobal})
		assert.True(t, allowed)
	}

	allowed, _, _ := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{ratelimit.ScopeGlobal})
	assert.False(t, allowed)

	// Client 2 should still have their full limit
	for range 2 {
		allowed, _, _ := limiter.Allow(context.Background(), "client2", []ratelimit.Scope{ratelimit.ScopeGlobal})
		assert.True(t, allowed)
	}
}

func TestPolicyLimiter_SkipsUndefinedScopes(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	// Policy only defines global scope
	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeGlobal, 5, time.Minute).
		Build()

	limiter := ratelimit.NewPolicyLimiter(store, policy)

	// Request with undefined scope should only check global
	scopes := []ratelimit.Scope{ratelimit.ScopeGlobal, ratelimit.ScopeWrite}
	allowed, _, err := limiter.Allow(context.Background(), "client1", scopes)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestPolicyLimiter_MultipleWindowsPerScope(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeWrite, 5, time.Minute). // 5 per minute
		AddLimit(ratelimit.ScopeWrite, 10, time.Hour).  // 10 per hour
		Build()

	limiter := ratelimit.NewPolicyLimiter(store, policy)

	// First 5 should be allowed (per-minute limit)
	for range 5 {
		allowed, _, _ := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{ratelimit.ScopeWrite})
		assert.True(t, allowed)
	}

	// 6th should be denied by per-minute limit
	allowed, exceeded, _ := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{ratelimit.ScopeWrite})
	assert.False(t, allowed)
	assert.Equal(t, time.Minute, exceeded.Config.Window)
}

func TestPolicyLimiter_PropagatesStoreErrors(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.err = errors.New("store error")

	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeGlobal, 10, time.Minute).
		Build()

	limiter := ratelimit.NewPolicyLimiter(store, policy)

	allowed, exceeded, err := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{ratelimit.ScopeGlobal})
	assert.False(t, allowed)
	assert.Nil(t, exceeded)
	require.Error(t, err)
	assert.Equal(t, "store error", err.Error())
}

func TestPolicyLimiter_EmptyScopes(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeGlobal, 1, time.Minute).
		Build()

	limiter := ratelimit.NewPolicyLimiter(store, policy)

	// Empty scopes should allow (no limits to check)
	allowed, exceeded, err := limiter.Allow(context.Background(), "client1", []ratelimit.Scope{})
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Nil(t, exceeded)
}
