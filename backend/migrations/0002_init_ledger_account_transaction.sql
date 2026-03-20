CREATE TABLE IF NOT EXISTS ledgers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (user_id, is_default)
);

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    initial_balance NUMERIC NOT NULL,
    archived_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    ledger_id UUID NOT NULL,
    account_id UUID,
    type TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    transfer_pair_id UUID,
    version INT NOT NULL DEFAULT 1,
    occurred_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_transactions_ledger FOREIGN KEY (ledger_id) REFERENCES ledgers (id),
    CONSTRAINT fk_transactions_account FOREIGN KEY (account_id) REFERENCES accounts (id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_occurred_at_desc
    ON transactions (user_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_transfer_pair_id
    ON transactions (transfer_pair_id);
