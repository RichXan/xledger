package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorMapping maps an error code string to an HTTP status and response envelope.
type ErrorMapping struct {
	Status  int
	Code    string
	Message string
}

// RecoveryMiddleware recovers from panics and returns a proper 500 error.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ErrorHandler is a helper that writes an error response based on a mapping table.
// It checks the error's ErrorCode (if it implements the errorCodeGetter interface)
// against the provided mappings, falling back to defaultMapping if no match found.
type ErrorHandler struct {
	mappings       map[string]ErrorMapping
	defaultMapping ErrorMapping
}

type errorCodeGetter interface {
	ErrorCode() string
}

// NewErrorHandler creates an ErrorHandler with the given mappings and default.
func NewErrorHandler(mappings map[string]ErrorMapping, defaultMapping ErrorMapping) *ErrorHandler {
	return &ErrorHandler{
		mappings:       mappings,
		defaultMapping: defaultMapping,
	}
}

// Handle looks up the error code and writes the appropriate response.
// If the error implements errorCodeGetter, it uses that code; otherwise it uses err.Error().
func (h *ErrorHandler) Handle(c *gin.Context, err error) {
	code := ""
	if eg, ok := err.(errorCodeGetter); ok {
		code = eg.ErrorCode()
	} else {
		code = err.Error()
	}

	if mapping, ok := h.mappings[code]; ok {
		JSON(c, mapping.Status, mapping.Code, mapping.Message, nil)
		return
	}
	JSON(c, h.defaultMapping.Status, h.defaultMapping.Code, h.defaultMapping.Message, nil)
}

// WriteError writes a simple error response with the given status, code, and message.
// This is a convenience function for straightforward error cases.
func WriteError(c *gin.Context, status int, code string, message string) {
	JSON(c, status, code, message, nil)
}
