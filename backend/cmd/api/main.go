package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xledger/backend/internal/bootstrap/config"
	bootstraphttp "xledger/backend/internal/bootstrap/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	router := bootstraphttp.NewRouter()
	server := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err == nil || err == http.ErrServerClosed {
			errCh <- nil
			return
		}

		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server stopped: %v", err)
		}
		return
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}

	if err := <-errCh; err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
