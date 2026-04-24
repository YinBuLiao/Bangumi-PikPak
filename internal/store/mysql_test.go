package store

import (
	"context"
	"testing"

	"bangumi-pikpak/internal/config"
)

func TestOpenMySQLReturnsNilWhenNotConfigured(t *testing.T) {
	s, err := OpenMySQL(context.Background(), config.Config{})
	if err != nil {
		t.Fatalf("OpenMySQL returned error: %v", err)
	}
	if s != nil {
		t.Fatalf("expected nil store when mysql is not configured")
	}
}

func TestConfigDetectsMySQLConfigured(t *testing.T) {
	cfg := config.Config{MySQLHost: "127.0.0.1", MySQLDatabase: "anime", MySQLUsername: "anime"}
	if !cfg.MySQLConfigured() {
		t.Fatal("expected mysql configured")
	}
}
