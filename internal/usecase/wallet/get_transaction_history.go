package wallet

import (
	"context"
	"fmt"
	"time"

	"quasarflow-api/internal/domain/wallet"
	"quasarflow-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon/operations"
)

type TransactionRecord struct {
	ID             string    `json:"id"`
	Hash           string    `json:"hash"`
	Ledger         int32     `json:"ledger"`
	CreatedAt      time.Time `json:"created_at"`
	SourceAccount  string    `json:"source_account"`
	Type           string    `json:"type"`
	TypeI          int32     `json:"type_i"`
	OperationCount int32     `json:"operation_count"`
	Successful     bool      `json:"successful"`
	MaxFee         int64     `json:"max_fee"`
	FeeCharged     int64     `json:"fee_charged"`
	MemoType       string    `json:"memo_type,omitempty"`
	Memo           string    `json:"memo,omitempty"`
}

type PaymentOperation struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	CreatedAt     time.Time `json:"created_at"`
	TransactionID string    `json:"transaction_id"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	Amount        string    `json:"amount"`
	AssetType     string    `json:"asset_type"`
	AssetCode     string    `json:"asset_code,omitempty"`
	AssetIssuer   string    `json:"asset_issuer,omitempty"`
}

type GetTransactionHistoryInput struct {
	WalletID uuid.UUID `json:"wallet_id" validate:"required"`
	Limit    uint      `json:"limit,omitempty"`  // Default: 10, Max: 200
	Order    string    `json:"order,omitempty"`  // "asc" or "desc", default: "desc"
	Cursor   string    `json:"cursor,omitempty"` // For pagination
}

type GetTransactionHistoryOutput struct {
	WalletID     string              `json:"wallet_id"`
	PublicKey    string              `json:"public_key"`
	Network      string              `json:"network"`
	Transactions []TransactionRecord `json:"transactions"`
	Operations   []PaymentOperation  `json:"operations"`
	HasNext      bool                `json:"has_next"`
	NextCursor   string              `json:"next_cursor,omitempty"`
}

type GetTransactionHistoryUseCase struct {
	repo          wallet.Repository
	horizonClient *horizonclient.Client
	logger        logger.Logger
}

func NewGetTransactionHistoryUseCase(
	repo wallet.Repository,
	horizonClient *horizonclient.Client,
	logger logger.Logger,
) *GetTransactionHistoryUseCase {
	return &GetTransactionHistoryUseCase{
		repo:          repo,
		horizonClient: horizonClient,
		logger:        logger,
	}
}

func (uc *GetTransactionHistoryUseCase) Execute(ctx context.Context, input GetTransactionHistoryInput) (*GetTransactionHistoryOutput, error) {
	// 1. Find wallet in database
	w, err := uc.repo.FindByID(ctx, input.WalletID)
	if err != nil {
		uc.logger.Error("failed to find wallet", logger.Error(err))
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// 2. Set default values
	limit := input.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}

	order := input.Order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// 3. Build transaction request
	txRequest := horizonclient.TransactionRequest{
		ForAccount: w.PublicKey,
		Limit:      limit,
		Order:      horizonclient.Order(order),
	}

	if input.Cursor != "" {
		txRequest.Cursor = input.Cursor
	}

	uc.logger.Info("fetching transaction history",
		logger.String("wallet_id", w.ID.String()),
		logger.String("public_key", w.PublicKey),
		logger.Int("limit", int(limit)),
		logger.String("order", order),
	)

	// 4. Get transactions from Horizon
	transactionsPage, err := uc.horizonClient.Transactions(txRequest)
	if err != nil {
		uc.logger.Error("failed to fetch transactions", logger.Error(err))
		return nil, fmt.Errorf("failed to fetch transaction history: %w", err)
	}

	// 5. Convert transactions to our format
	transactions := make([]TransactionRecord, 0, len(transactionsPage.Embedded.Records))
	for _, tx := range transactionsPage.Embedded.Records {
		record := TransactionRecord{
			ID:             tx.ID,
			Hash:           tx.Hash,
			Ledger:         tx.Ledger,
			CreatedAt:      tx.LedgerCloseTime,
			SourceAccount:  "",
			OperationCount: tx.OperationCount,
			Successful:     tx.Successful,
			MaxFee:         tx.MaxFee,
			FeeCharged:     tx.FeeCharged,
			MemoType:       tx.MemoType,
		}

		// Handle memo based on type
		switch tx.MemoType {
		case "text":
			if tx.Memo != "" {
				record.Memo = tx.Memo
			}
		case "id":
			if tx.Memo != "" {
				record.Memo = tx.Memo
			}
		case "hash", "return":
			if tx.Memo != "" {
				record.Memo = tx.Memo
			}
		}

		transactions = append(transactions, record)
	}

	// 6. Get payment operations for more detailed view
	opRequest := horizonclient.OperationRequest{
		ForAccount: w.PublicKey,
		Limit:      limit,
		Order:      horizonclient.Order(order),
	}

	if input.Cursor != "" {
		opRequest.Cursor = input.Cursor
	}

	operationsPage, err := uc.horizonClient.Operations(opRequest)
	if err != nil {
		uc.logger.Warn("failed to fetch operations", logger.Error(err))
		// Continue without operations if this fails
	}

	// 7. Convert payment operations to our format
	paymentOps := make([]PaymentOperation, 0)
	if err == nil {
		for _, op := range operationsPage.Embedded.Records {
			// Only include payment operations
			if payment, ok := op.(operations.Payment); ok {
				paymentOp := PaymentOperation{
					ID:            payment.GetID(),
					Type:          payment.GetType(),
					CreatedAt:     payment.LedgerCloseTime,
					TransactionID: payment.GetTransactionHash(),
					From:          payment.From,
					To:            payment.To,
					Amount:        payment.Amount,
					AssetType:     payment.Asset.Type,
				}

				if payment.Asset.Code != "" {
					paymentOp.AssetCode = payment.Asset.Code
				}
				if payment.Asset.Issuer != "" {
					paymentOp.AssetIssuer = payment.Asset.Issuer
				}

				paymentOps = append(paymentOps, paymentOp)
			}
		}
	}

	// 8. Determine pagination info
	hasNext := len(transactionsPage.Embedded.Records) == int(limit)
	nextCursor := ""
	if hasNext && len(transactions) > 0 {
		nextCursor = transactions[len(transactions)-1].ID
	}

	uc.logger.Info("transaction history fetched successfully",
		logger.String("wallet_id", w.ID.String()),
		logger.Int("transaction_count", len(transactions)),
		logger.Int("payment_count", len(paymentOps)),
		logger.Bool("has_next", hasNext),
	)

	return &GetTransactionHistoryOutput{
		WalletID:     w.ID.String(),
		PublicKey:    w.PublicKey,
		Network:      w.Network,
		Transactions: transactions,
		Operations:   paymentOps,
		HasNext:      hasNext,
		NextCursor:   nextCursor,
	}, nil
}
