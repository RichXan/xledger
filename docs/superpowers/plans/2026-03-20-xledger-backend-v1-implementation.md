# Xledger Backend v1 Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build Xledger backend v1 core capabilities (Auth, Accounting, Classification, Reporting, Portability) and optional automation, with verification emails sent directly by SMTP.

**Architecture:** Implement bounded contexts under `backend/internal/*` and keep contracts explicit at HTTP boundary. Keep write consistency in Accounting with DB transactions, keep Reporting read-only with fixed financial invariants, and isolate optional automation behind n8n webhook contract so core manual bookkeeping remains fully available without AI.

**Tech Stack:** Go, Gin, PostgreSQL, Redis, GORM, golang-migrate, SMTP, optional n8n + LLM, Docker Compose

---

## Scope Check

The approved specs were split into independent subsystems:
- Plan-A Auth
- Plan-B Accounting
- Plan-C Classification
- Plan-D Reporting
- Plan-E Portability
- Plan-F Optional Automation

Execution order is mandatory: `A -> B -> C -> D -> E -> F`.

---

## File Structure

**Bootstrap / shared**
- Create: `backend/go.mod`
- Create: `backend/cmd/api/main.go`
- Create: `backend/internal/bootstrap/config/config.go`
- Create: `backend/internal/bootstrap/http/router.go`
- Create: `backend/internal/bootstrap/http/middleware.go`
- Create: `backend/internal/common/httpx/errors.go`
- Create: `backend/internal/common/httpx/response.go`
- Create: `backend/internal/common/timex/timezone.go`

**Migrations**
- Create: `backend/migrations/0001_init_users_auth.sql`
- Create: `backend/migrations/0002_init_ledger_account_transaction.sql`
- Create: `backend/migrations/0003_init_category_tag.sql`
- Create: `backend/migrations/0004_init_pat_import_jobs.sql`
- Create: `backend/migrations/0005_indexes_and_constraints.sql`

**Plan-A (Auth)**
- Create: `backend/internal/auth/code_service.go`
- Create: `backend/internal/auth/session_service.go`
- Create: `backend/internal/auth/oauth_service.go`
- Create: `backend/internal/auth/smtp_sender.go`
- Create: `backend/internal/auth/repository.go`
- Create: `backend/internal/auth/handler.go`
- Create: `backend/internal/auth/auth_test.go`

**Plan-B (Accounting)**
- Create: `backend/internal/accounting/ledger_service.go`
- Create: `backend/internal/accounting/account_service.go`
- Create: `backend/internal/accounting/transaction_service.go`
- Create: `backend/internal/accounting/transfer_service.go`
- Create: `backend/internal/accounting/repository_ledger.go`
- Create: `backend/internal/accounting/repository_account.go`
- Create: `backend/internal/accounting/repository_transaction.go`
- Create: `backend/internal/accounting/handler.go`
- Create: `backend/internal/accounting/accounting_test.go`

**Plan-C (Classification)**
- Create: `backend/internal/classification/template_service.go`
- Create: `backend/internal/classification/category_service.go`
- Create: `backend/internal/classification/tag_service.go`
- Create: `backend/internal/classification/repository.go`
- Create: `backend/internal/classification/handler.go`
- Create: `backend/internal/classification/category_lifecycle_test.go`
- Create: `backend/internal/classification/tag_lifecycle_test.go`

**Plan-D (Reporting)**
- Create: `backend/internal/reporting/overview_service.go`
- Create: `backend/internal/reporting/trend_service.go`
- Create: `backend/internal/reporting/category_service.go`
- Create: `backend/internal/reporting/repository.go`
- Create: `backend/internal/reporting/handler.go`
- Create: `backend/internal/reporting/reporting_overview_test.go`
- Create: `backend/internal/reporting/reporting_trend_test.go`
- Create: `backend/internal/reporting/reporting_category_test.go`

**Plan-E (Portability)**
- Create: `backend/internal/portability/import_preview_service.go`
- Create: `backend/internal/portability/import_confirm_service.go`
- Create: `backend/internal/portability/export_service.go`
- Create: `backend/internal/portability/pat_service.go`
- Create: `backend/internal/portability/repository.go`
- Create: `backend/internal/portability/handler.go`
- Create: `backend/internal/portability/import_test.go`
- Create: `backend/internal/portability/pat_test.go`

**Plan-F (Optional automation)**
- Create: `backend/internal/automation/quick_entry_contract.go`
- Create: `backend/internal/automation/quick_entry_adapter.go`
- Create: `backend/internal/automation/automation_test.go`

**Docs + ops**
- Create: `backend/openapi/openapi.yaml`
- Create: `backend/.env.example`
- Create: `backend/docker-compose.backend.yml`
- Create: `backend/README.md`
- Create: `backend/tests/integration/e2e_test.go`

---

## Chunk 1: Plan-A Auth + Foundation

### Task 1: Bootstrap app skeleton and auth base migration

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/api/main.go`
- Create: `backend/internal/bootstrap/config/config.go`
- Create: `backend/internal/bootstrap/http/router.go`
- Create: `backend/migrations/0001_init_users_auth.sql`
- Test: `backend/internal/bootstrap/bootstrap_test.go`

- [ ] **Step 1: Write failing tests for health route + config load**

```go
func TestRouter_Healthz(t *testing.T) { /* expect 200 */ }
func TestConfig_RequiresSMTPEnv(t *testing.T) { /* expect error on missing SMTP_HOST */ }
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/bootstrap -v`  
Expected: FAIL with missing router/config symbols.

- [ ] **Step 3: Implement minimal bootstrap + auth migration**

```sql
CREATE TABLE users (id uuid primary key, email text unique not null, created_at timestamptz not null);
CREATE TABLE refresh_tokens (id uuid primary key, user_id uuid not null, token_hash text not null, expires_at timestamptz not null);
CREATE TABLE ledgers (id uuid primary key, user_id uuid not null, name text not null, is_default boolean not null default false);
```

- [ ] **Step 4: Re-run bootstrap tests + migration dry-run**

Run: `cd backend && go test ./internal/bootstrap -v`  
Expected: PASS with `TestRouter_Healthz` and `TestConfig_RequiresSMTPEnv`.

Run: `cd backend && migrate -path migrations -database "$TEST_DATABASE_URL" up`  
Expected: migration `0001_init_users_auth` applied successfully.

- [ ] **Step 5: Commit**

```bash
git add backend/go.mod backend/cmd/api/main.go backend/internal/bootstrap backend/migrations/0001_init_users_auth.sql
git commit -m "feat(bootstrap): initialize api skeleton and auth base schema"
```

### Task 2: Implement SMTP send-code/verify-code flow with rate-limit

**Files:**
- Create: `backend/internal/auth/code_service.go`
- Create: `backend/internal/auth/smtp_sender.go`
- Create: `backend/internal/auth/repository.go`
- Create: `backend/internal/auth/handler.go`
- Modify: `backend/internal/bootstrap/http/router.go`
- Test: `backend/internal/auth/auth_test.go`

- [ ] **Step 1: Write failing tests for `send-code` + `verify-code` + error codes**

```go
func TestSendCode_UsesSMTPAndRedis(t *testing.T) {}
func TestSendCode_GeneratesSixDigitCode(t *testing.T) {}
func TestVerifyCode_Success_ReturnsAccessAndRefresh(t *testing.T) {}
func TestVerifyCode_Invalid_ReturnsAUTH_CODE_INVALID(t *testing.T) {}
func TestVerifyCode_Expired_ReturnsAUTH_CODE_EXPIRED(t *testing.T) {}
func TestSendCode_SMTPFailure_ReturnsAUTH_CODE_SEND_FAILED(t *testing.T) {}
func TestSendCode_SMTPFailure_CreatesNoSession(t *testing.T) {}
func TestSendCode_TooFrequent_ReturnsAUTH_CODE_RATE_LIMIT(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/auth -run 'SendCode|VerifyCode' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement SMTP sender + code TTL + rate-limit**

```go
type SMTPSender interface { Send(to, subject, body string) error }
// store code with 10m TTL, rate-limit 60s by email + hourly cap by IP
```

- [ ] **Step 4: Re-run tests + API smoke check**

Run: `cd backend && go test ./internal/auth -run 'SendCode|VerifyCode' -v`  
Expected: PASS with explicit `AUTH_CODE_*` assertions.

Run: `curl -s -X POST localhost:8080/api/auth/send-code -H 'Content-Type: application/json' -d '{"email":"a@b.com"}'`  
Expected: status `200` and JSON field `code_sent=true`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/auth/code_service.go backend/internal/auth/smtp_sender.go backend/internal/auth/repository.go backend/internal/auth/handler.go backend/internal/auth/auth_test.go backend/internal/bootstrap/http/router.go
git commit -m "feat(auth): add smtp verification code login flow"
```

### Task 3: Implement OAuth callback, refresh rotation, logout blacklist, first-login bootstrap hook

**Files:**
- Create: `backend/internal/auth/oauth_service.go`
- Create: `backend/internal/auth/session_service.go`
- Modify: `backend/internal/auth/handler.go`
- Modify: `backend/internal/auth/repository.go`
- Modify: `backend/internal/bootstrap/http/router.go`
- Test: `backend/internal/auth/auth_session_test.go`

- [ ] **Step 1: Write failing tests for callback/refresh/logout/strict-mode**

```go
func TestGoogleCallback_ValidatesStateNonce(t *testing.T) {}
func TestGoogleCallback_ReplayedNonceRejected(t *testing.T) {}
func TestGoogleCallback_HTTPContract_IsGET_api_auth_google_callback(t *testing.T) {}
func TestGoogleCallback_InvalidProviderResponse_ReturnsAUTH_OAUTH_FAILED(t *testing.T) {}
func TestRefresh_RotationInvalidatesOldToken(t *testing.T) {}
func TestRefresh_ExpiryIsSevenDays(t *testing.T) {}
func TestRefresh_Expired_ReturnsAUTH_REFRESH_EXPIRED(t *testing.T) {}
func TestRefresh_RejectsAccessTokenType(t *testing.T) {}
func TestLogout_BlacklistsRefreshWithinSLA(t *testing.T) {}
func TestLogout_BlacklistPropagationLTE5s(t *testing.T) {}
func TestLogout_RejectsRefreshTokenType(t *testing.T) {}
func TestLogout_Unauthorized_ReturnsAUTH_UNAUTHORIZED(t *testing.T) {}
func TestRefresh_BlacklistStrictMode_ReturnsAUTH_REFRESH_REVOKED(t *testing.T) {}
func TestRefresh_BlacklistStrictMode_EmitsAlertEvent(t *testing.T) {}
func TestFirstLogin_CreatesDefaultLedger(t *testing.T) {}
func TestOAuthFailure_DoesNotAffectSendCodeFlow(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/auth -run 'GoogleCallback|Refresh|Logout|FirstLogin' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement callback + token lifecycle + ledger bootstrap hook**

```go
// callback: validate state+nonce, resolve user, issue tokens
// refresh: rotate refresh token, blacklist old, expires_at = now + 7d
// logout: blacklist refresh
// first login: ensure default ledger exists (idempotent)
// strict mode: reject refresh and emit alert event
```

- [ ] **Step 4: Re-run tests + auth smoke chain**

Run: `cd backend && go test ./internal/auth -v`  
Expected: PASS.

Run: `curl -s -X POST localhost:8080/api/auth/refresh -H "Authorization: Bearer ${REFRESH_TOKEN}"`  
Expected: status `200` and new access token.

Run: `curl -s -X POST localhost:8080/api/auth/logout -H "Authorization: Bearer ${ACCESS_TOKEN}"`  
Expected: status `200` and `revoked=true`.

Run: `curl -s -X POST localhost:8080/api/auth/refresh -H "Authorization: Bearer ${OLD_REFRESH_TOKEN}"`  
Expected: status `401` and `AUTH_REFRESH_REVOKED`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/auth/oauth_service.go backend/internal/auth/session_service.go backend/internal/auth/handler.go backend/internal/auth/repository.go backend/internal/auth/auth_session_test.go backend/internal/bootstrap/http/router.go
git commit -m "feat(auth): add oauth callback refresh rotation and logout blacklist"
```

---

## Chunk 2: Plan-B Accounting

### Task 4: Implement ledger/account schema and CRUD contracts

**Files:**
- Create: `backend/migrations/0002_init_ledger_account_transaction.sql`
- Create: `backend/internal/accounting/repository_ledger.go`
- Create: `backend/internal/accounting/repository_account.go`
- Create: `backend/internal/accounting/ledger_service.go`
- Create: `backend/internal/accounting/account_service.go`
- Create: `backend/internal/accounting/handler.go`
- Test: `backend/internal/accounting/ledger_account_test.go`

**Responsibilities:**
- `repository_*`: persistence and query composition only.
- `*_service.go`: invariants and orchestration only.
- `handler.go`: HTTP binding, auth context, error mapping only.

- [ ] **Step 1: Write failing CRUD and invariant tests**

```go
func TestDeleteDefaultLedger_ReturnsLEDGER_DEFAULT_IMMUTABLE(t *testing.T) {}
func TestAccountCRUD_OwnershipScopedByUser(t *testing.T) {}
func TestAccountGet_NotFound_ReturnsACCOUNT_NOT_FOUND(t *testing.T) {}
func TestAccountCreate_InvalidPayload_ReturnsACCOUNT_INVALID(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/accounting -run 'Ledger|Account' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement explicit migration + services**

```sql
CREATE TABLE ledgers (id uuid primary key, user_id uuid not null, name text not null, is_default boolean not null default false, unique(user_id,is_default));
CREATE TABLE accounts (id uuid primary key, user_id uuid not null, name text not null, type text not null, initial_balance numeric not null, archived_at timestamptz);
CREATE TABLE transactions (id uuid primary key, user_id uuid not null, ledger_id uuid not null, account_id uuid, type text not null, amount numeric not null, transfer_pair_id uuid, version int not null default 1, occurred_at timestamptz not null);
CREATE INDEX idx_txn_user_time ON transactions(user_id, occurred_at desc);
CREATE INDEX idx_txn_transfer_pair ON transactions(transfer_pair_id);
```

- [ ] **Step 4: Re-run tests + API contract smoke**

Run: `cd backend && go test ./internal/accounting -run 'Ledger|Account' -v`  
Expected: PASS for `LEDGER_DEFAULT_IMMUTABLE`, `ACCOUNT_NOT_FOUND`, `ACCOUNT_INVALID`.

Run: `curl -X DELETE localhost:8080/api/ledgers/{defaultId}`  
Expected: `409` + `LEDGER_DEFAULT_IMMUTABLE`.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/0002_init_ledger_account_transaction.sql backend/internal/accounting/repository_ledger.go backend/internal/accounting/repository_account.go backend/internal/accounting/ledger_service.go backend/internal/accounting/account_service.go backend/internal/accounting/handler.go backend/internal/accounting/ledger_account_test.go
git commit -m "feat(accounting): add ledger account schema and ownership-safe crud"
```

### Task 5: Implement income/expense validation rules and transaction query base

**Files:**
- Create: `backend/internal/accounting/repository_transaction.go`
- Create: `backend/internal/accounting/transaction_service.go`
- Modify: `backend/internal/accounting/handler.go`
- Test: `backend/internal/accounting/transaction_rule_test.go`

**Responsibilities:**
- `repository_transaction.go`: transaction read/write SQL and locks.
- `transaction_service.go`: validation and recalculation orchestration.
- `handler.go`: endpoint-level request/response mapping.

- [ ] **Step 1: Write failing tests for rule set**

```go
func TestCreateTxn_LedgerRequired(t *testing.T) {}
func TestCreateTxn_AccountNullableForExpenseIncome(t *testing.T) {}
func TestCreateTransfer_RequiresBothAccounts(t *testing.T) {}
func TestEditTxn_RecalculatesBalances(t *testing.T) {}
func TestDeleteTxn_RecalculatesBalances(t *testing.T) {}
func TestEditDeleteTxn_RecalculatesStatsInput(t *testing.T) {}
func TestEditTxn_NotFound_ReturnsTXN_NOT_FOUND(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/accounting -run 'LedgerRequired|AccountNullable|RequiresBothAccounts' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement rules + error mapping**

```go
if in.LedgerID == uuid.Nil { return ErrTXNValidation }
if in.Type == Transfer && (in.FromAccountID == nil || in.ToAccountID == nil) { return ErrTXNValidation }
```

- [ ] **Step 4: Re-run tests + endpoint smoke**

Run: `cd backend && go test ./internal/accounting -run 'LedgerRequired|AccountNullable|RequiresBothAccounts' -v`  
Expected: PASS.

Run: `curl -s -X POST localhost:8080/api/transactions -H 'Content-Type: application/json' -H "Authorization: Bearer ${ACCESS_TOKEN}" -d '{"type":"expense","amount":10}'`  
Expected: `400` + `TXN_VALIDATION_FAILED`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/accounting/repository_transaction.go backend/internal/accounting/transaction_service.go backend/internal/accounting/handler.go backend/internal/accounting/transaction_rule_test.go
git commit -m "feat(accounting): enforce transaction validation and nullable account rules"
```

### Task 6: Implement transfer pair atomicity, conflict handling, cross-ledger transfer behavior

**Files:**
- Create: `backend/internal/accounting/transfer_service.go`
- Modify: `backend/internal/accounting/transaction_service.go`
- Modify: `backend/internal/accounting/repository_transaction.go`
- Test: `backend/internal/accounting/transfer_atomicity_test.go`

**Responsibilities:**
- `transfer_service.go`: pair lifecycle and conflict policy.
- `transaction_service.go`: delegates transfer actions.
- `repository_transaction.go`: pair lock and tx persistence primitives.

- [ ] **Step 1: Write failing tests for pair/rollback/conflict/cross-ledger**

```go
func TestTransfer_CreateEditDeletePairAtomically(t *testing.T) {}
func TestTransfer_ConflictReturnsTXN_CONFLICT(t *testing.T) {}
func TestTransfer_VersionConflictReturnsTXN_CONFLICT(t *testing.T) {}
func TestTransfer_BilateralMismatchReturnsTXN_CONFLICT(t *testing.T) {}
func TestTransfer_CrossLedgerAllowed(t *testing.T) {}
func TestTransfer_CrossLedger_KeepsLedgerScopedAggregationInputs(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/accounting -run 'Transfer_' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement tx pair lock + rollback policy**

```go
return repo.WithTx(ctx, func(tx TxRepo) error {
  tx.LockTransferPair(pairID)
  return tx.UpsertBothSides(...)
})
```

- [ ] **Step 4: Re-run tests + API smoke for conflict**

Run: `cd backend && go test ./internal/accounting -run 'Transfer_' -v`  
Expected: PASS.

Run: `cd backend && go test ./internal/accounting -run 'Transfer_Conflict|Transfer_VersionConflict|Transfer_BilateralMismatch' -v`  
Expected: PASS and each conflict test asserts `TXN_CONFLICT`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/accounting/transfer_service.go backend/internal/accounting/transaction_service.go backend/internal/accounting/repository_transaction.go backend/internal/accounting/transfer_atomicity_test.go
git commit -m "feat(accounting): add atomic transfer pair lifecycle and conflict handling"
```

---

## Chunk 3: Plan-C Classification

### Task 7: Implement default category template copy on user bootstrap

**Files:**
- Create: `backend/internal/classification/template_service.go`
- Modify: `backend/internal/auth/session_service.go`
- Test: `backend/internal/classification/template_bootstrap_test.go`

- [ ] **Step 1: Write failing test for first-login template copy (idempotent)**

```go
func TestFirstLogin_CopiesDefaultCategoryTemplateOnce(t *testing.T) {}
```

- [ ] **Step 2: Run test to verify fail**

Run: `cd backend && go test ./internal/classification -run FirstLogin_CopiesDefaultCategoryTemplateOnce -v`  
Expected: FAIL.

- [ ] **Step 3: Implement template copy service and hook**

```go
func (s *TemplateService) EnsureUserDefaults(ctx context.Context, userID uuid.UUID) error
```

- [ ] **Step 4: Re-run tests**

Run: `cd backend && go test ./internal/classification -run FirstLogin_CopiesDefaultCategoryTemplateOnce -v`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/classification/template_service.go backend/internal/auth/session_service.go backend/internal/classification/template_bootstrap_test.go
git commit -m "feat(classification): copy default category template on first login"
```

### Task 8: Implement category/tag lifecycle + archive semantics

**Files:**
- Create: `backend/migrations/0003_init_category_tag.sql`
- Create: `backend/internal/classification/repository.go`
- Create: `backend/internal/classification/category_service.go`
- Create: `backend/internal/classification/tag_service.go`
- Create: `backend/internal/classification/handler.go`
- Test: `backend/internal/classification/category_lifecycle_test.go`
- Test: `backend/internal/classification/tag_lifecycle_test.go`

**Responsibilities:**
- `repository.go`: category/tag persistence operations.
- `category_service.go`: hierarchy/archiving/history-name invariants.
- `tag_service.go`: tag uniqueness and not-found semantics.
- `handler.go`: `/api/categories` and `/api/tags` contract mapping.

- [ ] **Step 1: Write failing tests for key semantics and errors**

```go
func TestDeleteReferencedCategory_Archives_ReturnsCAT_IN_USE_ARCHIVED(t *testing.T) {}
func TestCreateCategory_InvalidParent_ReturnsCAT_INVALID_PARENT(t *testing.T) {}
func TestCreateCategory_DepthExceeded_ReturnsCAT_INVALID_PARENT(t *testing.T) {}
func TestArchivedCategory_NotSelectableInTxnEdit(t *testing.T) {}
func TestStatsAndExport_KeepHistoricalCategoryNameAfterArchive(t *testing.T) {}
func TestCreateTag_Duplicate_ReturnsTAG_DUPLICATED(t *testing.T) {}
func TestUpdateTag_NotFound_ReturnsTAG_NOT_FOUND(t *testing.T) {}
func TestCategoriesAndTags_AcceptsAccessAndPAT(t *testing.T) {}
func TestTag_AssociationWithTransactionVisibleInListFilter(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/classification -v`  
Expected: FAIL.

- [ ] **Step 3: Implement migration + repository/service lifecycle logic**

```go
if referenced { archive(categoryID); return CAT_IN_USE_ARCHIVED }
```

- [ ] **Step 4: Wire handler routes and re-run tests + category API smoke**

Run: `cd backend && go test ./internal/classification -v`  
Expected: PASS for `CAT_IN_USE_ARCHIVED`, `CAT_INVALID_PARENT`, `TAG_DUPLICATED`, `TAG_NOT_FOUND`.

Run: `curl -s -X DELETE localhost:8080/api/categories/${REFERENCED_CATEGORY_ID} -H "Authorization: Bearer ${ACCESS_TOKEN}"`  
Expected: status `200` and response contains `archived=true`.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/0003_init_category_tag.sql backend/internal/classification/repository.go backend/internal/classification/category_service.go backend/internal/classification/tag_service.go backend/internal/classification/handler.go backend/internal/classification/category_lifecycle_test.go backend/internal/classification/tag_lifecycle_test.go
git commit -m "feat(classification): implement category archive and tag lifecycle contracts"
```

### Task 9: Implement multi-filter query semantics and performance check

**Files:**
- Modify: `backend/internal/accounting/repository_transaction.go`
- Modify: `backend/internal/accounting/handler.go`
- Test: `backend/internal/accounting/transaction_filter_test.go`
- Test: `backend/internal/accounting/transaction_filter_perf_test.go`

- [ ] **Step 1: Write failing filter and invalid-param tests**

```go
func TestListTransactions_MultiFilterStableOrder(t *testing.T) {}
func TestListTransactions_InvalidRange_ReturnsBadRequest(t *testing.T) {}
func TestListTransactions_InvalidPage_ReturnsBadRequest(t *testing.T) {}
func TestListTransactions_InvalidPageSize_ReturnsBadRequest(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/accounting -run 'MultiFilter|InvalidRange' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement filter builder + param validation**

```go
// filters: ledger/account/category/tag/date, order by occurred_at desc, id desc
// reject invalid range, invalid page/page_size
```

- [ ] **Step 4: Re-run correctness + performance test**

Run: `cd backend && go test ./internal/accounting -run 'MultiFilter|InvalidRange|Perf1K' -v`  
Expected: PASS and perf assertion `P95 <= 1s` with page size 50 and 1k seeded records.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/accounting/repository_transaction.go backend/internal/accounting/handler.go backend/internal/accounting/transaction_filter_test.go backend/internal/accounting/transaction_filter_perf_test.go
git commit -m "feat(accounting): add multi-filter transaction query with validation and perf guard"
```

---

## Chunk 4: Plan-D Reporting

### Task 10: Implement overview/trend/category endpoints with accounting invariants

**Files:**
- Create: `backend/internal/reporting/repository.go`
- Create: `backend/internal/reporting/overview_service.go`
- Create: `backend/internal/reporting/trend_service.go`
- Create: `backend/internal/reporting/category_service.go`
- Create: `backend/internal/reporting/handler.go`
- Modify: `backend/internal/bootstrap/http/router.go`
- Test: `backend/internal/reporting/reporting_overview_test.go`
- Test: `backend/internal/reporting/reporting_trend_test.go`
- Test: `backend/internal/reporting/reporting_category_test.go`

**Responsibilities:**
- `overview_service.go`: overview invariant computation only.
- `trend_service.go`: series aggregation only.
- `category_service.go`: category distribution only.
- `handler.go`: auth and error mapping only.

- [ ] **Step 1: Write failing invariant tests**

```go
func TestOverview_TotalAssetsIndependentFromLedgerFilter(t *testing.T) {}
func TestOverview_AccountNullIncludedInIncomeExpenseNotAssets(t *testing.T) {}
func TestOverview_TransferOffsetsExcludedFromIncomeExpense(t *testing.T) {}
func TestStatsEndpoints_AcceptsAccessAndPAT(t *testing.T) {}
func TestTrend_BasicWindowAggregation(t *testing.T) {}
func TestCategoryStats_UsesHistoricalCategoryNames(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/reporting -run 'Overview|Category' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement read-only services and endpoint error mapping**

```go
// /api/stats/overview, /api/stats/trend, /api/stats/category
// errors: STAT_QUERY_INVALID, STAT_TIMEOUT
```

- [ ] **Step 4: Re-run tests + API smoke checks**

Run: `cd backend && go test ./internal/reporting -run 'Overview|Category' -v`  
Expected: PASS.

Run: `curl -s -H "Authorization: Bearer ${ACCESS_TOKEN}" "localhost:8080/api/stats/overview?ledger_id=${LEDGER_A}"` and `curl -s -H "Authorization: Bearer ${ACCESS_TOKEN}" "localhost:8080/api/stats/overview"`  
Expected: both responses have identical `total_assets`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/reporting/repository.go backend/internal/reporting/overview_service.go backend/internal/reporting/trend_service.go backend/internal/reporting/category_service.go backend/internal/reporting/handler.go backend/internal/reporting/reporting_overview_test.go backend/internal/reporting/reporting_category_test.go
git commit -m "feat(reporting): add stats endpoints with fixed asset and historical category invariants"
```

### Task 11: Implement timezone/empty-window/error behavior and reporting perf gate

**Files:**
- Create: `backend/internal/common/timex/timezone.go`
- Modify: `backend/internal/reporting/trend_service.go`
- Modify: `backend/internal/reporting/handler.go`
- Test: `backend/internal/reporting/reporting_trend_test.go`
- Test: `backend/internal/reporting/reporting_perf_test.go`

- [ ] **Step 1: Write failing tests for timezone boundary and empty-window response**

```go
func TestTrend_UsesUserTimezoneOrUTC8Default(t *testing.T) {}
func TestTrend_EmptyWindowReturnsZeroSeries(t *testing.T) {}
func TestTrend_InvalidParams_ReturnsSTAT_QUERY_INVALID(t *testing.T) {}
func TestTrend_TimeoutReturnsSTAT_TIMEOUT_NoPartialPayload(t *testing.T) {}
func TestTrend_ReadOnlyDegradationRejectsHeavyWindow(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/reporting -run 'Trend_|InvalidParams' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement timezone resolver and timeout handling**

```go
func ResolveUserTZ(v string) *time.Location // user tz else UTC+8
```

- [ ] **Step 4: Re-run trend/perf tests**

Run: `cd backend && go test ./internal/reporting -run 'Trend_|Perf1K' -v`  
Expected: PASS and perf assertion `P95 <= 1s`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/common/timex/timezone.go backend/internal/reporting/trend_service.go backend/internal/reporting/handler.go backend/internal/reporting/reporting_trend_test.go backend/internal/reporting/reporting_perf_test.go
git commit -m "feat(reporting): add timezone and empty-window semantics with performance guard"
```

---

## Chunk 5: Plan-E Portability

### Task 12: Implement CSV import preview endpoint (`/api/import/csv`)

**Files:**
- Create: `backend/internal/portability/import_preview_service.go`
- Create: `backend/internal/portability/handler.go`
- Test: `backend/internal/portability/import_preview_test.go`

- [ ] **Step 1: Write failing tests for upload parse + preview + invalid-file error**

```go
func TestImportPreview_ReturnsDetectedColumnsAndSampleRows(t *testing.T) {}
func TestImportPreview_IncludesMappingSlotsForConfirmStage(t *testing.T) {}
func TestImportPreview_InvalidFile_ReturnsIMPORT_INVALID_FILE(t *testing.T) {}
func TestImportPreview_AcceptsAccessAndPAT(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/portability -run 'ImportPreview' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement preview parser and response schema**

```go
type PreviewResponse struct {
  Columns []string
  SampleRows [][]string
  MappingSlots []string // amount/date/description/category/account/tag...
  MappingCandidates map[string][]string // non-AI candidates for manual selection
}
```

- [ ] **Step 4: Re-run tests + upload API smoke**

Run: `cd backend && go test ./internal/portability -run 'ImportPreview' -v`  
Expected: PASS.

Run: `curl -s -X POST localhost:8080/api/import/csv -H "Authorization: Bearer ${ACCESS_TOKEN}" -F "file=@./testdata/invalid.csv"`  
Expected: `400` + `IMPORT_INVALID_FILE`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/portability/import_preview_service.go backend/internal/portability/handler.go backend/internal/portability/import_preview_test.go
git commit -m "feat(portability): add csv import preview endpoint"
```

### Task 13: Implement CSV import confirm with idempotency + triple dedup + partial failure report

**Files:**
- Create: `backend/internal/portability/import_confirm_service.go`
- Create: `backend/internal/portability/repository.go`
- Modify: `backend/internal/portability/handler.go`
- Modify: `backend/internal/bootstrap/http/router.go`
- Create: `backend/migrations/0004_init_pat_import_jobs.sql`
- Test: `backend/internal/portability/import_confirm_test.go`

- [ ] **Step 1: Write failing tests for rule precedence and partial failure summary**

```go
func TestImportConfirm_UsesIdempotencyBeforeTripleDedup(t *testing.T) {}
func TestImportConfirm_ReturnsSuccessSkipFailBreakdown(t *testing.T) {}
func TestImportConfirm_DuplicateRequest_ReturnsIMPORT_DUPLICATE_REQUEST(t *testing.T) {}
func TestImportConfirm_PartialFailure_ReturnsIMPORT_PARTIAL_FAILED(t *testing.T) {}
func TestImportConfirm_ResponseIncludesPerRowReason(t *testing.T) {}
func TestImportConfirm_AcceptsAccessAndPAT(t *testing.T) {}
func TestImportConfirm_MissingIdempotencyKey_ReturnsBadRequest(t *testing.T) {}
func TestImportConfirm_IdempotencyWindowBoundary24h(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/portability -run 'ImportConfirm' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement confirm pipeline**

```go
// 1) X-Idempotency-Key (24h)
// 2) amount+date+description dedup
// 3) write rows and collect success/skip/fail reasons
```

- [ ] **Step 4: Re-run tests + idempotency replay smoke**

Run: `cd backend && go test ./internal/portability -run 'ImportConfirm' -v`  
Expected: PASS.

Run: `curl -s -X POST localhost:8080/api/import/csv/confirm -H "Authorization: Bearer ${ACCESS_TOKEN}" -H 'X-Idempotency-Key: import-1' -H 'Content-Type: application/json' -d @confirm.json` (run twice)  
Expected: second call no extra writes and returns `IMPORT_DUPLICATE_REQUEST`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/portability/import_confirm_service.go backend/internal/portability/repository.go backend/internal/portability/handler.go backend/internal/bootstrap/http/router.go backend/migrations/0004_init_pat_import_jobs.sql backend/internal/portability/import_confirm_test.go
git commit -m "feat(portability): add import confirm with idempotency dedup and partial-failure reporting"
```

### Task 14: Implement export contracts and timeout/performance behavior

**Files:**
- Create: `backend/internal/portability/export_service.go`
- Modify: `backend/internal/portability/handler.go`
- Test: `backend/internal/portability/export_test.go`

**Responsibilities:**
- `export_service.go`: CSV/JSON generation and timeout policy.
- `handler.go`: `/api/export` request validation and error mapping.

- [ ] **Step 1: Write failing tests for PAT boundaries and export semantics**

```go
func TestExport_PreservesHistoricalCategoryNames(t *testing.T) {}
func TestExport_AcceptsAccessAndPAT(t *testing.T) {}
func TestExport_InvalidRange_ReturnsEXPORT_INVALID_RANGE(t *testing.T) {}
func TestExport_Timeout_ReturnsEXPORT_TIMEOUT(t *testing.T) {}
func TestExport_SupportsCSVAndJSON(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/portability -run 'Export_' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement export service and endpoint mapping**

```go
// validate date range and format
// enforce timeout path -> EXPORT_TIMEOUT
```

- [ ] **Step 4: Re-run tests + SLO checks**

Run: `cd backend && go test ./internal/portability -run 'Export_' -v`  
Expected: PASS.

Run: `cd backend && go test ./internal/portability -run ExportPerf10K -v`  
Expected: benchmark assertion passes with export `<=10s`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/portability/export_service.go backend/internal/portability/handler.go backend/internal/portability/export_test.go
git commit -m "feat(portability): add export contracts with timeout and performance guards"
```

### Task 15: Implement PAT lifecycle, route boundary, and strict-mode propagation

**Files:**
- Create: `backend/internal/portability/pat_service.go`
- Modify: `backend/internal/bootstrap/http/middleware.go`
- Modify: `backend/internal/bootstrap/http/router.go`
- Modify: `backend/internal/portability/handler.go`
- Test: `backend/internal/portability/pat_test.go`

**Responsibilities:**
- `pat_service.go`: PAT create/list/revoke/hash/ttl logic.
- `middleware.go`: PAT vs auth endpoint boundary enforcement.
- `handler.go`: PAT endpoint contracts and error codes.

- [ ] **Step 1: Write failing PAT tests for lifecycle and error contracts**

```go
func TestPAT_CannotAccessAuthEndpoints(t *testing.T) {}
func TestPATEndpoints_AccessOnly_GETPOSTDELETE_On_api_settings_tokens(t *testing.T) {}
func TestPAT_DefaultTTL90Days_StoredAsHash(t *testing.T) {}
func TestPAT_Expired_ReturnsPAT_EXPIRED(t *testing.T) {}
func TestPAT_Revoked_ReturnsPAT_REVOKED(t *testing.T) {}
func TestPAT_RevokePropagationOver5s_TriggersStrictModeAlert(t *testing.T) {}
```

- [ ] **Step 2: Run PAT tests to verify fail**

Run: `cd backend && go test ./internal/portability -run 'PAT_' -v`  
Expected: FAIL.

- [ ] **Step 3: Implement PAT lifecycle and boundary guard**

```go
if isAuthPath(path) && tokenType == PAT { return ErrPATForbiddenOnAuth }
```

- [ ] **Step 4: Re-run tests + propagation SLA check**

Run: `cd backend && go test ./internal/portability -run 'PAT_' -v`  
Expected: PASS including `PAT_EXPIRED` and `PAT_REVOKED` assertions.

Run: revoke PAT then poll protected endpoint for 5 seconds  
Expected: requests rejected within `<=5s`; if not, strict mode + alert event emitted.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/portability/pat_service.go backend/internal/bootstrap/http/middleware.go backend/internal/bootstrap/http/router.go backend/internal/portability/handler.go backend/internal/portability/pat_test.go
git commit -m "feat(portability): add pat lifecycle boundary enforcement and propagation strict mode"
```

---

## Chunk 6: Plan-F Optional Automation + Final Readiness

### Task 16: Implement quick-entry contract with full error and idempotency semantics

**Files:**
- Create: `backend/internal/automation/quick_entry_contract.go`
- Create: `backend/internal/automation/quick_entry_adapter.go`
- Create: `backend/internal/automation/automation_test.go`
- Modify: `backend/internal/accounting/transaction_service.go`

**Responsibilities:**
- `quick_entry_contract.go`: request/response schema and error code mapping.
- `quick_entry_adapter.go`: Go-side contract validation + safe create-transaction call; no retry/orchestration/PAT-validation ownership.
- `automation_test.go`: contract-level acceptance tests.
- `transaction_service.go`: minimal integration point for safe write path, no manual-flow behavior change.

- [ ] **Step 1: Write failing tests for all contract outcomes**

```go
func TestQuickEntry_ForwardedPATInvalidSignal_ReturnsQE_PAT_INVALID_NoWrite(t *testing.T) {}
func TestQuickEntry_LLMUnavailable_ReturnsQE_LLM_UNAVAILABLE_NoWrite(t *testing.T) {}
func TestQuickEntry_ParseFailed_ReturnsQE_PARSE_FAILED_NoWrite(t *testing.T) {}
func TestQuickEntry_Timeout_ReturnsQE_TIMEOUT_NoWrite(t *testing.T) {}
func TestQuickEntry_Success_ReturnsStructuredConfirmation(t *testing.T) {}
func TestQuickEntry_AmbiguousAccountHint_DowngradesToNullAccount_WithHint(t *testing.T) {}
func TestQuickEntry_Idempotency_SamePATSameNormalizedTextSameKey_Dedupes(t *testing.T) {}
func TestQuickEntry_Idempotency_SamePATDifferentNormalizedText_NoDedup(t *testing.T) {}
func TestQuickEntry_Idempotency_ReplayAfter24h_NotDeduped(t *testing.T) {}
func TestQuickEntry_GoValidationError_IsPassthrough(t *testing.T) {}
func TestManualBookkeepingFlow_UnchangedWhenAutomationEnabled(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/automation -v`  
Expected: FAIL.

- [ ] **Step 3: Implement Go adapter as strict contract guard (n8n owns orchestration)**

```go
// validate payload and map QE_* errors
// call accounting service once, pass through domain validation errors
// strict no-write on unresolved parse/timeout/auth failures
// do not change core manual create/edit/delete transaction behavior
// PAT validation remains in n8n; Go only consumes forwarded auth context/signal
```

- [ ] **Step 4: Re-run tests + n8n webhook contract smoke**

Run: `cd backend && go test ./internal/automation -v`  
Expected: PASS.

Run: `curl -s -X POST "${N8N_BASE_URL}/webhook/quick-entry" -H "Authorization: Bearer ${PAT_TOKEN}" -H "X-Idempotency-Key: qe-1" -H 'Content-Type: application/json' -d '{"text":"午饭25元 微信"}'` (run twice)  
Expected: second call returns same result or duplicate-request semantics, and DB transaction count increases by 1 only.

Run: `cd backend && npm run test:n8n-contract`  
Expected: n8n workflow assertions pass: timeout `5s`, retry `<=2`, then returns `QE_TIMEOUT` and writes `0` transactions.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/automation/quick_entry_contract.go backend/internal/automation/quick_entry_adapter.go backend/internal/automation/automation_test.go backend/internal/accounting/transaction_service.go
git commit -m "feat(automation): implement quick-entry contract with strict no-write failure policy"
```

### Task 17: Implement optional CSV mapping suggestions contract

**Files:**
- Create: `backend/internal/automation/csv_mapping_contract.go`
- Create: `backend/internal/automation/csv_mapping_adapter.go`
- Modify: `backend/internal/portability/import_preview_service.go`
- Test: `backend/internal/automation/csv_mapping_test.go`

**Responsibilities:**
- `csv_mapping_contract.go`: AI suggestion request/response schema.
- `csv_mapping_adapter.go`: optional call to n8n mapping flow with fallback behavior.
- `import_preview_service.go`: merge suggestion output into preview response.

- [ ] **Step 1: Write failing tests for suggestion and fallback semantics**

```go
func TestCSVMappingSuggestion_ReturnsSuggestedMappingWhenEnabled(t *testing.T) {}
func TestCSVMappingSuggestion_DisabledOrUnavailable_FallsBackToManualOnly(t *testing.T) {}
func TestCSVMappingSuggestion_Timeout_NoImportBlock(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd backend && go test ./internal/automation -run CSVMapping -v`  
Expected: FAIL.

- [ ] **Step 3: Implement optional suggestion adapter and preview merge**

```go
// if AI enabled: call n8n mapping flow and attach suggested_mapping
// if AI disabled/unavailable: return preview with empty suggested_mapping
```

- [ ] **Step 4: Re-run tests + n8n contract check**

Run: `cd backend && go test ./internal/automation -run CSVMapping -v`  
Expected: PASS.

Run: `cd backend && npm run test:n8n-csv-mapping-contract`  
Expected: timeout/retry policy assertion passes and fallback remains non-blocking.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/automation/csv_mapping_contract.go backend/internal/automation/csv_mapping_adapter.go backend/internal/portability/import_preview_service.go backend/internal/automation/csv_mapping_test.go
git commit -m "feat(automation): add optional csv mapping suggestion with safe fallback"
```

### Task 18: Final docs/OpenAPI/compose/e2e quality gate

**Files:**
- Create: `backend/openapi/openapi.yaml`
- Create: `backend/.env.example`
- Create: `backend/docker-compose.backend.yml`
- Create: `backend/README.md`
- Create: `backend/tests/integration/e2e_test.go`

- [ ] **Step 1: Write failing e2e tests for core path without optional automation**

```go
func TestE2E_CorePath_AuthToExport_WithoutN8N(t *testing.T) {}
```

- [ ] **Step 2: Run e2e test to verify fail**

Run: `cd backend && go test ./tests/integration -run TestE2E_CorePath_AuthToExport_WithoutN8N -v`  
Expected: FAIL.

- [ ] **Step 3: Implement docs and contract artifacts**

```yaml
paths:
  /api/auth/send-code: { post: {} }
  /api/import/csv: { post: {} }
  /api/import/csv/confirm: { post: {} }
```

- [ ] **Step 4: Run full verification suite**

Run: `cd backend && go test ./...`  
Expected: `PASS` and `FAIL: 0`.

Run: `cd backend && golangci-lint run`  
Expected: exit code `0` and no lint diagnostics.

Run: `cd backend && docker compose -f docker-compose.backend.yml up -d && go test ./tests/integration -v`  
Expected: compose services report healthy and `TestE2E_CorePath_AuthToExport_WithoutN8N` passes.

Run: `cd backend && N8N_BASE_URL=http://127.0.0.1:9 go test ./tests/integration -run TestE2E_CorePath_AuthToExport_WithoutN8N -v`  
Expected: core path still PASS while automation endpoint returns enhancement-unavailable semantics.

- [ ] **Step 5: Commit**

```bash
git add backend/openapi/openapi.yaml backend/.env.example backend/docker-compose.backend.yml backend/README.md backend/tests/integration/e2e_test.go
git commit -m "chore(backend): finalize openapi docs compose and integration verification"
```

---

## Cross-Chunk Done Criteria

- [ ] All chunk tests and e2e pass (`go test ./...`).
- [ ] Lint passes (`golangci-lint run`) with no warnings.
- [ ] SMTP auth flow works without n8n.
- [ ] Optional automation can be disabled without affecting core path tests.

---

## Execution Rules

- Do not skip failing-test-first steps.
- Do not combine multiple chunk commits into one giant commit.
- Keep each service/repository file single-purpose; split early when growth appears.
- If spec conflict appears during execution, update the spec first, then resume coding.
