package main

import (
	"bytes"
	"net/http/httptest"
	"fmt"
	"github.com/gin-gonic/gin"
	"xledger/backend/internal/portability"
)

func main() {
	pat := portability.NewPATService(nil)
	recorder := portability.NewInMemoryCallbackRecorder()
	h := portability.NewShortcutHandler(pat, nil, nil, nil, recorder)
	
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"name": "Quick Entry"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	
	// Simulate middleware
	c.Set("user_id", "testuser@example.com")
	
	h.GenerateShortcut(c)
	fmt.Println("Done")
}
