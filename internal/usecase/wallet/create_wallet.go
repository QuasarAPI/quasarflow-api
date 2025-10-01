package wallet

import (
	"context"
	"fmt"

	"quasarflow-api/internal/domain/wallet"
	"quasarflow-api/internal/infrastructure/crypto"
	"quasarflow-api/pkg/logger"

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
		uc.logger.Error("failed to generate keypair", logger.Error(err))
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	// 2. Encrypt private key before storing
	encryptedKey, err := uc.crypto.Encrypt(pair.Seed())
	if err != nil {
		uc.logger.Error("failed to encrypt private key", logger.Error(err))
		return nil, fmt.Errorf("failed to encrypt key: %w", err)
	}

	// 3. Create wallet entity
	w, err := wallet.NewWallet(pair.Address(), encryptedKey, input.Network)
	if err != nil {
		uc.logger.Error("failed to create wallet entity", logger.Error(err))
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// 4. Save to database
	if err := uc.repo.Create(ctx, w); err != nil {
		uc.logger.Error("failed to save wallet", logger.Error(err))
		return nil, fmt.Errorf("failed to save wallet: %w", err)
	}

	uc.logger.Info("wallet created successfully", logger.String("id", w.ID.String()), logger.String("public_key", w.PublicKey))

	return &CreateWalletOutput{
		ID:        w.ID.String(),
		PublicKey: w.PublicKey,
		Network:   w.Network,
		CreatedAt: w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
