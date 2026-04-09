# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repo structure

- `backend/` — Go API server, business logic, SQL migrations, OpenAPI contract, integration tests, Docker Compose for local infra.
- `frontend/app/` — Vite + React + TypeScript web client.
- `frontend/stitch/` — exported design/artifact directories used as UI references.
- `docs/superpowers/specs/` and `docs/superpowers/plans/` — architecture/spec and implementation-plan documents created during earlier planning.
- `PRD.md` — product requirements and core domain rules for v1.

There is no root workspace script setup; backend and frontend are developed independently from their own directories.

## Common commands

### Backend (`/backend`)

- Run all tests:
  - `go test ./...`
- Run integration tests:
  - `go test ./tests/integration -count=1`
- Run a single Go test:
  - `go test ./internal/accounting -run TestLedgerService_Create -count=1`
- Run the API with minimal local env (in-memory repositories):
  - Copy `config/config.yaml.example` to `config/config.yaml`, set `auth.code_pepper` and `auth.token_secret`, remove `database_url`, then:
  - `go run ./cmd/api`
- Run the API against local Postgres + Redis:
  - Ensure `config/config.yaml` has `database_url` and `redis_url` set, then:
  - `CONFIG_FILE=config/config.yaml go run ./cmd/api`
- Start local Postgres + Redis only:
  - `docker compose -f docker-compose.backend.yml up -d`
- Stop local Postgres + Redis:
  - `docker compose -f docker-compose.backend.yml down -v`
- Start full backend stack (API + Postgres + Redis):
  - `docker compose -f docker-compose.yml up -d`
- Migration DDL tests (requires `TEST_DATABASE_URL`):
  - `TEST_DATABASE_URL=postgres://... go test ./migrations -count=1`
- Format / vet:
  - `go fmt ./...`
  - `go vet ./...`

### Frontend (`/frontend/app`)

- Install dependencies:
  - `npm install`
- Start dev server:
  - `npm run dev`
- Build production bundle:
  - `npm run build`
- Preview production build:
  - `npm run preview`
- Run all tests:
  - `npm run test`
- Run tests in watch mode:
  - `npm run test:watch`
- Run a single Vitest file:
  - `npm run test -- src/App.test.tsx`
- Typecheck:
  - `npm run lint`
- Run the main frontend verification pass:
  - `npm run check`

## Product/domain context

Read `PRD.md` before changing core accounting behavior. The most important v1 rules are:

- Accounts are global to the user, not attached to a ledger.
- Every transaction belongs to a ledger; the default ledger is created automatically.
- `account_id = NULL` transactions still affect ledger statistics, but do not affect account balances or total assets.
- Dashboard/reporting/export behavior must keep that accounting meaning consistent.

## Backend architecture

### Entry and wiring

- API entrypoint: `backend/cmd/api/main.go`
- Config loader: `backend/internal/bootstrap/config/config.go`
- Infra connections: `backend/internal/bootstrap/infrastructure/connections.go`
- Router composition: `backend/internal/bootstrap/http/router.go`

`main.go` loads env config, optionally connects Postgres and Redis, and chooses between:

- in-memory wiring when `DATABASE_URL` is unset
- Postgres-backed wiring via `NewRouterWithPostgreSQL(...)` when `DATABASE_URL` is present

That split matters: some tests and local development intentionally rely on the in-memory mode.

### HTTP/API shape

- Base API prefix: `/api`
- Auth routes: `/api/auth/*`
- Business routes live directly under `/api`
- OpenAPI contract: `backend/openapi/openapi.yaml`
- Business JSON responses use a unified envelope: `{ code, message, data }`
- Paginated lists use `data.items` and `data.pagination`

### Main backend domains

The Go code is organized by domain/context rather than by technical layer:

- `internal/auth` — email code login, refresh/logout, OAuth, session/token handling
- `internal/accounting` — ledgers, accounts, transactions, transfers, core balance logic
- `internal/classification` — categories, tags, templates/history
- `internal/reporting` — overview, trend, category aggregations
- `internal/portability` — CSV import/export and personal access token flows
- `internal/common/httpx` — shared HTTP response helpers

Each domain generally contains handlers, services, and repository implementations together.

### Persistence and infrastructure

- Postgres repositories are constructed inside `backend/internal/bootstrap/http/router.go`
- SQL schema lives in `backend/migrations/`
- Docker local infra lives in `backend/docker-compose.backend.yml`
- Full containerized backend stack lives in `backend/docker-compose.yml`

Important behavior: the server reads all config from `config/config.yaml` (path overridable via `CONFIG_FILE` env var). Fields `auth.code_pepper` and `auth.token_secret` are required. `database_url` and `redis_url` are optional; omitting them enables in-memory mode.

### Tests

- Most unit tests live beside implementation files in `internal/...`
- API/integration coverage lives in `backend/tests/integration/`
- Migration lock-down tests live in `backend/migrations/`

When changing API behavior, verify both the implementation and the OpenAPI/integration-test expectations.

## Frontend architecture

### Entry and app shell

- Frontend entrypoint: `frontend/app/src/main.tsx`
- Route tree: `frontend/app/src/App.tsx`
- Shared layout shell: `frontend/app/src/components/layout/app-shell.tsx`

The app is a client-side React SPA using:

- `react-router-dom` for routing
- `@tanstack/react-query` for server state
- a custom auth context for session bootstrap and logout

Routes are rendered inside `AppShell` for authenticated pages and guarded with `RequireAuth`.

### Frontend feature layout

The frontend is split mainly by feature:

- `src/features/auth` — auth API, auth context, localStorage session persistence, route guard
- `src/features/transactions` — transaction list, transaction form options, import preview, mutation/query hooks
- `src/features/management` — account management, PAT flows, CSV export hooks
- `src/features/reporting` — overview/trend/category stats hooks and API calls
- `src/pages` — page-level composition
- `src/components` — reusable UI/layout building blocks
- `src/lib` — API utilities, formatting, helpers

A common pattern is:

1. feature-specific `*-api.ts` wraps backend endpoints
2. feature-specific `*-hooks.ts` wraps those calls in React Query
3. pages consume the hooks directly

### Auth and API assumptions

- API base is hardcoded in `frontend/app/src/lib/api.ts` as `/api`
- Auth session is stored in `localStorage` via `frontend/app/src/features/auth/auth-storage.ts`
- `AuthProvider` attempts `/auth/me`, then refreshes via `/auth/refresh` if needed
- Protected pages depend on a bearer token in the frontend session state
- Vite dev server proxies `/api` to `http://127.0.0.1:8080` in `frontend/app/vite.config.ts`

### Frontend tests

- Test runner: Vitest with jsdom
- Config: `frontend/app/vitest.config.ts`
- Setup file: `frontend/app/src/test/setup.ts`
- Tests are mostly colocated in `src/*.test.tsx`

## Important files

- `PRD.md` — source of truth for product scope and accounting rules.
- `backend/README.md` — backend run/test/API notes.
- `backend/cmd/api/main.go` — process startup and mode selection.
- `backend/internal/bootstrap/http/router.go` — route registration and dependency wiring.
- `backend/openapi/openapi.yaml` — API contract.
- `backend/migrations/` — database schema and migration tests.
- `frontend/app/package.json` — frontend scripts.
- `frontend/app/src/main.tsx` — React root, router, query client, auth provider.
- `frontend/app/src/App.tsx` — route structure.
- `frontend/app/src/lib/api.ts` — shared HTTP request helpers and response envelope assumptions.
- `frontend/app/src/features/auth/auth-context.tsx` — session bootstrap and refresh flow.
- `frontend/app/src/features/transactions/transactions-api.ts` — core transaction-facing client endpoints.
- `frontend/app/src/features/management/management-api.ts` — PAT and export client endpoints.
- `frontend/app/src/features/reporting/reporting-api.ts` — dashboard/analytics data endpoints.

## Existing docs worth checking before major changes

- `docs/superpowers/specs/2026-03-19-xledger-backend-architecture-tech-solution-design.md`
- `docs/superpowers/specs/2026-03-19-plan-a-auth-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-b-accounting-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-c-classification-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-d-reporting-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-e-portability-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-f-automation-spec.md`

Use these when a change touches domain rules or API shape across multiple subsystems.
