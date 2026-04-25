package web

import (
	"context"
	"log/slog"

	"bangumi-pikpak/internal/bangumi"
	"bangumi-pikpak/internal/torrent"
)

type CoverResolver interface {
	SearchMetadata(title string) (bangumi.Metadata, error)
}

func EnrichLibraryCovers(ctx context.Context, lib *LibraryResponse, resolver CoverResolver, torrentRoot string) error {
	return enrichLibraryCovers(ctx, lib, resolver, torrentRoot, slog.Default())
}

func enrichLibraryCovers(ctx context.Context, lib *LibraryResponse, resolver CoverResolver, torrentRoot string, log *slog.Logger) error {
	if resolver == nil || lib == nil {
		return nil
	}
	for i := range lib.Bangumi {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if lib.Bangumi[i].CoverURL != "" && lib.Bangumi[i].Summary != "" {
			continue
		}
		queryTitle := ExtractBangumiTitle(lib.Bangumi[i].Title)
		if queryTitle != "" {
			lib.Bangumi[i].Title = queryTitle
		}
		meta, err := resolver.SearchMetadata(lib.Bangumi[i].Title)
		if err != nil {
			if log != nil {
				log.Warn("bangumi.tv metadata lookup failed", "bangumi", lib.Bangumi[i].Title, "error", err)
			}
			continue
		}
		if meta.CoverURL == "" && meta.Summary == "" {
			continue
		}
		if meta.CoverURL != "" {
			lib.Bangumi[i].CoverURL = meta.CoverURL
		}
		if meta.Summary != "" {
			lib.Bangumi[i].Summary = meta.Summary
		}
		if err := torrent.SaveBangumiMetadata(torrentRoot, lib.Bangumi[i].Title, torrent.BangumiMetadata{Title: lib.Bangumi[i].Title, CoverURL: lib.Bangumi[i].CoverURL, Summary: lib.Bangumi[i].Summary}); err != nil && log != nil {
			log.Warn("write bangumi metadata failed", "bangumi", lib.Bangumi[i].Title, "error", err)
		}
	}
	return nil
}
