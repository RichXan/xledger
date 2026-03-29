CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    email text UNIQUE NOT NULL,
    created_at timestamptz NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS ledgers (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    is_default boolean NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ledgers_user_default_true
    ON ledgers (user_id)
    WHERE is_default = TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ledgers_id_user
    ON ledgers (id, user_id);

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    initial_balance NUMERIC NOT NULL,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_id_user
    ON accounts (id, user_id);

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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    category_id UUID,
    category_name TEXT,
    from_account_id UUID,
    to_account_id UUID,
    transfer_side TEXT,
    deleted_at TIMESTAMPTZ,
    memo TEXT,
    CONSTRAINT fk_transactions_ledger FOREIGN KEY (ledger_id, user_id) REFERENCES ledgers (id, user_id),
    CONSTRAINT fk_transactions_account FOREIGN KEY (account_id, user_id) REFERENCES accounts (id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_occurred_at_desc
    ON transactions (user_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_transfer_pair_id
    ON transactions (transfer_pair_id);

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    parent_id UUID,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    usage_count INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_id_user
    ON categories (id, user_id);

CREATE INDEX IF NOT EXISTS idx_categories_user_parent
    ON categories (user_id, parent_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_categories_parent_user'
    ) THEN
        ALTER TABLE categories
            ADD CONSTRAINT fk_categories_parent_user
            FOREIGN KEY (parent_id, user_id) REFERENCES categories (id, user_id);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_transactions_category_user'
    ) THEN
        ALTER TABLE transactions
            ADD CONSTRAINT fk_transactions_category_user
            FOREIGN KEY (category_id, user_id) REFERENCES categories (id, user_id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_transactions_user_category
    ON transactions (user_id, category_id);

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_id_user
    ON tags (id, user_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_user_name_lower
    ON tags (user_id, lower(name));

CREATE TABLE IF NOT EXISTS transaction_tags (
    transaction_id UUID NOT NULL,
    tag_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (transaction_id, tag_id, user_id),
    CONSTRAINT fk_transaction_tags_tag_user FOREIGN KEY (tag_id, user_id) REFERENCES tags (id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_transaction_tags_user_tag
    ON transaction_tags (user_id, tag_id);
