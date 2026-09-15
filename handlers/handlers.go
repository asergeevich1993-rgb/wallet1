package handler

import (
	"encoding/json"
	"net/http"
	"wallet/services"

	"github.com/google/uuid"
)

type Handlers struct {
	services services.WalletOperationInt
}

func (hh *Handlers) HandleCreateWallet(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	uuid, err := hh.services.CreateWallet(rctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"Ваш uuid": uuid})

}
func (hh *Handlers) HandleChangeBalance(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	var dto DTOWallet
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := dto.ValidateForCreate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	balance, err := hh.services.ChangeBalance(rctx, dto.ID, dto.Operation, dto.Amount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"Ваш баланс изменен": balance})
}
func (hh *Handlers) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	strid := r.PathValue("WALLET_UUID")
	id, err := uuid.Parse(strid)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	wallet, err := hh.services.GetBalance(rctx, id.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wallet)
}

func writeError(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(err)
}
