package handler

import (
	"encoding/json"
	"net/http"

	"quasarflow-api/internal/interface/http/middleware"
	"quasarflow-api/internal/interface/http/response"
	"quasarflow-api/pkg/logger"

	"go.uber.org/zap"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authMiddleware *middleware.AuthMiddleware
	logger         logger.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authMiddleware *middleware.AuthMiddleware, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn string `json:"expires_in"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
}

// Login handles user login and returns a JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid login request", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Basic validation
	if req.Username == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// TODO: Implement proper user authentication against database
	// For now, we'll use a simple demo authentication
	// In production, you should:
	// 1. Hash passwords using bcrypt
	// 2. Verify credentials against database
	// 3. Implement proper user management
	if !h.authenticateUser(req.Username, req.Password) {
		h.logger.Warn("authentication failed",
			zap.String("username", req.Username),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	// TODO: Get actual user ID and role from database
	userID := "demo-user-id"
	role := "user"

	token, err := h.authMiddleware.GenerateToken(userID, role)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.logger.Info("user logged in successfully",
		zap.String("username", req.Username),
		zap.String("user_id", userID),
		zap.String("ip", r.RemoteAddr))

	// Return token response
	loginResp := LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: "24h", // TODO: Get from config
		UserID:    userID,
		Role:      role,
	}

	response.Success(w, http.StatusOK, loginResp)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// In a stateless JWT implementation, logout is typically handled client-side
	// by removing the token. For enhanced security, you might want to implement
	// a token blacklist or use refresh tokens.

	h.logger.Info("user logged out")
	response.Success(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Me returns information about the current authenticated user
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	role, _ := middleware.GetUserRoleFromContext(r.Context())

	userInfo := map[string]interface{}{
		"user_id": userID,
		"role":    role,
	}

	response.Success(w, http.StatusOK, userInfo)
}

// authenticateUser performs basic authentication
// TODO: Replace with proper database authentication
func (h *AuthHandler) authenticateUser(username, password string) bool {
	// Demo credentials - replace with proper authentication logic
	demoCredentials := map[string]string{
		"admin": "admin123",
		"user":  "user123",
	}

	if storedPassword, exists := demoCredentials[username]; exists {
		return storedPassword == password
	}

	return false
}
