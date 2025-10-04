package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"quasarflow-api/internal/interface/http/response"
	"quasarflow-api/internal/usecase/wallet"
	pkgErrors "quasarflow-api/pkg/errors"
	"quasarflow-api/pkg/logger"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// AccountHandler handles account-related operations including ownership verification
type AccountHandler struct {
	verifyOwnership    *wallet.VerifyOwnershipUseCase
	getBalance         *wallet.GetBalanceUseCase
	getTransactionHist *wallet.GetTransactionHistoryUseCase
	logger             logger.Logger
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(
	verifyOwnership *wallet.VerifyOwnershipUseCase,
	getBalance *wallet.GetBalanceUseCase,
	getTransactionHist *wallet.GetTransactionHistoryUseCase,
	logger logger.Logger,
) *AccountHandler {
	return &AccountHandler{
		verifyOwnership:    verifyOwnership,
		getBalance:         getBalance,
		getTransactionHist: getTransactionHist,
		logger:             logger,
	}
}

// GetChallenge generates a SEP-10 compliant challenge for ownership verification
// GET /api/v1/accounts/{public_key}/challenge
func (h *AccountHandler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	publicKey := mux.Vars(r)["public_key"]

	// Validate public key format
	if !h.isValidStellarPublicKey(publicKey) {
		h.logger.Warn("invalid public key format",
			zap.String("public_key", publicKey),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, "Invalid Stellar public key format")
		return
	}

	challenge := h.verifyOwnership.GenerateChallenge(publicKey)

	h.logger.Info("challenge generated",
		zap.String("public_key", publicKey),
		zap.String("challenge", challenge.Challenge),
		zap.String("ip", r.RemoteAddr))

	response.Success(w, http.StatusOK, challenge)
}

// VerifyOwnership verifies wallet ownership using SEP-10 message signing
// POST /api/v1/accounts/{public_key}/verify-ownership
func (h *AccountHandler) VerifyOwnership(w http.ResponseWriter, r *http.Request) {
	publicKey := mux.Vars(r)["public_key"]

	var input struct {
		Signature string `json:"signature"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request body for verify ownership",
			zap.String("public_key", publicKey),
			zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if input.Signature == "" || input.Message == "" {
		h.logger.Warn("missing required fields for verify ownership",
			zap.String("public_key", publicKey),
			zap.Bool("has_signature", input.Signature != ""),
			zap.Bool("has_message", input.Message != ""))
		response.Error(w, http.StatusBadRequest, "signature and message are required")
		return
	}

	verifyInput := wallet.VerifyOwnershipInput{
		PublicKey: publicKey,
		Signature: input.Signature,
		Message:   input.Message,
	}

	output, err := h.verifyOwnership.Execute(r.Context(), verifyInput)
	if err != nil {
		h.handleUseCaseError(w, r, err, "verify_ownership")
		return
	}

	statusCode := http.StatusOK
	if !output.IsOwner {
		statusCode = http.StatusUnauthorized
	}

	h.logger.Info("ownership verification completed",
		zap.String("public_key", publicKey),
		zap.Bool("is_owner", output.IsOwner),
		zap.Int("status_code", statusCode),
		zap.String("ip", r.RemoteAddr))

	response.Success(w, statusCode, output)
}

// VerifyOwnershipByTransaction verifies ownership via a signed transaction
// POST /api/v1/accounts/{public_key}/verify-transaction
func (h *AccountHandler) VerifyOwnershipByTransaction(w http.ResponseWriter, r *http.Request) {
	publicKey := mux.Vars(r)["public_key"]

	var input struct {
		TransactionHash string `json:"transaction_hash"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request body for verify transaction",
			zap.String("public_key", publicKey),
			zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if input.TransactionHash == "" {
		h.logger.Warn("missing transaction hash for verify transaction",
			zap.String("public_key", publicKey))
		response.Error(w, http.StatusBadRequest, "transaction_hash is required")
		return
	}

	output, err := h.verifyOwnership.VerifyOwnershipByTransaction(r.Context(), publicKey, input.TransactionHash)
	if err != nil {
		h.handleUseCaseError(w, r, err, "verify_ownership_transaction")
		return
	}

	statusCode := http.StatusOK
	if !output.IsOwner {
		statusCode = http.StatusUnauthorized
	}

	h.logger.Info("transaction ownership verification completed",
		zap.String("public_key", publicKey),
		zap.String("transaction_hash", input.TransactionHash),
		zap.Bool("is_owner", output.IsOwner),
		zap.Int("status_code", statusCode),
		zap.String("ip", r.RemoteAddr))

	response.Success(w, statusCode, output)
}

// VerifyOwnershipByAccount verifies ownership by checking account existence and activity
// GET /api/v1/accounts/{public_key}/verify-account
func (h *AccountHandler) VerifyOwnershipByAccount(w http.ResponseWriter, r *http.Request) {
	publicKey := mux.Vars(r)["public_key"]

	// Validate public key format
	if !h.isValidStellarPublicKey(publicKey) {
		h.logger.Warn("invalid public key format for account verification",
			zap.String("public_key", publicKey),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, "Invalid Stellar public key format")
		return
	}

	output, err := h.verifyOwnership.VerifyOwnershipByAccount(r.Context(), publicKey)
	if err != nil {
		h.handleUseCaseError(w, r, err, "verify_ownership_account")
		return
	}

	statusCode := http.StatusOK
	if !output.IsOwner {
		statusCode = http.StatusUnauthorized
	}

	h.logger.Info("account ownership verification completed",
		zap.String("public_key", publicKey),
		zap.Bool("is_owner", output.IsOwner),
		zap.Int("status_code", statusCode),
		zap.String("ip", r.RemoteAddr))

	response.Success(w, statusCode, output)
}

// GetAccountBalance retrieves the balance for a public key (external wallet)
// GET /api/v1/accounts/{public_key}/balance
func (h *AccountHandler) GetAccountBalance(w http.ResponseWriter, r *http.Request) {
	publicKey := mux.Vars(r)["public_key"]

	// Validate public key format
	if !h.isValidStellarPublicKey(publicKey) {
		h.logger.Warn("invalid public key format for balance",
			zap.String("public_key", publicKey),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, "Invalid Stellar public key format")
		return
	}

	// For now, return a simple response indicating this endpoint needs implementation
	// TODO: Implement direct balance fetching using stellar client
	h.logger.Info("account balance requested (not implemented yet)",
		zap.String("public_key", publicKey),
		zap.String("ip", r.RemoteAddr))

	response.Success(w, http.StatusOK, map[string]interface{}{
		"public_key": publicKey,
		"message":    "Balance endpoint not yet implemented for external wallets",
		"note":       "Use the wallet endpoints for registered wallets",
	})
}

// GetAccountTransactionHistory retrieves transaction history for a public key (external wallet)
// GET /api/v1/accounts/{public_key}/transactions
func (h *AccountHandler) GetAccountTransactionHistory(w http.ResponseWriter, r *http.Request) {
	publicKey := mux.Vars(r)["public_key"]

	// Validate public key format
	if !h.isValidStellarPublicKey(publicKey) {
		h.logger.Warn("invalid public key format for transaction history",
			zap.String("public_key", publicKey),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, "Invalid Stellar public key format")
		return
	}

	// Parse query parameters
	limit := h.parseQueryLimit(r, 10)
	offset := h.parseQueryOffset(r, 0)

	// For now, return a simple response indicating this endpoint needs implementation
	// TODO: Implement direct transaction history fetching using stellar client
	h.logger.Info("account transaction history requested (not implemented yet)",
		zap.String("public_key", publicKey),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("ip", r.RemoteAddr))

	response.Success(w, http.StatusOK, map[string]interface{}{
		"public_key": publicKey,
		"limit":      limit,
		"offset":     offset,
		"message":    "Transaction history endpoint not yet implemented for external wallets",
		"note":       "Use the wallet endpoints for registered wallets",
	})
}

// Helper methods

// isValidStellarPublicKey validates Stellar public key format
func (h *AccountHandler) isValidStellarPublicKey(publicKey string) bool {
	return len(publicKey) == 56 && publicKey[0] == 'G'
}

// parseQueryLimit parses the limit query parameter
func (h *AccountHandler) parseQueryLimit(r *http.Request, defaultValue int) int {
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			return limit
		}
	}
	return defaultValue
}

// parseQueryOffset parses the offset query parameter
func (h *AccountHandler) parseQueryOffset(r *http.Request, defaultValue int) int {
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			return offset
		}
	}
	return defaultValue
}

// handleUseCaseError handles errors from use cases with appropriate HTTP status codes
func (h *AccountHandler) handleUseCaseError(w http.ResponseWriter, r *http.Request, err error, operation string) {
	var appErr *pkgErrors.AppError
	if errors.As(err, &appErr) {
		h.logger.Warn("use case error",
			zap.String("operation", operation),
			zap.String("error_type", string(appErr.Type)),
			zap.String("ip", r.RemoteAddr),
			zap.Error(err))
		response.Error(w, appErr.StatusCode, appErr.Message)
		return
	}

	// Handle generic errors
	h.logger.Error("unexpected error",
		zap.String("operation", operation),
		zap.String("ip", r.RemoteAddr),
		zap.Error(err))
	response.Error(w, http.StatusInternalServerError, "Internal server error")
}
