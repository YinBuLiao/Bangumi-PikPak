package storage

import (
	"context"
	"fmt"

	"bangumi-pikpak/internal/pikpak"
)

type PikPakClient interface {
	Login() error
	EnsureFolder(parentID, name string) (string, error)
	HasOriginalURL(parentID, targetURL string) (bool, error)
	HasChildren(parentID string) (bool, error)
	OfflineDownload(name, fileURL, parentID string) (pikpak.RemoteTask, error)
	DeleteFile(id string) error
}

type PikPakProvider struct {
	client PikPakClient
	rootID string
}

func NewPikPakProvider(client PikPakClient, rootID string) *PikPakProvider {
	return &PikPakProvider{client: client, rootID: rootID}
}

func (p *PikPakProvider) Name() string { return "pikpak" }

func (p *PikPakProvider) Login(ctx context.Context) error {
	if p == nil || p.client == nil {
		return fmt.Errorf("pikpak client is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return p.client.Login()
}

func (p *PikPakProvider) EnsureBangumi(ctx context.Context, title string) (Folder, error) {
	if err := ctx.Err(); err != nil {
		return Folder{}, err
	}
	id, err := p.client.EnsureFolder(p.rootID, title)
	if err != nil {
		return Folder{}, err
	}
	return Folder{ID: id, Name: title}, nil
}

func (p *PikPakProvider) EnsureEpisode(ctx context.Context, bangumi Folder, episode string) (Folder, error) {
	if err := ctx.Err(); err != nil {
		return Folder{}, err
	}
	id, err := p.client.EnsureFolder(bangumi.ID, episode)
	if err != nil {
		return Folder{}, err
	}
	return Folder{ID: id, Name: episode}, nil
}

func (p *PikPakProvider) HasOriginalURL(ctx context.Context, folder Folder, sourceURL string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return p.client.HasOriginalURL(folder.ID, sourceURL)
}

func (p *PikPakProvider) HasChildren(ctx context.Context, folder Folder) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return p.client.HasChildren(folder.ID)
}

func (p *PikPakProvider) SubmitDownload(ctx context.Context, task DownloadTask) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	remote, err := p.client.OfflineDownload(task.Name, task.SourceURL, task.Folder.ID)
	if err != nil {
		return Task{}, err
	}
	return Task{ID: remote.ID, Name: task.Name, Status: "submitted"}, nil
}

func (p *PikPakProvider) DeleteBangumi(ctx context.Context, folder Folder) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p == nil || p.client == nil {
		return fmt.Errorf("pikpak client is nil")
	}
	if folder.ID == "" {
		return fmt.Errorf("pikpak folder id is empty")
	}
	return p.client.DeleteFile(folder.ID)
}
