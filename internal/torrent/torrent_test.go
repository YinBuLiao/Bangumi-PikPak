package torrent

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalPathUsesTorrentFilename(t *testing.T) {
	path, err := LocalPath("torrent", "进击/巨人", "https://example.test/Download/abc123.torrent")
	if err != nil {
		t.Fatalf("LocalPath returned error: %v", err)
	}
	want := filepath.Join("torrent", "进击_巨人", "abc123.torrent")
	if path != want {
		t.Fatalf("path mismatch: got %q want %q", path, want)
	}
}

func TestLocalEpisodePathSupportsMagnet(t *testing.T) {
	got, err := LocalEpisodePath("torrent", "测试番剧", "第01集", "magnet:?xt=urn:btih:abcdef")
	if err != nil {
		t.Fatalf("LocalEpisodePath returned error: %v", err)
	}
	if !strings.HasSuffix(got, ".magnet") || !strings.Contains(got, "第01集") {
		t.Fatalf("magnet local path mismatch: %s", got)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.torrent")
	if Exists(file) {
		t.Fatal("file should not exist yet")
	}
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !Exists(file) {
		t.Fatal("file should exist")
	}
}

func TestDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("torrent-data"))
	}))
	defer server.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "a.torrent")
	if err := Download(server.Client(), server.URL+"/a.torrent", path); err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "torrent-data" {
		t.Fatalf("content mismatch: %q", string(b))
	}
}

func TestBangumiMetadataRoundTrip(t *testing.T) {
	dir := t.TempDir()
	meta := BangumiMetadata{Title: "测试番剧", CoverURL: "https://example.test/cover.webp", Summary: "来自 Bangumi.tv 的简介"}
	if err := SaveBangumiMetadata(dir, "测试番剧", meta); err != nil {
		t.Fatalf("SaveBangumiMetadata returned error: %v", err)
	}
	got, err := LoadBangumiMetadata(dir, "测试番剧")
	if err != nil {
		t.Fatalf("LoadBangumiMetadata returned error: %v", err)
	}
	if got.Title != meta.Title || got.CoverURL != meta.CoverURL || got.Summary != meta.Summary {
		t.Fatalf("metadata mismatch: %#v", got)
	}
}
