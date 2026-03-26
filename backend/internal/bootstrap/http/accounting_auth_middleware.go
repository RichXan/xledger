package http

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/auth"
)

type userIDResolver func(ctx context.Context, email string) (string, error)

func accountingAuthMiddleware(resolver userIDResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[len("Bearer "):])
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
			return
		}

		email, ok := parseBusinessAuthEmail(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
			return
		}

		resolvedUserID := stableUserIDFromEmail(email)
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
		}

		c.Set("user_id", resolvedUserID)
		c.Next()
	}
}

func parseBusinessAuthEmail(token string) (string, bool) {
	parsed, err := auth.ParseSessionToken(token)
	if err == nil && parsed.Type == "access" && strings.TrimSpace(parsed.Email) != "" && time.Now().UTC().Before(parsed.ExpiresAt) {
		return strings.TrimSpace(parsed.Email), true
	}

	lowerToken := strings.ToLower(token)
	if !strings.HasPrefix(lowerToken, "pat:") {
		return "", false
	}

	rawEmail := strings.TrimSpace(token[len("pat:"):])
	if rawEmail == "" {
		return "", false
	}
	if idx := strings.Index(rawEmail, ":"); idx >= 0 {
		rawEmail = strings.TrimSpace(rawEmail[:idx])
	}
	if rawEmail == "" {
		return "", false
	}
	return rawEmail, true
}

func stableUserIDFromEmail(email string) string {
	normalized := strings.TrimSpace(strings.ToLower(email))
	sum := sha1.Sum([]byte(normalized))
	bytes := sum[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x50
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	hexID := hex.EncodeToString(bytes)
	return hexID[0:8] + "-" + hexID[8:12] + "-" + hexID[12:16] + "-" + hexID[16:20] + "-" + hexID[20:32]
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
