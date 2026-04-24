package cache

import (
	"context"
	"testing"

	"bangumi-pikpak/internal/config"
)

func TestOpenRedisReturnsNilWhenNotConfigured(t *testing.T) {
	c, err := OpenRedis(context.Background(), config.Config{})
	if err != nil {
		t.Fatalf("OpenRedis returned error: %v", err)
	}
	if c != nil {
		t.Fatal("expected nil redis cache")
	}
}
