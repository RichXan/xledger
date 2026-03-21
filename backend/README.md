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

## API contract

- Base prefix: `/api`
- JSON responses use unified envelope: `{ code, message, data }`
- Paginated lists use `data.items` and `data.pagination`
- Partial update resources use `PATCH`
- PAT resource path: `/api/personal-access-tokens`
- OpenAPI spec: `openapi/openapi.yaml`

## Environment

Copy `.env.example` and fill SMTP / optional N8N values.
