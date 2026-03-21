# Xledger Backend

## Run tests

```bash
go test ./...
```

## Integration test

```bash
go test ./tests/integration -run TestE2E_CorePath_AuthToExport_WithoutN8N -count=1
```

## API contract

OpenAPI spec is available at `openapi/openapi.yaml`.

## Environment

Copy `.env.example` and fill SMTP / optional N8N values.
