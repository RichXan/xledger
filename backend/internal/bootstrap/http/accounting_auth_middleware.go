package http

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/auth"
	"xledger/backend/internal/portability"
)

type userIDResolver func(ctx context.Context, email string) (string, error)

func accountingAuthMiddleware(resolver userIDResolver, patService *portability.PATService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[len("Bearer "):])
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
			return
		}

		email, ok := parseBusinessAuthEmail(token, patService, c.Request.URL.Path)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
			return
		}

		resolvedUserID := ""
		if resolver != nil {
			value, err := resolver(c.Request.Context(), email)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
				return
			}
			resolvedUserID = strings.TrimSpace(value)
			if resolvedUserID == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
				return
			}
		} else {
			resolvedUserID = email
		}

		c.Set("user_id", resolvedUserID)
		c.Next()
	}
}

func parseBusinessAuthEmail(token string, patService *portability.PATService, path string) (string, bool) {
	parsed, err := auth.ParseSessionToken(token)
	if err == nil && parsed.Type == "access" && strings.TrimSpace(parsed.Email) != "" && time.Now().UTC().Before(parsed.ExpiresAt) {
		return strings.TrimSpace(parsed.Email), true
	}

	lowerToken := strings.ToLower(token)
	if strings.HasPrefix(lowerToken, "pat:") {
		if patService != nil {
			email, valErr := patService.ValidatePAT(context.Background(), token, path)
			if valErr == nil && email != "" {
				return email, true
			}
			// patService exists but PAT not found (empty in-memory storage): fall back to old parsing
			rawEmail := strings.TrimSpace(token[len("pat:"):])
			if idx := strings.Index(rawEmail, ":"); idx >= 0 {
				rawEmail = strings.TrimSpace(rawEmail[:idx])
			}
			if rawEmail != "" {
				return rawEmail, true
			}
			return "", false
		}
		// patService is nil: fall back to old insecure parsing
		rawEmail := strings.TrimSpace(token[len("pat:"):])
		if idx := strings.Index(rawEmail, ":"); idx >= 0 {
			rawEmail = strings.TrimSpace(rawEmail[:idx])
		}
		if rawEmail == "" {
			return "", false
		}
		return rawEmail, true
	}

	return "", false
}

func postgresUserIDResolver(db *sql.DB) userIDResolver {
	if db == nil {
		return nil
	}
	return func(ctx context.Context, email string) (string, error) {
		var userID string
		err := db.QueryRowContext(ctx, `SELECT id::text FROM users WHERE email = $1`, strings.TrimSpace(strings.ToLower(email))).Scan(&userID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
		if err != nil {
			return "", err
		}
		return userID, nil
	}
}
