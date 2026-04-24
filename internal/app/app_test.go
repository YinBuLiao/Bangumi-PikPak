package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
}

func (f *fakePikPak) Login() error { f.loginCalls++; return nil }
func (f *fakePikPak) EnsureFolder(parentID, name string) (string, error) {
	f.folderCalls++
	return "folder-id", nil
}
func (f *fakePikPak) HasOriginalURL(parentID, targetURL string) (bool, error) {
	return f.duplicate, nil
}
func (f *fakePikPak) OfflineDownload(name, fileURL, parentID string) (pikpak.RemoteTask, error) {
	f.offlineCalls++
	return pikpak.RemoteTask{ID: "task", Name: name}, nil
}

func TestRunOnceNoNewTorrentSkipsLogin(t *testing.T) {
	dir := t.TempDir()
	localPath := filepath.Join(dir, "torrent", "Bangumi", "a.torrent")
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localPath, []byte("cached"), 0o600); err != nil {
		t.Fatal(err)
	}
	pp := &fakePikPak{}
	runner := Runner{
		Config:      config.Config{Username: "u", Password: "p", Path: "parent", RSS: "rss"},
		HTTPClient:  http.DefaultClient,
		TorrentRoot: filepath.Join(dir, "torrent"),
		PikPak:      pp,
		EntriesFunc: func(context.Context) ([]ResolvedEntry, error) {
			return []ResolvedEntry{{Entry: rss.Entry{Title: "e", Link: "l", TorrentURL: "https://example.test/a.torrent"}, BangumiTitle: "Bangumi"}}, nil
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
			return []ResolvedEntry{{Entry: rss.Entry{Title: "e", Link: "l", TorrentURL: server.URL + "/a.torrent"}, BangumiTitle: "Bangumi"}}, nil
		},
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if pp.loginCalls != 1 || pp.folderCalls != 1 || pp.offlineCalls != 1 {
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
			return []ResolvedEntry{{Entry: rss.Entry{Title: "e", Link: "l", TorrentURL: server.URL + "/a.torrent"}, BangumiTitle: "Bangumi"}}, nil
		},
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if pp.loginCalls != 1 || pp.folderCalls != 1 || pp.offlineCalls != 0 {
		t.Fatalf("call mismatch: %#v", pp)
	}
}
