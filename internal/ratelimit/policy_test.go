package ratelimit_test

import (
	"testing"
	"time"

	"github.com/serroba/web-demo-go/internal/ratelimit"
	"github.com/stretchr/testify/assert"
)

func TestPolicyBuilder_Build(t *testing.T) {
	t.Parallel()

	policy := ratelimit.NewPolicyBuilder().
		AddLimit(ratelimit.ScopeGlobal, 1000, 24*time.Hour).
		AddLimit(ratelimit.ScopeRead, 100, time.Minute).
		AddLimit(ratelimit.ScopeWrite, 10, time.Minute).
		AddLimit(ratelimit.ScopeWrite, 50, time.Hour).
		Build()

	assert.Len(t, policy.Limits[ratelimit.ScopeGlobal], 1)
	assert.Len(t, policy.Limits[ratelimit.ScopeRead], 1)
	assert.Len(t, policy.Limits[ratelimit.ScopeWrite], 2)

	assert.Equal(t, int64(1000), policy.Limits[ratelimit.ScopeGlobal][0].Max)
	assert.Equal(t, 24*time.Hour, policy.Limits[ratelimit.ScopeGlobal][0].Window)

	assert.Equal(t, int64(100), policy.Limits[ratelimit.ScopeRead][0].Max)
	assert.Equal(t, time.Minute, policy.Limits[ratelimit.ScopeRead][0].Window)

	assert.Equal(t, int64(10), policy.Limits[ratelimit.ScopeWrite][0].Max)
	assert.Equal(t, time.Minute, policy.Limits[ratelimit.ScopeWrite][0].Window)

	assert.Equal(t, int64(50), policy.Limits[ratelimit.ScopeWrite][1].Max)
	assert.Equal(t, time.Hour, policy.Limits[ratelimit.ScopeWrite][1].Window)
}

func TestPolicyBuilder_EmptyPolicy(t *testing.T) {
	t.Parallel()

	policy := ratelimit.NewPolicyBuilder().Build()

	assert.Empty(t, policy.Limits)
}

func TestNewPolicy(t *testing.T) {
	t.Parallel()

	limits := map[ratelimit.Scope][]ratelimit.LimitConfig{
		ratelimit.ScopeGlobal: {
			{Window: 24 * time.Hour, Max: 5000},
		},
		ratelimit.ScopeRead: {
			{Window: time.Minute, Max: 100},
		},
	}

	policy := ratelimit.NewPolicy(limits)

	assert.Equal(t, limits, policy.Limits)
}
