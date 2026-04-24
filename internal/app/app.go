package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"bangumi-pikpak/internal/config"
	"bangumi-pikpak/internal/episode"
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
	HasChildren(parentID string) (bool, error)
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

type plannedEntry struct {
	ResolvedEntry
	EpisodeLabel string
	LocalPath    string
}

func (r Runner) RunOnce(ctx context.Context) error {
	entries, err := r.entries(ctx)
	if err != nil {
		return err
	}
	r.log().Info("resolved RSS entries", "count", len(entries))

	newEntries := make([]plannedEntry, 0, len(entries))
	seenEpisodes := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		episodeLabel, ok := episode.LabelFromTitle(entry.Entry.Title)
		if !ok {
			r.log().Warn("skip entry because episode number cannot be parsed", "bangumi", entry.BangumiTitle, "entry", entry.Entry.Title)
			continue
		}
		folderName := sanitize.Name(entry.BangumiTitle)
		episodeLabel = sanitize.Name(episodeLabel)
		key := folderName + "\x00" + episodeLabel
		if _, exists := seenEpisodes[key]; exists {
			r.log().Info("skip duplicate episode release by RSS order", "bangumi", folderName, "episode", episodeLabel, "entry", entry.Entry.Title, "torrent", entry.Entry.TorrentURL)
			continue
		}
		seenEpisodes[key] = struct{}{}

		if torrent.MarkerExists(r.torrentRoot(), folderName, episodeLabel) {
			r.log().Info("episode already processed locally", "bangumi", folderName, "episode", episodeLabel)
			continue
		}
		localPath, err := torrent.LocalEpisodePath(r.torrentRoot(), folderName, episodeLabel, entry.Entry.TorrentURL)
		if err != nil {
			r.log().Warn("skip entry with invalid torrent url", "title", entry.Entry.Title, "error", err)
			continue
		}
		r.log().Info("detected new torrent", "bangumi", folderName, "episode", episodeLabel, "entry", entry.Entry.Title, "torrent", entry.Entry.TorrentURL, "local_path", localPath)
		newEntries = append(newEntries, plannedEntry{ResolvedEntry: entry, EpisodeLabel: episodeLabel, LocalPath: localPath})
	}
	if len(newEntries) == 0 {
		r.log().Info("RSS source has no new updates", "checked", len(entries))
		return nil
	}

	r.log().Info("new torrents detected, logging in to PikPak", "count", len(newEntries), "username", r.Config.Username)
	if err := r.PikPak.Login(); err != nil {
		r.log().Error("PikPak login failed", "username", r.Config.Username, "error", err)
		return fmt.Errorf("pikpak login: %w", err)
	}
	r.log().Info("PikPak login succeeded", "username", r.Config.Username)

	for _, entry := range newEntries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		folderName := sanitize.Name(entry.BangumiTitle)
		r.log().Info("processing new bangumi torrent", "bangumi", folderName, "episode", entry.EpisodeLabel, "entry", entry.Entry.Title, "torrent", entry.Entry.TorrentURL)

		bangumiFolderID, err := r.PikPak.EnsureFolder(r.Config.Path, folderName)
		if err != nil {
			r.log().Error("ensure pikpak bangumi folder failed", "bangumi", folderName, "error", err)
			continue
		}
		r.log().Info("PikPak bangumi folder ready", "bangumi", folderName, "folder_id", bangumiFolderID)

		episodeFolderID, err := r.PikPak.EnsureFolder(bangumiFolderID, entry.EpisodeLabel)
		if err != nil {
			r.log().Error("ensure pikpak episode folder failed", "bangumi", folderName, "episode", entry.EpisodeLabel, "error", err)
			continue
		}
		r.log().Info("PikPak episode folder ready", "bangumi", folderName, "episode", entry.EpisodeLabel, "folder_id", episodeFolderID)

		hasChildren, err := r.PikPak.HasChildren(episodeFolderID)
		if err != nil {
			r.log().Warn("remote episode folder check failed", "bangumi", folderName, "episode", entry.EpisodeLabel, "error", err)
		}
		if hasChildren {
			r.log().Info("skip episode because remote folder already has files", "bangumi", folderName, "episode", entry.EpisodeLabel)
			if err := torrent.MarkDownloaded(r.torrentRoot(), folderName, entry.EpisodeLabel, entry.Entry.TorrentURL); err != nil {
				r.log().Warn("write local episode marker failed", "bangumi", folderName, "episode", entry.EpisodeLabel, "error", err)
			}
			continue
		}

		if err := torrent.Download(r.httpClient(), entry.Entry.TorrentURL, entry.LocalPath); err != nil {
			r.log().Error("download torrent failed", "bangumi", folderName, "episode", entry.EpisodeLabel, "url", entry.Entry.TorrentURL, "error", err)
			continue
		}
		r.log().Info("torrent downloaded", "bangumi", folderName, "episode", entry.EpisodeLabel, "local_path", entry.LocalPath)

		duplicate, err := r.PikPak.HasOriginalURL(episodeFolderID, entry.Entry.TorrentURL)
		if err != nil {
			r.log().Warn("remote duplicate check failed", "bangumi", folderName, "episode", entry.EpisodeLabel, "error", err)
		}
		if duplicate {
			r.log().Info("skip duplicate remote torrent", "bangumi", folderName, "episode", entry.EpisodeLabel, "torrent", entry.Entry.TorrentURL)
			if err := torrent.MarkDownloaded(r.torrentRoot(), folderName, entry.EpisodeLabel, entry.Entry.TorrentURL); err != nil {
				r.log().Warn("write local episode marker failed", "bangumi", folderName, "episode", entry.EpisodeLabel, "error", err)
			}
			continue
		}

		name := filepath.Base(entry.LocalPath)
		if _, err := r.PikPak.OfflineDownload(name, entry.Entry.TorrentURL, episodeFolderID); err != nil {
			r.log().Error("offline download failed", "bangumi", folderName, "episode", entry.EpisodeLabel, "error", err)
			continue
		}
		if err := torrent.MarkDownloaded(r.torrentRoot(), folderName, entry.EpisodeLabel, entry.Entry.TorrentURL); err != nil {
			r.log().Warn("write local episode marker failed", "bangumi", folderName, "episode", entry.EpisodeLabel, "error", err)
		}
		r.log().Info("submitted offline download", "bangumi", folderName, "episode", entry.EpisodeLabel, "torrent", entry.Entry.TorrentURL)
	}
	return nil
}

func (r Runner) entries(ctx context.Context) ([]ResolvedEntry, error) {
	if r.EntriesFunc != nil {
		return r.EntriesFunc(ctx)
	}
	r.log().Info("fetching RSS feed", "rss", r.Config.RSS)
	entries, err := rss.Fetch(r.httpClient(), r.Config.RSS)
	if err != nil {
		return nil, err
	}
	r.log().Info("RSS feed parsed", "count", len(entries))

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
		r.log().Info("recognized bangumi", "entry", entry.Title, "bangumi", title, "torrent", entry.TorrentURL)
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
