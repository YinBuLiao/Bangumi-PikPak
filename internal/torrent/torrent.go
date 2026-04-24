package torrent

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"bangumi-pikpak/internal/sanitize"
)

func LocalPath(root, bangumiTitle, torrentURL string) (string, error) {
	parsed, err := url.Parse(torrentURL)
	if err != nil {
		return "", fmt.Errorf("parse torrent url: %w", err)
	}
	name := path.Base(parsed.Path)
	if name == "." || name == "/" || name == "" {
		return "", fmt.Errorf("torrent url has no filename: %s", torrentURL)
	}
	return filepath.Join(root, sanitize.Name(bangumiTitle), sanitize.Name(name)), nil
}

func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func Download(client *http.Client, url, target string) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download torrent: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download torrent: status %s", resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create torrent dir: %w", err)
	}
	f, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create torrent file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write torrent file: %w", err)
	}
	return nil
}
