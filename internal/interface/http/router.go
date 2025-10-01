package http

import (
	"github.com/QuasarAPI/quasarflow-api/internal/interface/http/handler"
	"github.com/QuasarAPI/quasarflow-api/internal/interface/http/middleware"
	"github.com/gorilla/mux"
)

func SetupRouter(walletHandler *handler.WalletHandler, healthHandler *handler.HealthHandler) *mux.Router {
	r := mux.NewRouter()

	// Middleware global
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)
	r.Use(middleware.RequestID)

	// Health check
	r.HandleFunc("/health", healthHandler.Check).Methods("GET")

	// API v1
	api := r.PathPrefix("/api/v1").Subrouter()

	// Wallet endpoints
	api.HandleFunc("/wallets", walletHandler.Create).Methods("POST")
	api.HandleFunc("/wallets/{id}", walletHandler.GetByID).Methods("GET")
	api.HandleFunc("/wallets/{id}/balance", walletHandler.GetBalance).Methods("GET")
	api.HandleFunc("/wallets", walletHandler.List).Methods("GET")

	return r
}
