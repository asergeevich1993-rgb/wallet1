package services

import (
	"context"
	"wallet/storage"

	"github.com/google/uuid"
)

type WalletOperationInt interface {
	CreateWallet(ctx context.Context) (string, error)
	ChangeBalance(ctx context.Context, id, operation string, amount int64) (int64, error)
	GetBalance(ctx context.Context, id string) (storage.MWallet, error)
}

type WalletOperation struct {
	storage storage.StorageInt
}

func NewWalett(si storage.StorageInt) *WalletOperation {
	return &WalletOperation{
		storage: si,
	}
}

func (wo *WalletOperation) CreateWallet(ctx context.Context) (string, error) {
	uuID := uuid.New().String()
	id, err := wo.storage.CreateWallet(ctx, uuID)
	if err != nil {
		return "", err
	}
	return id, nil

}

func (wo *WalletOperation) ChangeBalance(ctx context.Context, id, operation string, amount int64) (int64, error) {
	switch operation {
	case "DEPOSIT":
		if amount > 0 {
			balance, err := wo.storage.Deposit(ctx, id, amount)
			if err != nil {
				return 0, err
			}
			return balance, nil
		} else {
			return 0, ErrInvalidAmount
		}

	case "WITHDRAW":
		if amount > 0 {
			balance, err := wo.storage.Withdraw(ctx, id, amount)
			if err != nil {
				return 0, err
			}
			return balance, nil
		} else {
			return 0, ErrInvalidAmount
		}

	}
	return 0, ErrInvalidOperation
}

func (wo *WalletOperation) GetBalance(ctx context.Context, id string) (storage.MWallet, error) {

	wallet, err := wo.storage.GetBalance(ctx, id)
	if err != nil {
		return storage.MWallet{}, err
	}

	return wallet, nil
}
