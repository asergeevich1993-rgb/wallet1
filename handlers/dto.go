package handler

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type DTOWallet struct {
	ID        string `json:"wallet_id"`
	Operation string `json:"operationType"`
	Amount    int64  `json:"amount"`
}

func (dw *DTOWallet) Validate() error {
	dw.ID = strings.TrimSpace(dw.ID)
	dw.Operation = strings.TrimSpace(dw.Operation)

	if dw.ID == "" {
		return errors.New("uuid is empty")
	}
	_, err := uuid.Parse(dw.ID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	if dw.Operation == "" {
		return errors.New("operation is empty")
	}
	if dw.Amount <= 0 {
		return errors.New("invalid amount")
	}
	return nil
}
