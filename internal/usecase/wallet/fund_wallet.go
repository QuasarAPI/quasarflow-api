package wallet

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"quasarflow-api/internal/domain/wallet"
	"quasarflow-api/pkg/logger"

	"github.com/google/uuid"
)

type FundWalletInput struct {
	WalletID uuid.UUID `json:"wallet_id" validate:"required"`
	Amount   string    `json:"amount,omitempty"` // Optional, defaults to Friendbot default
}

type FundWalletOutput struct {
	WalletID      string `json:"wallet_id"`
	PublicKey     string `json:"public_key"`
	Network       string `json:"network"`
	TransactionID string `json:"transaction_id,omitempty"`
	Message       string `json:"message"`
	Success       bool   `json:"success"`
}

type FundWalletUseCase struct {
	repo         wallet.Repository
	friendbotURL string
	logger       logger.Logger
}

func NewFundWalletUseCase(
	repo wallet.Repository,
	friendbotURL string,
	logger logger.Logger,
) *FundWalletUseCase {
	return &FundWalletUseCase{
		repo:         repo,
		friendbotURL: friendbotURL,
		logger:       logger,
	}
}

func (uc *FundWalletUseCase) Execute(ctx context.Context, input FundWalletInput) (*FundWalletOutput, error) {
	// 1. Find wallet in database
	w, err := uc.repo.FindByID(ctx, input.WalletID)
	if err != nil {
		uc.logger.Error("failed to find wallet", logger.Error(err), logger.String("wallet_id", input.WalletID.String()))
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// 2. Check if wallet is on a network that supports Friendbot
	if w.Network == "mainnet" {
		return &FundWalletOutput{
			WalletID:  w.ID.String(),
			PublicKey: w.PublicKey,
			Network:   w.Network,
			Message:   "Friendbot is not available on mainnet. Please fund this wallet manually.",
			Success:   false,
		}, nil
	}

	// 3. Use Friendbot URL from config (centralized)
	friendbotEndpoint := uc.friendbotURL
	if friendbotEndpoint == "" {
		return nil, fmt.Errorf("Friendbot URL not configured for network: %s", w.Network)
	}

	// 4. Build request URL
	requestURL, err := url.Parse(friendbotEndpoint)
	if err != nil {
		uc.logger.Error("failed to parse friendbot URL", logger.Error(err))
		return nil, fmt.Errorf("invalid friendbot URL: %w", err)
	}

	query := requestURL.Query()
	query.Add("addr", w.PublicKey)
	if input.Amount != "" {
		query.Add("amount", input.Amount)
	}
	requestURL.RawQuery = query.Encode()

	// 5. Make request to Friendbot
	uc.logger.Info("funding wallet via friendbot",
		logger.String("wallet_id", w.ID.String()),
		logger.String("public_key", w.PublicKey),
		logger.String("network", w.Network),
		logger.String("url", requestURL.String()),
	)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL.String(), nil)
	if err != nil {
		uc.logger.Error("failed to create friendbot request", logger.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		uc.logger.Error("failed to call friendbot", logger.Error(err))
		return nil, fmt.Errorf("failed to fund wallet: %w", err)
	}
	defer resp.Body.Close()

	// 6. Check response
	if resp.StatusCode != http.StatusOK {
		uc.logger.Error("friendbot request failed",
			logger.Int("status_code", resp.StatusCode),
			logger.String("status", resp.Status),
		)

		var message string
		switch resp.StatusCode {
		case http.StatusBadRequest:
			message = "Invalid wallet address or request parameters"
		case http.StatusNotFound:
			message = "Friendbot service not available"
		case http.StatusTooManyRequests:
			message = "Rate limit exceeded. Please try again later"
		default:
			message = fmt.Sprintf("Friendbot request failed with status: %s", resp.Status)
		}

		return &FundWalletOutput{
			WalletID:  w.ID.String(),
			PublicKey: w.PublicKey,
			Network:   w.Network,
			Message:   message,
			Success:   false,
		}, nil
	}

	// 7. Success response
	uc.logger.Info("wallet funded successfully",
		logger.String("wallet_id", w.ID.String()),
		logger.String("public_key", w.PublicKey),
	)

	fundingAmount := "10000" // Default Friendbot amount
	if input.Amount != "" {
		fundingAmount = input.Amount
	}

	return &FundWalletOutput{
		WalletID:  w.ID.String(),
		PublicKey: w.PublicKey,
		Network:   w.Network,
		Message:   fmt.Sprintf("Wallet successfully funded with %s XLM", fundingAmount),
		Success:   true,
	}, nil
}
