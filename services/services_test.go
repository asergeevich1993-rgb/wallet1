package services

import (
	"context"
	"errors"
	"testing"

	"wallet/storage"
)

type mockStorage struct {
	createWalletResult string
	createWalletErr    error

	getBalanceResult storage.MWallet
	getBalanceErr    error

	depositResult int64
	depositErr    error

	withdrawResult int64
	withdrawErr    error
}

func (m *mockStorage) CreateWallet(ctx context.Context, id string) (string, error) {
	return m.createWalletResult, m.createWalletErr
}

func (m *mockStorage) GetBalance(ctx context.Context, id string) (storage.MWallet, error) {
	return m.getBalanceResult, m.getBalanceErr
}

func (m *mockStorage) Deposit(ctx context.Context, id string, amount int64) (int64, error) {
	return m.depositResult, m.depositErr
}

func (m *mockStorage) Withdraw(ctx context.Context, id string, amount int64) (int64, error) {
	return m.withdrawResult, m.withdrawErr
}

func TestChangeBalance(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		amount      int64
		mockStorage *mockStorage
		wantErr     error
		wantBalance int64
	}{
		{
			name:        "успешный депозит",
			operation:   "DEPOSIT",
			amount:      100,
			mockStorage: &mockStorage{depositResult: 1100},
			wantBalance: 1100,
		},
		{
			name:        "депозит с нулевой суммой",
			operation:   "DEPOSIT",
			amount:      0,
			mockStorage: &mockStorage{},
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "депозит с отрицательной суммой",
			operation:   "DEPOSIT",
			amount:      -50,
			mockStorage: &mockStorage{},
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "успешный вывод",
			operation:   "WITHDRAW",
			amount:      50,
			mockStorage: &mockStorage{withdrawResult: 450},
			wantBalance: 450,
		},
		{
			name:        "вывод с нулевой суммой",
			operation:   "WITHDRAW",
			amount:      0,
			mockStorage: &mockStorage{},
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "неизвестная операция",
			operation:   "TRANSFER",
			amount:      100,
			mockStorage: &mockStorage{},
			wantErr:     ErrInvalidOperation,
		},
		{
			name:        "storage вернул ошибку при депозите",
			operation:   "DEPOSIT",
			amount:      100,
			mockStorage: &mockStorage{depositErr: storage.ErrNotFound},
			wantErr:     storage.ErrNotFound,
		},
		{
			name:        "storage вернул ошибку недостатка средств",
			operation:   "WITHDRAW",
			amount:      99999,
			mockStorage: &mockStorage{withdrawErr: storage.ErrLowBalance},
			wantErr:     storage.ErrLowBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wo := NewWalett(tt.mockStorage)

			balance, err := wo.ChangeBalance(context.Background(), "some-id", tt.operation, tt.amount)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ожидали ошибку %v, получили %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("не ожидали ошибку, получили: %v", err)
			}
			if balance != tt.wantBalance {
				t.Errorf("ожидали баланс %d, получили %d", tt.wantBalance, balance)
			}
		})
	}
}

func TestCreateWallet(t *testing.T) {
	ms := &mockStorage{createWalletResult: "11111111-1111-1111-1111-111111111111"}
	wo := NewWalett(ms)

	id, err := wo.CreateWallet(context.Background())

	if err != nil {
		t.Fatalf("не ожидали ошибку, получили: %v", err)
	}
	if id != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("ожидали id 11111111-1111-1111-1111-111111111111, получили %s", id)
	}
}

func TestCreateWallet_StorageError(t *testing.T) {
	ms := &mockStorage{createWalletErr: errors.New("db down")}
	wo := NewWalett(ms)

	_, err := wo.CreateWallet(context.Background())

	if err == nil {
		t.Error("ожидали ошибку, получили nil")
	}
}

func TestGetBalance(t *testing.T) {
	ms := &mockStorage{
		getBalanceResult: storage.MWallet{ID: "some-id", Balance: 700},
	}
	wo := NewWalett(ms)

	wallet, err := wo.GetBalance(context.Background(), "some-id")

	if err != nil {
		t.Fatalf("не ожидали ошибку, получили: %v", err)
	}
	if wallet.Balance != 700 {
		t.Errorf("ожидали баланс 700, получили %d", wallet.Balance)
	}
}
func TestGetBalance_NotFound(t *testing.T) {
	ms := &mockStorage{getBalanceErr: storage.ErrNotFound}
	wo := NewWalett(ms)

	_, err := wo.GetBalance(context.Background(), "unknown-id")

	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("ожидали ErrNotFound, получили %v", err)
	}
}
