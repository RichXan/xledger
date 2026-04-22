# Xledger Backend

## Run tests

```bash
go test ./...
```

## Integration tests

```bash
go test ./tests/integration -count=1
N8N_BASE_URL=http://127.0.0.1:9 go test ./tests/integration -run TestE2E_CorePath_AuthToExport_WithoutN8N -count=1
```

## Docker Compose infra + validation

```bash
docker compose -f docker-compose.backend.yml up -d
DATABASE_URL=postgres://xledger:xledger_secret@127.0.0.1:5432/xledger?sslmode=disable \
REDIS_URL=redis://127.0.0.1:6379/0 \
SMTP_HOST=smtp.example.com AUTH_CODE_PEPPER=local-pepper \
go test ./... -count=1
DATABASE_URL=postgres://xledger:xledger_secret@127.0.0.1:5432/xledger?sslmode=disable \
REDIS_URL=redis://127.0.0.1:6379/0 \
SMTP_HOST=smtp.example.com AUTH_CODE_PEPPER=local-pepper \
timeout 8s go run ./cmd/api || [ $? -eq 124 ]
docker compose -f docker-compose.backend.yml down -v
```

This flow will:
- start `postgres` and `redis`
- run `go test ./...`
- smoke-run `go run ./cmd/api` to verify startup can connect to both Postgres and Redis

## Full deployment stack

If you want to run the full application stack in a deployment-oriented way, use the root `deploy/` directory instead of the backend-only compose files:

```bash
cd ../deploy
docker compose up -d --build
docker compose ps
docker compose down
```

This flow will:
- start `postgres`, `redis`, `xledger-backend`, and `xledger-frontend`
- serve the frontend through Nginx on port `4173`
- expose the backend API on port `8080`
- use `backend/config/config.yaml` as the backend config source for deployment

## API contract

- Base prefix: `/api`
- JSON responses use unified envelope: `{ code, message, data }`
- Paginated lists use `data.items` and `data.pagination`
- Partial update resources use `PATCH`
- PAT resource path: `/api/personal-access-tokens`
- OpenAPI spec: `openapi/openapi.yaml`

## Environment

Copy `.env.example` and fill SMTP / optional N8N values.
