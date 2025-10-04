package wallet

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"quasarflow-api/internal/infrastructure/stellar"
	"quasarflow-api/pkg/errors"
	"quasarflow-api/pkg/logger"

	"github.com/stellar/go/keypair"
)

// VerifyOwnershipUseCase handles wallet ownership verification using Stellar SDK
type VerifyOwnershipUseCase struct {
	stellarClient stellar.Client
	logger        logger.Logger
	domain        string
}

// VerifyOwnershipInput represents the input for ownership verification
type VerifyOwnershipInput struct {
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

// VerifyOwnershipOutput represents the result of ownership verification
type VerifyOwnershipOutput struct {
	IsOwner bool   `json:"is_owner"`
	Message string `json:"message"`
}

// ChallengeOutput represents a generated challenge
type ChallengeOutput struct {
	Challenge    string `json:"challenge"`
	Message      string `json:"message"`
	PublicKey    string `json:"public_key"`
	Instructions string `json:"instructions"`
}

// NewVerifyOwnershipUseCase creates a new ownership verification use case
func NewVerifyOwnershipUseCase(stellarClient stellar.Client, logger logger.Logger, domain string) *VerifyOwnershipUseCase {
	return &VerifyOwnershipUseCase{
		stellarClient: stellarClient,
		logger:        logger,
		domain:        domain,
	}
}

// Execute verifies wallet ownership using message signature
func (uc *VerifyOwnershipUseCase) Execute(ctx context.Context, input VerifyOwnershipInput) (*VerifyOwnershipOutput, error) {
	uc.logger.Info("verifying wallet ownership",
		logger.String("public_key", input.PublicKey),
		logger.String("message", input.Message))

	// 1. Validate Stellar public key format
	if !uc.isValidStellarPublicKey(input.PublicKey) {
		uc.logger.Warn("invalid stellar public key format",
			logger.String("public_key", input.PublicKey))
		return &VerifyOwnershipOutput{
			IsOwner: false,
			Message: "Invalid Stellar public key format",
		}, nil
	}

	// 2. Verify signature using Stellar Go SDK
	isValid, err := uc.verifySignatureWithSDK(input.PublicKey, input.Message, input.Signature)
	if err != nil {
		uc.logger.Error("failed to verify signature",
			logger.Error(err),
			logger.String("public_key", input.PublicKey))
		return nil, errors.NewInternalError("Failed to verify signature", err)
	}

	if !isValid {
		uc.logger.Warn("ownership verification failed",
			logger.String("public_key", input.PublicKey),
			logger.String("reason", "invalid signature"))
		return &VerifyOwnershipOutput{
			IsOwner: false,
			Message: "Invalid signature or message",
		}, nil
	}

	uc.logger.Info("ownership verified successfully",
		logger.String("public_key", input.PublicKey))

	return &VerifyOwnershipOutput{
		IsOwner: true,
		Message: "Ownership verified successfully",
	}, nil
}

// GenerateChallenge creates a SEP-10 compliant challenge for ownership verification
func (uc *VerifyOwnershipUseCase) GenerateChallenge(publicKey string) *ChallengeOutput {
	if !uc.isValidStellarPublicKey(publicKey) {
		return &ChallengeOutput{
			Challenge:    "",
			Message:      "Invalid Stellar public key format",
			PublicKey:    publicKey,
			Instructions: "",
		}
	}

	// SEP-10 format: timestamp.nonce.domain.public_key
	timestamp := time.Now().Unix()
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	challenge := fmt.Sprintf("%d.%s.%s.%s", timestamp, nonce, uc.domain, publicKey)

	uc.logger.Info("generated SEP-10 challenge",
		logger.String("public_key", publicKey),
		logger.String("challenge", challenge))

	return &ChallengeOutput{
		Challenge:    challenge,
		Message:      "Sign this challenge with your private key to verify ownership",
		PublicKey:    publicKey,
		Instructions: "Use Stellar SDK to sign the challenge with your private key",
	}
}

// VerifyOwnershipByTransaction verifies ownership via a signed transaction
func (uc *VerifyOwnershipUseCase) VerifyOwnershipByTransaction(ctx context.Context, publicKey, transactionHash string) (*VerifyOwnershipOutput, error) {
	uc.logger.Info("verifying ownership via transaction",
		logger.String("public_key", publicKey),
		logger.String("transaction_hash", transactionHash))

	// Validate public key format
	if !uc.isValidStellarPublicKey(publicKey) {
		return &VerifyOwnershipOutput{
			IsOwner: false,
			Message: "Invalid Stellar public key format",
		}, nil
	}

	// Fetch transaction from Horizon
	tx, err := uc.stellarClient.GetTransaction(transactionHash)
	if err != nil {
		uc.logger.Error("failed to fetch transaction",
			logger.Error(err),
			logger.String("transaction_hash", transactionHash))
		return nil, errors.NewBlockchainError("Failed to fetch transaction", err)
	}

	// Verify if the transaction was signed by the specified wallet
	if tx.SourceAccount != publicKey {
		uc.logger.Warn("transaction not signed by specified wallet",
			logger.String("public_key", publicKey),
			logger.String("source_account", tx.SourceAccount),
			logger.String("transaction_hash", transactionHash))
		return &VerifyOwnershipOutput{
			IsOwner: false,
			Message: "Transaction was not signed by the specified wallet",
		}, nil
	}

	// Additional verification: check if it's a recent transaction (within last 24 hours)
	// This prevents replay attacks with old transactions
	if time.Since(tx.LedgerCloseTime) > 24*time.Hour {
		uc.logger.Warn("transaction too old for ownership verification",
			logger.String("public_key", publicKey),
			logger.String("ledger_close_time", tx.LedgerCloseTime.Format(time.RFC3339)),
			logger.String("transaction_hash", transactionHash))
		return &VerifyOwnershipOutput{
			IsOwner: false,
			Message: "Transaction is too old for ownership verification",
		}, nil
	}

	uc.logger.Info("ownership verified via transaction",
		logger.String("public_key", publicKey),
		logger.String("transaction_hash", transactionHash))

	return &VerifyOwnershipOutput{
		IsOwner: true,
		Message: "Ownership verified via transaction",
	}, nil
}

// VerifyOwnershipByAccount verifies ownership by checking account existence and recent activity
func (uc *VerifyOwnershipUseCase) VerifyOwnershipByAccount(ctx context.Context, publicKey string) (*VerifyOwnershipOutput, error) {
	uc.logger.Info("verifying ownership via account details",
		logger.String("public_key", publicKey))

	// Validate public key format
	if !uc.isValidStellarPublicKey(publicKey) {
		return &VerifyOwnershipOutput{
			IsOwner: false,
			Message: "Invalid Stellar public key format",
		}, nil
	}

	// Check if account exists on Stellar network
	account, err := uc.stellarClient.GetAccount(publicKey)
	if err != nil {
		uc.logger.Warn("account not found on Stellar network",
			logger.String("public_key", publicKey),
			logger.Error(err))
		return &VerifyOwnershipOutput{
			IsOwner: false,
			Message: "Account not found on Stellar network",
		}, nil
	}

	// Check if account has recent activity (within last 30 days)
	// This indicates the account is active and likely owned by the requester
	recentActivity := time.Since(account.LastModifiedTime) < 30*24*time.Hour

	if !recentActivity {
		uc.logger.Warn("account has no recent activity",
			logger.String("public_key", publicKey),
			logger.String("last_modified", account.LastModifiedTime.Format(time.RFC3339)))
		return &VerifyOwnershipOutput{
			IsOwner: false,
			Message: "Account has no recent activity",
		}, nil
	}

	uc.logger.Info("ownership verified via account activity",
		logger.String("public_key", publicKey),
		logger.String("last_modified", account.LastModifiedTime.Format(time.RFC3339)))

	return &VerifyOwnershipOutput{
		IsOwner: true,
		Message: "Account exists and has recent activity",
	}, nil
}

// isValidStellarPublicKey validates the format of a Stellar public key
func (uc *VerifyOwnershipUseCase) isValidStellarPublicKey(publicKey string) bool {
	// Basic format validation
	if len(publicKey) != 56 || publicKey[0] != 'G' {
		return false
	}

	// Try to parse as Stellar keypair using SDK
	_, err := keypair.ParseAddress(publicKey)
	return err == nil
}

// verifySignatureWithSDK verifies signature using Stellar Go SDK
func (uc *VerifyOwnershipUseCase) verifySignatureWithSDK(publicKey, message, signature string) (bool, error) {
	// Parse the public key using SDK
	kp, err := keypair.ParseAddress(publicKey)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Decode the signature (base64)
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Convert message to bytes
	messageBytes := []byte(message)

	// Verify signature using SDK Go (most robust method)
	err = kp.Verify(messageBytes, sigBytes)
	if err != nil {
		return false, nil // Invalid signature, but not a system error
	}

	return true, nil
}
