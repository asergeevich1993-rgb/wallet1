package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"wallet/services"
	"wallet/storage"
)

type mockService struct {
	createWalletResult string
	createWalletErr    error

	changeBalanceResult int64
	changeBalanceErr    error

	getBalanceResult storage.MWallet
	getBalanceErr    error
}

func (m *mockService) CreateWallet(ctx context.Context) (string, error) {
	return m.createWalletResult, m.createWalletErr
}

func (m *mockService) ChangeBalance(ctx context.Context, id, operation string, amount int64) (int64, error) {
	return m.changeBalanceResult, m.changeBalanceErr
}

func (m *mockService) GetBalance(ctx context.Context, id string) (storage.MWallet, error) {
	return m.getBalanceResult, m.getBalanceErr
}

func TestHandleChangeBalance_Success(t *testing.T) {
	mock := &mockService{
		changeBalanceResult: 1100,
		changeBalanceErr:    nil,
	}
	hh := CreateHandlers(mock)

	reqBody := `{"wallet_id":"11111111-1111-1111-1111-111111111111","operationType":"DEPOSIT","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидали статус 200, получили %d, тело: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}
	found := false
	for _, v := range resp {
		if v == 1100 {
			found = true
		}
	}
	if !found {
		t.Errorf("ожидали баланс 1100 в ответе, получили: %v", resp)
	}
}

func TestHandleChangeBalance_InvalidJSON(t *testing.T) {
	mock := &mockService{}
	hh := CreateHandlers(mock)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(`{не json`))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", rec.Code)
	}
}

func TestHandleChangeBalance_EmptyWalletID(t *testing.T) {
	mock := &mockService{}
	hh := CreateHandlers(mock)

	reqBody := `{"wallet_id":"","operationType":"DEPOSIT","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", rec.Code)
	}
}

func TestHandleChangeBalance_InvalidUUID(t *testing.T) {
	mock := &mockService{}
	hh := CreateHandlers(mock)

	reqBody := `{"wallet_id":"not-a-uuid","operationType":"DEPOSIT","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", rec.Code)
	}
}

func TestHandleChangeBalance_NegativeAmount(t *testing.T) {
	mock := &mockService{}
	hh := CreateHandlers(mock)

	reqBody := `{"wallet_id":"11111111-1111-1111-1111-111111111111","operationType":"DEPOSIT","amount":-100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", rec.Code)
	}
}

func TestHandleChangeBalance_InvalidAmountFromService(t *testing.T) {
	mock := &mockService{
		changeBalanceErr: services.ErrInvalidAmount,
	}
	hh := CreateHandlers(mock)

	reqBody := `{"wallet_id":"11111111-1111-1111-1111-111111111111","operationType":"DEPOSIT","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", rec.Code)
	}
}

func TestHandleChangeBalance_WalletNotFound(t *testing.T) {
	mock := &mockService{
		changeBalanceErr: storage.ErrNotFound,
	}
	hh := CreateHandlers(mock)
	reqBody := `{"wallet_id":"11111111-1111-1111-1111-111111111111","operationType":"DEPOSIT","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("ожидали статус 404, получили %d", rec.Code)
	}
}

func TestHandleChangeBalance_LowBalance(t *testing.T) {
	mock := &mockService{
		changeBalanceErr: storage.ErrLowBalance,
	}
	hh := CreateHandlers(mock)

	reqBody := `{"wallet_id":"11111111-1111-1111-1111-111111111111","operationType":"WITHDRAW","amount":999999}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("ожидали статус 422, получили %d", rec.Code)
	}
}

func TestHandleChangeBalance_InvalidOperation(t *testing.T) {
	mock := &mockService{
		changeBalanceErr: services.ErrInvalidOperation,
	}
	hh := CreateHandlers(mock)

	reqBody := `{"wallet_id":"11111111-1111-1111-1111-111111111111","operationType":"TRANSFER","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", rec.Code)
	}
}

func TestHandleChangeBalance_InternalError(t *testing.T) {
	mock := &mockService{
		changeBalanceErr: errors.New("непредвиденная ошибка базы"),
	}
	hh := CreateHandlers(mock)

	reqBody := `{"wallet_id":"11111111-1111-1111-1111-111111111111","operationType":"DEPOSIT","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	hh.HandleChangeBalance(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("ожидали статус 500, получили %d", rec.Code)
	}
}
