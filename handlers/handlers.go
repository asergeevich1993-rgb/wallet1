package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"wallet/services"
	"wallet/storage"

	"github.com/google/uuid"
)

type HandlersInt interface {
	HandleCreateWallet(w http.ResponseWriter, r *http.Request)
	HandleChangeBalance(w http.ResponseWriter, r *http.Request)
	HandleGetBalance(w http.ResponseWriter, r *http.Request)
}

type Handlers struct {
	services services.WalletOperationInt
}

func CreateHandlers(wo services.WalletOperationInt) *Handlers {
	return &Handlers{
		services: wo,
	}
}

func (hh *Handlers) HandleCreateWallet(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	uuid, err := hh.services.CreateWallet(rctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"uuid": uuid})

}
func (hh *Handlers) HandleChangeBalance(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	var dto DTOWallet
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := dto.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	balance, err := hh.services.ChangeBalance(rctx, dto.ID, dto.Operation, dto.Amount)
	if err != nil {
		if errors.Is(err, services.ErrInvalidAmount) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		} else if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		} else if errors.Is(err, storage.ErrLowBalance) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		} else if errors.Is(err, services.ErrInvalidOperation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"new balance": balance})
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
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
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
