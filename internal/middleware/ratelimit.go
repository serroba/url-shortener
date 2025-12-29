package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/serroba/web-demo-go/internal/ratelimit"
)

// RateLimiter returns a Huma middleware that limits requests based on client IP and User-Agent.
func RateLimiter(api huma.API, limiter ratelimit.Limiter) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		key := clientKey(ctx)

		allowed, err := limiter.Allow(ctx.Context(), key)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error", err)

			return
		}

		if !allowed {
			_ = huma.WriteErr(api, ctx, http.StatusTooManyRequests, "rate limit exceeded")

			return
		}

		next(ctx)
	}
}

// clientKey generates a unique key for rate limiting based on IP and User-Agent.
func clientKey(ctx huma.Context) string {
	ip := clientIP(ctx)
	ua := ctx.Header("User-Agent")

	hash := sha256.Sum256([]byte(ip + "|" + ua))

	return hex.EncodeToString(hash[:])
}

// clientIP extracts the client IP from the request, considering proxies.
func clientIP(ctx huma.Context) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	if xff := ctx.Header("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}

		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := ctx.Header("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to Host (which contains remote addr in Huma context)
	host := ctx.Host()

	ip, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}

	return ip
}

// PolicyRateLimiter returns a Huma middleware that applies policy-based rate limiting.
// It uses a ScopeResolver to determine which scopes apply to each request,
// then checks all applicable limits from the policy.
//
// Per-endpoint configuration can be provided via operation metadata using
// ratelimit.MetadataKey. This allows endpoints to:
//   - Disable rate limiting entirely (Disabled: true)
//   - Override the scope detection (Scope: ratelimit.ScopeRead)
//   - Define custom limits (Limits: []ratelimit.LimitConfig{...})
func PolicyRateLimiter(
	api huma.API,
	limiter *ratelimit.PolicyLimiter,
	resolver ratelimit.ScopeResolver,
) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Check for per-endpoint configuration
		if cfg := ratelimit.GetEndpointConfig(ctx); cfg != nil {
			// Skip rate limiting if disabled for this endpoint
			if cfg.Disabled {
				next(ctx)

				return
			}

			// If custom limits are defined, use them directly
			if len(cfg.Limits) > 0 {
				if !checkCustomLimits(api, ctx, limiter.Store(), cfg.Limits) {
					return
				}

				next(ctx)

				return
			}
		}

		// Default behavior: use policy-based rate limiting
		key := clientKey(ctx)
		scopes := resolver.Resolve(ctx)

		allowed, exceeded, err := limiter.Allow(ctx.Context(), key, scopes)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error", err)

			return
		}

		if !allowed {
			msg := "rate limit exceeded"
			if exceeded != nil {
				msg = fmt.Sprintf("rate limit exceeded: %s scope, %d/%d requests in %s",
					exceeded.Scope, exceeded.Count, exceeded.Config.Max, exceeded.Config.Window)
			}

			_ = huma.WriteErr(api, ctx, http.StatusTooManyRequests, msg)

			return
		}

		next(ctx)
	}
}

// checkCustomLimits applies custom rate limits defined in endpoint config.
// Returns true if request is allowed, false if rate limited.
func checkCustomLimits(
	api huma.API,
	ctx huma.Context,
	store ratelimit.Store,
	limits []ratelimit.LimitConfig,
) bool {
	clientK := clientKey(ctx)
	path := ctx.Operation().Path

	for _, limit := range limits {
		// Build key combining client, path, and window for unique tracking
		key := fmt.Sprintf("%s:custom:%s:%d", clientK, path, limit.Window.Milliseconds())

		count, err := store.Record(ctx.Context(), key, limit.Window)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error", err)

			return false
		}

		if count > limit.Max {
			msg := fmt.Sprintf("rate limit exceeded: %d/%d requests in %s",
				count, limit.Max, limit.Window)
			_ = huma.WriteErr(api, ctx, http.StatusTooManyRequests, msg)

			return false
		}
	}

	return true
}
