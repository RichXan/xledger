DROP INDEX IF EXISTS idx_transactions_user_deleted_at;

DROP TABLE IF EXISTS import_dedup;
DROP TABLE IF EXISTS import_rows;
DROP TABLE IF EXISTS import_jobs;
DROP TABLE IF EXISTS category_history;
DROP TABLE IF EXISTS stats_recalc_log;
DROP TABLE IF EXISTS balance_recalc_log;
DROP TABLE IF EXISTS user_category_templates;
DROP TABLE IF EXISTS default_categories CASCADE;
