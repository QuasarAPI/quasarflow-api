package database

import (
	"context"
	"database/sql"
	"fmt"

	"quasarflow-api/internal/domain/wallet"

	"github.com/google/uuid"
)

type PostgresWalletRepository struct {
	db *sql.DB
}

func NewPostgresWalletRepository(db *sql.DB) *PostgresWalletRepository {
	return &PostgresWalletRepository{db: db}
}

func (r *PostgresWalletRepository) Create(ctx context.Context, w *wallet.Wallet) error {
	query := `
        INSERT INTO wallets (id, public_key, encrypted_key, network, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.ExecContext(ctx, query,
		w.ID,
		w.PublicKey,
		w.EncryptedKey,
		w.Network,
		w.CreatedAt,
		w.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	return nil
}

func (r *PostgresWalletRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet.Wallet, error) {
	query := `
        SELECT id, public_key, encrypted_key, network, created_at, updated_at
        FROM wallets
        WHERE id = $1
    `

	w := &wallet.Wallet{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID,
		&w.PublicKey,
		&w.EncryptedKey,
		&w.Network,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	return w, nil
}

func (r *PostgresWalletRepository) List(ctx context.Context, limit, offset int) ([]*wallet.Wallet, error) {
	query := `
        SELECT id, public_key, encrypted_key, network, created_at, updated_at
        FROM wallets
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}
	defer rows.Close()

	wallets := make([]*wallet.Wallet, 0)
	for rows.Next() {
		w := &wallet.Wallet{}
		if err := rows.Scan(
			&w.ID,
			&w.PublicKey,
			&w.EncryptedKey,
			&w.Network,
			&w.CreatedAt,
			&w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}

	return wallets, nil
}

func (r *PostgresWalletRepository) FindByPublicKey(ctx context.Context, publicKey string) (*wallet.Wallet, error) {
	query := `
        SELECT id, public_key, encrypted_key, network, created_at, updated_at
        FROM wallets
        WHERE public_key = $1
    `

	w := &wallet.Wallet{}
	err := r.db.QueryRowContext(ctx, query, publicKey).Scan(
		&w.ID,
		&w.PublicKey,
		&w.EncryptedKey,
		&w.Network,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	return w, nil
}

func (r *PostgresWalletRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM wallets`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count wallets: %w", err)
	}

	return count, nil
}
