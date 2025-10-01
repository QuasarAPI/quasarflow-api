package handler

import (
	"database/sql"
	"net/http"

	"quasarflow-api/internal/interface/http/response"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// Check handles the health check endpoint
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	healthStatus := HealthResponse{
		Status:   "healthy",
		Database: "unknown",
	}

	// Check database connection
	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			healthStatus.Database = "unhealthy"
			healthStatus.Status = "unhealthy"
			response.Error(w, http.StatusServiceUnavailable, "Database connection failed")
			return
		}
		healthStatus.Database = "healthy"
	}

	response.Success(w, http.StatusOK, healthStatus)
}
