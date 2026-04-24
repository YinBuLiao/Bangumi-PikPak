package web

import (
	"context"
	"net/url"
	"os"
	"sort"
	"strings"

	"bangumi-pikpak/internal/pikpak"
	"bangumi-pikpak/internal/torrent"
)

type DriveClient interface {
	Login() error
	List(parentID string) ([]pikpak.RemoteFile, error)
	DownloadURL(id string) (string, error)
}

type LibraryResponse struct {
	Bangumi []Bangumi `json:"bangumi"`
}

type Bangumi struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	CoverURL string    `json:"cover_url,omitempty"`
	Summary  string    `json:"summary,omitempty"`
	Episodes []Episode `json:"episodes"`
}

type Episode struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Files []PlayableFile `json:"files"`
}

type PlayableFile struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	StreamURL    string `json:"stream_url"`
}

func BuildLibrary(ctx context.Context, drive DriveClient, rootID, torrentRoot string) (LibraryResponse, error) {
	rootFiles, err := drive.List(rootID)
	if err != nil {
		return LibraryResponse{}, err
	}
	lib := LibraryResponse{Bangumi: make([]Bangumi, 0)}
	for _, folder := range rootFiles {
		select {
		case <-ctx.Done():
			return LibraryResponse{}, ctx.Err()
		default:
		}
		if folder.Kind != pikpak.KindFolder {
			continue
		}
		b := Bangumi{ID: folder.ID, Title: folder.Name}
		if meta, err := torrent.LoadBangumiMetadata(torrentRoot, folder.Name); err == nil {
			if meta.Title != "" {
				b.Title = meta.Title
			}
			b.CoverURL = meta.CoverURL
			b.Summary = meta.Summary
		} else if !os.IsNotExist(err) {
			// metadata is optional; malformed metadata should not hide the drive library
		}
		children, err := drive.List(folder.ID)
		if err != nil {
			return LibraryResponse{}, err
		}
		direct := Episode{ID: folder.ID + ":direct", Label: "文件"}
		for _, child := range children {
			if child.Kind != pikpak.KindFolder {
				if isPlayable(child) {
					direct.Files = append(direct.Files, playableFromRemote(child))
					if b.CoverURL == "" && child.ThumbnailLink != "" {
						b.CoverURL = child.ThumbnailLink
					}
				}
				continue
			}
			ep := Episode{ID: child.ID, Label: child.Name}
			files, err := drive.List(child.ID)
			if err != nil {
				return LibraryResponse{}, err
			}
			for _, file := range files {
				if file.Kind == pikpak.KindFolder || !isPlayable(file) {
					continue
				}
				ep.Files = append(ep.Files, playableFromRemote(file))
				if b.CoverURL == "" && file.ThumbnailLink != "" {
					b.CoverURL = file.ThumbnailLink
				}
			}
			// 媒体库要展示 PikPak 上的所有番剧目录；即使某集还在离线下载中、暂时没有可播文件，也保留集目录。
			b.Episodes = append(b.Episodes, ep)
		}
		if len(direct.Files) > 0 {
			b.Episodes = append([]Episode{direct}, b.Episodes...)
		}
		sort.SliceStable(b.Episodes, func(i, j int) bool { return b.Episodes[i].Label < b.Episodes[j].Label })
		lib.Bangumi = append(lib.Bangumi, b)
	}
	sort.SliceStable(lib.Bangumi, func(i, j int) bool { return lib.Bangumi[i].Title < lib.Bangumi[j].Title })
	return lib, nil
}

func playableFromRemote(file pikpak.RemoteFile) PlayableFile {
	return PlayableFile{
		ID:           file.ID,
		Name:         file.Name,
		Size:         file.Size,
		MimeType:     file.MimeType,
		ThumbnailURL: file.ThumbnailLink,
		StreamURL:    "/api/stream?id=" + url.QueryEscape(file.ID),
	}
}

func isPlayable(file pikpak.RemoteFile) bool {
	mime := strings.ToLower(file.MimeType)
	category := strings.ToLower(file.FileCategory)
	ext := strings.ToLower(strings.TrimPrefix(file.FileExtension, "."))
	if strings.HasPrefix(mime, "video/") || category == "video" {
		return true
	}
	switch ext {
	case "mp4", "mkv", "webm", "m4v", "mov", "avi":
		return true
	default:
		return false
	}
}
