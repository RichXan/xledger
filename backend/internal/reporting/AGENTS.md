# Reporting Package

Dashboard stats, trends, and category breakdowns.

## Files

| File | Purpose |
|------|---------|
| `handler.go` | HTTP handlers for stats endpoints |
| `overview_service.go` | Total income/expense/assets |
| `trend_service.go` | Time-series aggregations |
| `category_service.go` | Category breakdown |
| `repository.go` | Postgres queries |

## Endpoints

- `GET /api/stats/overview` — Income, expense, total assets
- `GET /api/stats/trend` — Time-series (day/week/month granularity)
- `GET /api/stats/category` — Expense breakdown by category

## Key Rules

- Assets = sum of all account balances
- Overview supports optional date range filtering
- Trend buckets use `bucket_start` as key
- Category stats = expense-only (no income)

## Don't

- Don't include deleted transactions in stats
- Don't cache without invalidation strategy
- Don't return negative balances without context
