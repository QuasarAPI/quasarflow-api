package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"quasarflow-api/internal/config"
	"quasarflow-api/internal/infrastructure/crypto"
	"quasarflow-api/internal/infrastructure/database"
	"quasarflow-api/internal/infrastructure/stellar"
	httpHandler "quasarflow-api/internal/interface/http"
	"quasarflow-api/internal/interface/http/handler"
	"quasarflow-api/internal/usecase/wallet"
	"quasarflow-api/pkg/logger"

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
		log.Fatal("failed to connect to database", logger.Error(err))
	}
	defer db.Close()

	// Setup repositories
	walletRepo := database.NewPostgresWalletRepository(db)

	// Setup Stellar client
	stellarClient := stellar.NewClient(cfg.StellarHorizonURL)

	// Setup crypto
	encryptor, err := crypto.NewAESEncryptor(cfg.EncryptionKey)
	if err != nil {
		log.Fatal("failed to create encryptor", logger.Error(err))
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
	router := httpHandler.SetupRouter(walletHandler, healthHandler)

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
		log.Info("starting server", logger.String("address", cfg.ServerAddress))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", logger.Error(err))
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
		log.Fatal("server forced to shutdown", logger.Error(err))
	}

	log.Info("server exited")
}
