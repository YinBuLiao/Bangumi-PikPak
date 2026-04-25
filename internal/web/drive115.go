package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bangumi-pikpak/internal/pikpak"
)

type drive115DriveClient struct {
	cookie string
	hc     *http.Client
}

func newDrive115DriveClient(cookie string, hc *http.Client) drive115DriveClient {
	if hc == nil {
		hc = http.DefaultClient
	}
	return drive115DriveClient{cookie: cookie, hc: hc}
}

func (d drive115DriveClient) Login() error {
	if strings.TrimSpace(d.cookie) == "" {
		return fmt.Errorf("115 cookie is required")
	}
	return nil
}

func (d drive115DriveClient) List(parentID string) ([]pikpak.RemoteFile, error) {
	values := url.Values{}
	values.Set("cid", parentID)
	values.Set("limit", "200")
	values.Set("offset", "0")
	endpoint := "https://webapi.115.com/files?" + values.Encode()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	d.setHeaders(req)
	resp, err := d.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("115 list status %d", resp.StatusCode)
	}
	var decoded struct {
		State bool `json:"state"`
		Data  []struct {
			ID       string          `json:"id"`
			CID      string          `json:"cid"`
			FID      string          `json:"fid"`
			Name     string          `json:"n"`
			FileName string          `json:"file_name"`
			Size     json.RawMessage `json:"s"`
			PickCode string          `json:"pick_code"`
			Category string          `json:"category"`
			IsDir    json.RawMessage `json:"is_dir"`
		} `json:"data"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if !decoded.State && decoded.Error != "" {
		return nil, fmt.Errorf("115 list: %s", decoded.Error)
	}
	files := make([]pikpak.RemoteFile, 0, len(decoded.Data))
	for _, item := range decoded.Data {
		name := firstNonEmpty(item.Name, item.FileName)
		id := firstNonEmpty(item.PickCode, item.FID, item.ID, item.CID)
		kind := ""
		if isDrive115Folder(item.IsDir, item.Category) {
			kind = pikpak.KindFolder
			id = firstNonEmpty(item.CID, item.ID, item.FID)
		}
		files = append(files, pikpak.RemoteFile{ID: id, Name: name, Kind: kind, Size: parseDrive115Size(item.Size), FileExtension: fileExt(name), FileCategory: categoryFromName(name)})
	}
	return files, nil
}

func (d drive115DriveClient) DownloadURL(id string) (string, error) {
	values := url.Values{}
	values.Set("pickcode", id)
	req, err := http.NewRequest(http.MethodGet, "https://webapi.115.com/files/download?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}
	d.setHeaders(req)
	resp, err := d.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("115 download status %d", resp.StatusCode)
	}
	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	if u, _ := decoded["file_url"].(string); u != "" {
		return u, nil
	}
	if u, _ := decoded["url"].(string); u != "" {
		return u, nil
	}
	if nested, ok := decoded["url"].(map[string]any); ok {
		if u, _ := nested["url"].(string); u != "" {
			return u, nil
		}
	}
	return "", fmt.Errorf("115 download url is empty")
}

func (d drive115DriveClient) setHeaders(req *http.Request) {
	req.Header.Set("Cookie", d.cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 AnimeX/1.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
}

func isDrive115Folder(raw json.RawMessage, category string) bool {
	if strings.EqualFold(category, "0") || strings.EqualFold(category, "folder") {
		return true
	}
	var b bool
	if json.Unmarshal(raw, &b) == nil {
		return b
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s == "1" || strings.EqualFold(s, "true")
	}
	var n int
	if json.Unmarshal(raw, &n) == nil {
		return n == 1
	}
	return false
}

func parseDrive115Size(raw json.RawMessage) int64 {
	var n int64
	if json.Unmarshal(raw, &n) == nil {
		return n
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		v, _ := strconv.ParseInt(s, 10, 64)
		return v
	}
	return 0
}

func fileExt(name string) string {
	idx := strings.LastIndex(name, ".")
	if idx < 0 {
		return ""
	}
	return strings.ToLower(name[idx+1:])
}

func categoryFromName(name string) string {
	switch fileExt(name) {
	case "mp4", "mkv", "avi", "mov", "flv", "webm", "m4v":
		return "video"
	default:
		return ""
	}
}
