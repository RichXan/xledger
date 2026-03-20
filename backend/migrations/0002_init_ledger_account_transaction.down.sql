DROP INDEX IF EXISTS idx_transactions_transfer_pair_id;
DROP INDEX IF EXISTS idx_transactions_user_occurred_at_desc;
DROP TABLE IF EXISTS transactions;

DROP INDEX IF EXISTS idx_accounts_id_user;
DROP TABLE IF EXISTS accounts;

DROP INDEX IF EXISTS idx_ledgers_id_user;
DROP INDEX IF EXISTS idx_ledgers_user_default_true;
