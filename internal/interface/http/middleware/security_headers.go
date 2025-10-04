package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	CSPConnectSources []string
	EnableHSTS        bool
}

// SecurityHeaders is a middleware that adds security headers to responses
func SecurityHeaders(config SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking attacks
			w.Header().Set("X-Frame-Options", "DENY")

			// Enable XSS protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Force HTTPS (only when enabled and in HTTPS context)
			if config.EnableHSTS && (r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https") {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			// Prevent information disclosure
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

			// Control referrer information
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy (configurable)
			connectSources := "'self'"
			if len(config.CSPConnectSources) > 0 {
				connectSources += " " + strings.Join(config.CSPConnectSources, " ")
			}

			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline'; "+
					"style-src 'self' 'unsafe-inline'; "+
					"img-src 'self' data: https:; "+
					"font-src 'self'; "+
					"connect-src "+connectSources+"; "+
					"frame-ancestors 'none'; "+
					"base-uri 'self'; "+
					"form-action 'self'")

			// Remove server information
			w.Header().Set("Server", "")

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}
