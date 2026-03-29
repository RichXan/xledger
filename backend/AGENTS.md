# Backend (`/backend`)

Go API server for xledger personal finance app.

## Structure

```
backend/
├── cmd/api/              # Entry point, CLI (serve, migrate)
├── internal/
│   ├── accounting/       # Ledgers, accounts, transactions, transfers
│   ├── auth/             # Email code login, OAuth, sessions
│   ├── classification/   # Categories, tags, templates
│   ├── reporting/        # Stats, trends, aggregations
│   ├── portability/      # CSV import/export, PATs
│   ├── bootstrap/        # DI wiring, HTTP router, infra
│   └── common/           # Shared HTTP helpers
├── migrations/           # SQL migrations (*.sql, append-only)
├── openapi/              # API contract (openapi.yaml)
└── tests/integration/    # API integration tests
```

## Commands

```bash
go test ./...                                    # All tests
go test ./internal/accounting -run TestCreate     # Single test
go fmt ./...                                     # Format
go vet ./...                                     # Lint
```

## Conventions

- **Domain packages**: Each `internal/` dir = one domain. Contains handler, service, repository.
- **Error codes**: Package-level constants (`LEDGER_INVALID`), wrapped in `contractError`.
- **DI**: All wiring in `bootstrap/http/router.go`. Repos constructed there.
- **Migrations**: Append-only SQL files. Never modify existing.
- **Response envelope**: `{ code, message, data }`

## Anti-Patterns

- Don't use `interface{}` for handler/service params
- Don't skip error codes — always use domain constants
- Don't modify migration files after merge
- Don't bypass `router.go` wiring for new dependencies

## Key Files

| File | Purpose |
|------|---------|
| `cmd/api/main.go` | Entry point, mode selection |
| `internal/bootstrap/http/router.go` | Route registration, DI |
| `openapi/openapi.yaml` | API contract |
| `migrations/*.sql` | Database schema |
