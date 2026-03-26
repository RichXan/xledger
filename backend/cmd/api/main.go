package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"xledger/backend/internal/bootstrap/config"
	bootstraphttp "xledger/backend/internal/bootstrap/http"
	"xledger/backend/internal/bootstrap/infrastructure"
)

func main() {
	_ = godotenv.Load(".env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()

	var db *sql.DB
	if cfg.DatabaseURL != "" {
		db, err = infrastructure.ConnectPostgres(ctx, infrastructure.PostgresConfig{
			URL:             cfg.DatabaseURL,
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			PingTimeout:     5 * time.Second,
		})
		if err != nil {
			log.Fatalf("database connection error: %v", err)
		}
		if err := infrastructure.ApplyMigrations(ctx, db, "migrations"); err != nil {
			log.Fatalf("database migration error: %v", err)
		}
		defer db.Close()
		log.Println("Connected to PostgreSQL")
	} else {
		log.Println("DATABASE_URL not set, using in-memory repositories")
	}

	if cfg.RedisURL != "" {
		redisClient, redisErr := infrastructure.ConnectRedis(ctx, infrastructure.RedisConfig{
			URL:         cfg.RedisURL,
			PingTimeout: 3 * time.Second,
		})
		if redisErr != nil {
			log.Fatalf("redis connection error: %v", redisErr)
		}
		defer redisClient.Close()
		log.Println("Connected to Redis")
	} else {
		log.Println("REDIS_URL not set, skipping Redis connection")
	}

	var router http.Handler
	if db != nil {
		router = bootstraphttp.NewRouterWithPostgreSQL(db, cfg)
	} else {
		router, err = bootstraphttp.NewRouter(cfg.TrustedProxies)
		if err != nil {
			log.Fatalf("startup error: %v", err)
		}
	}

	server := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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
	case <-shutdownCtx.Done():
	}

	gracefulCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(gracefulCtx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}

	if err := <-errCh; err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
