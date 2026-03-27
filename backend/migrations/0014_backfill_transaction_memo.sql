WITH import_ranked AS (
    SELECT
        id,
        user_id,
        amount,
        description,
        created_at,
        ROW_NUMBER() OVER (PARTITION BY user_id, amount ORDER BY created_at, id) AS rn
    FROM import_rows
),
transaction_ranked AS (
    SELECT
        id,
        user_id,
        amount,
        created_at,
        ROW_NUMBER() OVER (PARTITION BY user_id, amount ORDER BY created_at, id) AS rn
    FROM transactions
    WHERE deleted_at IS NULL
      AND (memo IS NULL OR memo = '')
)
UPDATE transactions AS t
SET memo = import_ranked.description
FROM transaction_ranked
JOIN import_ranked
  ON import_ranked.user_id = transaction_ranked.user_id
 AND import_ranked.amount = transaction_ranked.amount
 AND import_ranked.rn = transaction_ranked.rn
WHERE t.id = transaction_ranked.id
  AND (t.memo IS NULL OR t.memo = '')
  AND ABS(EXTRACT(EPOCH FROM (transaction_ranked.created_at - import_ranked.created_at))) <= 5;
