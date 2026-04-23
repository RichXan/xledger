# xledger e2e-tests

Midscene + Playwright based end-to-end tests for Xledger.

## Quick Start

```bash
cd /home/xan/paramita/xledger/e2e-tests
pnpm install
pnpm run e2e:install
```

Use one of the model configurations:

```bash
# Codex app-server mode
export MIDSCENE_MODEL_BASE_URL="codex://app-server"
export MIDSCENE_MODEL_NAME="gpt-5.4"
export MIDSCENE_MODEL_FAMILY="gpt-5"
```

Then run:

```bash
E2E_BASE_URL="http://127.0.0.1:4173" \
E2E_API_BASE_URL="http://127.0.0.1:8080/api" \
pnpm run e2e
```

## Specs

- `e2e/accounts.spec.ts`
- `e2e/account-ledger-management.spec.ts`
- `e2e/transactions.spec.ts`
- `e2e/transfer-lifecycle.spec.ts`
- `e2e/reporting.spec.ts`
- `e2e/classification-export.spec.ts`
- `e2e/pat.spec.ts`
- `e2e/quality.visual.spec.ts` (`@visual`)
- `e2e/quality.a11y.spec.ts` (`@a11y`)

## CI Parallel

`playwright.config.ts` uses `workers=4` when `CI=true`, and the specs are independent so they can run in parallel safely.

## Quality Gates

- Visual baseline compare: `pnpm run e2e:visual`
- Visual baseline update: `pnpm run e2e:visual:update`
- Accessibility audit (axe): `pnpm run e2e:a11y`

`pnpm run e2e` and `pnpm run e2e:ci` exclude `@visual` and `@a11y` by default.
