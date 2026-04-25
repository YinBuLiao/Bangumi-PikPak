package web

import (
	"net/http"
	"sync"

	"bangumi-pikpak/internal/cache"
	"bangumi-pikpak/internal/config"
	"bangumi-pikpak/internal/storage"
	"bangumi-pikpak/internal/store"
)

type RuntimeState struct {
	mu          sync.RWMutex
	Config      config.Config
	Installed   bool
	InstallOnly bool
	HTTPClient  *http.Client
	PikPak      PikPakClient
	Storage     storage.Provider
	Store       *store.MySQLStore
	Cache       *cache.RedisCache
}

func NewRuntimeState(cfg config.Config, installed, installOnly bool, httpClient *http.Client, pikpak PikPakClient, storageProvider storage.Provider, mysqlStore *store.MySQLStore, redisCache *cache.RedisCache) *RuntimeState {
	return &RuntimeState{Config: cfg, Installed: installed, InstallOnly: installOnly, HTTPClient: httpClient, PikPak: pikpak, Storage: storageProvider, Store: mysqlStore, Cache: redisCache}
}

func (r *RuntimeState) Snapshot() (config.Config, bool, bool, *http.Client, PikPakClient, storage.Provider, *store.MySQLStore, *cache.RedisCache) {
	if r == nil {
		return config.Config{}, false, false, nil, nil, nil, nil, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Config, r.Installed, r.InstallOnly, r.HTTPClient, r.PikPak, r.Storage, r.Store, r.Cache
}

func (r *RuntimeState) Update(cfg config.Config, installed, installOnly bool, httpClient *http.Client, pikpak PikPakClient, storageProvider storage.Provider, mysqlStore *store.MySQLStore, redisCache *cache.RedisCache) {
	if r == nil {
		return
	}
	r.mu.Lock()
	oldStore := r.Store
	oldCache := r.Cache
	r.Config = cfg
	r.Installed = installed
	r.InstallOnly = installOnly
	r.HTTPClient = httpClient
	r.PikPak = pikpak
	r.Storage = storageProvider
	r.Store = mysqlStore
	r.Cache = redisCache
	r.mu.Unlock()
	if oldStore != nil && oldStore != mysqlStore {
		_ = oldStore.Close()
	}
	if oldCache != nil && oldCache != redisCache {
		_ = oldCache.Close()
	}
}
