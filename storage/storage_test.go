//go:build integration

package storage

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

func TestConcurrentDeposit(t *testing.T) {
	testDSN := "postgres://postgres:pass@127.0.0.1:5432/walletdb?sslmode=disable"

	strg, err := NewDataBase(context.Background(), testDSN)
	if err != nil {
		t.Fatalf("неудалось подключиться к БД %v", err)
	}
	var wg sync.WaitGroup
	var operations int64 = 1000
	var amount int64 = 10
	wantBalance := operations * amount
	uuId := uuid.New().String()
	_, err = strg.CreateWallet(context.Background(), uuId)
	if err != nil {
		t.Errorf("неожидали ошибку в создании кошелька %v", err)
	}

	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()
			_, err := strg.Deposit(context.Background(), uuId, amount)
			if err != nil {
				t.Errorf("Неожидали ошибку при депозите %v на итерации %d", err, iteration)
			}
		}(i)
	}
	wg.Wait()

	wallet, err := strg.GetBalance(context.Background(), uuId)
	if err != nil {
		t.Errorf("Неожидали ошибку в выводе баланса %v", err)
	}

	if wantBalance != wallet.Balance {
		t.Errorf("Неожидали ошибку в балансе %d", wallet.Balance)
	}

}
func TestConcurrentWithdraw(t *testing.T) {
	testDSN := "postgres://postgres:pass@127.0.0.1:5432/walletdb?sslmode=disable"
	strg, err := NewDataBase(context.Background(), testDSN)
	if err != nil {
		t.Fatalf("неудалось подключиться к БД %v", err)
	}
	var wg sync.WaitGroup
	var operations int64 = 1000
	var amount int64 = 1500
	uuId := uuid.New().String()
	_, err = strg.CreateWallet(context.Background(), uuId)
	if err != nil {
		t.Fatalf("неожидали ошибку в создании кошелька %v", err)
	}
	_, err = strg.Deposit(context.Background(), uuId, amount)
	if err != nil {
		t.Fatalf("Неожидали ошибку при депозите %v", err)
	}

	var withdraw int64 = 1
	wantBalance := amount - operations*withdraw
	var i int64 = 1
	for i = 1; i <= operations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := strg.Withdraw(context.Background(), uuId, withdraw)
			if err != nil {
				t.Errorf("Неожидали ошибку в выводе %v", err)
			}
		}()
	}
	wg.Wait()

	wallet, err := strg.GetBalance(context.Background(), uuId)
	if err != nil {
		t.Errorf("Неожидали ошибку в выводе баланса %v", err)
	}

	if wantBalance != wallet.Balance {
		t.Errorf("Неожидали ошибку в балансе %d", wallet.Balance)
	}

}
func TestConcurrentWithdrawWithErr(t *testing.T) {
	testDSN := "postgres://postgres:pass@127.0.0.1:5432/walletdb?sslmode=disable"
	strg, err := NewDataBase(context.Background(), testDSN)
	if err != nil {
		t.Fatalf("неудалось подключиться к БД %v", err)
	}
	var wg sync.WaitGroup
	var operations int64 = 1000
	var amount int64 = 500
	uuId := uuid.New().String()
	_, err = strg.CreateWallet(context.Background(), uuId)
	if err != nil {
		t.Fatalf("неожидали ошибку в создании кошелька %v", err)
	}
	_, err = strg.Deposit(context.Background(), uuId, amount)
	if err != nil {
		t.Fatalf("Неожидали ошибку при депозите %v", err)
	}

	var withdraw int64 = 1
	var i int64 = 1
	var success, rejected atomic.Int64
	for i = 1; i <= operations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := strg.Withdraw(context.Background(), uuId, withdraw)
			if err == nil {
				success.Add(1)
			} else if errors.Is(err, ErrLowBalance) {
				rejected.Add(1)
			} else {
				t.Errorf("Неожиданная ошибка: %v", err)
			}
		}()
	}
	wg.Wait()

	if success.Load() != 500 {
		t.Errorf("Успехов: ожидали 500 получили %d", success.Load())
	}
	if rejected.Load() != 500 {
		t.Errorf("Отказов: ожидали 500 получили %d", rejected.Load())
	}
	wallet, err := strg.GetBalance(context.Background(), uuId)
	if err != nil {
		t.Fatalf("Неожидали ошибку в выводе баланса %v", err)
	}

	if wallet.Balance != 0 {
		t.Errorf("Ожидали 0 получили %d", wallet.Balance)
	}

}
