package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisReportingCache implements reporting.Cache backed by Redis.
type RedisReportingCache struct {
	client *redis.Client
}

func NewRedisReportingCache(client *redis.Client) *RedisReportingCache {
	return &RedisReportingCache{client: client}
}

func (c *RedisReportingCache) Get(key string) ([]byte, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	val, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

func (c *RedisReportingCache) Set(key string, value []byte, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisReportingCache) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	return c.client.Del(ctx, key).Err()
}
