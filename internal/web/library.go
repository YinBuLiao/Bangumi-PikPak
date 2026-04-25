package web

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"bangumi-pikpak/internal/episode"
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
		title := ExtractBangumiTitle(folder.Name)
		b := Bangumi{ID: folder.ID, Title: title}
		if meta, err := loadLibraryMetadata(torrentRoot, folder.Name, title); err == nil {
			if meta.Title != "" {
				b.Title = ExtractBangumiTitle(meta.Title)
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
		direct := Episode{ID: folder.ID + ":direct", Label: "文件", Files: []PlayableFile{}}
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
			files, cover, err := collectRemotePlayableFiles(ctx, drive, child.ID)
			if err != nil {
				return LibraryResponse{}, err
			}
			if b.CoverURL == "" && cover != "" {
				b.CoverURL = cover
			}
			// 媒体库要展示 PikPak 上的所有番剧目录；即使某集还在离线下载中、暂时没有可播文件，也保留集目录。
			b.Episodes = append(b.Episodes, expandEpisodes(child.ID, child.Name, files)...)
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

func collectRemotePlayableFiles(ctx context.Context, drive DriveClient, parentID string) ([]PlayableFile, string, error) {
	children, err := drive.List(parentID)
	if err != nil {
		return nil, "", err
	}
	files := make([]PlayableFile, 0)
	cover := ""
	for _, child := range children {
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		default:
		}
		if child.Kind == pikpak.KindFolder {
			nested, nestedCover, err := collectRemotePlayableFiles(ctx, drive, child.ID)
			if err != nil {
				return nil, "", err
			}
			files = append(files, nested...)
			if cover == "" {
				cover = nestedCover
			}
			continue
		}
		if !isPlayable(child) {
			continue
		}
		files = append(files, playableFromRemote(child))
		if cover == "" && child.ThumbnailLink != "" {
			cover = child.ThumbnailLink
		}
	}
	return files, cover, nil
}

var fileEpisodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)s\d{1,2}e0*(\d{1,3})`),
	regexp.MustCompile(`(?i)(?:^|[\s._-])ep(?:isode)?\s*0*(\d{1,3})(?:[\s._-]|\(|\[|$)`),
	regexp.MustCompile(`(?:^|[\s._-])0*(\d{1,3})(?:\s*(?:END|Fin))?(?:[\s._-]|\(|\[|$)`),
}

var episodeNumberPattern = regexp.MustCompile(`\d+`)

func expandEpisodes(folderID, folderLabel string, files []PlayableFile) []Episode {
	folderLabel = strings.TrimSpace(folderLabel)
	if folderLabel == "" {
		folderLabel = "文件"
	}
	if len(files) == 0 {
		return []Episode{{ID: folderID, Label: folderLabel, Files: []PlayableFile{}}}
	}
	groups := make(map[string][]PlayableFile)
	order := make([]string, 0)
	for _, file := range files {
		label, ok := labelFromPlayableFile(file.Name)
		if !ok {
			label = folderLabel
		}
		if _, exists := groups[label]; !exists {
			order = append(order, label)
		}
		groups[label] = append(groups[label], file)
	}
	if len(order) == 1 {
		return []Episode{{ID: folderID, Label: order[0], Files: groups[order[0]]}}
	}
	episodes := make([]Episode, 0, len(order))
	for _, label := range order {
		episodes = append(episodes, Episode{ID: folderID + ":" + label, Label: label, Files: groups[label]})
	}
	sort.SliceStable(episodes, func(i, j int) bool { return episodeSortKey(episodes[i].Label) < episodeSortKey(episodes[j].Label) })
	return episodes
}

func labelFromPlayableFile(name string) (string, bool) {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	if label, ok := episode.LabelFromTitle(base); ok {
		if !strings.Contains(label, "-") && label != "合集" {
			return label, true
		}
	}
	for _, pattern := range fileEpisodePatterns {
		match := pattern.FindStringSubmatch(base)
		if len(match) < 2 {
			continue
		}
		n, err := strconv.Atoi(match[1])
		if err != nil || n <= 0 {
			continue
		}
		return fmt.Sprintf("第%02d集", n), true
	}
	return "", false
}

func episodeSortKey(label string) string {
	match := episodeNumberPattern.FindString(label)
	if match == "" {
		return "999999:" + label
	}
	n, _ := strconv.Atoi(match)
	return fmt.Sprintf("%06d:%s", n, label)
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
	if ext == "" {
		ext = strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Name), "."))
	}
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

func BuildLocalLibrary(ctx context.Context, rootPath, torrentRoot string) (LibraryResponse, error) {
	rootPath = strings.TrimSpace(rootPath)
	if rootPath == "" {
		return LibraryResponse{}, nil
	}
	rootPath, _ = filepath.Abs(rootPath)
	entries, err := os.ReadDir(rootPath)
	if os.IsNotExist(err) {
		return LibraryResponse{Bangumi: []Bangumi{}}, nil
	}
	if err != nil {
		return LibraryResponse{}, err
	}
	lib := LibraryResponse{Bangumi: make([]Bangumi, 0)}
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return LibraryResponse{}, ctx.Err()
		default:
		}
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		bangumiPath := filepath.Join(rootPath, entry.Name())
		title := ExtractBangumiTitle(entry.Name())
		b := Bangumi{ID: localStreamID(bangumiPath), Title: title}
		if meta, err := loadLibraryMetadata(torrentRoot, entry.Name(), title); err == nil {
			if meta.Title != "" {
				b.Title = ExtractBangumiTitle(meta.Title)
			}
			b.CoverURL = meta.CoverURL
			b.Summary = meta.Summary
		}
		children, err := os.ReadDir(bangumiPath)
		if err != nil {
			continue
		}
		direct := Episode{ID: b.ID + ":direct", Label: "文件", Files: []PlayableFile{}}
		for _, child := range children {
			childPath := filepath.Join(bangumiPath, child.Name())
			if child.IsDir() {
				b.Episodes = append(b.Episodes, expandEpisodes(localStreamID(childPath), child.Name(), collectLocalPlayableFiles(ctx, childPath))...)
				continue
			}
			if isPlayableLocal(child.Name()) {
				if pf, ok := playableFromLocal(childPath); ok {
					direct.Files = append(direct.Files, pf)
				}
			}
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

func loadLibraryMetadata(torrentRoot, rawTitle, extractedTitle string) (torrent.BangumiMetadata, error) {
	meta, err := torrent.LoadBangumiMetadata(torrentRoot, rawTitle)
	if err == nil {
		return meta, nil
	}
	if extractedTitle != "" && extractedTitle != rawTitle {
		if extractedMeta, extractedErr := torrent.LoadBangumiMetadata(torrentRoot, extractedTitle); extractedErr == nil {
			return extractedMeta, nil
		}
	}
	return torrent.BangumiMetadata{}, err
}

func collectLocalPlayableFiles(ctx context.Context, rootPath string) []PlayableFile {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return []PlayableFile{}
	}
	files := make([]PlayableFile, 0)
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return files
		default:
		}
		path := filepath.Join(rootPath, entry.Name())
		if entry.IsDir() {
			files = append(files, collectLocalPlayableFiles(ctx, path)...)
			continue
		}
		if !isPlayableLocal(entry.Name()) {
			continue
		}
		if pf, ok := playableFromLocal(path); ok {
			files = append(files, pf)
		}
	}
	return files
}

func isPlayableLocal(name string) bool {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(name), ".")) {
	case "mp4", "mkv", "avi", "mov", "flv", "webm", "m4v":
		return true
	default:
		return false
	}
}

func playableFromLocal(path string) (PlayableFile, bool) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return PlayableFile{}, false
	}
	id := localStreamID(path)
	return PlayableFile{
		ID:        id,
		Name:      info.Name(),
		Size:      info.Size(),
		MimeType:  "video/" + strings.TrimPrefix(strings.ToLower(filepath.Ext(info.Name())), "."),
		StreamURL: "/api/stream?id=" + url.QueryEscape(id),
	}, true
}

func localStreamID(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return "local:" + base64.RawURLEncoding.EncodeToString([]byte(abs))
}

func decodeLocalStreamID(id string) (string, error) {
	raw := strings.TrimPrefix(id, "local:")
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
