package main

import (
	"bytes"
	"net/http/httptest"
	"fmt"
	"xledger/backend/internal/bootstrap/http"
	"xledger/backend/internal/bootstrap/config"
	"io"
)

func main() {
	cfg := config.Config{}
	r, _ := http.NewRouter(cfg.TrustedProxies)
	
	// Create a POST request
	req := httptest.NewRequest("POST", "/api/shortcuts/generate", bytes.NewBufferString(`{"name": "Quick Entry"}`))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Status: %d\nBody: %s\n", res.StatusCode, string(body))
}
