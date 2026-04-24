package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bangumi-pikpak/internal/app"
	"bangumi-pikpak/internal/config"
	"bangumi-pikpak/internal/logger"
	"bangumi-pikpak/internal/pikpak"
	"bangumi-pikpak/internal/proxy"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config.json")
	intervalSeconds := flag.Int("interval", 600, "RSS polling interval in seconds")
	once := flag.Bool("once", false, "run one sync cycle and exit")
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

	api, err := pikpak.NewGoAPI(cfg.Username, cfg.Password)
	if err != nil {
		log.Error("create pikpak client failed", "error", err)
		os.Exit(1)
	}
	runner := app.Runner{Config: cfg, HTTPClient: proxy.HTTPClient(), Logger: log, TorrentRoot: "torrent", PikPak: pikpak.NewAdapter(api)}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	run := func() bool {
		if err := runner.RunOnce(ctx); err != nil {
			log.Error("sync cycle failed", "error", err)
			return false
		}
		_ = pikpak.SaveState(*statePath, pikpak.State{Username: cfg.Username, LastLoginTime: time.Now(), LastRefreshTime: time.Now(), Client: "pikpak-go"})
		return true
	}

	if *once {
		if !run() {
			os.Exit(1)
		}
		return
	}

	interval := time.Duration(*intervalSeconds) * time.Second
	for {
		run()
		select {
		case <-ctx.Done():
			log.Info("shutdown requested")
			return
		case <-time.After(interval):
		}
	}
}
