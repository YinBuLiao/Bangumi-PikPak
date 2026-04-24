package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Path        string `json:"path"`
	RSS         string `json:"rss"`
	HTTPProxy   string `json:"http_proxy"`
	HTTPSProxy  string `json:"https_proxy"`
	SocksProxy  string `json:"socks_proxy"`
	EnableProxy bool   `json:"enable_proxy"`
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

func (c Config) Validate() error {
	var missing []string
	if strings.TrimSpace(c.Username) == "" {
		missing = append(missing, "username")
	}
	if strings.TrimSpace(c.Password) == "" {
		missing = append(missing, "password")
	}
	if strings.TrimSpace(c.Path) == "" {
		missing = append(missing, "path")
	}
	if strings.TrimSpace(c.RSS) == "" {
		missing = append(missing, "rss")
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
	if c.Username == examples["username"] {
		exampleFields = append(exampleFields, "username")
	}
	if c.Password == examples["password"] {
		exampleFields = append(exampleFields, "password")
	}
	if c.Path == examples["path"] {
		exampleFields = append(exampleFields, "path")
	}
	if strings.Contains(c.RSS, "your_token_here") {
		exampleFields = append(exampleFields, "rss")
	}
	if len(exampleFields) > 0 {
		return errors.New("config contains example values: " + strings.Join(exampleFields, ", "))
	}
	return nil
}
