package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"bangumi-pikpak/internal/app"
	"bangumi-pikpak/internal/cache"
	"bangumi-pikpak/internal/config"
	"bangumi-pikpak/internal/logger"
	"bangumi-pikpak/internal/pikpak"
	"bangumi-pikpak/internal/proxy"
	"bangumi-pikpak/internal/store"
	"bangumi-pikpak/internal/web"
)

func main() {
	configPath := flag.String("config", ".env", "path to .env config")
	intervalSeconds := flag.Int("interval", 600, "RSS polling interval in seconds")
	once := flag.Bool("once", false, "run one sync cycle and exit")
	webMode := flag.Bool("web", true, "serve the Vue web UI by default; set -web=false for CLI-only sync")
	addr := flag.String("addr", ":8080", "web UI listen address")
	staticDir := flag.String("static", "frontend/dist", "Vue web UI dist directory")
	logFile := flag.String("log", "rss-pikpak.log", "log file path")
	statePath := flag.String("state", "pikpak.json", "PikPak runtime state file path")
	flag.Parse()

	log := logger.New(*logFile)
	slog.SetDefault(log)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("load config failed", "error", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		log.Error("config validation failed", "error", err)
		os.Exit(1)
	}
	if proxy.Apply(cfg) {
		log.Info("proxy configuration enabled")
	}

	mysqlStore, err := store.OpenMySQL(context.Background(), cfg)
	if err != nil {
		log.Error("mysql initialization failed", "error", err)
		os.Exit(1)
	}
	if mysqlStore != nil {
		defer mysqlStore.Close()
		log.Info("mysql state management enabled", "database", cfg.MySQLDatabase)
	}
	redisCache, err := cache.OpenRedis(context.Background(), cfg)
	if err != nil {
		log.Warn("redis initialization failed, continue without redis cache", "error", err)
	}
	if redisCache != nil {
		defer redisCache.Close()
		log.Info("redis cache enabled", "addr", cfg.RedisAddr, "ttl", "15m")
	}

	api, err := pikpak.NewGoAPI(cfg.Username, cfg.Password)
	if err != nil {
		log.Error("create pikpak client failed", "error", err)
		os.Exit(1)
	}
	pp := &lockedPikPak{inner: pikpak.NewAdapter(api)}
	runner := app.Runner{Config: cfg, HTTPClient: proxy.HTTPClient(), Logger: log, TorrentRoot: "torrent", PikPak: pp, Store: mysqlStore}
	scheduler := &syncScheduler{runner: runner, log: log, statePath: *statePath, username: cfg.Username}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if *webMode {
		webServer := web.Server{Config: cfg, HTTPClient: proxy.HTTPClient(), Logger: log, TorrentRoot: "torrent", StaticDir: *staticDir, PikPak: pp, Store: mysqlStore, Cache: redisCache}
		httpServer := &http.Server{Addr: *addr, Handler: webServer.Handler()}
		errCh := make(chan error, 1)
		go func() {
			log.Info("web UI listening", "addr", *addr)
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
			}
		}()
		if !*once {
			interval := time.Duration(*intervalSeconds) * time.Second
			go scheduler.Loop(ctx, interval, 15*time.Second)
		}
		select {
		case <-ctx.Done():
			log.Info("shutdown requested")
		case err := <-errCh:
			log.Error("web UI server failed", "error", err)
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("web UI shutdown failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if *once {
		if !scheduler.Run(ctx) {
			os.Exit(1)
		}
		return
	}

	interval := time.Duration(*intervalSeconds) * time.Second
	scheduler.Loop(ctx, interval, 0)
}

type syncScheduler struct {
	mu        sync.Mutex
	running   bool
	runner    app.Runner
	log       *slog.Logger
	statePath string
	username  string
}

func (s *syncScheduler) Run(ctx context.Context) bool {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		s.log.Info("skip RSS sync because previous cycle is still running")
		return true
	}
	s.running = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	if err := s.runner.RunOnce(ctx); err != nil {
		s.log.Error("sync cycle failed", "error", err)
		return false
	}
	_ = pikpak.SaveState(s.statePath, pikpak.State{Username: s.username, LastLoginTime: time.Now(), LastRefreshTime: time.Now(), Client: "pikpak-go"})
	return true
}

func (s *syncScheduler) Loop(ctx context.Context, interval, initialDelay time.Duration) {
	if initialDelay > 0 {
		s.log.Info("RSS sync scheduled after web startup", "initial_delay", initialDelay.String(), "interval", interval.String())
		select {
		case <-ctx.Done():
			return
		case <-time.After(initialDelay):
		}
	}
	for {
		s.Run(ctx)
		select {
		case <-ctx.Done():
			s.log.Info("RSS scheduler stopped")
			return
		case <-time.After(interval):
		}
	}
}

type lockedPikPak struct {
	mu        sync.Mutex
	inner     *pikpak.Adapter
	lastLogin time.Time
}

func (l *lockedPikPak) Login() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.lastLogin.IsZero() && time.Since(l.lastLogin) < 20*time.Minute {
		return nil
	}
	if err := l.inner.Login(); err != nil {
		return err
	}
	l.lastLogin = time.Now()
	return nil
}

func (l *lockedPikPak) EnsureFolder(parentID, name string) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.inner.EnsureFolder(parentID, name)
}

func (l *lockedPikPak) HasOriginalURL(parentID, targetURL string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.inner.HasOriginalURL(parentID, targetURL)
}

func (l *lockedPikPak) HasChildren(parentID string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.inner.HasChildren(parentID)
}

func (l *lockedPikPak) OfflineDownload(name, fileURL, parentID string) (pikpak.RemoteTask, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.inner.OfflineDownload(name, fileURL, parentID)
}

func (l *lockedPikPak) List(parentID string) ([]pikpak.RemoteFile, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.inner.List(parentID)
}

func (l *lockedPikPak) DownloadURL(id string) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.inner.DownloadURL(id)
}
