package wallet

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID           uuid.UUID
	PublicKey    string // Stellar public key (G...)
	EncryptedKey string // Encrypted private key
	Network      string // "testnet" ou "mainnet"
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewWallet(publicKey, encryptedKey, network string) (*Wallet, error) {
	if !strings.HasPrefix(publicKey, "G") {
		return nil, fmt.Errorf("invalid public key format: must start with 'G'")
	}

	if len(publicKey) != 56 {
		return nil, fmt.Errorf("invalid public key length: must be 56 characters")
	}

	if encryptedKey == "" {
		return nil, fmt.Errorf("encrypted key is required")
	}

	if network != "testnet" && network != "mainnet" {
		return nil, fmt.Errorf("invalid network: must be 'testnet' or 'mainnet'")
	}

	return &Wallet{
		ID:           uuid.New(),
		PublicKey:    publicKey,
		EncryptedKey: encryptedKey,
		Network:      network,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}
