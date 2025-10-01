package stellar

import (
	"context"
	"fmt"

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
