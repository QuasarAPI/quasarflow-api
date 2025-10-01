package wallet

import (
	"context"
	"fmt"

	"quasarflow-api/internal/domain/wallet"
)

// WalletListItem represents a single wallet in the list
type WalletListItem struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	Network   string `json:"network"`
	CreatedAt string `json:"created_at"`
}

// ListWalletsOutput represents the output of the list wallets use case
type ListWalletsOutput struct {
	Wallets []WalletListItem `json:"wallets"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// ListWalletsUseCase handles listing wallets with pagination
type ListWalletsUseCase struct {
	repo wallet.Repository
}

// NewListWalletsUseCase creates a new list wallets use case
func NewListWalletsUseCase(repo wallet.Repository) *ListWalletsUseCase {
	return &ListWalletsUseCase{
		repo: repo,
	}
}

// Execute retrieves a paginated list of wallets
func (uc *ListWalletsUseCase) Execute(ctx context.Context, limit, offset int) (*ListWalletsOutput, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}
	if offset < 0 {
		offset = 0
	}

	// Get wallets from database
	wallets, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	// Get total count
	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count wallets: %w", err)
	}

	// Convert to output format
	items := make([]WalletListItem, 0, len(wallets))
	for _, w := range wallets {
		items = append(items, WalletListItem{
			ID:        w.ID.String(),
			PublicKey: w.PublicKey,
			Network:   w.Network,
			CreatedAt: w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &ListWalletsOutput{
		Wallets: items,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}
