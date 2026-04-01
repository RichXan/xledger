package httpx

import "github.com/gin-gonic/gin"

// UserIDFromContext extracts the user ID from the gin context.
// Returns the user ID and true if present, or empty string and false if not.
func UserIDFromContext(c *gin.Context) (string, bool) {
	if value, exists := c.Get("user_id"); exists {
		if userID, ok := value.(string); ok && userID != "" {
			return userID, true
		}
	}
	return "", false
}
