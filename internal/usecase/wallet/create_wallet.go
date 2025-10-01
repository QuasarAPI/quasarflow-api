package wallet

import (
	"context"
	"fmt"

	"github.com/QuasarAPI/quasarflow-api/internal/domain/wallet"
	"github.com/QuasarAPI/quasarflow-api/internal/infrastructure/crypto"
	"github.com/QuasarAPI/quasarflow-api/pkg/logger"
	"github.com/stellar/go/keypair"
)

type CreateWalletInput struct {
	Network string `json:"network" validate:"required,oneof=testnet mainnet"`
}

type CreateWalletOutput struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	Network   string `json:"network"`
	CreatedAt string `json:"created_at"`
}

type CreateWalletUseCase struct {
	repo   wallet.Repository
	crypto crypto.Encryptor
	logger logger.Logger
}

func NewCreateWalletUseCase(
	repo wallet.Repository,
	crypto crypto.Encryptor,
	logger logger.Logger,
) *CreateWalletUseCase {
	return &CreateWalletUseCase{
		repo:   repo,
		crypto: crypto,
		logger: logger,
	}
}

func (uc *CreateWalletUseCase) Execute(ctx context.Context, input CreateWalletInput) (*CreateWalletOutput, error) {
	// 1. Generate Stellar keypair
	pair, err := keypair.Random()
	if err != nil {
		uc.logger.Error("failed to generate keypair", "error", err)
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	// 2. Encrypt private key before storing
	encryptedKey, err := uc.crypto.Encrypt(pair.Seed())
	if err != nil {
		uc.logger.Error("failed to encrypt private key", "error", err)
		return nil, fmt.Errorf("failed to encrypt key: %w", err)
	}

	// 3. Create wallet entity
	w := wallet.NewWallet(pair.Address(), encryptedKey, input.Network)

	// 4. Save to database
	if err := uc.repo.Create(ctx, w); err != nil {
		uc.logger.Error("failed to save wallet", "error", err)
		return nil, fmt.Errorf("failed to save wallet: %w", err)
	}

	uc.logger.Info("wallet created successfully", "id", w.ID, "public_key", w.PublicKey)

	return &CreateWalletOutput{
		ID:        w.ID.String(),
		PublicKey: w.PublicKey,
		Network:   w.Network,
		CreatedAt: w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
