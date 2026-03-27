WITH parsed_import_rows AS (
    SELECT
        user_id,
        amount,
        description,
        CASE
            WHEN date ~ '^\d{4}/\d{2}/\d{2} \d{2}:\d{2}$' THEN to_timestamp(date, 'YYYY/MM/DD HH24:MI')
            WHEN date ~ '^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$' THEN to_timestamp(date, 'YYYY-MM-DD HH24:MI:SS')
            WHEN date ~ '^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$' THEN to_timestamp(date, 'YYYY-MM-DD HH24:MI')
            WHEN date ~ '^\d{4}/\d{2}/\d{2}$' THEN to_timestamp(date, 'YYYY/MM/DD')
            WHEN date ~ '^\d{4}-\d{2}-\d{2}$' THEN to_timestamp(date, 'YYYY-MM-DD')
            ELSE NULL
        END AS occurred_at
    FROM import_rows
)
UPDATE transactions AS t
SET memo = parsed_import_rows.description
FROM parsed_import_rows
WHERE t.user_id = parsed_import_rows.user_id
  AND t.amount = parsed_import_rows.amount
  AND t.occurred_at = parsed_import_rows.occurred_at
  AND (t.memo IS NULL OR t.memo = '');
