package main

import (
	"log"
	"os"

	"xledger/backend/internal/bootstrap/config"
	bootstraphttp "xledger/backend/internal/bootstrap/http"
)

func main() {
	_, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	router := bootstraphttp.NewRouter()

	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
