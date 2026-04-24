package proxy

import (
	"os"
	"testing"

	"bangumi-pikpak/internal/config"
)

func TestApplyDisabledClearsNothingAndReturnsFalse(t *testing.T) {
	t.Setenv("HTTP_PROXY", "existing")
	applied := Apply(config.Config{EnableProxy: false})
	if applied {
		t.Fatal("expected Apply to return false")
	}
	if os.Getenv("HTTP_PROXY") != "existing" {
		t.Fatal("disabled proxy should not overwrite existing environment")
	}
}

func TestApplyEnabledSetsEnvironment(t *testing.T) {
	cfg := config.Config{EnableProxy: true, HTTPProxy: "http://127.0.0.1:7890", HTTPSProxy: "http://127.0.0.1:7891", SocksProxy: "socks5://127.0.0.1:7892"}
	applied := Apply(cfg)
	if !applied {
		t.Fatal("expected Apply to return true")
	}
	if os.Getenv("HTTP_PROXY") != cfg.HTTPProxy || os.Getenv("http_proxy") != cfg.HTTPProxy {
		t.Fatal("HTTP proxy env mismatch")
	}
	if os.Getenv("HTTPS_PROXY") != cfg.HTTPSProxy || os.Getenv("https_proxy") != cfg.HTTPSProxy {
		t.Fatal("HTTPS proxy env mismatch")
	}
	if os.Getenv("SOCKS_PROXY") != cfg.SocksProxy || os.Getenv("socks_proxy") != cfg.SocksProxy {
		t.Fatal("SOCKS proxy env mismatch")
	}
}
