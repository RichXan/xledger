# Bootstrap Package

Infrastructure wiring: config, HTTP router, database connections, migrations.

## Structure

```
bootstrap/
├── config/
│   └── config.go           # Env-based config loader
├── http/
│   ├── router.go           # Route registration, DI wiring
│   ├── middleware.go        # Auth middleware
│   ├── *_wiring.go          # Per-domain handler wiring
│   └── router_test.go      # Integration-style tests
└── infrastructure/
    ├── connections.go       # Postgres + Redis connections
    └── migrations.go        # Migration runner
```

## Key Behaviors

- `DATABASE_URL` unset → in-memory mode
- `DATABASE_URL` set → Postgres-backed mode
- `SMTP_HOST` + `AUTH_CODE_PEPPER` always required
- Migrations run on startup, tracked in `schema_migrations`

## Wiring Pattern

Each domain has a `*_wiring.go` file that:
1. Creates repository (in-memory or Postgres)
2. Creates service
3. Creates handler
4. Registers routes on router

## Don't

- Don't add routes outside `*_wiring.go` files
- Don't skip migration version tracking
- Don't create circular dependencies between domains
