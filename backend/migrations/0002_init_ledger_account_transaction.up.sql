CREATE UNIQUE INDEX idx_ledgers_user_default_true
    ON ledgers (user_id)
    WHERE is_default = TRUE;

CREATE UNIQUE INDEX idx_ledgers_id_user
    ON ledgers (id, user_id);

CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    initial_balance NUMERIC NOT NULL,
    archived_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_accounts_id_user
    ON accounts (id, user_id);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    ledger_id UUID NOT NULL,
    account_id UUID,
    type TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    transfer_pair_id UUID,
    version INT NOT NULL DEFAULT 1,
    occurred_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_transactions_ledger FOREIGN KEY (ledger_id, user_id) REFERENCES ledgers (id, user_id),
    CONSTRAINT fk_transactions_account FOREIGN KEY (account_id, user_id) REFERENCES accounts (id, user_id)
);

CREATE INDEX idx_transactions_user_occurred_at_desc
    ON transactions (user_id, occurred_at DESC);

CREATE INDEX idx_transactions_transfer_pair_id
    ON transactions (transfer_pair_id);
