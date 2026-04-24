package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"bangumi-pikpak/internal/app"
	"bangumi-pikpak/internal/bangumi"
	"bangumi-pikpak/internal/cache"
	"bangumi-pikpak/internal/config"
	"bangumi-pikpak/internal/mikan"
	"bangumi-pikpak/internal/rss"
	"bangumi-pikpak/internal/store"
)

type PikPakClient interface {
	app.PikPakClient
	DriveClient
}

type Server struct {
	Config      config.Config
	HTTPClient  *http.Client
	Logger      *slog.Logger
	TorrentRoot string
	StaticDir   string
	PikPak      PikPakClient
	Store       *store.MySQLStore
	Cache       *cache.RedisCache
}

type downloadRequest struct {
	Title        string `json:"title"`
	BangumiTitle string `json:"bangumi_title"`
	CoverURL     string `json:"cover_url"`
	Summary      string `json:"summary"`
	Link         string `json:"link"`
	TorrentURL   string `json:"torrent_url"`
	Magnet       string `json:"magnet"`
}

type searchResponse struct {
	Results      []mikan.SearchResult `json:"results"`
	Query        string               `json:"query,omitempty"`
	MatchedQuery string               `json:"matched_query,omitempty"`
}

const librarySnapshotKey = "pikpak_library"
const librarySnapshotTTL = 10 * time.Minute
const externalCacheTTL = 15 * time.Minute

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/favicon.ico", s.handleFavicon)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/library", s.handleLibrary)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/bangumi/discover", s.handleBangumiDiscover)
	mux.HandleFunc("/api/mikan/schedule", s.handleMikanSchedule)
	mux.HandleFunc("/api/mikan/subscribe", s.handleMikanSubscribe)
	mux.HandleFunc("/api/download", s.handleDownload)
	mux.HandleFunc("/api/sync", s.handleSync)
	mux.HandleFunc("/api/stream", s.handleStream)
	return loggingMiddleware(s.log(), mux)
}

func (s Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	staticDir := s.staticDir()
	requested := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if requested == "." {
		requested = ""
	}
	if requested != "" {
		target := filepath.Join(staticDir, requested)
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			http.ServeFile(w, r, target)
			return
		}
	}
	index := filepath.Join(staticDir, "index.html")
	if _, err := os.Stat(index); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`<!doctype html><meta charset="utf-8"><title>Bangumi PikPak</title><h1>前端还没有构建</h1><p>请先运行：<code>cd frontend && npm install && npm run build</code></p>`))
		return
	}
	http.ServeFile(w, r, index)
}

func (s Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"rss_configured":  strings.TrimSpace(s.Config.RSS) != "",
		"path_configured": strings.TrimSpace(s.Config.Path) != "",
		"torrent_root":    s.torrentRoot(),
		"time":            time.Now().Format(time.RFC3339),
	})
}

func (s Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	forceRefresh := r.URL.Query().Get("refresh") == "1"
	if !forceRefresh {
		if lib, updatedAt, ok := s.loadCachedLibrary(ctx); ok && time.Since(updatedAt) < librarySnapshotTTL {
			s.log().Info("serve library from mysql snapshot", "age", time.Since(updatedAt).String())
			writeJSON(w, http.StatusOK, lib)
			return
		}
	}
	if err := s.PikPak.Login(); err != nil {
		if lib, updatedAt, ok := s.loadCachedLibrary(ctx); ok {
			s.log().Warn("pikpak login failed, serve stale library snapshot", "age", time.Since(updatedAt).String(), "error", err)
			writeJSON(w, http.StatusOK, lib)
			return
		}
		writeError(w, http.StatusBadGateway, fmt.Errorf("pikpak login: %w", err))
		return
	}
	lib, err := BuildLibrary(ctx, s.PikPak, s.Config.Path, s.torrentRoot())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if err := s.applyStoredLibraryMetadata(ctx, &lib); err != nil {
		s.log().Warn("mysql library metadata enrichment failed", "error", err)
	}
	if err := enrichLibraryCovers(ctx, &lib, s.coverResolver(), s.torrentRoot(), s.log()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if err := s.saveLibraryToStore(ctx, lib); err != nil {
		s.log().Warn("mysql library upsert failed", "error", err)
	}
	if s.Store != nil {
		if err := s.Store.SaveSnapshot(ctx, librarySnapshotKey, lib); err != nil {
			s.log().Warn("mysql library snapshot save failed", "error", err)
		}
	}
	writeJSON(w, http.StatusOK, lib)
}

func (s Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, map[string]any{"results": []mikan.SearchResult{}})
		return
	}
	cacheKey := "search:mikan:v2:" + q
	if s.Cache != nil {
		var cached searchResponse
		if ok, err := s.Cache.GetJSON(r.Context(), cacheKey, &cached); err == nil && ok {
			s.log().Info("serve mikan search from redis", "query", q)
			writeJSON(w, http.StatusOK, cached)
			return
		} else if err != nil {
			s.log().Warn("redis mikan search read failed", "query", q, "error", err)
		}
	}
	results, matchedQuery, err := s.searchMikanWithFallback(r.Context(), q, 18)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	for i := range results {
		if i >= 10 {
			break
		}
		meta, err := mikan.FetchEpisodeMetadata(s.httpClient(), results[i].Link)
		if err != nil {
			s.log().Warn("mikan metadata enrichment failed", "title", results[i].Title, "error", err)
			continue
		}
		results[i].BangumiTitle = meta.Title
		if bgmMeta, err := s.coverResolver().SearchMetadata(meta.Title); err == nil && (bgmMeta.CoverURL != "" || bgmMeta.Summary != "") {
			results[i].CoverURL = bgmMeta.CoverURL
			results[i].Summary = bgmMeta.Summary
		} else {
			results[i].CoverURL = meta.CoverURL
		}
	}
	resp := searchResponse{Results: results, Query: q, MatchedQuery: matchedQuery}
	if s.Cache != nil {
		if err := s.Cache.SetJSON(r.Context(), cacheKey, resp, externalCacheTTL); err != nil {
			s.log().Warn("redis mikan search write failed", "query", q, "error", err)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s Server) searchMikanWithFallback(ctx context.Context, q string, limit int) ([]mikan.SearchResult, string, error) {
	candidates := []string{q}
	if subjects, err := s.coverResolver().SearchSubjects(q, 5); err == nil {
		for _, subject := range subjects {
			candidates = append(candidates, subject.Name, subject.NameCN)
		}
	} else {
		s.log().Warn("bangumi fallback keyword lookup failed", "query", q, "error", err)
	}
	candidates = append(candidates, cjkFallbackKeywords(q)...)
	candidates = uniqueNonEmpty(candidates)
	var lastErr error
	for _, candidate := range candidates {
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		default:
		}
		results, err := mikan.Search(s.httpClient(), candidate, limit)
		if err != nil {
			lastErr = err
			s.log().Warn("mikan search candidate failed", "query", q, "candidate", candidate, "error", err)
			continue
		}
		if len(results) > 0 {
			if candidate != q {
				s.log().Info("mikan search fallback matched", "query", q, "matched_query", candidate, "count", len(results))
			}
			return results, candidate, nil
		}
	}
	if lastErr != nil {
		return nil, "", lastErr
	}
	return []mikan.SearchResult{}, "", nil
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

func cjkFallbackKeywords(q string) []string {
	normalized := make([]rune, 0, len(q))
	for _, r := range q {
		if unicode.Is(unicode.Han, r) {
			normalized = append(normalized, r)
		}
	}
	if len(normalized) < 4 || len(normalized) > 12 {
		return nil
	}
	out := make([]string, 0, 6)
	for size := minInt(4, len(normalized)-1); size >= 2; size-- {
		for i := 0; i+size <= len(normalized); i++ {
			out = append(out, string(normalized[i:i+size]))
			if len(out) >= 6 {
				return out
			}
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s Server) handleBangumiDiscover(w http.ResponseWriter, r *http.Request) {
	section := strings.TrimSpace(r.URL.Query().Get("section"))
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	limit := queryInt(r, "limit", 24)
	offset := queryInt(r, "offset", 0)
	if limit <= 0 || limit > 50 {
		limit = 24
	}
	cacheKey := fmt.Sprintf("bangumi:discover:%s:%s:%d:%d", section, tag, limit, offset)
	if s.Cache != nil {
		var cached map[string]any
		if ok, err := s.Cache.GetJSON(r.Context(), cacheKey, &cached); err == nil && ok {
			s.log().Info("serve bangumi discover from redis", "section", section, "tag", tag, "offset", offset)
			writeJSON(w, http.StatusOK, cached)
			return
		} else if err != nil {
			s.log().Warn("redis bangumi discover read failed", "error", err)
		}
	}
	items, err := s.coverResolver().DiscoverPage(section, tag, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	resp := map[string]any{"subjects": items, "limit": limit, "offset": offset, "has_more": len(items) == limit}
	if s.Cache != nil {
		if err := s.Cache.SetJSON(r.Context(), cacheKey, resp, externalCacheTTL); err != nil {
			s.log().Warn("redis bangumi discover write failed", "error", err)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s Server) handleMikanSchedule(w http.ResponseWriter, r *http.Request) {
	year := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("year")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &year)
	}
	season := strings.TrimSpace(r.URL.Query().Get("season"))
	coverMode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("cover")))
	limitParam := queryInt(r, "limit", 0)
	cacheKey := fmt.Sprintf("mikan:schedule:%d:%s:%s:%d", year, season, coverMode, limitParam)
	if s.Cache != nil {
		var cached mikan.ScheduleResponse
		if ok, err := s.Cache.GetJSON(r.Context(), cacheKey, &cached); err == nil && ok {
			s.log().Info("serve mikan schedule from redis", "year", year, "season", season, "cover", coverMode)
			writeJSON(w, http.StatusOK, cached)
			return
		} else if err != nil {
			s.log().Warn("redis mikan schedule read failed", "error", err)
		}
	}
	schedule, err := mikan.FetchSchedule(s.httpClient(), year, season)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if coverMode == "bangumi" {
		limit := limitParam
		if limit == 0 {
			limit = len(schedule.Items)
		}
		if limit <= 0 || limit > len(schedule.Items) {
			limit = len(schedule.Items)
		}
		s.enrichMikanScheduleCovers(r.Context(), &schedule, limit)
	}
	if s.Cache != nil {
		if err := s.Cache.SetJSON(r.Context(), cacheKey, schedule, externalCacheTTL); err != nil {
			s.log().Warn("redis mikan schedule write failed", "error", err)
		}
	}
	writeJSON(w, http.StatusOK, schedule)
}

func (s Server) enrichMikanScheduleCovers(ctx context.Context, schedule *mikan.ScheduleResponse, limit int) {
	if schedule == nil || limit <= 0 {
		return
	}
	resolver := s.coverResolver()
	coverByID := make(map[int]string)
	coverByTitle := make(map[string]string)
	for i := range schedule.Items {
		if i >= limit {
			break
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
		var meta bangumi.Metadata
		if schedule.Items[i].PageURL != "" {
			subjectID, err := mikan.FetchBangumiSubjectID(s.httpClient(), schedule.Items[i].PageURL)
			if err == nil && subjectID > 0 {
				subject, err := resolver.GetSubject(subjectID)
				if err == nil {
					meta = bangumi.Metadata{ID: subject.ID, Title: firstNonEmpty(subject.NameCN, subject.Name), CoverURL: subject.CoverURL(), Summary: subject.Summary}
				} else {
					s.log().Warn("bangumi.tv subject cover lookup failed", "title", schedule.Items[i].Title, "subject_id", subjectID, "error", err)
				}
			}
		}
		if meta.CoverURL == "" {
			var err error
			meta, err = resolver.SearchMetadata(schedule.Items[i].Title)
			if err != nil {
				s.log().Warn("bangumi.tv schedule cover lookup failed", "title", schedule.Items[i].Title, "error", err)
				continue
			}
		}
		if meta.CoverURL == "" {
			continue
		}
		schedule.Items[i].CoverURL = meta.CoverURL
		schedule.Items[i].CoverFrom = "bangumi"
		if schedule.Items[i].ID > 0 {
			coverByID[schedule.Items[i].ID] = meta.CoverURL
		}
		coverByTitle[schedule.Items[i].Title] = meta.CoverURL
	}
	for d := range schedule.Days {
		for i := range schedule.Days[d].Items {
			if cover, ok := coverByID[schedule.Days[d].Items[i].ID]; ok {
				schedule.Days[d].Items[i].CoverURL = cover
				schedule.Days[d].Items[i].CoverFrom = "bangumi"
				continue
			}
			if cover, ok := coverByTitle[schedule.Days[d].Items[i].Title]; ok {
				schedule.Days[d].Items[i].CoverURL = cover
				schedule.Days[d].Items[i].CoverFrom = "bangumi"
			}
		}
	}
}

type subscribeRequest struct {
	SubjectID int    `json:"subject_id"`
	Title     string `json:"title"`
	CoverURL  string `json:"cover_url"`
	Summary   string `json:"summary"`
	Language  int    `json:"language"`
}

func (s Server) handleMikanSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	if !s.Config.MikanConfigured() {
		writeError(w, http.StatusBadRequest, errors.New("mikan username/password is not configured"))
		return
	}
	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	client := mikan.NewSession(s.httpClient())
	if err := mikan.Login(client, s.Config.MikanUsername, s.Config.MikanPassword); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	candidate, err := mikan.FindBangumiCandidate(client, req.Title)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if err := mikan.SubscribeBangumi(client, candidate.ID, req.Language); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if s.Store != nil {
		if err := s.Store.SaveSubscription(r.Context(), store.BangumiRecord{Title: req.Title, CoverURL: firstNonEmpty(req.CoverURL, candidate.CoverURL), Summary: req.Summary, BangumiSubjectID: req.SubjectID, MikanBangumiID: candidate.ID}, req.Language); err != nil {
			s.log().Warn("mysql subscription save failed", "title", req.Title, "error", err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "已在 Mikan 订阅", "mikan_bangumi": candidate})
}

func (s Server) applyStoredLibraryMetadata(ctx context.Context, lib *LibraryResponse) error {
	if s.Store == nil || lib == nil {
		return nil
	}
	for i := range lib.Bangumi {
		rec, ok, err := s.Store.Metadata(ctx, lib.Bangumi[i].Title)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if lib.Bangumi[i].CoverURL == "" {
			lib.Bangumi[i].CoverURL = rec.CoverURL
		}
		if lib.Bangumi[i].Summary == "" {
			lib.Bangumi[i].Summary = rec.Summary
		}
	}
	return nil
}

func (s Server) saveLibraryToStore(ctx context.Context, lib LibraryResponse) error {
	if s.Store == nil {
		return nil
	}
	for _, item := range lib.Bangumi {
		if err := s.Store.UpsertBangumi(ctx, store.BangumiRecord{Title: item.Title, PikPakFolderID: item.ID, CoverURL: item.CoverURL, Summary: item.Summary}); err != nil {
			return err
		}
	}
	return nil
}

func (s Server) loadCachedLibrary(ctx context.Context) (LibraryResponse, time.Time, bool) {
	if s.Store == nil {
		return LibraryResponse{}, time.Time{}, false
	}
	var lib LibraryResponse
	updatedAt, ok, err := s.Store.LoadSnapshot(ctx, librarySnapshotKey, &lib)
	if err != nil {
		s.log().Warn("mysql library snapshot load failed", "error", err)
		return LibraryResponse{}, time.Time{}, false
	}
	return lib, updatedAt, ok
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func (s Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req downloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.TorrentURL = strings.TrimSpace(req.TorrentURL)
	req.Magnet = strings.TrimSpace(req.Magnet)
	downloadURL := req.TorrentURL
	if downloadURL == "" {
		downloadURL = req.Magnet
	}
	if req.Title == "" || downloadURL == "" {
		writeError(w, http.StatusBadRequest, errors.New("title and torrent_url/magnet are required"))
		return
	}
	if req.BangumiTitle == "" && req.Link != "" {
		if meta, err := mikan.FetchEpisodeMetadata(s.httpClient(), req.Link); err == nil {
			req.BangumiTitle = meta.Title
			req.CoverURL = meta.CoverURL
		} else {
			s.log().Warn("mikan metadata resolve failed before download", "title", req.Title, "error", err)
		}
	}
	if req.BangumiTitle == "" {
		req.BangumiTitle = req.Title
	}
	if bgmMeta, err := s.coverResolver().SearchMetadata(req.BangumiTitle); err == nil && (bgmMeta.CoverURL != "" || bgmMeta.Summary != "") {
		req.CoverURL = bgmMeta.CoverURL
		req.Summary = bgmMeta.Summary
	} else if err != nil {
		s.log().Warn("bangumi.tv metadata lookup failed before download", "bangumi", req.BangumiTitle, "error", err)
	}
	runner := app.Runner{
		Config:      s.Config,
		HTTPClient:  s.httpClient(),
		Logger:      s.log(),
		TorrentRoot: s.torrentRoot(),
		PikPak:      s.PikPak,
		Store:       s.Store,
		EntriesFunc: func(context.Context) ([]app.ResolvedEntry, error) {
			return []app.ResolvedEntry{{
				Entry:        rss.Entry{Title: req.Title, Link: req.Link, TorrentURL: downloadURL},
				BangumiTitle: req.BangumiTitle,
				CoverURL:     req.CoverURL,
				Summary:      req.Summary,
			}}, nil
		},
	}
	if err := runner.RunOnce(r.Context()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "已提交到 PikPak 离线下载队列"})
}

func (s Server) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	runner := app.Runner{Config: s.Config, HTTPClient: s.httpClient(), Logger: s.log(), TorrentRoot: s.torrentRoot(), PikPak: s.PikPak, Store: s.Store}
	if err := runner.RunOnce(r.Context()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "RSS 搜刮同步已完成"})
}

func (s Server) handleStream(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}
	if err := s.PikPak.Login(); err != nil {
		writeError(w, http.StatusBadGateway, fmt.Errorf("pikpak login: %w", err))
		return
	}
	u, err := s.PikPak.DownloadURL(id)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	http.Redirect(w, r, u, http.StatusFound)
}

func (s Server) httpClient() *http.Client {
	if s.HTTPClient != nil {
		return s.HTTPClient
	}
	return http.DefaultClient
}

func (s Server) torrentRoot() string {
	if s.TorrentRoot != "" {
		return s.TorrentRoot
	}
	return "torrent"
}

func (s Server) staticDir() string {
	if s.StaticDir != "" {
		return s.StaticDir
	}
	return filepath.Join("frontend", "dist")
}

func (s Server) log() *slog.Logger {
	if s.Logger != nil {
		return s.Logger
	}
	return slog.Default()
}

func (s Server) coverResolver() bangumi.Client {
	return bangumi.Client{HTTPClient: s.httpClient()}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
}

func loggingMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Info("web request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
		}
	})
}
