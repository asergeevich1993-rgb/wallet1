package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StorageInt interface {
	CreateWallet(ctx context.Context, id string) (string, error)
	GetBalance(ctx context.Context, id string) (MWallet, error)
	Deposit(ctx context.Context, id string, amount int64) (int64, error)
	Withdraw(ctx context.Context, id string, amount int64) (int64, error)
}

type Storage struct {
	pool *pgxpool.Pool
}

func NewDataBase(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Storage{
		pool: pool,
	}, nil
}

func (s *Storage) CreateWallet(ctx context.Context, id string) (string, error) {
	var uuid string
	sql := `INSERT INTO wallets (wallet_uuid,balance) VALUES ($1,0) RETURNING wallet_uuid`
	err := s.pool.QueryRow(ctx, sql, id).Scan(&uuid)
	if err != nil {
		return "", err
	}
	return uuid, nil
}

func (s *Storage) GetBalance(ctx context.Context, id string) (MWallet, error) {
	var mw MWallet
	sql := `SELECT wallet_uuid,balance FROM wallets WHERE wallet_uuid = $1`
	err := s.pool.QueryRow(ctx, sql, id).Scan(&mw.ID, &mw.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MWallet{}, ErrNotFound
		}
		return MWallet{}, err
	}
	return mw, nil

}
func (s *Storage) Deposit(ctx context.Context, id string, amount int64) (int64, error) {

	var mw MWallet
	query := `UPDATE wallets SET balance=balance+$1 WHERE wallet_uuid = $2 RETURNING balance`
	err := s.pool.QueryRow(ctx, query, amount, id).Scan(&mw.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	return mw.Balance, nil
}

func (s *Storage) Withdraw(ctx context.Context, id string, amount int64) (int64, error) {

	var mw MWallet
	sql := `UPDATE wallets SET balance=balance-$1 WHERE wallet_uuid = $2 AND balance >=$1 RETURNING balance`
	err := s.pool.QueryRow(ctx, sql, amount, id).Scan(&mw.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			exists, checkErr := s.walletExists(ctx, id)
			if checkErr != nil {
				return 0, checkErr
			}
			if !exists {
				return 0, ErrNotFound
			}
			return 0, ErrLowBalance
		}
		return 0, err
	}

	return mw.Balance, nil
}

func (s *Storage) walletExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM wallets WHERE wallet_uuid=$1)`
	err := s.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return exists, err
	}
	return exists, nil
}
