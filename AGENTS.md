# AGENTS.md (Root)

This file provides guidance to AI coding agents when working with code in this repository.

## Quick Reference

```bash
# Backend (from /backend)
go test ./...                                    # All tests
go test ./internal/accounting -run TestCreate     # Single test
go fmt ./... && go vet ./...                      # Format + lint

# Frontend (from /frontend/app)
pnpm run dev                                     # Dev server
pnpm run test                                    # All tests
pnpm run test -- src/App.test.tsx                # Single test
pnpm run check                                   # Lint + test

# Root Makefile
make setup | make backend | make frontend | make migrate-up | make clean
```

## Hierarchy

| Path | Domain |
|------|--------|
| `backend/AGENTS.md` | Backend overview, commands, conventions |
| `backend/internal/accounting/AGENTS.md` | Ledgers, accounts, transactions, transfers |
| `backend/internal/auth/AGENTS.md` | Auth, sessions, OAuth |
| `backend/internal/bootstrap/AGENTS.md` | DI wiring, router, infra |
| `backend/internal/classification/AGENTS.md` | Categories, tags |
| `backend/internal/portability/AGENTS.md` | Import/export, PATs |
| `backend/internal/reporting/AGENTS.md` | Stats, trends |
| `frontend/app/src/features/AGENTS.md` | Feature pattern, API/hooks convention |

## Code Style

### TypeScript/React
- Imports: `@/` alias, group external → internal → relative
- Components: PascalCase. Functions: camelCase.
- Types: `interface` for objects, `type` for unions
- Hooks: `use` prefix, colocate in `*-hooks.ts`
- Exports: Named preferred, default for pages only

### Go
- Error codes: Package constants (`LEDGER_INVALID`)
- Error types: `contractError` with code
- Packages: Organize by domain
- Testing: Table-driven, test files beside implementation

## Anti-Patterns

1. **No `as any` or `@ts-ignore`** — Fix types properly
2. **No empty catch blocks** — Always handle errors
3. **Migrations are append-only** — Never modify existing
4. **PRD.md is source of truth** — Read before accounting changes

## Key Files

| File | Purpose |
|------|---------|
| `PRD.md` | Product requirements, accounting rules |
| `backend/openapi/openapi.yaml` | API contract |
| `frontend/app/src/lib/api.ts` | HTTP client, response types |
| `backend/internal/bootstrap/http/router.go` | Route registration, DI |
| `Makefile` | Common dev commands |

## Environment Variables

Required: `SMTP_HOST`, `AUTH_CODE_PEPPER`
Optional: `DATABASE_URL` (omit for in-memory), `REDIS_URL`
