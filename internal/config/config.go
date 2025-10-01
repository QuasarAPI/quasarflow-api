package config

import (
	"fmt"
	"log"
	"os"

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

	// Security configuration
	EncryptionKey string
	JWTSecret     string

	// Logging configuration
	LogLevel string

	// Rate limiting configuration
	RateLimit      int
	RateLimitBurst int
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

		// Security
		EncryptionKey: getEnvRequired("ENCRYPTION_KEY"),
		JWTSecret:     getEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Rate limiting
		RateLimit:      getEnvInt("RATE_LIMIT", 100),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 200),
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
