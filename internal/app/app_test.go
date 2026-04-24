package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bangumi-pikpak/internal/config"
	"bangumi-pikpak/internal/pikpak"
	"bangumi-pikpak/internal/rss"
)

type fakePikPak struct {
	loginCalls   int
	folderCalls  int
	offlineCalls int
	duplicate    bool
	loginErr     error
	folderNames  []string
	offlineURLs  []string
	hasChildren  map[string]bool
}

func (f *fakePikPak) Login() error { f.loginCalls++; return f.loginErr }
func (f *fakePikPak) EnsureFolder(parentID, name string) (string, error) {
	f.folderCalls++
	f.folderNames = append(f.folderNames, name)
	return parentID + "/" + name, nil
}
func (f *fakePikPak) HasOriginalURL(parentID, targetURL string) (bool, error) {
	return f.duplicate, nil
}
func (f *fakePikPak) HasChildren(parentID string) (bool, error) {
	return f.hasChildren[parentID], nil
}
func (f *fakePikPak) OfflineDownload(name, fileURL, parentID string) (pikpak.RemoteTask, error) {
	f.offlineCalls++
	f.offlineURLs = append(f.offlineURLs, fileURL)
	return pikpak.RemoteTask{ID: "task", Name: name}, nil
}

func testLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func TestRunOnceNoNewTorrentSkipsLogin(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "torrent", "Bangumi", "第01集", ".downloaded")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("cached"), 0o600); err != nil {
		t.Fatal(err)
	}
	pp := &fakePikPak{}
	runner := Runner{
		Config:      config.Config{Username: "u", Password: "p", Path: "parent", RSS: "rss"},
		HTTPClient:  http.DefaultClient,
		TorrentRoot: filepath.Join(dir, "torrent"),
		PikPak:      pp,
		EntriesFunc: func(context.Context) ([]ResolvedEntry, error) {
			return []ResolvedEntry{{Entry: rss.Entry{Title: "Episode 01", Link: "l", TorrentURL: "https://example.test/a.torrent"}, BangumiTitle: "Bangumi"}}, nil
		},
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if pp.loginCalls != 0 || pp.offlineCalls != 0 {
		t.Fatalf("unexpected calls: %#v", pp)
	}
}

func TestRunOnceNewTorrentSubmitsOfflineTask(t *testing.T) {
	dir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("torrent")) }))
	defer server.Close()
	pp := &fakePikPak{}
	runner := Runner{
		Config:      config.Config{Username: "u", Password: "p", Path: "parent", RSS: "rss"},
		HTTPClient:  server.Client(),
		TorrentRoot: filepath.Join(dir, "torrent"),
		PikPak:      pp,
		EntriesFunc: func(context.Context) ([]ResolvedEntry, error) {
			return []ResolvedEntry{{Entry: rss.Entry{Title: "Episode 01", Link: "l", TorrentURL: server.URL + "/a.torrent"}, BangumiTitle: "Bangumi"}}, nil
		},
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if pp.loginCalls != 1 || pp.folderCalls != 2 || pp.offlineCalls != 1 {
		t.Fatalf("call mismatch: %#v", pp)
	}
}

func TestRunOnceDuplicateRemoteSkipsOfflineDownload(t *testing.T) {
	dir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("torrent")) }))
	defer server.Close()
	pp := &fakePikPak{duplicate: true}
	runner := Runner{
		Config:      config.Config{Username: "u", Password: "p", Path: "parent", RSS: "rss"},
		HTTPClient:  server.Client(),
		TorrentRoot: filepath.Join(dir, "torrent"),
		PikPak:      pp,
		EntriesFunc: func(context.Context) ([]ResolvedEntry, error) {
			return []ResolvedEntry{{Entry: rss.Entry{Title: "Episode 01", Link: "l", TorrentURL: server.URL + "/a.torrent"}, BangumiTitle: "Bangumi"}}, nil
		},
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if pp.loginCalls != 1 || pp.folderCalls != 2 || pp.offlineCalls != 0 {
		t.Fatalf("call mismatch: %#v", pp)
	}
}

func TestRunOnceLogsRSSNewTorrentAndPikPakLogin(t *testing.T) {
	dir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("torrent")) }))
	defer server.Close()
	pp := &fakePikPak{}
	var logs bytes.Buffer
	runner := Runner{
		Config:      config.Config{Username: "user@example.com", Password: "p", Path: "parent", RSS: "rss"},
		HTTPClient:  server.Client(),
		Logger:      testLogger(&logs),
		TorrentRoot: filepath.Join(dir, "torrent"),
		PikPak:      pp,
		EntriesFunc: func(context.Context) ([]ResolvedEntry, error) {
			return []ResolvedEntry{{Entry: rss.Entry{Title: "Episode 01", Link: "l", TorrentURL: server.URL + "/a.torrent"}, BangumiTitle: "Bangumi"}}, nil
		},
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	out := logs.String()
	for _, want := range []string{
		"resolved RSS entries",
		"detected new torrent",
		"new torrents detected, logging in to PikPak",
		"PikPak login succeeded",
		"submitted offline download",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected logs to contain %q, got:\n%s", want, out)
		}
	}
}

func TestRunOnceWrapsPikPakLoginError(t *testing.T) {
	dir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("torrent")) }))
	defer server.Close()
	pp := &fakePikPak{loginErr: errors.New("invalid_argument")}
	runner := Runner{
		Config:      config.Config{Username: "user@example.com", Password: "p", Path: "parent", RSS: "rss"},
		HTTPClient:  server.Client(),
		TorrentRoot: filepath.Join(dir, "torrent"),
		PikPak:      pp,
		EntriesFunc: func(context.Context) ([]ResolvedEntry, error) {
			return []ResolvedEntry{{Entry: rss.Entry{Title: "Episode 01", Link: "l", TorrentURL: server.URL + "/a.torrent"}, BangumiTitle: "Bangumi"}}, nil
		},
	}
	err := runner.RunOnce(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "pikpak login") || !strings.Contains(err.Error(), "invalid_argument") {
		t.Fatalf("expected wrapped login error, got %v", err)
	}
}

func TestRunOnceDeduplicatesSameBangumiEpisodeByRSSOrder(t *testing.T) {
	dir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("torrent")) }))
	defer server.Close()
	pp := &fakePikPak{}
	firstURL := server.URL + "/first.torrent"
	secondURL := server.URL + "/second.torrent"
	runner := Runner{
		Config:      config.Config{Username: "u", Password: "p", Path: "root", RSS: "rss"},
		HTTPClient:  server.Client(),
		TorrentRoot: filepath.Join(dir, "torrent"),
		PikPak:      pp,
		EntriesFunc: func(context.Context) ([]ResolvedEntry, error) {
			return []ResolvedEntry{
				{Entry: rss.Entry{Title: "[GroupA] Test Bangumi [03][1080P]", Link: "l1", TorrentURL: firstURL}, BangumiTitle: "Test Bangumi"},
				{Entry: rss.Entry{Title: "[GroupB] Test Bangumi - 03 (1080P)", Link: "l2", TorrentURL: secondURL}, BangumiTitle: "Test Bangumi"},
			}, nil
		},
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if pp.offlineCalls != 1 {
		t.Fatalf("expected one offline task, got %d", pp.offlineCalls)
	}
	if len(pp.offlineURLs) != 1 || pp.offlineURLs[0] != firstURL {
		t.Fatalf("expected first RSS torrent to be selected, got %#v", pp.offlineURLs)
	}
	wantFolders := []string{"Test Bangumi", "第03集"}
	if strings.Join(pp.folderNames, ",") != strings.Join(wantFolders, ",") {
		t.Fatalf("folder hierarchy mismatch: got %#v want %#v", pp.folderNames, wantFolders)
	}
}

func TestRunOnceSkipsEpisodeWithLocalMarker(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "torrent", "Test Bangumi", "第03集", ".downloaded")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("done"), 0o600); err != nil {
		t.Fatal(err)
	}
	pp := &fakePikPak{}
	runner := Runner{
		Config:      config.Config{Username: "u", Password: "p", Path: "root", RSS: "rss"},
		HTTPClient:  http.DefaultClient,
		TorrentRoot: filepath.Join(dir, "torrent"),
		PikPak:      pp,
		EntriesFunc: func(context.Context) ([]ResolvedEntry, error) {
			return []ResolvedEntry{{Entry: rss.Entry{Title: "[GroupA] Test Bangumi [03][1080P]", Link: "l1", TorrentURL: "https://example.test/first.torrent"}, BangumiTitle: "Test Bangumi"}}, nil
		},
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if pp.loginCalls != 0 || pp.offlineCalls != 0 {
		t.Fatalf("expected marker to skip login/offline, got %#v", pp)
	}
}
