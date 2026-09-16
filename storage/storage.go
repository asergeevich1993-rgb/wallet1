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
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	var mw MWallet
	sql := `SELECT wallet_uuid,balance FROM wallets WHERE wallet_uuid = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, sql, id).Scan(&mw.ID, &mw.Balance)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	newbalance := mw.Balance + amount
	var balance int64
	sqlu := `UPDATE wallets SET balance=$1 WHERE wallet_uuid = $2 RETURNING balance`
	err = tx.QueryRow(ctx, sqlu, newbalance, mw.ID).Scan(&balance)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return balance, nil
}

func (s *Storage) Withdraw(ctx context.Context, id string, amount int64) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	var mw MWallet
	sql := `SELECT wallet_uuid,balance FROM wallets WHERE wallet_uuid = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, sql, id).Scan(&mw.ID, &mw.Balance)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	if mw.Balance < amount {
		tx.Rollback(ctx)
		return 0, ErrLowBalance
	}
	newbalance := mw.Balance - amount

	var balance int64
	sqlu := `UPDATE wallets SET balance=$1 WHERE wallet_uuid = $2 RETURNING balance`
	err = tx.QueryRow(ctx, sqlu, newbalance, mw.ID).Scan(&balance)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return balance, nil
}
