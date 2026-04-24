package app

import (
	"context"
	"log/slog"
	"net/http"
	"path/filepath"

	"bangumi-pikpak/internal/config"
	"bangumi-pikpak/internal/mikan"
	"bangumi-pikpak/internal/pikpak"
	"bangumi-pikpak/internal/rss"
	"bangumi-pikpak/internal/sanitize"
	"bangumi-pikpak/internal/torrent"
)

type ResolvedEntry struct {
	Entry        rss.Entry
	BangumiTitle string
}

type PikPakClient interface {
	Login() error
	EnsureFolder(parentID, name string) (string, error)
	HasOriginalURL(parentID, targetURL string) (bool, error)
	OfflineDownload(name, fileURL, parentID string) (pikpak.RemoteTask, error)
}

type Runner struct {
	Config      config.Config
	HTTPClient  *http.Client
	Logger      *slog.Logger
	TorrentRoot string
	PikPak      PikPakClient
	EntriesFunc func(context.Context) ([]ResolvedEntry, error)
}

func (r Runner) RunOnce(ctx context.Context) error {
	entries, err := r.entries(ctx)
	if err != nil {
		return err
	}
	newEntries := make([]ResolvedEntry, 0, len(entries))
	paths := make(map[string]string, len(entries))
	for _, entry := range entries {
		localPath, err := torrent.LocalPath(r.torrentRoot(), entry.BangumiTitle, entry.Entry.TorrentURL)
		if err != nil {
			r.log().Warn("skip entry with invalid torrent url", "title", entry.Entry.Title, "error", err)
			continue
		}
		paths[entry.Entry.TorrentURL] = localPath
		if !torrent.Exists(localPath) {
			newEntries = append(newEntries, entry)
		}
	}
	if len(newEntries) == 0 {
		r.log().Info("RSS source has no new updates")
		return nil
	}
	if err := r.PikPak.Login(); err != nil {
		return err
	}
	for _, entry := range newEntries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		localPath := paths[entry.Entry.TorrentURL]
		folderName := sanitize.Name(entry.BangumiTitle)
		folderID, err := r.PikPak.EnsureFolder(r.Config.Path, folderName)
		if err != nil {
			r.log().Error("ensure pikpak folder failed", "title", folderName, "error", err)
			continue
		}
		if err := torrent.Download(r.httpClient(), entry.Entry.TorrentURL, localPath); err != nil {
			r.log().Error("download torrent failed", "url", entry.Entry.TorrentURL, "error", err)
			continue
		}
		duplicate, err := r.PikPak.HasOriginalURL(folderID, entry.Entry.TorrentURL)
		if err != nil {
			r.log().Warn("remote duplicate check failed", "title", folderName, "error", err)
		}
		if duplicate {
			r.log().Info("skip duplicate remote torrent", "title", folderName)
			continue
		}
		name := filepath.Base(localPath)
		if _, err := r.PikPak.OfflineDownload(name, entry.Entry.TorrentURL, folderID); err != nil {
			r.log().Error("offline download failed", "title", folderName, "error", err)
			continue
		}
		r.log().Info("submitted offline download", "title", folderName, "torrent", entry.Entry.TorrentURL)
	}
	return nil
}

func (r Runner) entries(ctx context.Context) ([]ResolvedEntry, error) {
	if r.EntriesFunc != nil {
		return r.EntriesFunc(ctx)
	}
	entries, err := rss.Fetch(r.httpClient(), r.Config.RSS)
	if err != nil {
		return nil, err
	}
	resolved := make([]ResolvedEntry, 0, len(entries))
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		title, err := mikan.FetchTitle(r.httpClient(), entry.Link)
		if err != nil {
			r.log().Warn("skip entry because mikan title cannot be resolved", "entry", entry.Title, "error", err)
			continue
		}
		resolved = append(resolved, ResolvedEntry{Entry: entry, BangumiTitle: title})
	}
	return resolved, nil
}

func (r Runner) httpClient() *http.Client {
	if r.HTTPClient != nil {
		return r.HTTPClient
	}
	return http.DefaultClient
}

func (r Runner) torrentRoot() string {
	if r.TorrentRoot != "" {
		return r.TorrentRoot
	}
	return "torrent"
}

func (r Runner) log() *slog.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	return slog.Default()
}
