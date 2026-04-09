package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"xledger/backend/internal/bootstrap/config"
	bootstraphttp "xledger/backend/internal/bootstrap/http"
	"xledger/backend/internal/bootstrap/infrastructure"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			handleMigrateCommand()
			return
		case "serve":
		default:
			printUsage()
			return
		}
	}

	startServer()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd/api serve          - Start the HTTP server")
	fmt.Println("  go run ./cmd/api migrate up     - Run all pending migrations")
	fmt.Println("  go run ./cmd/api migrate down   - Rollback the last migration")
	fmt.Println("  go run ./cmd/api migrate down N - Rollback the last N migrations")
	fmt.Println("  go run ./cmd/api migrate status - Show migration status")
}

func handleMigrateCommand() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required for migration commands")
	}

	ctx := context.Background()
	db, err := infrastructure.ConnectPostgres(ctx, infrastructure.PostgresConfig{
		URL:             cfg.DatabaseURL,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		PingTimeout:     5 * time.Second,
	})
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer db.Close()

	switch os.Args[2] {
	case "up":
		if err := infrastructure.ApplyMigrations(ctx, db, "migrations"); err != nil {
			log.Fatalf("migration up error: %v", err)
		}
		log.Println("All migrations applied successfully")

	case "down":
		steps := 1
		if len(os.Args) > 3 {
			if _, err := fmt.Sscanf(os.Args[3], "%d", &steps); err != nil {
				log.Fatalf("invalid steps value: %v", err)
			}
		}
		n, err := infrastructure.RollbackMigrations(ctx, db, "migrations", steps)
		if err != nil {
			log.Fatalf("migration down error: %v", err)
		}
		log.Printf("Rolled back %d migration(s) successfully", n)

	case "status":
		statuses, err := infrastructure.GetMigrationStatus(ctx, db)
		if err != nil {
			log.Fatalf("migration status error: %v", err)
		}
		if len(statuses) == 0 {
			log.Println("No migrations applied")
			return
		}
		log.Println("Applied migrations:")
		for _, status := range statuses {
			log.Printf("  %s (applied at %s)", status.Filename, status.AppliedAt.Format(time.RFC3339))
		}

	default:
		printUsage()
	}
}

func startServer() {
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

	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		redisClient, err = infrastructure.ConnectRedis(ctx, infrastructure.RedisConfig{
			URL:         cfg.RedisURL,
			PingTimeout: 3 * time.Second,
		})
		if err != nil {
			log.Fatalf("redis connection error: %v", err)
		}
		defer redisClient.Close()
		log.Println("Connected to Redis")
	} else {
		log.Println("REDIS_URL not set, skipping Redis connection")
	}

	var router http.Handler
	if db != nil {
		router = bootstraphttp.NewRouterWithPostgreSQL(db, cfg, redisClient)
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
