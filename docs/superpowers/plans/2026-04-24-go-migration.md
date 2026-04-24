# Go Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert Bangumi-PikPak from Python to Go while preserving existing behavior and adding a cleaner modular runtime.

**Architecture:** `cmd/bangumi-pikpak` owns CLI flags, startup, signal handling, and loop control. Focused `internal/*` packages own config, logging, proxy setup, RSS parsing, Mikan title extraction, torrent cache handling, PikPak integration, and one-cycle orchestration. PikPak calls go through a local adapter around `github.com/kanghengliu/pikpak-go v0.1.0`.

**Tech Stack:** Go 1.22+, `github.com/kanghengliu/pikpak-go`, `github.com/PuerkitoBio/goquery`, `github.com/mmcdole/gofeed`, `gopkg.in/natefinch/lumberjack.v2`, Go `testing`.

---

## File Map

- Create: `go.mod`, `go.sum`
- Create: `cmd/bangumi-pikpak/main.go`
- Create: `internal/config/config.go`, `internal/config/config_test.go`
- Create: `internal/logger/logger.go`
- Create: `internal/proxy/proxy.go`, `internal/proxy/proxy_test.go`
- Create: `internal/sanitize/sanitize.go`, `internal/sanitize/sanitize_test.go`
- Create: `internal/rss/rss.go`, `internal/rss/rss_test.go`
- Create: `internal/mikan/mikan.go`, `internal/mikan/mikan_test.go`
- Create: `internal/torrent/torrent.go`, `internal/torrent/torrent_test.go`
- Create: `internal/pikpak/client.go`, `internal/pikpak/state.go`, `internal/pikpak/client_test.go`
- Create: `internal/app/app.go`, `internal/app/app_test.go`
- Create: `Dockerfile`, `docs/examples/bangumi-pikpak.service`
- Modify: `README.md`

---

### Task 1: Module and Config

**Files:**
- Create: `go.mod`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Initialize module**

Run: `go mod init bangumi-pikpak`

Expected: `go.mod` contains `module bangumi-pikpak`.

- [ ] **Step 2: Write failing tests**

Create tests for:

```go
func TestLoadExistingConfig(t *testing.T) {
  // write JSON with username/password/path/rss/http_proxy/https_proxy/socks_proxy/enable_proxy
  // Load(path) must preserve all fields
}

func TestValidateRejectsMissingRequiredFields(t *testing.T) {
  // Config{Username:"u", Password:"p"}.Validate() must mention missing path and rss
}

func TestSaveRoundTrip(t *testing.T) {
  // Save(path, cfg), then Load(path), must preserve Username, Password, Path, RSS, proxy fields
}
```

Run: `go test ./internal/config`

Expected: FAIL because package is missing.

- [ ] **Step 3: Implement config**

Implement:

```go
type Config struct {
  Username string `json:"username"`
  Password string `json:"password"`
  Path string `json:"path"`
  RSS string `json:"rss"`
  HTTPProxy string `json:"http_proxy"`
  HTTPSProxy string `json:"https_proxy"`
  SocksProxy string `json:"socks_proxy"`
  EnableProxy bool `json:"enable_proxy"`
}
func Load(path string) (Config, error)
func Save(path string, cfg Config) error
func (c Config) Validate() error
```

Validation requires non-empty `username`, `password`, `path`, and `rss`.

- [ ] **Step 4: Verify and commit**

Run:

```powershell
go test ./internal/config
git add go.mod internal/config
git commit -m "feat: add go config loading"
```

Expected: tests pass and commit succeeds.

---

### Task 2: Runtime Utilities

**Files:**
- Create: `internal/logger/logger.go`
- Create: `internal/proxy/proxy.go`, `internal/proxy/proxy_test.go`
- Create: `internal/sanitize/sanitize.go`, `internal/sanitize/sanitize_test.go`

- [ ] **Step 1: Add dependency**

Run: `go get gopkg.in/natefinch/lumberjack.v2@latest`

- [ ] **Step 2: Write failing tests**

Sanitizer tests:

```go
Name(`a<b>c:d"e/f\g|h?i*j`) == "a_b_c_d_e_f_g_h_i_j"
Name("  Bangumi.  ") == "Bangumi"
Name(`<>:"/\|?* ...`) == "untitled"
```

Proxy tests:

```go
Apply(config.Config{EnableProxy:false}) == false
Apply(config.Config{EnableProxy:true, HTTPProxy:"http://127.0.0.1:7890"}) == true
os.Getenv("HTTP_PROXY") == "http://127.0.0.1:7890"
os.Getenv("http_proxy") == "http://127.0.0.1:7890"
```

Run: `go test ./internal/sanitize ./internal/proxy`

Expected: FAIL before implementation.

- [ ] **Step 3: Implement utilities**

Implement:

```go
package sanitize
func Name(s string) string
```

Rules: replace `< > : " / \ | ? *` and control characters with `_`, collapse repeated underscores, trim leading/trailing spaces, periods, and underscores, return `untitled` when empty.

Implement:

```go
package proxy
func Apply(cfg config.Config) bool
func HTTPClient() *http.Client
```

`Apply` sets `HTTP_PROXY`, `http_proxy`, `HTTPS_PROXY`, `https_proxy`, `SOCKS_PROXY`, `socks_proxy` only when enabled.

Implement:

```go
package logger
func New(logFile string) *slog.Logger
```

Use `io.MultiWriter(os.Stdout, &lumberjack.Logger{Filename: logFile, MaxSize: 10, MaxBackups: 5})`.

- [ ] **Step 4: Verify and commit**

Run:

```powershell
go test ./internal/sanitize ./internal/proxy
git add go.mod go.sum internal/logger internal/proxy internal/sanitize
git commit -m "feat: add runtime utilities"
```

---

### Task 3: RSS and Mikan Parsing

**Files:**
- Create: `internal/rss/rss.go`, `internal/rss/rss_test.go`
- Create: `internal/mikan/mikan.go`, `internal/mikan/mikan_test.go`

- [ ] **Step 1: Add dependencies**

Run: `go get github.com/mmcdole/gofeed@latest github.com/PuerkitoBio/goquery@latest`

- [ ] **Step 2: Write failing tests**

RSS fixture contains one `<item>` with `title`, `link`, `pubDate`, and torrent `enclosure`. Test `Parse(reader)` returns:

```go
Entry{Title:"[Group] Test 01", Link:"https://mikanani.me/Home/Episode/abc", TorrentURL:"https://mikanani.me/Download/test.torrent", PublishedDate:"2026-04-24"}
```

Mikan HTML fixture:

```html
<p class="bangumi-title">  进击的巨人 最终季  </p>
```

Test `ParseTitle(reader)` returns `进击的巨人 最终季`, and missing selector returns an error.

Run: `go test ./internal/rss ./internal/mikan`

Expected: FAIL before implementation.

- [ ] **Step 3: Implement parser packages**

Implement:

```go
package rss
type Entry struct { Title, Link, TorrentURL, PublishedDate string }
func Parse(r io.Reader) ([]Entry, error)
func Fetch(client *http.Client, url string) ([]Entry, error)
```

Use `gofeed.NewParser().Parse(r)`. Use first non-empty enclosure URL.

Implement:

```go
package mikan
func ParseTitle(r io.Reader) (string, error)
func FetchTitle(client *http.Client, url string) (string, error)
```

Use `goquery` selector `p.bangumi-title`.

- [ ] **Step 4: Verify and commit**

Run:

```powershell
go test ./internal/rss ./internal/mikan
git add go.mod go.sum internal/rss internal/mikan
git commit -m "feat: add feed and mikan parsers"
```

---

### Task 4: Torrent Cache

**Files:**
- Create: `internal/torrent/torrent.go`, `internal/torrent/torrent_test.go`

- [ ] **Step 1: Write failing tests**

Test:

```go
LocalPath("torrent", "进击/巨人", "https://example.test/abc123.torrent")
// returns filepath.Join("torrent", "进击_巨人", "abc123.torrent")
Exists(path) reflects file presence
Download(client, server.URL+"/a.torrent", target) writes response body and creates parent directories
```

Run: `go test ./internal/torrent`

Expected: FAIL before implementation.

- [ ] **Step 2: Implement torrent package**

Implement:

```go
package torrent
func LocalPath(root, bangumiTitle, torrentURL string) (string, error)
func Exists(path string) bool
func Download(client *http.Client, url, target string) error
```

`LocalPath` parses URL path basename and sanitizes both folder and file name.

- [ ] **Step 3: Verify and commit**

Run:

```powershell
go test ./internal/torrent
git add internal/torrent
git commit -m "feat: add torrent cache handling"
```

---

### Task 5: PikPak Adapter

**Files:**
- Create: `internal/pikpak/client.go`, `internal/pikpak/state.go`, `internal/pikpak/client_test.go`

- [ ] **Step 1: Add dependency**

Run: `go get github.com/kanghengliu/pikpak-go@v0.1.0`

Known API signatures from `v0.1.0`:

```go
func NewPikPakClient(username, password string) (*PikPakClient, error)
func (c *PikPakClient) Login() error
func (c *PikPakClient) FileListAll(fileId string) ([]*File, error)
func (c *PikPakClient) CreateFolder(name string, parentId string) (*File, error)
func (c *PikPakClient) OfflineDownload(name string, fileUrl string, parentId string) (*NewTask, error)
```

- [ ] **Step 2: Write failing adapter tests**

Use a fake API to verify:

```go
EnsureFolder(parentID, name) returns existing folder ID when FileListAll has RemoteFile{Name:name, Kind:"drive#folder"}
EnsureFolder(parentID, name) calls CreateFolder when missing
HasOriginalURL(parentID, targetURL) checks RemoteFile.OriginalURL and RemoteFile.ParamURL
OfflineDownload(name, url, parentID) delegates to fake API
```

Run: `go test ./internal/pikpak`

Expected: FAIL before implementation.

- [ ] **Step 3: Implement adapter and state**

Implement:

```go
const KindFolder = "drive#folder"
type RemoteFile struct { ID, Name, Kind, OriginalURL, ParamURL string }
type RemoteTask struct { ID, Name string }
type API interface {
  Login() error
  FileListAll(parentID string) ([]RemoteFile, error)
  CreateFolder(name, parentID string) (RemoteFile, error)
  OfflineDownload(name, fileURL, parentID string) (RemoteTask, error)
}
type Adapter struct { api API }
func NewAdapter(api API) *Adapter
func (a *Adapter) Login() error
func (a *Adapter) EnsureFolder(parentID, name string) (string, error)
func (a *Adapter) HasOriginalURL(parentID, targetURL string) (bool, error)
func (a *Adapter) OfflineDownload(name, fileURL, parentID string) (RemoteTask, error)
```

Add `GoAPI` wrapper that converts `pikpak-go` `File` and `NewTask` values into project types.

Implement state:

```go
type State struct {
  Username string `json:"username"`
  LastLoginTime time.Time `json:"last_login_time"`
  LastRefreshTime time.Time `json:"last_refresh_time"`
  Client string `json:"client"`
}
func LoadState(path string) (State, error)
func SaveState(path string, state State) error
```

- [ ] **Step 4: Verify and commit**

Run:

```powershell
go test ./internal/pikpak
git add go.mod go.sum internal/pikpak
git commit -m "feat: add pikpak adapter"
```

---

### Task 6: App Orchestration

**Files:**
- Create: `internal/app/app.go`, `internal/app/app_test.go`

- [ ] **Step 1: Write failing tests**

Use fake PikPak client and injected entry resolver:

```go
TestRunOnceNoNewTorrentSkipsLogin
// pre-create local torrent file; RunOnce must not call Login or OfflineDownload

TestRunOnceNewTorrentSubmitsOfflineTask
// no local file; RunOnce must call Login, EnsureFolder, Download, HasOriginalURL, OfflineDownload

TestRunOnceDuplicateRemoteSkipsOfflineDownload
// fake HasOriginalURL returns true; RunOnce downloads local torrent but skips OfflineDownload
```

Run: `go test ./internal/app`

Expected: FAIL before implementation.

- [ ] **Step 2: Implement runner**

Implement:

```go
type ResolvedEntry struct { Entry rss.Entry; BangumiTitle string }
type PikPakClient interface {
  Login() error
  EnsureFolder(parentID, name string) (string, error)
  HasOriginalURL(parentID, targetURL string) (bool, error)
  OfflineDownload(name, fileURL, parentID string) (pikpak.RemoteTask, error)
}
type Runner struct {
  Config config.Config
  HTTPClient *http.Client
  Logger *slog.Logger
  TorrentRoot string
  PikPak PikPakClient
  EntriesFunc func(context.Context) ([]ResolvedEntry, error)
}
func (r Runner) RunOnce(ctx context.Context) error
```

`RunOnce` must first resolve RSS/Mikan entries, check local cache, skip PikPak login when no new torrent exists, then process new entries serially.

- [ ] **Step 3: Verify and commit**

Run:

```powershell
go test ./internal/app
git add internal/app
git commit -m "feat: add sync orchestration"
```

---

### Task 7: CLI Entrypoint

**Files:**
- Create: `cmd/bangumi-pikpak/main.go`

- [ ] **Step 1: Implement flags and loop**

Required flags:

```text
-config config.json
-interval 600
-once
-log rss-pikpak.log
-state pikpak.json
```

Startup order:

1. Create logger.
2. Load and validate config.
3. Apply proxy settings.
4. Create `pikpak.NewGoAPI`.
5. Create `app.Runner`.
6. Use `signal.NotifyContext` for `SIGINT` and `SIGTERM`.
7. If `-once`, run once and exit.
8. Otherwise run forever and sleep `interval` seconds between cycles.
9. Save Go runtime state after each successful cycle.

- [ ] **Step 2: Verify CLI**

Run:

```powershell
go test ./...
go build ./cmd/bangumi-pikpak
go run ./cmd/bangumi-pikpak -config example.config.json -once
```

Expected: tests pass, build succeeds, sample config run exits with a readable error and no panic.

- [ ] **Step 3: Commit**

```powershell
git add cmd/bangumi-pikpak/main.go
git commit -m "feat: add go cli entrypoint"
```

---

### Task 8: Documentation and Deployment

**Files:**
- Modify: `README.md`
- Create: `Dockerfile`
- Create: `docs/examples/bangumi-pikpak.service`

- [ ] **Step 1: Add Dockerfile**

Use a multi-stage build:

```dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bangumi-pikpak ./cmd/bangumi-pikpak

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/bangumi-pikpak /usr/local/bin/bangumi-pikpak
COPY example.config.json /app/example.config.json
VOLUME ["/app/data"]
CMD ["bangumi-pikpak", "-config", "/app/data/config.json", "-log", "/app/data/rss-pikpak.log"]
```

- [ ] **Step 2: Add systemd example**

Create `docs/examples/bangumi-pikpak.service` with `WorkingDirectory=/opt/bangumi-pikpak` and `ExecStart=/opt/bangumi-pikpak/bangumi-pikpak -config /opt/bangumi-pikpak/config.json -log /opt/bangumi-pikpak/rss-pikpak.log`.

- [ ] **Step 3: Update README**

README must document:

```bash
go build ./cmd/bangumi-pikpak
./bangumi-pikpak -config config.json
go run ./cmd/bangumi-pikpak -config config.json -once
```

Also document flags, Docker usage, systemd usage, existing `config.json` compatibility, and that `main.py` is retained temporarily as the legacy Python implementation.

- [ ] **Step 4: Verify and commit**

Run:

```powershell
go test ./...
go build ./cmd/bangumi-pikpak
docker build -t bangumi-pikpak-go .
```

If Docker is not installed, record the exact Docker error and treat Go tests/build as the required verification.

Commit:

```powershell
git add README.md Dockerfile docs/examples/bangumi-pikpak.service
git commit -m "docs: add go deployment instructions"
```

---

### Task 9: Final Verification

**Files:**
- Modify only files needed to fix verification output.

- [ ] **Step 1: Format**

Run: `gofmt -w cmd internal`

- [ ] **Step 2: Full test**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 3: Build**

Run: `go build ./cmd/bangumi-pikpak`

Expected: PASS.

- [ ] **Step 4: Sample one-shot**

Run: `go run ./cmd/bangumi-pikpak -config example.config.json -once`

Expected: readable error or auth failure, no panic.

- [ ] **Step 5: Diff review**

Run:

```powershell
git status --short
git diff --stat HEAD
```

Expected: only intentional files remain changed. Commit formatting or verification fixes if any:

```powershell
git add .
git commit -m "chore: verify go migration"
```

---

## Self-Review

- Config compatibility is covered by Task 1.
- Proxy, logging, and sanitization are covered by Task 2.
- RSS and Mikan parsing are covered by Task 3.
- Local torrent cache behavior is covered by Task 4.
- PikPak integration through `pikpak-go` is covered by Task 5.
- End-to-end sync behavior is covered by Task 6.
- CLI continuous and one-shot modes are covered by Task 7.
- README, Docker, and systemd are covered by Task 8.
- Final verification is covered by Task 9.

The plan keeps `main.py` as a legacy reference and does not remove Python files.
