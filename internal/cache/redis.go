package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bangumi-pikpak/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	prefix string
}

func OpenRedis(ctx context.Context, cfg config.Config) (*RedisCache, error) {
	if !cfg.RedisConfigured() {
		return nil, nil
	}
	client := redis.NewClient(&redis.Options{Addr: strings.TrimSpace(cfg.RedisAddr), Password: cfg.RedisPassword, DB: cfg.RedisDB})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &RedisCache{client: client, prefix: "bangumi-pikpak:"}, nil
}

func (c *RedisCache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *RedisCache) GetJSON(ctx context.Context, key string, target any) (bool, error) {
	if c == nil || c.client == nil || strings.TrimSpace(key) == "" {
		return false, nil
	}
	raw, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return false, err
	}
	return true, nil
}

func (c *RedisCache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c == nil || c.client == nil || strings.TrimSpace(key) == "" || ttl <= 0 {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.prefix+key, raw, ttl).Err()
}
