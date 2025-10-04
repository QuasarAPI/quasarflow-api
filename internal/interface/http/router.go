package http

import (
	"time"

	"quasarflow-api/internal/config"
	"quasarflow-api/internal/interface/http/handler"
	"quasarflow-api/internal/interface/http/middleware"
	"quasarflow-api/pkg/logger"

	"github.com/gorilla/mux"
)

func SetupRouter(
	walletHandler *handler.WalletHandler,
	accountHandler *handler.AccountHandler,
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	cfg *config.Config,
	log logger.Logger,
) *mux.Router {
	r := mux.NewRouter()

	// Initialize security middlewares
	authConfig := middleware.AuthConfig{
		SecretKey:     cfg.JWTSecret,
		TokenDuration: parseDuration(cfg.JWTExpiration),
		Issuer:        cfg.JWTIssuer,
	}
	authMiddleware := middleware.NewAuthMiddleware(authConfig, log)

	rateLimitConfig := middleware.RateLimitConfig{
		RequestsPerSecond: cfg.RateLimitRequestsPerSecond,
		BurstSize:         cfg.RateLimitBurst,
		CleanupInterval:   parseDuration(cfg.RateLimitCleanupInterval),
	}
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimitConfig, log)

	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID", "Accept", "Origin"},
		MaxAge:         3600,
	}

	// validationMiddleware := middleware.NewValidationMiddleware(log) // TODO: Use for request validation

	// Security headers configuration
	securityHeadersConfig := middleware.SecurityHeadersConfig{
		CSPConnectSources: cfg.CSPConnectSources,
		EnableHSTS:        cfg.Environment == "production",
	}

	// Global middleware (order matters!)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.SecurityHeaders(securityHeadersConfig))
	r.Use(middleware.CORS(corsConfig))
	r.Use(middleware.RequestID)
	r.Use(rateLimitMiddleware.RateLimit)

	// Health check (public endpoint)
	r.HandleFunc("/health", healthHandler.Check).Methods("GET")

	// Authentication endpoints (public)
	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")
	auth.HandleFunc("/logout", authHandler.Logout).Methods("POST")

	// Public account verification endpoints (no authentication required)
	// These endpoints allow external users to verify wallet ownership
	accounts := r.PathPrefix("/api/v1/accounts").Subrouter()
	accounts.HandleFunc("/{public_key}/challenge", accountHandler.GetChallenge).Methods("GET")
	accounts.HandleFunc("/{public_key}/verify-ownership", accountHandler.VerifyOwnership).Methods("POST")
	accounts.HandleFunc("/{public_key}/verify-transaction", accountHandler.VerifyOwnershipByTransaction).Methods("POST")
	accounts.HandleFunc("/{public_key}/verify-account", accountHandler.VerifyOwnershipByAccount).Methods("GET")
	accounts.HandleFunc("/{public_key}/balance", accountHandler.GetAccountBalance).Methods("GET")
	accounts.HandleFunc("/{public_key}/transactions", accountHandler.GetAccountTransactionHistory).Methods("GET")

	// API v1 (protected routes)
	api := r.PathPrefix("/api/v1").Subrouter()

	// Apply authentication middleware to all API routes
	api.Use(authMiddleware.RequireAuth)

	// User info endpoint
	api.HandleFunc("/me", authHandler.Me).Methods("GET")

	// Wallet endpoints (all require authentication)
	api.HandleFunc("/wallets", walletHandler.Create).Methods("POST")
	api.HandleFunc("/wallets", walletHandler.List).Methods("GET")
	api.HandleFunc("/wallets/{id}", walletHandler.GetByID).Methods("GET")
	api.HandleFunc("/wallets/{id}/balance", walletHandler.GetBalance).Methods("GET")
	api.HandleFunc("/wallets/{id}/fund", walletHandler.Fund).Methods("POST")
	api.HandleFunc("/wallets/{id}/payment", walletHandler.SendPayment).Methods("POST")
	api.HandleFunc("/wallets/{id}/transactions", walletHandler.GetTransactionHistory).Methods("GET")

	return r
}

// parseDuration parses a duration string and returns a time.Duration
func parseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		// Default to 24 hours if parsing fails
		return 24 * time.Hour
	}
	return duration
}
