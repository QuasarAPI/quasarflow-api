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

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	// Default pagination values
	defaultLimit  = 10
	defaultOffset = 0

	// Query parameter names
	paramLimit  = "limit"
	paramOffset = "offset"
	paramOrder  = "order"
	paramCursor = "cursor"

	// Order values
	orderAsc  = "asc"
	orderDesc = "desc"

	// Error messages
	errMsgInvalidWalletID    = "invalid wallet id"
	errMsgInvalidRequestBody = "invalid request body"
	errMsgToAddressRequired  = "to_address is required"
	errMsgAmountRequired     = "amount is required"
)

type WalletHandler struct {
	createWallet       *wallet.CreateWalletUseCase
	getWallet          *wallet.GetWalletUseCase
	getBalance         *wallet.GetBalanceUseCase
	listWallets        *wallet.ListWalletsUseCase
	fundWallet         *wallet.FundWalletUseCase
	sendPayment        *wallet.SendPaymentUseCase
	getTransactionHist *wallet.GetTransactionHistoryUseCase
	logger             logger.Logger
}

func NewWalletHandler(
	createWallet *wallet.CreateWalletUseCase,
	getWallet *wallet.GetWalletUseCase,
	getBalance *wallet.GetBalanceUseCase,
	listWallets *wallet.ListWalletsUseCase,
	fundWallet *wallet.FundWalletUseCase,
	sendPayment *wallet.SendPaymentUseCase,
	getTransactionHist *wallet.GetTransactionHistoryUseCase,
	logger logger.Logger,
) *WalletHandler {
	return &WalletHandler{
		createWallet:       createWallet,
		getWallet:          getWallet,
		getBalance:         getBalance,
		listWallets:        listWallets,
		fundWallet:         fundWallet,
		sendPayment:        sendPayment,
		getTransactionHist: getTransactionHist,
		logger:             logger,
	}
}

// validateWalletID validates and parses wallet ID from URL parameters
func (h *WalletHandler) validateWalletID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.logger.Warn("invalid wallet id",
			zap.String("wallet_id", vars["id"]),
			zap.String("ip", r.RemoteAddr),
			zap.Error(err))
		response.Error(w, http.StatusBadRequest, errMsgInvalidWalletID)
		return uuid.Nil, false
	}
	return id, true
}

// parseQueryLimit parses limit parameter from query string
func (h *WalletHandler) parseQueryLimit(r *http.Request, defaultValue int) int {
	if limitStr := r.URL.Query().Get(paramLimit); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			return parsedLimit
		}
	}
	return defaultValue
}

// parseQueryOffset parses offset parameter from query string
func (h *WalletHandler) parseQueryOffset(r *http.Request, defaultValue int) int {
	if offsetStr := r.URL.Query().Get(paramOffset); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			return parsedOffset
		}
	}
	return defaultValue
}

// handleUseCaseError handles errors from use cases with appropriate HTTP status codes
func (h *WalletHandler) handleUseCaseError(w http.ResponseWriter, r *http.Request, err error, operation string) {
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

func (h *WalletHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input wallet.CreateWalletInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request body",
			zap.String("ip", r.RemoteAddr),
			zap.Error(err))
		response.Error(w, http.StatusBadRequest, errMsgInvalidRequestBody)
		return
	}

	output, err := h.createWallet.Execute(r.Context(), input)
	if err != nil {
		h.handleUseCaseError(w, r, err, "create_wallet")
		return
	}

	h.logger.Info("wallet created successfully",
		zap.String("wallet_id", output.ID),
		zap.String("ip", r.RemoteAddr))
	response.Success(w, http.StatusCreated, output)
}

func (h *WalletHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := h.validateWalletID(w, r)
	if !ok {
		return
	}

	output, err := h.getWallet.Execute(r.Context(), id)
	if err != nil {
		h.handleUseCaseError(w, r, err, "get_wallet")
		return
	}

	response.Success(w, http.StatusOK, output)
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	id, ok := h.validateWalletID(w, r)
	if !ok {
		return
	}

	output, err := h.getBalance.Execute(r.Context(), id)
	if err != nil {
		h.handleUseCaseError(w, r, err, "get_balance")
		return
	}

	response.Success(w, http.StatusOK, output)
}

func (h *WalletHandler) Fund(w http.ResponseWriter, r *http.Request) {
	id, ok := h.validateWalletID(w, r)
	if !ok {
		return
	}

	var input wallet.FundWalletInput
	input.WalletID = id

	// Parse optional amount from request body
	if r.Body != nil {
		var body struct {
			Amount string `json:"amount,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.logger.Warn("invalid request body for fund wallet",
				zap.String("wallet_id", id.String()),
				zap.String("ip", r.RemoteAddr),
				zap.Error(err))
			response.Error(w, http.StatusBadRequest, errMsgInvalidRequestBody)
			return
		}
		input.Amount = body.Amount
	}

	output, err := h.fundWallet.Execute(r.Context(), input)
	if err != nil {
		h.handleUseCaseError(w, r, err, "fund_wallet")
		return
	}

	if !output.Success {
		response.Error(w, http.StatusBadRequest, output.Message)
		return
	}

	h.logger.Info("wallet funded successfully",
		zap.String("wallet_id", id.String()),
		zap.String("amount", input.Amount),
		zap.String("ip", r.RemoteAddr))
	response.Success(w, http.StatusOK, output)
}

func (h *WalletHandler) SendPayment(w http.ResponseWriter, r *http.Request) {
	id, ok := h.validateWalletID(w, r)
	if !ok {
		return
	}

	var input wallet.SendPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request body for send payment",
			zap.String("wallet_id", id.String()),
			zap.String("ip", r.RemoteAddr),
			zap.Error(err))
		response.Error(w, http.StatusBadRequest, errMsgInvalidRequestBody)
		return
	}

	// Set the from wallet ID from the URL parameter
	input.FromWalletID = id

	// Validate required fields
	if input.ToAddress == "" {
		h.logger.Warn("missing to_address in send payment",
			zap.String("wallet_id", id.String()),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, errMsgToAddressRequired)
		return
	}
	if input.Amount == "" {
		h.logger.Warn("missing amount in send payment",
			zap.String("wallet_id", id.String()),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, errMsgAmountRequired)
		return
	}

	output, err := h.sendPayment.Execute(r.Context(), input)
	if err != nil {
		h.handleUseCaseError(w, r, err, "send_payment")
		return
	}

	h.logger.Info("payment sent successfully",
		zap.String("from_wallet_id", id.String()),
		zap.String("to_address", input.ToAddress),
		zap.String("amount", input.Amount),
		zap.String("ip", r.RemoteAddr))
	response.Success(w, http.StatusOK, output)
}

func (h *WalletHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := h.validateWalletID(w, r)
	if !ok {
		return
	}

	input := wallet.GetTransactionHistoryInput{
		WalletID: id,
	}

	// Parse query parameters
	if limitStr := r.URL.Query().Get(paramLimit); limitStr != "" {
		if parsedLimit, err := strconv.ParseUint(limitStr, 10, 32); err == nil && parsedLimit > 0 {
			input.Limit = uint(parsedLimit)
		}
	}

	if order := r.URL.Query().Get(paramOrder); order == orderAsc || order == orderDesc {
		input.Order = order
	}

	if cursor := r.URL.Query().Get(paramCursor); cursor != "" {
		input.Cursor = cursor
	}

	output, err := h.getTransactionHist.Execute(r.Context(), input)
	if err != nil {
		h.handleUseCaseError(w, r, err, "get_transaction_history")
		return
	}

	response.Success(w, http.StatusOK, output)
}

func (h *WalletHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters using helper methods
	limit := h.parseQueryLimit(r, defaultLimit)
	offset := h.parseQueryOffset(r, defaultOffset)

	output, err := h.listWallets.Execute(r.Context(), limit, offset)
	if err != nil {
		h.handleUseCaseError(w, r, err, "list_wallets")
		return
	}

	h.logger.Debug("wallets listed",
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("ip", r.RemoteAddr))
	response.Success(w, http.StatusOK, output)
}
