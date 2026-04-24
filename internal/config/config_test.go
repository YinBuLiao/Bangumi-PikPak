package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadExistingConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := `{
		"username":"user@example.com",
		"password":"secret",
		"path":"folder-id",
		"rss":"https://mikanani.me/RSS/MyBangumi?token=abc",
		"http_proxy":"http://127.0.0.1:7890",
		"https_proxy":"http://127.0.0.1:7890",
		"socks_proxy":"socks5://127.0.0.1:7890",
		"enable_proxy":true
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Username != "user@example.com" || cfg.Password != "secret" || cfg.Path != "folder-id" || cfg.RSS == "" {
		t.Fatalf("loaded config mismatch: %#v", cfg)
	}
	if !cfg.EnableProxy || cfg.SocksProxy == "" {
		t.Fatalf("proxy config mismatch: %#v", cfg)
	}
}

func TestValidateRejectsMissingRequiredFields(t *testing.T) {
	cfg := Config{Username: "user@example.com", Password: "secret"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "path") || !strings.Contains(err.Error(), "rss") {
		t.Fatalf("validation error should mention missing fields, got %v", err)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg := Config{Username: "user@example.com", Password: "secret", Path: "folder-id", RSS: "https://example.test/rss", HTTPProxy: "http://127.0.0.1:7890", EnableProxy: true}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.Username != cfg.Username || loaded.Password != cfg.Password || loaded.Path != cfg.Path || loaded.RSS != cfg.RSS || loaded.HTTPProxy != cfg.HTTPProxy || loaded.EnableProxy != cfg.EnableProxy {
		t.Fatalf("round trip mismatch: %#v", loaded)
	}
}

func TestValidateRejectsExampleValues(t *testing.T) {
	cfg := Config{
		Username: "your_email@example.com",
		Password: "your_password",
		Path:     "your_pikpak_folder_id",
		RSS:      "https://mikanani.me/RSS/MyBangumi?token=your_token_here",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for example values")
	}
	if !strings.Contains(err.Error(), "example") {
		t.Fatalf("validation error should mention example values, got %v", err)
	}
}
