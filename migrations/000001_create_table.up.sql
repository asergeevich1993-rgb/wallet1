CREATE TABLE IF NOT EXISTS wallets (
    wallet_uuid UUID PRIMARY KEY,
    balance     BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);