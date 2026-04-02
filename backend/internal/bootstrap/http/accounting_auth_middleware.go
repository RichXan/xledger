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
	"xledger/backend/internal/common/httpx"
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
			httpx.JSON(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "未认证或凭证无效", nil)
			c.Abort()
			return
		}

		email, ok := parseBusinessAuthEmail(token, patService, c.Request.URL.Path)
		if !ok {
			httpx.JSON(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "未认证或凭证无效", nil)
			c.Abort()
			return
		}

		resolvedUserID := ""
		if resolver != nil {
			value, err := resolver(c.Request.Context(), email)
			if err != nil {
				httpx.JSON(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "未认证或凭证无效", nil)
				c.Abort()
				return
			}
			resolvedUserID = strings.TrimSpace(value)
			if resolvedUserID == "" {
				httpx.JSON(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "未认证或凭证无效", nil)
				c.Abort()
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
		if patService == nil {
			return "", false
		}
		email, valErr := patService.ValidatePAT(context.Background(), token, path)
		if valErr == nil && email != "" {
			return email, true
		}
		return "", false
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
