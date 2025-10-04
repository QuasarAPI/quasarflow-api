package wallet

import (
	"context"
	"fmt"

	"quasarflow-api/internal/domain/wallet"
	"quasarflow-api/internal/infrastructure/crypto"
	"quasarflow-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

type SendPaymentInput struct {
	FromWalletID uuid.UUID `json:"from_wallet_id" validate:"required"`
	ToAddress    string    `json:"to_address" validate:"required"`
	Amount       string    `json:"amount" validate:"required"`
	AssetCode    string    `json:"asset_code,omitempty"` // Optional, defaults to XLM
	AssetIssuer  string    `json:"asset_issuer,omitempty"`
	Memo         string    `json:"memo,omitempty"`
}

type SendPaymentOutput struct {
	TransactionHash string `json:"transaction_hash"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	AssetCode       string `json:"asset_code"`
	AssetIssuer     string `json:"asset_issuer,omitempty"`
	Memo            string `json:"memo,omitempty"`
	Network         string `json:"network"`
	Ledger          int32  `json:"ledger"`
	Success         bool   `json:"success"`
}

type SendPaymentUseCase struct {
	repo          wallet.Repository
	horizonClient *horizonclient.Client
	encryptor     crypto.Encryptor
	logger        logger.Logger
}

func NewSendPaymentUseCase(
	repo wallet.Repository,
	horizonClient *horizonclient.Client,
	encryptor crypto.Encryptor,
	logger logger.Logger,
) *SendPaymentUseCase {
	return &SendPaymentUseCase{
		repo:          repo,
		horizonClient: horizonClient,
		encryptor:     encryptor,
		logger:        logger,
	}
}

func (uc *SendPaymentUseCase) Execute(ctx context.Context, input SendPaymentInput) (*SendPaymentOutput, error) {
	// 1. Find source wallet
	sourceWallet, err := uc.repo.FindByID(ctx, input.FromWalletID)
	if err != nil {
		uc.logger.Error("failed to find source wallet", logger.Error(err))
		return nil, fmt.Errorf("source wallet not found: %w", err)
	}

	// 2. Decrypt private key
	decryptedSeed, err := uc.encryptor.Decrypt(sourceWallet.EncryptedKey)
	if err != nil {
		uc.logger.Error("failed to decrypt private key", logger.Error(err))
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// 3. Create keypair from seed
	sourceKeypair, err := keypair.ParseFull(decryptedSeed)
	if err != nil {
		uc.logger.Error("failed to parse keypair", logger.Error(err))
		return nil, fmt.Errorf("failed to parse keypair: %w", err)
	}

	// 4. Get network passphrase
	var networkPassphrase string
	switch sourceWallet.Network {
	case "mainnet":
		networkPassphrase = network.PublicNetworkPassphrase
	case "testnet":
		networkPassphrase = network.TestNetworkPassphrase
	case "local":
		networkPassphrase = "Standalone Network ; February 2017"
	default:
		return nil, fmt.Errorf("unsupported network: %s", sourceWallet.Network)
	}

	// 5. Load source account
	accountRequest := horizonclient.AccountRequest{
		AccountID: sourceKeypair.Address(),
	}

	sourceAccount, err := uc.horizonClient.AccountDetail(accountRequest)
	if err != nil {
		uc.logger.Error("failed to load source account", logger.Error(err))
		return nil, fmt.Errorf("failed to load source account: %w", err)
	}

	// 6. Create asset (default to native XLM)
	var asset txnbuild.Asset
	if input.AssetCode == "" || input.AssetCode == "XLM" {
		asset = txnbuild.NativeAsset{}
	} else {
		if input.AssetIssuer == "" {
			return nil, fmt.Errorf("asset issuer is required for non-native assets")
		}
		asset = txnbuild.CreditAsset{
			Code:   input.AssetCode,
			Issuer: input.AssetIssuer,
		}
	}

	// 7. Create payment operation
	paymentOp := &txnbuild.Payment{
		Destination: input.ToAddress,
		Amount:      input.Amount,
		Asset:       asset,
	}

	// 8. Build transaction
	txParams := txnbuild.TransactionParams{
		SourceAccount:        &sourceAccount,
		IncrementSequenceNum: true,
		Operations:           []txnbuild.Operation{paymentOp},
		BaseFee:              txnbuild.MinBaseFee,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
	}

	// Add memo if provided
	if input.Memo != "" {
		txParams.Memo = txnbuild.MemoText(input.Memo)
	}

	tx, err := txnbuild.NewTransaction(txParams)
	if err != nil {
		uc.logger.Error("failed to build transaction", logger.Error(err))
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// 9. Sign transaction
	tx, err = tx.Sign(networkPassphrase, sourceKeypair)
	if err != nil {
		uc.logger.Error("failed to sign transaction", logger.Error(err))
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 10. Submit transaction
	uc.logger.Info("submitting payment transaction",
		logger.String("from", sourceKeypair.Address()),
		logger.String("to", input.ToAddress),
		logger.String("amount", input.Amount),
		logger.String("asset", input.AssetCode),
	)

	resp, err := uc.horizonClient.SubmitTransaction(tx)
	if err != nil {
		uc.logger.Error("failed to submit transaction", logger.Error(err))

		// Try to get more details from Horizon error
		if horizonErr, ok := err.(*horizonclient.Error); ok {
			return nil, fmt.Errorf("transaction failed: %s (code: %d)", horizonErr.Problem.Detail, horizonErr.Problem.Status)
		}

		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	// 11. Parse response
	assetCode := "XLM"
	assetIssuer := ""
	if input.AssetCode != "" && input.AssetCode != "XLM" {
		assetCode = input.AssetCode
		assetIssuer = input.AssetIssuer
	}

	uc.logger.Info("payment transaction successful",
		logger.String("hash", resp.Hash),
		logger.Int32("ledger", resp.Ledger),
		logger.String("from", sourceKeypair.Address()),
		logger.String("to", input.ToAddress),
	)

	return &SendPaymentOutput{
		TransactionHash: resp.Hash,
		FromAddress:     sourceKeypair.Address(),
		ToAddress:       input.ToAddress,
		Amount:          input.Amount,
		AssetCode:       assetCode,
		AssetIssuer:     assetIssuer,
		Memo:            input.Memo,
		Network:         sourceWallet.Network,
		Ledger:          resp.Ledger,
		Success:         resp.Successful,
	}, nil
}
