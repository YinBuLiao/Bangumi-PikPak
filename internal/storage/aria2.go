package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bangumi-pikpak/internal/sanitize"
)

type Aria2Provider struct {
	name       string
	rootPath   string
	rpcURL     string
	rpcSecret  string
	httpClient *http.Client
}

func NewAria2LocalProvider(rootPath, rpcURL, rpcSecret string, hc *http.Client) *Aria2Provider {
	return newAria2Provider("local", rootPath, rpcURL, rpcSecret, hc)
}

func NewAria2NASProvider(rootPath, rpcURL, rpcSecret string, hc *http.Client) *Aria2Provider {
	return newAria2Provider("nas", rootPath, rpcURL, rpcSecret, hc)
}

func newAria2Provider(name, rootPath, rpcURL, rpcSecret string, hc *http.Client) *Aria2Provider {
	if strings.TrimSpace(rpcURL) == "" {
		rpcURL = "http://127.0.0.1:6800/jsonrpc"
	}
	if hc == nil {
		hc = http.DefaultClient
	}
	return &Aria2Provider{name: name, rootPath: rootPath, rpcURL: rpcURL, rpcSecret: rpcSecret, httpClient: hc}
}

func (p *Aria2Provider) Name() string { return p.name }
func (p *Aria2Provider) Login(ctx context.Context) error {
	if strings.TrimSpace(p.rootPath) == "" {
		return fmt.Errorf("%s storage path is empty", p.name)
	}
	return os.MkdirAll(p.rootPath, 0o755)
}

func (p *Aria2Provider) EnsureBangumi(ctx context.Context, title string) (Folder, error) {
	if err := ctx.Err(); err != nil {
		return Folder{}, err
	}
	name := sanitize.Name(title)
	dir := filepath.Join(p.rootPath, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Folder{}, err
	}
	return Folder{ID: dir, Name: name, Path: dir}, nil
}

func (p *Aria2Provider) EnsureEpisode(ctx context.Context, bangumi Folder, episode string) (Folder, error) {
	if err := ctx.Err(); err != nil {
		return Folder{}, err
	}
	name := sanitize.Name(episode)
	parent := bangumi.Path
	if parent == "" {
		parent = bangumi.ID
	}
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Folder{}, err
	}
	return Folder{ID: dir, Name: name, Path: dir}, nil
}

func (p *Aria2Provider) HasOriginalURL(ctx context.Context, folder Folder, sourceURL string) (bool, error) {
	return false, ctx.Err()
}

func (p *Aria2Provider) HasChildren(ctx context.Context, folder Folder) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	dir := folder.Path
	if dir == "" {
		dir = folder.ID
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		return true, nil
	}
	return false, nil
}

func (p *Aria2Provider) SubmitDownload(ctx context.Context, task DownloadTask) (Task, error) {
	dir := task.Folder.Path
	if dir == "" {
		dir = task.Folder.ID
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Task{}, err
	}
	options := map[string]string{
		"dir":              dir,
		"out":              task.Name,
		"bt-save-metadata": "true",
		"continue":         "true",
		"allow-overwrite":  "false",
	}
	if isMagnet(task.SourceURL) {
		var gid string
		if err := p.rpc(ctx, "aria2.addUri", []any{[]string{task.SourceURL}, options}, &gid); err != nil {
			return Task{}, err
		}
		return Task{ID: gid, Name: task.Name, Status: "submitted"}, nil
	}
	torrentBytes, err := p.fetchTorrent(ctx, task.SourceURL)
	if err != nil {
		return Task{}, err
	}
	encoded := base64.StdEncoding.EncodeToString(torrentBytes)
	var gid string
	if err := p.rpc(ctx, "aria2.addTorrent", []any{encoded, []string{}, options}, &gid); err != nil {
		return Task{}, err
	}
	return Task{ID: gid, Name: task.Name, Status: "submitted"}, nil
}

func (p *Aria2Provider) DeleteBangumi(ctx context.Context, folder Folder) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(p.rootPath) == "" {
		return fmt.Errorf("%s storage path is empty", p.name)
	}
	root, err := filepath.Abs(p.rootPath)
	if err != nil {
		return err
	}
	target := firstNonEmptyPath(folder.Path, folder.ID)
	if strings.TrimSpace(target) == "" {
		target = filepath.Join(root, sanitize.Name(folder.Name))
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." || filepath.IsAbs(rel) {
		return fmt.Errorf("refuse to delete path outside storage root: %s", target)
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.RemoveAll(target)
}

func (p *Aria2Provider) fetchTorrent(ctx context.Context, raw string) ([]byte, error) {
	if strings.HasPrefix(raw, "file://") {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(u.Path)
	}
	if _, err := os.Stat(raw); err == nil {
		return os.ReadFile(raw)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download torrent status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 64<<20))
}

type aria2Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type aria2Response struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (p *Aria2Provider) rpc(ctx context.Context, method string, params []any, result any) error {
	if strings.TrimSpace(p.rpcSecret) != "" {
		params = append([]any{"token:" + p.rpcSecret}, params...)
	}
	payload, err := json.Marshal(aria2Request{JSONRPC: "2.0", ID: fmt.Sprintf("animex-%d", time.Now().UnixNano()), Method: method, Params: params})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.rpcURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("aria2 rpc status %d", resp.StatusCode)
	}
	var decoded aria2Response
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}
	if decoded.Error != nil {
		return fmt.Errorf("aria2 rpc %s failed: %s", method, decoded.Error.Message)
	}
	if result != nil && len(decoded.Result) > 0 {
		return json.Unmarshal(decoded.Result, result)
	}
	return nil
}

func isMagnet(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && strings.EqualFold(u.Scheme, "magnet")
}

func firstNonEmptyPath(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
