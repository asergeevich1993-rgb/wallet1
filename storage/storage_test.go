//go:build integration

package storage

import (
	"context"
	"sync"
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
