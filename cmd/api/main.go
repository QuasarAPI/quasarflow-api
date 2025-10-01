package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/QuasarAPI/quasarflow-api/internal/config"
	"github.com/QuasarAPI/quasarflow-api/internal/infrastructure/crypto"
	"github.com/QuasarAPI/quasarflow-api/internal/infrastructure/database"
	"github.com/QuasarAPI/quasarflow-api/internal/infrastructure/stellar"
	"github.com/QuasarAPI/quasarflow-api/internal/interface/http"
	"github.com/QuasarAPI/quasarflow-api/internal/interface/http/handler"
	"github.com/QuasarAPI/quasarflow-api/internal/usecase/wallet"
	"github.com/QuasarAPI/quasarflow-api/pkg/logger"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New(cfg.LogLevel)

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database", "error", err)
	}
	defer db.Close()

	// Setup repositories
	walletRepo := database.NewPostgresWalletRepository(db)

	// Setup Stellar client
	stellarClient := stellar.NewClient(cfg.StellarHorizonURL)

	// Setup crypto
	encryptor, err := crypto.NewAESEncryptor(cfg.EncryptionKey)
	if err != nil {
		log.Fatal("failed to create encryptor", "error", err)
	}

	// Setup use cases
	createWalletUC := wallet.NewCreateWalletUseCase(walletRepo, encryptor, log)
	getWalletUC := wallet.NewGetWalletUseCase(walletRepo)
	getBalanceUC := wallet.NewGetBalanceUseCase(walletRepo, stellarClient)
	listWalletsUC := wallet.NewListWalletsUseCase(walletRepo)

	// Setup handlers
	walletHandler := handler.NewWalletHandler(createWalletUC, getWalletUC, getBalanceUC, listWalletsUC)
	healthHandler := handler.NewHealthHandler(db)

	// Setup router
	router := http.SetupRouter(walletHandler, healthHandler)

	// Setup HTTP server
	srv := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("starting server", "address", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", "error", err)
	}

	log.Info("server exited")
}
