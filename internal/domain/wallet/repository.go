package wallet

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, wallet *Wallet) error
	FindByID(ctx context.Context, id uuid.UUID) (*Wallet, error)
	FindByPublicKey(ctx context.Context, publicKey string) (*Wallet, error)
	List(ctx context.Context, limit, offset int) ([]*Wallet, error)
	Count(ctx context.Context) (int64, error)
}
