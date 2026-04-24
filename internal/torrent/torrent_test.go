package torrent

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
