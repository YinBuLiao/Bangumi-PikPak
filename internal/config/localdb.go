package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	DefaultLocalDBPath = "data/animex.db"
	RoleAdmin          = "admin"
	RoleUser           = "user"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

type InviteCode struct {
	Code      string `json:"code"`
	UsedBy    string `json:"used_by,omitempty"`
	UsedAt    string `json:"used_at,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
	CreatedAt string `json:"created_at"`
}

type DownloadRequest struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Title        string `json:"title"`
	BangumiTitle string `json:"bangumi_title"`
	EpisodeLabel string `json:"episode_label"`
	TorrentURL   string `json:"torrent_url"`
	Magnet       string `json:"magnet"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

type PikPakState struct {
	Username        string    `json:"username"`
	LastLoginTime   time.Time `json:"last_login_time"`
	LastRefreshTime time.Time `json:"last_refresh_time"`
	Client          string    `json:"client"`
	AccessToken     string    `json:"access_token,omitempty"`
	RefreshToken    string    `json:"refresh_token,omitempty"`
}

type LocalDB struct {
	path string
	db   *sql.DB
}

func OpenLocalDB(path string) (*LocalDB, error) {
	if path == "" {
		path = DefaultLocalDBPath
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create local config directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open local config db: %w", err)
	}
	local := &LocalDB{path: path, db: db}
	if err := local.EnsureSchema(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return local, nil
}

func (l *LocalDB) Close() error {
	if l == nil || l.db == nil {
		return nil
	}
	return l.db.Close()
}

func (l *LocalDB) EnsureSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS system_config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS install_state (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			installed INTEGER NOT NULL,
			installed_at TEXT NOT NULL,
			version TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS invite_codes (
			code TEXT PRIMARY KEY,
			used_by TEXT NOT NULL DEFAULT '',
			used_at TEXT NOT NULL DEFAULT '',
			expires_at TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS download_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			title TEXT NOT NULL,
			bangumi_title TEXT NOT NULL DEFAULT '',
			episode_label TEXT NOT NULL DEFAULT '',
			torrent_url TEXT NOT NULL DEFAULT '',
			magnet TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TEXT NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err := l.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("ensure local config schema: %w", err)
		}
	}
	if err := l.ensureUserRoleColumn(ctx); err != nil {
		return err
	}
	if err := l.ensureInviteExpiresColumn(ctx); err != nil {
		return err
	}
	return nil
}

func (l *LocalDB) ensureInviteExpiresColumn(ctx context.Context) error {
	rows, err := l.db.QueryContext(ctx, `PRAGMA table_info(invite_codes)`)
	if err != nil {
		return fmt.Errorf("inspect invite schema: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("read invite schema: %w", err)
		}
		if name == "expires_at" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read invite schema: %w", err)
	}
	if _, err := l.db.ExecContext(ctx, `ALTER TABLE invite_codes ADD COLUMN expires_at TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate invite expires column: %w", err)
	}
	return nil
}

func (l *LocalDB) ensureUserRoleColumn(ctx context.Context) error {
	rows, err := l.db.QueryContext(ctx, `PRAGMA table_info(admin_users)`)
	if err != nil {
		return fmt.Errorf("inspect user schema: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("read user schema: %w", err)
		}
		if name == "role" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read user schema: %w", err)
	}
	if _, err := l.db.ExecContext(ctx, `ALTER TABLE admin_users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'`); err != nil {
		return fmt.Errorf("migrate user role column: %w", err)
	}
	return nil
}

func NormalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case RoleAdmin:
		return RoleAdmin
	default:
		return RoleUser
	}
}

func (l *LocalDB) SaveUser(ctx context.Context, user User) error {
	if l == nil || l.db == nil {
		return fmt.Errorf("local config db is nil")
	}
	username := strings.TrimSpace(user.Username)
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if user.Password == "" {
		return fmt.Errorf("password is required")
	}
	now := time.Now().Format(time.RFC3339)
	if _, err := l.db.ExecContext(ctx, `INSERT INTO admin_users (username, password, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET password=excluded.password, role=excluded.role, updated_at=excluded.updated_at`, username, user.Password, NormalizeRole(user.Role), now, now); err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

func (l *LocalDB) FindUser(ctx context.Context, username string) (User, bool, error) {
	if l == nil || l.db == nil {
		return User{}, false, nil
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return User{}, false, nil
	}
	var user User
	err := l.db.QueryRowContext(ctx, `SELECT username, password, role FROM admin_users WHERE username = ?`, username).Scan(&user.Username, &user.Password, &user.Role)
	if err == sql.ErrNoRows {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, fmt.Errorf("find user: %w", err)
	}
	user.Role = NormalizeRole(user.Role)
	return user, true, nil
}

func (l *LocalDB) ListUsers(ctx context.Context) ([]User, error) {
	if l == nil || l.db == nil {
		return nil, fmt.Errorf("local config db is nil")
	}
	rows, err := l.db.QueryContext(ctx, `SELECT username, role FROM admin_users ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Username, &user.Role); err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		user.Role = NormalizeRole(user.Role)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (l *LocalDB) SaveInviteCode(ctx context.Context, code string, expiresAt string) error {
	if l == nil || l.db == nil {
		return fmt.Errorf("local config db is nil")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("invite code is required")
	}
	expiresAt = strings.TrimSpace(expiresAt)
	now := time.Now().Format(time.RFC3339)
	if _, err := l.db.ExecContext(ctx, `INSERT INTO invite_codes (code, expires_at, created_at) VALUES (?, ?, ?)`, code, expiresAt, now); err != nil {
		return fmt.Errorf("save invite code: %w", err)
	}
	return nil
}

func (l *LocalDB) ListInviteCodes(ctx context.Context) ([]InviteCode, error) {
	if l == nil || l.db == nil {
		return nil, fmt.Errorf("local config db is nil")
	}
	rows, err := l.db.QueryContext(ctx, `SELECT code, used_by, used_at, expires_at, created_at FROM invite_codes ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list invite codes: %w", err)
	}
	defer rows.Close()
	codes := []InviteCode{}
	for rows.Next() {
		var code InviteCode
		if err := rows.Scan(&code.Code, &code.UsedBy, &code.UsedAt, &code.ExpiresAt, &code.CreatedAt); err != nil {
			return nil, fmt.Errorf("list invite codes: %w", err)
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list invite codes: %w", err)
	}
	return codes, nil
}

func (l *LocalDB) DeleteInviteCodes(ctx context.Context, codes []string) (int64, error) {
	if l == nil || l.db == nil {
		return 0, fmt.Errorf("local config db is nil")
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var total int64
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		res, err := tx.ExecContext(ctx, `DELETE FROM invite_codes WHERE code = ?`, code)
		if err != nil {
			return 0, fmt.Errorf("delete invite code: %w", err)
		}
		n, _ := res.RowsAffected()
		total += n
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return total, nil
}

func (l *LocalDB) RegisterUser(ctx context.Context, user User, inviteCode string, requireInvite bool) error {
	if l == nil || l.db == nil {
		return fmt.Errorf("local config db is nil")
	}
	username := strings.TrimSpace(user.Username)
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if user.Password == "" {
		return fmt.Errorf("password is required")
	}
	inviteCode = strings.TrimSpace(inviteCode)
	if requireInvite && inviteCode == "" {
		return fmt.Errorf("invite code is required")
	}
	now := time.Now().Format(time.RFC3339)
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(1) FROM admin_users WHERE username = ?`, username).Scan(&exists); err != nil {
		return fmt.Errorf("check user exists: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("username already exists")
	}
	if requireInvite {
		var usedBy, expiresAt string
		err := tx.QueryRowContext(ctx, `SELECT used_by, expires_at FROM invite_codes WHERE code = ?`, inviteCode).Scan(&usedBy, &expiresAt)
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid invite code")
		}
		if err != nil {
			return fmt.Errorf("check invite code: %w", err)
		}
		if strings.TrimSpace(usedBy) != "" {
			return fmt.Errorf("invite code has been used")
		}
		if strings.TrimSpace(expiresAt) != "" {
			expires, err := time.Parse(time.RFC3339, expiresAt)
			if err != nil {
				return fmt.Errorf("invalid invite code expires time")
			}
			if time.Now().After(expires) {
				return fmt.Errorf("invite code has expired")
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE invite_codes SET used_by = ?, used_at = ? WHERE code = ? AND used_by = ''`, username, now, inviteCode); err != nil {
			return fmt.Errorf("consume invite code: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO admin_users (username, password, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, username, user.Password, RoleUser, now, now); err != nil {
		return fmt.Errorf("register user: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (l *LocalDB) CountDownloadRequestsSince(ctx context.Context, username string, since time.Time) (int, error) {
	if l == nil || l.db == nil {
		return 0, fmt.Errorf("local config db is nil")
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return 0, nil
	}
	var count int
	err := l.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM download_requests WHERE username = ? AND created_at >= ?`, username, since.Format(time.RFC3339)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count download requests: %w", err)
	}
	return count, nil
}

func (l *LocalDB) SaveDownloadRequest(ctx context.Context, req DownloadRequest) (int64, error) {
	if l == nil || l.db == nil {
		return 0, fmt.Errorf("local config db is nil")
	}
	username := strings.TrimSpace(req.Username)
	title := strings.TrimSpace(req.Title)
	if username == "" {
		return 0, fmt.Errorf("username is required")
	}
	if title == "" {
		return 0, fmt.Errorf("title is required")
	}
	now := time.Now().Format(time.RFC3339)
	res, err := l.db.ExecContext(ctx, `INSERT INTO download_requests (username, title, bangumi_title, episode_label, torrent_url, magnet, status, created_at) VALUES (?, ?, ?, ?, ?, ?, 'pending', ?)`,
		username, title, strings.TrimSpace(req.BangumiTitle), strings.TrimSpace(req.EpisodeLabel), strings.TrimSpace(req.TorrentURL), strings.TrimSpace(req.Magnet), now)
	if err != nil {
		return 0, fmt.Errorf("save download request: %w", err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func (l *LocalDB) ListDownloadRequests(ctx context.Context) ([]DownloadRequest, error) {
	if l == nil || l.db == nil {
		return nil, fmt.Errorf("local config db is nil")
	}
	rows, err := l.db.QueryContext(ctx, `SELECT id, username, title, bangumi_title, episode_label, torrent_url, magnet, status, created_at FROM download_requests WHERE status = 'pending' ORDER BY id DESC LIMIT 200`)
	if err != nil {
		return nil, fmt.Errorf("list download requests: %w", err)
	}
	defer rows.Close()
	items := []DownloadRequest{}
	for rows.Next() {
		var item DownloadRequest
		if err := rows.Scan(&item.ID, &item.Username, &item.Title, &item.BangumiTitle, &item.EpisodeLabel, &item.TorrentURL, &item.Magnet, &item.Status, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("list download requests: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list download requests: %w", err)
	}
	return items, nil
}

func (l *LocalDB) FindDownloadRequest(ctx context.Context, id int64) (DownloadRequest, bool, error) {
	if l == nil || l.db == nil {
		return DownloadRequest{}, false, fmt.Errorf("local config db is nil")
	}
	var item DownloadRequest
	err := l.db.QueryRowContext(ctx, `SELECT id, username, title, bangumi_title, episode_label, torrent_url, magnet, status, created_at FROM download_requests WHERE id = ?`, id).
		Scan(&item.ID, &item.Username, &item.Title, &item.BangumiTitle, &item.EpisodeLabel, &item.TorrentURL, &item.Magnet, &item.Status, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return DownloadRequest{}, false, nil
	}
	if err != nil {
		return DownloadRequest{}, false, fmt.Errorf("find download request: %w", err)
	}
	return item, true, nil
}

func (l *LocalDB) UpdateDownloadRequestStatus(ctx context.Context, id int64, status string) error {
	if l == nil || l.db == nil {
		return fmt.Errorf("local config db is nil")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		return fmt.Errorf("status is required")
	}
	if _, err := l.db.ExecContext(ctx, `UPDATE download_requests SET status = ? WHERE id = ?`, status, id); err != nil {
		return fmt.Errorf("update download request status: %w", err)
	}
	return nil
}

func (l *LocalDB) SaveConfig(ctx context.Context, cfg Config) error {
	if l == nil || l.db == nil {
		return fmt.Errorf("local config db is nil")
	}
	payload, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal local config: %w", err)
	}
	now := time.Now().Format(time.RFC3339)
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO system_config (key, value, updated_at) VALUES ('runtime_config', ?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`, string(payload), now); err != nil {
		return fmt.Errorf("write runtime config: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO admin_users (username, password, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET password=excluded.password, role=excluded.role, updated_at=excluded.updated_at`, cfg.AdminUsername, cfg.AdminPassword, RoleAdmin, now, now); err != nil {
		return fmt.Errorf("write admin user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO install_state (id, installed, installed_at, version) VALUES (1, 1, ?, '0.1.0')
		ON CONFLICT(id) DO UPDATE SET installed=1, installed_at=excluded.installed_at, version=excluded.version`, now); err != nil {
		return fmt.Errorf("write install state: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (l *LocalDB) LoadConfig(ctx context.Context) (Config, bool, error) {
	if l == nil || l.db == nil {
		return Config{}, false, nil
	}
	var payload string
	err := l.db.QueryRowContext(ctx, `SELECT value FROM system_config WHERE key='runtime_config'`).Scan(&payload)
	if err == sql.ErrNoRows {
		return Config{}, false, nil
	}
	if err != nil {
		return Config{}, false, fmt.Errorf("read runtime config: %w", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return Config{}, false, fmt.Errorf("parse runtime config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
		return Config{}, false, fmt.Errorf("parse runtime config: %w", err)
	}
	if _, ok := raw["require_login"]; !ok {
		cfg.RequireLogin = true
	}
	if _, ok := raw["enable_registration"]; !ok {
		cfg.EnableRegistration = true
	}
	defaults := Default()
	if _, ok := raw["user_daily_download_limit"]; !ok {
		cfg.UserDailyDownloadLimit = defaults.UserDailyDownloadLimit
	}
	if _, ok := raw["storage_provider"]; !ok || strings.TrimSpace(cfg.StorageProvider) == "" {
		cfg.StorageProvider = defaults.StorageProvider
	}
	if strings.TrimSpace(cfg.Drive115RootCID) == "" {
		cfg.Drive115RootCID = defaults.Drive115RootCID
	}
	if strings.TrimSpace(cfg.Aria2RPCURL) == "" {
		cfg.Aria2RPCURL = defaults.Aria2RPCURL
	}
	if strings.TrimSpace(cfg.LocalStoragePath) == "" {
		cfg.LocalStoragePath = defaults.LocalStoragePath
	}
	return cfg, true, nil
}

func (l *LocalDB) SavePikPakState(ctx context.Context, state PikPakState) error {
	if l == nil || l.db == nil {
		return fmt.Errorf("local config db is nil")
	}
	if strings.TrimSpace(state.AccessToken) == "" || strings.TrimSpace(state.RefreshToken) == "" {
		return nil
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal pikpak state: %w", err)
	}
	now := time.Now().Format(time.RFC3339)
	if _, err := l.db.ExecContext(ctx, `INSERT INTO system_config (key, value, updated_at) VALUES ('pikpak_state', ?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`, string(payload), now); err != nil {
		return fmt.Errorf("save pikpak state: %w", err)
	}
	return nil
}

func (l *LocalDB) LoadPikPakState(ctx context.Context) (PikPakState, bool, error) {
	if l == nil || l.db == nil {
		return PikPakState{}, false, nil
	}
	var payload string
	err := l.db.QueryRowContext(ctx, `SELECT value FROM system_config WHERE key='pikpak_state'`).Scan(&payload)
	if err == sql.ErrNoRows {
		return PikPakState{}, false, nil
	}
	if err != nil {
		return PikPakState{}, false, fmt.Errorf("read pikpak state: %w", err)
	}
	var state PikPakState
	if err := json.Unmarshal([]byte(payload), &state); err != nil {
		return PikPakState{}, false, fmt.Errorf("parse pikpak state: %w", err)
	}
	return state, true, nil
}

func (l *LocalDB) Installed(ctx context.Context) bool {
	if l == nil || l.db == nil {
		return false
	}
	var installed int
	if err := l.db.QueryRowContext(ctx, `SELECT installed FROM install_state WHERE id=1`).Scan(&installed); err != nil {
		return false
	}
	return installed == 1
}
