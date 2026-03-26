ALTER TABLE ledgers
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS from_account_id UUID;

ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS to_account_id UUID;

ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS transfer_side TEXT;

ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_transactions_user_deleted_at
ON transactions (user_id, deleted_at);
