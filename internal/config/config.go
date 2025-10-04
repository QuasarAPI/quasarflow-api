package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	ServerAddress string
	Environment   string

	// Database configuration
	DatabaseURL string

	// Stellar configuration
	StellarHorizonURL string
	StellarNetwork    string
	FriendbotURL      string

	// Security configuration
	EncryptionKey  string
	JWTSecret      string
	JWTExpiration  string
	JWTIssuer      string
	AllowedOrigins []string

	// Frontend and API URLs
	APIBaseURL        string
	FrontendURL       string
	CSPConnectSources []string

	// Logging configuration
	LogLevel string

	// Rate limiting configuration
	RateLimitRequestsPerSecond float64
	RateLimitBurst             int
	RateLimitCleanupInterval   string
}

// Load reads configuration from environment variables
// It attempts to load .env file first, then reads from environment
func Load() *Config {
	// Try to load .env file (optional, won't fail if not found)
	_ = godotenv.Load()

	return &Config{
		// Server
		ServerAddress: getEnv("PORT", ":8080"),
		Environment:   getEnv("ENV", "development"),

		// Database
		DatabaseURL: getEnvRequired("DATABASE_URL"),

		// Stellar
		StellarHorizonURL: getEnv("STELLAR_HORIZON_URL", "https://horizon-testnet.stellar.org"),
		StellarNetwork:    getEnv("STELLAR_NETWORK", "testnet"),
		FriendbotURL:      getEnv("FRIENDBOT_URL", "https://horizon-testnet.stellar.org/friendbot"),

		// Security
		EncryptionKey:  getEnvRequired("ENCRYPTION_KEY"),
		JWTSecret:      getEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),
		JWTExpiration:  getEnv("JWT_EXPIRATION", "24h"),
		JWTIssuer:      getEnv("JWT_ISSUER", "quasarflow-api"),
		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),

		// Frontend and API URLs
		APIBaseURL:        getEnv("API_BASE_URL", "http://localhost:8080"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		CSPConnectSources: getEnvSlice("CSP_CONNECT_SOURCES", []string{"https://horizon-testnet.stellar.org", "https://horizon.stellar.org", "http://localhost:8000"}),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Rate limiting
		RateLimitRequestsPerSecond: getEnvFloat64("RATE_LIMIT_REQUESTS_PER_SECOND", 100.0),
		RateLimitBurst:             getEnvInt("RATE_LIMIT_BURST", 200),
		RateLimitCleanupInterval:   getEnv("RATE_LIMIT_CLEANUP_INTERVAL", "10m"),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvRequired retrieves a required environment variable
// It will terminate the application if the variable is not set
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvFloat64 retrieves a float64 environment variable or returns a default value
func getEnvFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var floatValue float64
		if _, err := fmt.Sscanf(value, "%f", &floatValue); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvSlice retrieves a slice environment variable or returns a default value
func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim spaces
		values := strings.Split(value, ",")
		result := make([]string, 0, len(values))
		for _, v := range values {
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}
