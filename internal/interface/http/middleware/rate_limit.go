package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"quasarflow-api/internal/interface/http/response"
	"quasarflow-api/pkg/errors"
	"quasarflow-api/pkg/logger"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	// HTTP headers for client IP detection
	headerXForwardedFor = "X-Forwarded-For"
	headerXRealIP       = "X-Real-IP"

	// Default values
	defaultCleanupMultiplier = 2
)

var (
	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = &errors.AppError{
		Type:       errors.ErrorTypeInternal,
		Message:    "Rate limit exceeded. Please try again later.",
		StatusCode: http.StatusTooManyRequests,
	}
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	CleanupInterval   time.Duration
}

// RateLimiter represents a rate limiter for a specific key
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	config      RateLimitConfig
	limiters    map[string]*RateLimiter
	mu          sync.RWMutex
	logger      logger.Logger
	cleanupTick *time.Ticker
	done        chan bool
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(config RateLimitConfig, logger logger.Logger) *RateLimitMiddleware {
	middleware := &RateLimitMiddleware{
		config:   config,
		limiters: make(map[string]*RateLimiter),
		logger:   logger,
		done:     make(chan bool),
	}

	// Start cleanup goroutine
	middleware.startCleanup()

	return middleware
}

// RateLimit is a middleware that enforces rate limiting
func (rlm *RateLimitMiddleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP address
		clientIP := rlm.getClientIP(r)

		// Get or create rate limiter for this client
		limiter := rlm.getLimiter(clientIP)

		// Check if request is allowed
		if !limiter.Allow() {
			rlm.logger.Warn("rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.String("path", r.URL.Path),
				zap.String("user_agent", r.UserAgent()))

			response.Error(w, ErrRateLimitExceeded.StatusCode, ErrRateLimitExceeded.Message)
			return
		}

		// Update last seen time
		rlm.updateLastSeen(clientIP)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the real client IP address
func (rlm *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// Check for forwarded headers (in case of reverse proxy)
	if xff := r.Header.Get(headerXForwardedFor); xff != "" {
		// Take the first IP in the chain (most trusted)
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xri := r.Header.Get(headerXRealIP); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback to remote address (remove port if present)
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}
	return ip
}

// getLimiter gets or creates a rate limiter for the given key
func (rlm *RateLimitMiddleware) getLimiter(key string) *rate.Limiter {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	limiter, exists := rlm.limiters[key]
	if !exists {
		limiter = &RateLimiter{
			limiter:  rate.NewLimiter(rate.Limit(rlm.config.RequestsPerSecond), rlm.config.BurstSize),
			lastSeen: time.Now(),
		}
		rlm.limiters[key] = limiter
	}

	return limiter.limiter
}

// updateLastSeen updates the last seen time for a limiter
func (rlm *RateLimitMiddleware) updateLastSeen(key string) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	if limiter, exists := rlm.limiters[key]; exists {
		limiter.lastSeen = time.Now()
	}
}

// startCleanup starts the cleanup goroutine to remove old limiters
func (rlm *RateLimitMiddleware) startCleanup() {
	rlm.cleanupTick = time.NewTicker(rlm.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-rlm.cleanupTick.C:
				rlm.cleanup()
			case <-rlm.done:
				return
			}
		}
	}()
}

// cleanup removes old rate limiters that haven't been used recently
func (rlm *RateLimitMiddleware) cleanup() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	cutoff := time.Now().Add(-rlm.config.CleanupInterval * defaultCleanupMultiplier)
	for key, limiter := range rlm.limiters {
		if limiter.lastSeen.Before(cutoff) {
			delete(rlm.limiters, key)
			rlm.logger.Debug("cleaned up old rate limiter",
				zap.String("client_ip", key),
				zap.Time("last_seen", limiter.lastSeen))
		}
	}
}

// Close stops the cleanup goroutine
func (rlm *RateLimitMiddleware) Close() {
	if rlm.cleanupTick != nil {
		rlm.cleanupTick.Stop()
	}
	close(rlm.done)
}

// GetStats returns current rate limiting statistics
func (rlm *RateLimitMiddleware) GetStats() map[string]interface{} {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	return map[string]interface{}{
		"active_limiters":     len(rlm.limiters),
		"requests_per_second": rlm.config.RequestsPerSecond,
		"burst_size":          rlm.config.BurstSize,
	}
}
