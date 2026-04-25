package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"bangumi-pikpak/internal/sanitize"
)

type Drive115Provider struct {
	cookie     string
	rootCID    string
	httpClient *http.Client
	baseWebAPI string
	baseLixian string
}

func NewDrive115Provider(cookie, rootCID string, hc *http.Client) *Drive115Provider {
	if hc == nil {
		hc = http.DefaultClient
	}
	if strings.TrimSpace(rootCID) == "" {
		rootCID = "0"
	}
	return &Drive115Provider{cookie: cookie, rootCID: rootCID, httpClient: hc, baseWebAPI: "https://webapi.115.com", baseLixian: "https://115.com"}
}

func (p *Drive115Provider) Name() string { return "drive115" }
func (p *Drive115Provider) Login(ctx context.Context) error {
	if strings.TrimSpace(p.cookie) == "" {
		return fmt.Errorf("115 cookie is required")
	}
	if strings.TrimSpace(p.rootCID) == "" {
		p.rootCID = "0"
	}
	return ctx.Err()
}

func (p *Drive115Provider) EnsureBangumi(ctx context.Context, title string) (Folder, error) {
	id, err := p.ensureFolder(ctx, p.rootCID, sanitize.Name(title))
	if err != nil {
		return Folder{}, err
	}
	return Folder{ID: id, Name: sanitize.Name(title)}, nil
}

func (p *Drive115Provider) EnsureEpisode(ctx context.Context, bangumi Folder, episode string) (Folder, error) {
	id, err := p.ensureFolder(ctx, bangumi.ID, sanitize.Name(episode))
	if err != nil {
		return Folder{}, err
	}
	return Folder{ID: id, Name: sanitize.Name(episode)}, nil
}

func (p *Drive115Provider) HasOriginalURL(ctx context.Context, folder Folder, sourceURL string) (bool, error) {
	return false, ctx.Err()
}

func (p *Drive115Provider) HasChildren(ctx context.Context, folder Folder) (bool, error) {
	values := url.Values{}
	values.Set("cid", folder.ID)
	values.Set("limit", "1")
	values.Set("offset", "0")
	var resp struct {
		State bool              `json:"state"`
		Count int               `json:"count"`
		Data  []json.RawMessage `json:"data"`
		Error string            `json:"error"`
	}
	if err := p.getJSON(ctx, p.baseWebAPI+"/files?"+values.Encode(), &resp); err != nil {
		return false, err
	}
	if !resp.State && resp.Error != "" {
		return false, fmt.Errorf("115 list folder: %s", resp.Error)
	}
	return resp.Count > 0 || len(resp.Data) > 0, nil
}

func (p *Drive115Provider) SubmitDownload(ctx context.Context, task DownloadTask) (Task, error) {
	values := url.Values{}
	values.Set("url", task.SourceURL)
	values.Set("wp_path_id", task.Folder.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseLixian+"/web/lixian/?ct=lixian&ac=add_task_url", strings.NewReader(values.Encode()))
	if err != nil {
		return Task{}, err
	}
	p.setHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return Task{}, err
	}
	defer resp.Body.Close()
	var decoded struct {
		State    bool   `json:"state"`
		InfoHash string `json:"info_hash"`
		Error    string `json:"error"`
		ErrMsg   string `json:"err_msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Task{}, err
	}
	if !decoded.State {
		msg := firstNonEmpty(decoded.Error, decoded.ErrMsg, "115 offline task failed")
		return Task{}, fmt.Errorf("%s", msg)
	}
	return Task{ID: decoded.InfoHash, Name: task.Name, Status: "submitted"}, nil
}

func (p *Drive115Provider) DeleteBangumi(ctx context.Context, folder Folder) error {
	id := strings.TrimSpace(folder.ID)
	if id == "" {
		return fmt.Errorf("115 folder id is empty")
	}
	values := url.Values{}
	values.Set("fid[0]", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseWebAPI+"/rb/delete", strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	p.setHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("115 delete status %d", resp.StatusCode)
	}
	var decoded struct {
		State  bool   `json:"state"`
		Error  string `json:"error"`
		ErrMsg string `json:"err_msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}
	if !decoded.State {
		return fmt.Errorf("115 delete %q: %s", folder.Name, firstNonEmpty(decoded.Error, decoded.ErrMsg, "failed"))
	}
	return nil
}

func (p *Drive115Provider) ensureFolder(ctx context.Context, parentID, name string) (string, error) {
	if id, ok, err := p.findFolder(ctx, parentID, name); err != nil {
		return "", err
	} else if ok {
		return id, nil
	}
	values := url.Values{}
	values.Set("pid", parentID)
	values.Set("cname", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseWebAPI+"/files/add", strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	p.setHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var decoded struct {
		State  bool   `json:"state"`
		CID    string `json:"cid"`
		FileID string `json:"file_id"`
		Data   struct {
			CID    string `json:"cid"`
			FileID string `json:"file_id"`
		} `json:"data"`
		Error  string `json:"error"`
		ErrMsg string `json:"err_msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	if !decoded.State {
		return "", fmt.Errorf("115 create folder %q: %s", name, firstNonEmpty(decoded.Error, decoded.ErrMsg, "failed"))
	}
	id := firstNonEmpty(decoded.CID, decoded.FileID, decoded.Data.CID, decoded.Data.FileID)
	if id == "" {
		return "", fmt.Errorf("115 create folder %q returned empty cid", name)
	}
	return id, nil
}

func (p *Drive115Provider) findFolder(ctx context.Context, parentID, name string) (string, bool, error) {
	values := url.Values{}
	values.Set("cid", parentID)
	values.Set("search_value", name)
	values.Set("type", "0")
	values.Set("limit", "32")
	var resp struct {
		State bool `json:"state"`
		Data  []struct {
			ID       string `json:"id"`
			CID      string `json:"cid"`
			FileID   string `json:"file_id"`
			Name     string `json:"n"`
			FileName string `json:"file_name"`
			Category string `json:"category"`
			IsFolder string `json:"is_folder"`
		} `json:"data"`
		Error string `json:"error"`
	}
	if err := p.getJSON(ctx, p.baseWebAPI+"/files/search?"+values.Encode(), &resp); err != nil {
		return "", false, err
	}
	if !resp.State && resp.Error != "" {
		return "", false, fmt.Errorf("115 search folder: %s", resp.Error)
	}
	for _, item := range resp.Data {
		itemName := firstNonEmpty(item.Name, item.FileName)
		if itemName == name {
			return firstNonEmpty(item.ID, item.CID, item.FileID), true, nil
		}
	}
	return "", false, nil
}

func (p *Drive115Provider) getJSON(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	p.setHeaders(req)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("115 api status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (p *Drive115Provider) setHeaders(req *http.Request) {
	req.Header.Set("Cookie", p.cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 AnimeX/1.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
