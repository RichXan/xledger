ALTER TABLE import_jobs
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'succeeded',
    ADD COLUMN IF NOT EXISTS total_rows INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS processed_rows INTEGER NOT NULL DEFAULT 0;

UPDATE import_jobs
SET status = CASE
    WHEN error_code IS NULL OR error_code = '' THEN 'succeeded'
    ELSE 'failed'
END
WHERE status IS NULL OR status = '';
