package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Username               string `json:"username"`
	Password               string `json:"password"`
	PikPakAuthMode         string `json:"pikpak_auth_mode"`
	PikPakAccessToken      string `json:"pikpak_access_token"`
	PikPakRefreshToken     string `json:"pikpak_refresh_token"`
	PikPakEncodedToken     string `json:"pikpak_encoded_token"`
	Path                   string `json:"path"`
	RSS                    string `json:"rss"`
	HTTPProxy              string `json:"http_proxy"`
	HTTPSProxy             string `json:"https_proxy"`
	SocksProxy             string `json:"socks_proxy"`
	EnableProxy            bool   `json:"enable_proxy"`
	RequireLogin           bool   `json:"require_login"`
	EnableRegistration     bool   `json:"enable_registration"`
	RequireInvite          bool   `json:"require_invite"`
	UserDailyDownloadLimit int    `json:"user_daily_download_limit"`
	MikanUsername          string `json:"mikan_username"`
	MikanPassword          string `json:"mikan_password"`
	MySQLHost              string `json:"mysql_host"`
	MySQLPort              int    `json:"mysql_port"`
	MySQLDatabase          string `json:"mysql_database"`
	MySQLUsername          string `json:"mysql_username"`
	MySQLPassword          string `json:"mysql_password"`
	MySQLDSN               string `json:"mysql_dsn"`
	RedisAddr              string `json:"redis_addr"`
	RedisPassword          string `json:"redis_password"`
	RedisDB                int    `json:"redis_db"`
	AdminUsername          string `json:"admin_username"`
	AdminPassword          string `json:"admin_password"`
	StorageProvider        string `json:"storage_provider"`
	Drive115Cookie         string `json:"drive115_cookie"`
	Drive115RootCID        string `json:"drive115_root_cid"`
	Aria2RPCURL            string `json:"aria2_rpc_url"`
	Aria2RPCSecret         string `json:"aria2_rpc_secret"`
	LocalStoragePath       string `json:"local_storage_path"`
	NASStoragePath         string `json:"nas_storage_path"`
}

func Default() Config {
	return Config{
		PikPakAuthMode:         "password",
		HTTPProxy:              "http://127.0.0.1:7890",
		HTTPSProxy:             "http://127.0.0.1:7890",
		SocksProxy:             "socks5://127.0.0.1:7890",
		MySQLHost:              "127.0.0.1",
		MySQLPort:              3306,
		MySQLDatabase:          "anime",
		RedisAddr:              "127.0.0.1:6379",
		RedisDB:                0,
		RequireLogin:           true,
		EnableRegistration:     true,
		RequireInvite:          false,
		UserDailyDownloadLimit: 3,
		AdminUsername:          "admin",
		StorageProvider:        "pikpak",
		Drive115RootCID:        "0",
		Aria2RPCURL:            "http://127.0.0.1:6800/jsonrpc",
		LocalStoragePath:       "downloads",
	}
}

func InstallLockPath(configPath string) string {
	dir := filepath.Dir(strings.TrimSpace(configPath))
	if dir == "" || dir == "." {
		return ".install.lock"
	}
	return filepath.Join(dir, ".install.lock")
}

func HasInstallLock(configPath string) bool {
	info, err := os.Stat(InstallLockPath(configPath))
	return err == nil && !info.IsDir()
}

func WriteInstallLock(configPath string) error {
	lockPath := InstallLockPath(configPath)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return fmt.Errorf("create install lock directory: %w", err)
	}
	content := fmt.Sprintf("installed_at=%s\nconfig=%s\n", time.Now().Format(time.RFC3339), configPath)
	if err := os.WriteFile(lockPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write install lock %s: %w", lockPath, err)
	}
	return nil
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	b, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

func (c Config) MySQLConfigured() bool {
	return strings.TrimSpace(c.MySQLDSN) != "" || (strings.TrimSpace(c.MySQLHost) != "" && strings.TrimSpace(c.MySQLDatabase) != "" && strings.TrimSpace(c.MySQLUsername) != "")
}

func (c Config) MikanConfigured() bool {
	return strings.TrimSpace(c.MikanUsername) != "" && strings.TrimSpace(c.MikanPassword) != ""
}

func (c Config) RedisConfigured() bool {
	return strings.TrimSpace(c.RedisAddr) != ""
}

func (c Config) PikPakTokenAuth() bool {
	return strings.EqualFold(strings.TrimSpace(c.PikPakAuthMode), "token")
}

func (c Config) PikPakTokenConfigured() bool {
	if strings.TrimSpace(c.PikPakEncodedToken) != "" {
		return true
	}
	return strings.TrimSpace(c.PikPakAccessToken) != "" && strings.TrimSpace(c.PikPakRefreshToken) != ""
}

func (c Config) NormalizedStorageProvider() string {
	provider := strings.ToLower(strings.TrimSpace(c.StorageProvider))
	switch provider {
	case "", "pikpak":
		return "pikpak"
	case "115", "drive115":
		return "drive115"
	case "local":
		return "local"
	case "nas":
		return "nas"
	default:
		return provider
	}
}

func (c Config) Validate() error {
	var missing []string
	switch c.NormalizedStorageProvider() {
	case "pikpak":
		if c.PikPakTokenAuth() {
			if !c.PikPakTokenConfigured() {
				missing = append(missing, "pikpak token")
			}
		} else {
			if strings.TrimSpace(c.Username) == "" {
				missing = append(missing, "username")
			}
			if strings.TrimSpace(c.Password) == "" {
				missing = append(missing, "password")
			}
		}
		if strings.TrimSpace(c.Path) == "" {
			missing = append(missing, "path")
		}
	case "drive115":
		if strings.TrimSpace(c.Drive115Cookie) == "" {
			missing = append(missing, "drive115_cookie")
		}
	case "local":
		if strings.TrimSpace(c.Aria2RPCURL) == "" {
			missing = append(missing, "aria2_rpc_url")
		}
		if strings.TrimSpace(c.LocalStoragePath) == "" {
			missing = append(missing, "local_storage_path")
		}
	case "nas":
		if strings.TrimSpace(c.Aria2RPCURL) == "" {
			missing = append(missing, "aria2_rpc_url")
		}
		if strings.TrimSpace(c.NASStoragePath) == "" {
			missing = append(missing, "nas_storage_path")
		}
	default:
		missing = append(missing, "storage_provider")
	}
	if strings.TrimSpace(c.AdminUsername) == "" {
		missing = append(missing, "admin_username")
	}
	if strings.TrimSpace(c.AdminPassword) == "" {
		missing = append(missing, "admin_password")
	}
	if len(missing) > 0 {
		return errors.New("missing required config fields: " + strings.Join(missing, ", "))
	}
	examples := map[string]string{
		"username": "your_email@example.com",
		"password": "your_password",
		"path":     "your_pikpak_folder_id",
	}
	var exampleFields []string
	if c.NormalizedStorageProvider() == "pikpak" && !c.PikPakTokenAuth() && c.Username == examples["username"] {
		exampleFields = append(exampleFields, "username")
	}
	if c.NormalizedStorageProvider() == "pikpak" && !c.PikPakTokenAuth() && c.Password == examples["password"] {
		exampleFields = append(exampleFields, "password")
	}
	if c.NormalizedStorageProvider() == "pikpak" && c.Path == examples["path"] {
		exampleFields = append(exampleFields, "path")
	}
	if strings.Contains(c.RSS, "your_token_here") {
		exampleFields = append(exampleFields, "rss")
	}
	if c.MikanUsername == "your_mikan_email@example.com" {
		exampleFields = append(exampleFields, "mikan_username")
	}
	if c.MikanPassword == "your_mikan_password" {
		exampleFields = append(exampleFields, "mikan_password")
	}
	if c.MySQLUsername == "your_mysql_user" {
		exampleFields = append(exampleFields, "mysql_username")
	}
	if c.MySQLPassword == "your_mysql_password" {
		exampleFields = append(exampleFields, "mysql_password")
	}
	if c.AdminUsername == "your_admin_username" {
		exampleFields = append(exampleFields, "admin_username")
	}
	if c.AdminPassword == "your_admin_password" {
		exampleFields = append(exampleFields, "admin_password")
	}
	if len(exampleFields) > 0 {
		return errors.New("config contains example values: " + strings.Join(exampleFields, ", "))
	}
	return nil
}

func (c Config) ValidateInstall() error {
	var missing []string
	if !c.MySQLConfigured() {
		missing = append(missing, "mysql")
	}
	if strings.TrimSpace(c.AdminUsername) == "" {
		missing = append(missing, "admin_username")
	}
	if strings.TrimSpace(c.AdminPassword) == "" {
		missing = append(missing, "admin_password")
	}
	if len(missing) > 0 {
		return errors.New("missing required config fields: " + strings.Join(missing, ", "))
	}
	return nil
}
