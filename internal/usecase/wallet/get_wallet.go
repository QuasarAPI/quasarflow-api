package wallet

import (
	"context"
	"fmt"

	"quasarflow-api/internal/domain/wallet"

	"github.com/google/uuid"
)

// GetWalletOutput represents the output of the get wallet use case
type GetWalletOutput struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	Network   string `json:"network"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetWalletUseCase handles retrieving a wallet by ID
type GetWalletUseCase struct {
	repo wallet.Repository
}

// NewGetWalletUseCase creates a new get wallet use case
func NewGetWalletUseCase(repo wallet.Repository) *GetWalletUseCase {
	return &GetWalletUseCase{
		repo: repo,
	}
}

// Execute retrieves a wallet by its ID
func (uc *GetWalletUseCase) Execute(ctx context.Context, walletID uuid.UUID) (*GetWalletOutput, error) {
	// Find wallet in database
	w, err := uc.repo.FindByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	return &GetWalletOutput{
		ID:        w.ID.String(),
		PublicKey: w.PublicKey,
		Network:   w.Network,
		CreatedAt: w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: w.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
