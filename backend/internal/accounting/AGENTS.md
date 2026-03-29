# Accounting Package

Core domain: ledgers, accounts, transactions, transfers.

## Files

| File | Purpose |
|------|---------|
| `handler.go` | HTTP handlers for accounts/transactions |
| `ledger_service.go` | Ledger CRUD, default ledger logic |
| `account_service.go` | Account CRUD, balance calculations |
| `transaction_service.go` | Transaction CRUD, pagination |
| `transfer_service.go` | Atomic transfer pairs |
| `repository_*.go` | In-memory + Postgres implementations |

## Key Rules (from PRD.md)

- Accounts are global to user, NOT attached to ledger
- Every transaction belongs to a ledger
- `account_id = NULL` affects stats but NOT account balances
- Transfers create paired transactions via `transfer_pair_id`

## Patterns

- Table-driven tests (`*_test.go` beside implementation)
- Repository interface in same file as service
- Error codes: `LEDGER_INVALID`, `ACCOUNT_NOT_FOUND`, etc.
- Balance = `SUM(transactions)` where `account_id` matches

## Don't

- Don't bypass transfer pairing logic
- Don't allow transfers within same account
- Don't hardcode ledger selection — use user's default
