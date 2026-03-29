# Portability Package

Data import/export and Personal Access Tokens (PATs).

## Files

| File | Purpose |
|------|---------|
| `handler.go` | HTTP handlers for import/export/PAT |
| `import_preview_service.go` | CSV column detection |
| `import_confirm_service.go` | Transaction creation from CSV |
| `export_service.go` | Transaction export |
| `pat_service.go` | PAT CRUD for API access |
| `repository*.go` | In-memory + Postgres storage |

## Import Flow

1. `POST /api/import/csv` → Preview (detect columns)
2. `POST /api/import/csv/confirm` → Confirm (create transactions)
3. Idempotency key prevents duplicate imports

## Key Rules

- CSV must have header row + at least 1 data row
- Import creates transactions in user's default ledger
- PATs scoped to user, can be revoked
- Export includes all user transactions

## Don't

- Don't skip CSV validation
- Don't allow import without preview
- Don't expose PAT tokens after creation
