package wallet

import (
	"context"
	"fmt"

	domainStellar "quasarflow-api/internal/domain/stellar"
	"quasarflow-api/internal/domain/wallet"
	"quasarflow-api/internal/infrastructure/stellar"

	"github.com/google/uuid"
)

type GetBalanceOutput struct {
	PublicKey string                  `json:"public_key"`
	Network   string                  `json:"network"`
	Balances  []domainStellar.Balance `json:"balances"`
}

type GetBalanceUseCase struct {
	repo          wallet.Repository
	stellarClient *stellar.Client
}

func NewGetBalanceUseCase(
	repo wallet.Repository,
	stellarClient *stellar.Client,
) *GetBalanceUseCase {
	return &GetBalanceUseCase{
		repo:          repo,
		stellarClient: stellarClient,
	}
}

func (uc *GetBalanceUseCase) Execute(ctx context.Context, walletID uuid.UUID) (*GetBalanceOutput, error) {
	// 1. Find wallet in database
	w, err := uc.repo.FindByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// 2. Fetch balance from Stellar network
	balances, err := uc.stellarClient.GetAccountBalances(ctx, w.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balance: %w", err)
	}

	return &GetBalanceOutput{
		PublicKey: w.PublicKey,
		Network:   w.Network,
		Balances:  balances,
	}, nil
}
