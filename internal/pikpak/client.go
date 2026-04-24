package pikpak

import (
	"fmt"

	pikpakgo "github.com/kanghengliu/pikpak-go"
)

const KindFolder = "drive#folder"

type RemoteFile struct {
	ID          string
	Name        string
	Kind        string
	OriginalURL string
	ParamURL    string
}

type RemoteTask struct {
	ID   string
	Name string
}

type API interface {
	Login() error
	FileListAll(parentID string) ([]RemoteFile, error)
	CreateFolder(name, parentID string) (RemoteFile, error)
	OfflineDownload(name, fileURL, parentID string) (RemoteTask, error)
}

type Adapter struct {
	api API
}

func NewAdapter(api API) *Adapter {
	return &Adapter{api: api}
}

func (a *Adapter) Login() error {
	return a.api.Login()
}

func (a *Adapter) EnsureFolder(parentID, name string) (string, error) {
	files, err := a.api.FileListAll(parentID)
	if err != nil {
		return "", fmt.Errorf("list pikpak folder: %w", err)
	}
	for _, file := range files {
		if file.Name == name && file.Kind == KindFolder {
			return file.ID, nil
		}
	}
	created, err := a.api.CreateFolder(name, parentID)
	if err != nil {
		return "", fmt.Errorf("create pikpak folder %q: %w", name, err)
	}
	return created.ID, nil
}

func (a *Adapter) HasOriginalURL(parentID, targetURL string) (bool, error) {
	files, err := a.api.FileListAll(parentID)
	if err != nil {
		return false, fmt.Errorf("list pikpak folder: %w", err)
	}
	for _, file := range files {
		if file.OriginalURL == targetURL || file.ParamURL == targetURL {
			return true, nil
		}
	}
	return false, nil
}

func (a *Adapter) OfflineDownload(name, fileURL, parentID string) (RemoteTask, error) {
	task, err := a.api.OfflineDownload(name, fileURL, parentID)
	if err != nil {
		return RemoteTask{}, fmt.Errorf("create pikpak offline task: %w", err)
	}
	return task, nil
}

type GoAPI struct {
	client *pikpakgo.PikPakClient
}

func NewGoAPI(username, password string) (*GoAPI, error) {
	client, err := pikpakgo.NewPikPakClient(username, password)
	if err != nil {
		return nil, err
	}
	return &GoAPI{client: client}, nil
}

func (g *GoAPI) Login() error {
	return g.client.Login()
}

func (g *GoAPI) FileListAll(parentID string) ([]RemoteFile, error) {
	files, err := g.client.FileListAll(parentID)
	if err != nil {
		return nil, err
	}
	out := make([]RemoteFile, 0, len(files))
	for _, file := range files {
		remote := RemoteFile{ID: file.ID, Name: file.Name, Kind: file.Kind, OriginalURL: file.OriginalURL}
		if file.Params != nil {
			remote.ParamURL = file.Params.URL
		}
		out = append(out, remote)
	}
	return out, nil
}

func (g *GoAPI) CreateFolder(name, parentID string) (RemoteFile, error) {
	file, err := g.client.CreateFolder(name, parentID)
	if err != nil {
		return RemoteFile{}, err
	}
	return RemoteFile{ID: file.ID, Name: file.Name, Kind: file.Kind, OriginalURL: file.OriginalURL}, nil
}

func (g *GoAPI) OfflineDownload(name, fileURL, parentID string) (RemoteTask, error) {
	task, err := g.client.OfflineDownload(name, fileURL, parentID)
	if err != nil {
		return RemoteTask{}, err
	}
	if task.Task == nil {
		return RemoteTask{}, nil
	}
	return RemoteTask{ID: task.Task.ID, Name: task.Task.Name}, nil
}
