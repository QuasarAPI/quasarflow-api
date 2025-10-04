package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"quasarflow-api/internal/interface/http/response"
	"quasarflow-api/pkg/errors"
	"quasarflow-api/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthContextKey is a custom type for context keys to avoid collisions
type AuthContextKey string

const (
	// Context keys
	UserIDKey   AuthContextKey = "user_id"
	UserRoleKey AuthContextKey = "user_role"

	// JWT constants
	BearerPrefix = "Bearer "
	HMACMethod   = "HS256"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	SecretKey     string
	TokenDuration time.Duration
	Issuer        string
}

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	config AuthConfig
	logger logger.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config AuthConfig, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
		logger: logger,
	}
}

// RequireAuth is a middleware that validates JWT tokens
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			am.logger.Warn("missing authorization header",
				zap.String("ip", r.RemoteAddr),
				zap.String("path", r.URL.Path))
			response.Error(w, errors.ErrMissingAuthHeader.StatusCode, errors.ErrMissingAuthHeader.Message)
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			am.logger.Warn("invalid authorization header format",
				zap.String("ip", r.RemoteAddr),
				zap.String("path", r.URL.Path),
				zap.String("header", authHeader))
			response.Error(w, errors.ErrInvalidAuthFormat.StatusCode, errors.ErrInvalidAuthFormat.Message)
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		// Parse and validate token
		claims, err := am.validateToken(tokenString)
		if err != nil {
			am.logger.Warn("invalid token",
				zap.Error(err),
				zap.String("ip", r.RemoteAddr),
				zap.String("path", r.URL.Path))
			response.Error(w, errors.ErrInvalidToken.StatusCode, errors.ErrInvalidToken.Message)
			return
		}

		// Add user information to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		// Log successful authentication
		am.logger.Info("user authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("role", claims.Role),
			zap.String("ip", r.RemoteAddr))

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken validates a JWT token and returns the claims
func (am *AuthMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: %v", errors.ErrInvalidSigningMethod.Message, token.Header["alg"])
		}
		return []byte(am.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.ErrInvalidTokenClaims
	}

	// Validate issuer
	if claims.Issuer != am.config.Issuer {
		return nil, errors.ErrInvalidTokenIssuer
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.ErrTokenExpired
	}

	return claims, nil
}

// GenerateToken generates a new JWT token for a user
func (am *AuthMiddleware) GenerateToken(userID, role string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    am.config.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(am.config.TokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.config.SecretKey))
}

// RequireRole is a middleware that validates user roles
func (am *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				am.logger.Warn("user role not found in context",
					zap.String("ip", r.RemoteAddr),
					zap.String("path", r.URL.Path))
				response.Error(w, errors.ErrUserRoleNotFound.StatusCode, errors.ErrUserRoleNotFound.Message)
				return
			}

			if userRole != requiredRole {
				am.logger.Warn("insufficient permissions",
					zap.String("user_role", userRole),
					zap.String("required_role", requiredRole),
					zap.String("ip", r.RemoteAddr),
					zap.String("path", r.URL.Path))
				response.Error(w, errors.ErrInsufficientPermissions.StatusCode, errors.ErrInsufficientPermissions.Message)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUserRoleFromContext extracts user role from context
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}
