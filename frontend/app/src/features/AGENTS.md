# Frontend Features

Feature-based organization. Each feature = API + hooks + optional components.

## Structure

```
features/
├── auth/               # Session management, route guards
│   ├── auth-api.ts     # /auth/* endpoints
│   ├── auth-context.tsx # React context for session
│   ├── auth-storage.ts # localStorage persistence
│   └── require-auth.tsx # Route guard component
├── transactions/       # Transaction CRUD + import
│   ├── transactions-api.ts
│   └── transactions-hooks.ts
├── reporting/          # Dashboard stats
│   ├── reporting-api.ts
│   └── reporting-hooks.ts
├── management/         # PATs, export
│   ├── management-api.ts
│   └── management-hooks.ts
└── pwa/                # Service worker, install prompt
```

## Pattern

1. `*-api.ts` — Wraps backend endpoints via `requestEnvelope<T>()`
2. `*-hooks.ts` — Wraps API calls in React Query mutations/queries
3. Pages consume hooks directly

## Conventions

- Hooks: `use` prefix, return `{ data, isLoading, error }`
- API functions: accept `accessToken` as first param
- Error handling: `ApiError` class with `status` and `code`
- Query keys: `['feature', 'entity', ...params]`

## Don't

- Don't call API functions directly from components
- Don't skip `enabled` guard on queries (check `session?.accessToken`)
- Don't use `any` for response types
