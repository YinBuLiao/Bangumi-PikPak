package pikpak

import (
	"encoding/json"
	"testing"

	pikpakgo "github.com/kanghengliu/pikpak-go"
)

type fakeAPI struct {
	files         []RemoteFile
	createdName   string
	createdParent string
	offlineName   string
	offlineURL    string
	offlineParent string
}

func (f *fakeAPI) Login() error                                      { return nil }
func (f *fakeAPI) FileListAll(parentID string) ([]RemoteFile, error) { return f.files, nil }
func (f *fakeAPI) CreateFolder(name, parentID string) (RemoteFile, error) {
	f.createdName = name
	f.createdParent = parentID
	return RemoteFile{ID: "created-folder", Name: name, Kind: KindFolder}, nil
}
func (f *fakeAPI) OfflineDownload(name, fileURL, parentID string) (RemoteTask, error) {
	f.offlineName = name
	f.offlineURL = fileURL
	f.offlineParent = parentID
	return RemoteTask{ID: "task-id", Name: name}, nil
}

func TestEnsureFolderReturnsExisting(t *testing.T) {
	api := &fakeAPI{files: []RemoteFile{{ID: "existing", Name: "Bangumi", Kind: KindFolder}}}
	client := NewAdapter(api)
	id, err := client.EnsureFolder("parent", "Bangumi")
	if err != nil {
		t.Fatalf("EnsureFolder returned error: %v", err)
	}
	if id != "existing" || api.createdName != "" {
		t.Fatalf("expected existing folder, id=%q created=%q", id, api.createdName)
	}
}

func TestEnsureFolderCreatesMissing(t *testing.T) {
	api := &fakeAPI{}
	client := NewAdapter(api)
	id, err := client.EnsureFolder("parent", "Bangumi")
	if err != nil {
		t.Fatalf("EnsureFolder returned error: %v", err)
	}
	if id != "created-folder" || api.createdName != "Bangumi" || api.createdParent != "parent" {
		t.Fatalf("create mismatch: id=%q api=%#v", id, api)
	}
}

func TestHasOriginalURL(t *testing.T) {
	api := &fakeAPI{files: []RemoteFile{{ID: "file", Name: "x", OriginalURL: "magnet:?xt=urn:btih:abc"}}}
	client := NewAdapter(api)
	found, err := client.HasOriginalURL("parent", "magnet:?xt=urn:btih:abc")
	if err != nil {
		t.Fatalf("HasOriginalURL returned error: %v", err)
	}
	if !found {
		t.Fatal("expected original URL to be found")
	}
}

func TestOfflineDownloadDelegates(t *testing.T) {
	api := &fakeAPI{}
	client := NewAdapter(api)
	task, err := client.OfflineDownload("a.torrent", "https://example.test/a.torrent", "parent")
	if err != nil {
		t.Fatalf("OfflineDownload returned error: %v", err)
	}
	if task.ID != "task-id" || api.offlineName != "a.torrent" || api.offlineURL != "https://example.test/a.torrent" || api.offlineParent != "parent" {
		t.Fatalf("delegate mismatch: task=%#v api=%#v", task, api)
	}
}

func TestRequestLoginJSONIncludesCaptchaTokenForSignin(t *testing.T) {
	body, err := json.Marshal(pikpakgo.RequestLogin{ClientId: "client", ClientSecret: "secret", Username: "user", Password: "pass", CaptchaToken: "captcha"})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if decoded["client_id"] != "client" || decoded["client_secret"] != "secret" || decoded["captcha_token"] != "captcha" {
		t.Fatalf("signin JSON fields mismatch: %s", string(body))
	}
	if _, ok := decoded["grant_type"]; ok {
		t.Fatalf("signin request should not include grant_type: %s", string(body))
	}
}

func TestRequestNewTaskWithParentOmitsFolderTypeDownload(t *testing.T) {
	body, err := json.Marshal(pikpakgo.RequestNewTask{ParentID: "folder-id", FolderType: ""})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if decoded["parent_id"] != "folder-id" {
		t.Fatalf("expected parent_id, got %s", string(body))
	}
	if _, ok := decoded["folder_type"]; ok {
		t.Fatalf("folder_type must be omitted when parent_id is set, got %s", string(body))
	}
}
