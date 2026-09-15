package storage

type MWallet struct {
	ID      string `json:"wallet_id"`
	Balance int64  `json:"balance"`
}
