package wallet

import (
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

func NewWallet(publicKey, encryptedKey, network string) *Wallet {
	return &Wallet{
		ID:           uuid.New(),
		PublicKey:    publicKey,
		EncryptedKey: encryptedKey,
		Network:      network,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
