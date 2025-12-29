package ratelimit

import "github.com/danielgtaylor/huma/v2"

// Scope categorizes a request for rate limiting purposes.
// Different scopes can have different rate limits applied.
type Scope string

const (
	// ScopeGlobal applies to all requests regardless of type.
	ScopeGlobal Scope = "global"
	// ScopeRead applies to read operations (GET, HEAD, OPTIONS).
	ScopeRead Scope = "read"
	// ScopeWrite applies to write operations (POST, PUT, PATCH, DELETE).
	ScopeWrite Scope = "write"
)

// ScopeResolver determines which scopes apply to a given request.
type ScopeResolver interface {
	Resolve(ctx huma.Context) []Scope
}

// MethodScopeResolver resolves scopes based on HTTP method.
// GET, HEAD, OPTIONS are classified as read operations.
// All other methods are classified as write operations.
type MethodScopeResolver struct{}

// NewMethodScopeResolver creates a new method-based scope resolver.
func NewMethodScopeResolver() *MethodScopeResolver {
	return &MethodScopeResolver{}
}

// Resolve returns the scopes that apply to the request based on its HTTP method.
func (r *MethodScopeResolver) Resolve(ctx huma.Context) []Scope {
	scopes := []Scope{ScopeGlobal}

	switch ctx.Method() {
	case "GET", "HEAD", "OPTIONS":
		scopes = append(scopes, ScopeRead)
	default:
		scopes = append(scopes, ScopeWrite)
	}

	return scopes
}
