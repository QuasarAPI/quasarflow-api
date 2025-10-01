package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"quasarflow-api/internal/interface/http/response"
	"quasarflow-api/internal/usecase/wallet"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type WalletHandler struct {
	createWallet *wallet.CreateWalletUseCase
	getWallet    *wallet.GetWalletUseCase
	getBalance   *wallet.GetBalanceUseCase
	listWallets  *wallet.ListWalletsUseCase
}

func NewWalletHandler(
	createWallet *wallet.CreateWalletUseCase,
	getWallet *wallet.GetWalletUseCase,
	getBalance *wallet.GetBalanceUseCase,
	listWallets *wallet.ListWalletsUseCase,
) *WalletHandler {
	return &WalletHandler{
		createWallet: createWallet,
		getWallet:    getWallet,
		getBalance:   getBalance,
		listWallets:  listWallets,
	}
}

func (h *WalletHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input wallet.CreateWalletInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	output, err := h.createWallet.Execute(r.Context(), input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, output)
}

func (h *WalletHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid wallet id")
		return
	}

	output, err := h.getWallet.Execute(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.Success(w, http.StatusOK, output)
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid wallet id")
		return
	}

	output, err := h.getBalance.Execute(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, http.StatusOK, output)
}

func (h *WalletHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 10
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	output, err := h.listWallets.Execute(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, http.StatusOK, output)
}
