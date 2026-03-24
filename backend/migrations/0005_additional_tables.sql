ALTER TABLE accounts ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE categories ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE categories ADD COLUMN IF NOT EXISTS usage_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tags ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE TABLE IF NOT EXISTS balance_recalc_log (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    ledger_id UUID NOT NULL,
    recalculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stats_recalc_log (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    ledger_id UUID NOT NULL,
    recalculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS category_history (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    category_id UUID NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_category_history_user_category
    ON category_history (user_id, category_id);

CREATE TABLE IF NOT EXISTS import_jobs (
    user_id UUID NOT NULL,
    path TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    response_json TEXT NOT NULL DEFAULT '{}',
    error_code TEXT,
    PRIMARY KEY (user_id, path, idempotency_key)
);

CREATE TABLE IF NOT EXISTS import_rows (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    date TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    description TEXT NOT NULL,
    triple_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_import_rows_user_triple
    ON import_rows (user_id, triple_key);

CREATE TABLE IF NOT EXISTS import_dedup (
    user_id UUID NOT NULL,
    triple_key TEXT NOT NULL,
    PRIMARY KEY (user_id, triple_key)
);
