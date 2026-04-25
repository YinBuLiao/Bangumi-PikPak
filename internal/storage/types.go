package storage

import (
	"context"
)

type Folder struct {
	ID   string
	Name string
	Path string
}

type DownloadTask struct {
	Name         string
	SourceURL    string
	BangumiTitle string
	EpisodeLabel string
	Folder       Folder
}

type Task struct {
	ID     string
	Name   string
	Status string
}

type Provider interface {
	Name() string
	Login(ctx context.Context) error
	EnsureBangumi(ctx context.Context, title string) (Folder, error)
	EnsureEpisode(ctx context.Context, bangumi Folder, episode string) (Folder, error)
	HasOriginalURL(ctx context.Context, folder Folder, sourceURL string) (bool, error)
	HasChildren(ctx context.Context, folder Folder) (bool, error)
	SubmitDownload(ctx context.Context, task DownloadTask) (Task, error)
}

type DeleteCapable interface {
	DeleteBangumi(ctx context.Context, folder Folder) error
}
