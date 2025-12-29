package ratelimit

import "time"

// LimitConfig defines a single rate limit rule.
type LimitConfig struct {
	// Window is the time duration for the rate limit window.
	Window time.Duration
	// Max is the maximum number of requests allowed within the window.
	Max int64
}

// Policy defines rate limits for different scopes.
type Policy struct {
	// Limits maps scopes to their rate limit configurations.
	// A scope can have multiple limits (e.g., per-minute and per-day).
	Limits map[Scope][]LimitConfig
}

// NewPolicy creates a new policy with the given scope limits.
func NewPolicy(limits map[Scope][]LimitConfig) *Policy {
	return &Policy{
		Limits: limits,
	}
}

// PolicyBuilder provides a fluent API for building policies.
type PolicyBuilder struct {
	limits map[Scope][]LimitConfig
}

// NewPolicyBuilder creates a new policy builder.
func NewPolicyBuilder() *PolicyBuilder {
	return &PolicyBuilder{
		limits: make(map[Scope][]LimitConfig),
	}
}

// AddLimit adds a rate limit for a specific scope.
func (b *PolicyBuilder) AddLimit(scope Scope, maxReqs int64, window time.Duration) *PolicyBuilder {
	b.limits[scope] = append(b.limits[scope], LimitConfig{
		Window: window,
		Max:    maxReqs,
	})

	return b
}

// Build creates the policy from the builder configuration.
func (b *PolicyBuilder) Build() *Policy {
	return NewPolicy(b.limits)
}
