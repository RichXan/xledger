package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type PostgresConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	PingTimeout     time.Duration
}

func ConnectPostgres(ctx context.Context, cfg PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	pingCtx := ctx
	cancel := func() {}
	if cfg.PingTimeout > 0 {
		pingCtx, cancel = context.WithTimeout(ctx, cfg.PingTimeout)
	}
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

type RedisConfig struct {
	URL         string
	PingTimeout time.Duration
}

func ConnectRedis(ctx context.Context, cfg RedisConfig) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	pingCtx := ctx
	cancel := func() {}
	if cfg.PingTimeout > 0 {
		pingCtx, cancel = context.WithTimeout(ctx, cfg.PingTimeout)
	}
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
