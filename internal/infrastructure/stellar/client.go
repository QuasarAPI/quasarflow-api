package stellar

import (
	"context"
	"fmt"
	"time"

	"quasarflow-api/internal/domain/stellar"

	"github.com/shopspring/decimal"
	"github.com/stellar/go/clients/horizonclient"
)

type Client struct {
	horizon *horizonclient.Client
}

func NewClient(horizonURL string) *Client {
	return &Client{
		horizon: &horizonclient.Client{
			HorizonURL: horizonURL,
		},
	}
}

func (c *Client) GetAccountBalances(ctx context.Context, publicKey string) ([]stellar.Balance, error) {
	accountRequest := horizonclient.AccountRequest{
		AccountID: publicKey,
	}

	account, err := c.horizon.AccountDetail(accountRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	balances := make([]stellar.Balance, 0, len(account.Balances))
	for _, b := range account.Balances {
		balance := stellar.Balance{
			AssetType:   b.Asset.Type,
			AssetCode:   b.Asset.Code,
			AssetIssuer: b.Asset.Issuer,
			Amount:      decimal.RequireFromString(b.Balance),
		}

		if b.Limit != "" {
			limit := decimal.RequireFromString(b.Limit)
			balance.Limit = &limit
		}

		balances = append(balances, balance)
	}

	return balances, nil
}

func (c *Client) GetHorizonClient() *horizonclient.Client {
	return c.horizon
}

// GetTransaction retrieves a transaction by hash
func (c *Client) GetTransaction(transactionHash string) (*TransactionInfo, error) {
	tx, err := c.horizon.TransactionDetail(transactionHash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}

	return &TransactionInfo{
		Hash:            tx.Hash,
		SourceAccount:   tx.Account,
		LedgerCloseTime: tx.LedgerCloseTime,
		Memo:            tx.Memo,
	}, nil
}

// GetAccount retrieves account details
func (c *Client) GetAccount(publicKey string) (*AccountInfo, error) {
	accountRequest := horizonclient.AccountRequest{
		AccountID: publicKey,
	}

	account, err := c.horizon.AccountDetail(accountRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	// Handle nullable LastModifiedTime
	var lastModified time.Time
	if account.LastModifiedTime != nil {
		lastModified = *account.LastModifiedTime
	}

	return &AccountInfo{
		AccountID:        account.AccountID,
		LastModifiedTime: lastModified,
		Sequence:         fmt.Sprintf("%d", account.Sequence),
	}, nil
}

// TransactionInfo represents simplified transaction information
type TransactionInfo struct {
	Hash            string
	SourceAccount   string
	LedgerCloseTime time.Time
	Memo            string
}

// AccountInfo represents simplified account information
type AccountInfo struct {
	AccountID        string
	LastModifiedTime time.Time
	Sequence         string
}
