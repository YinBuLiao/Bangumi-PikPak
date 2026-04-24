package proxy

import (
	"net/http"
	"os"

	"bangumi-pikpak/internal/config"
)

func Apply(cfg config.Config) bool {
	if !cfg.EnableProxy {
		return false
	}
	if cfg.HTTPProxy != "" {
		os.Setenv("HTTP_PROXY", cfg.HTTPProxy)
		os.Setenv("http_proxy", cfg.HTTPProxy)
	}
	if cfg.HTTPSProxy != "" {
		os.Setenv("HTTPS_PROXY", cfg.HTTPSProxy)
		os.Setenv("https_proxy", cfg.HTTPSProxy)
	}
	if cfg.SocksProxy != "" {
		os.Setenv("SOCKS_PROXY", cfg.SocksProxy)
		os.Setenv("socks_proxy", cfg.SocksProxy)
	}
	return true
}

func HTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyFromEnvironment
	return &http.Client{Transport: transport}
}
