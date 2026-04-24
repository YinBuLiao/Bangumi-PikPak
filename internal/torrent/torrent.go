package torrent

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"bangumi-pikpak/internal/sanitize"
)

const markerFile = ".downloaded"
const metadataFile = ".metadata.json"

type BangumiMetadata struct {
	Title    string `json:"title"`
	CoverURL string `json:"cover_url,omitempty"`
	Summary  string `json:"summary,omitempty"`
}

func LocalPath(root, bangumiTitle, torrentURL string) (string, error) {
	return LocalEpisodePath(root, bangumiTitle, "", torrentURL)
}

func LocalEpisodePath(root, bangumiTitle, episodeLabel, torrentURL string) (string, error) {
	parsed, err := url.Parse(torrentURL)
	if err != nil {
		return "", fmt.Errorf("parse torrent url: %w", err)
	}
	name := path.Base(parsed.Path)
	if parsed.Scheme == "magnet" {
		sum := sha1.Sum([]byte(torrentURL))
		name = hex.EncodeToString(sum[:])[:16] + ".magnet"
	}
	if name == "." || name == "/" || name == "" {
		return "", fmt.Errorf("torrent url has no filename: %s", torrentURL)
	}
	parts := []string{root, sanitize.Name(bangumiTitle)}
	if episodeLabel != "" {
		parts = append(parts, sanitize.Name(episodeLabel))
	}
	parts = append(parts, sanitize.Name(name))
	return filepath.Join(parts...), nil
}

func EpisodeDir(root, bangumiTitle, episodeLabel string) string {
	return filepath.Join(root, sanitize.Name(bangumiTitle), sanitize.Name(episodeLabel))
}

func MarkerPath(root, bangumiTitle, episodeLabel string) string {
	return filepath.Join(EpisodeDir(root, bangumiTitle, episodeLabel), markerFile)
}

func MarkerExists(root, bangumiTitle, episodeLabel string) bool {
	return Exists(MarkerPath(root, bangumiTitle, episodeLabel))
}

func MarkDownloaded(root, bangumiTitle, episodeLabel, torrentURL string) error {
	marker := MarkerPath(root, bangumiTitle, episodeLabel)
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		return fmt.Errorf("create marker dir: %w", err)
	}
	content := []byte(fmt.Sprintf("downloaded_at=%s\ntorrent=%s\n", time.Now().Format(time.RFC3339), torrentURL))
	if err := os.WriteFile(marker, content, 0o600); err != nil {
		return fmt.Errorf("write marker: %w", err)
	}
	return nil
}

func MetadataPath(root, bangumiTitle string) string {
	return filepath.Join(root, sanitize.Name(bangumiTitle), metadataFile)
}

func SaveBangumiMetadata(root, bangumiTitle string, meta BangumiMetadata) error {
	if meta.Title == "" {
		meta.Title = bangumiTitle
	}
	b, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal bangumi metadata: %w", err)
	}
	b = append(b, '\n')
	target := MetadataPath(root, bangumiTitle)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create metadata dir: %w", err)
	}
	if err := os.WriteFile(target, b, 0o600); err != nil {
		return fmt.Errorf("write bangumi metadata: %w", err)
	}
	return nil
}

func LoadBangumiMetadata(root, bangumiTitle string) (BangumiMetadata, error) {
	b, err := os.ReadFile(MetadataPath(root, bangumiTitle))
	if err != nil {
		return BangumiMetadata{}, fmt.Errorf("read bangumi metadata: %w", err)
	}
	var meta BangumiMetadata
	if err := json.Unmarshal(b, &meta); err != nil {
		return BangumiMetadata{}, fmt.Errorf("parse bangumi metadata: %w", err)
	}
	return meta, nil
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
