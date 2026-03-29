# Auth Package

Authentication: email verification codes, OAuth, sessions, refresh tokens.

## Files

| File | Purpose |
|------|---------|
| `handler.go` | HTTP handlers for auth endpoints |
| `code_service.go` | Verification code generation/validation |
| `session_service.go` | Session + token management |
| `oauth_service.go` | OAuth flow orchestration |
| `google_oauth_provider.go` | Google OAuth implementation |
| `password_service.go` | Password hashing (optional) |
| `smtp_sender.go` | Email sending |
| `repository*.go` | In-memory + Postgres storage |

## Key Rules

- Auth code = 6-digit, expires in 10 min
- Refresh token rotation on use
- Sessions tracked in `refresh_tokens` table
- `AUTH_CODE_PEPPER` required for code hashing

## Patterns

- Rate limiting via `send_locks` + `ip_rate_limits` tables
- OAuth state stored in `oauth_states` table
- Token blacklist for revoked refresh tokens

## Don't

- Don't skip rate limiting on code send
- Don't reuse consumed refresh tokens
- Don't expose internal error details to client
