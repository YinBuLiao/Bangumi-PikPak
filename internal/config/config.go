package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Path          string `json:"path"`
	RSS           string `json:"rss"`
	HTTPProxy     string `json:"http_proxy"`
	HTTPSProxy    string `json:"https_proxy"`
	SocksProxy    string `json:"socks_proxy"`
	EnableProxy   bool   `json:"enable_proxy"`
	MikanUsername string `json:"mikan_username"`
	MikanPassword string `json:"mikan_password"`
	MySQLHost     string `json:"mysql_host"`
	MySQLPort     int    `json:"mysql_port"`
	MySQLDatabase string `json:"mysql_database"`
	MySQLUsername string `json:"mysql_username"`
	MySQLPassword string `json:"mysql_password"`
	MySQLDSN      string `json:"mysql_dsn"`
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	if strings.EqualFold(filepath.Ext(path), ".env") {
		cfg, err := parseEnvConfig(string(b))
		if err != nil {
			return Config{}, fmt.Errorf("parse config %s: %w", path, err)
		}
		return cfg, nil
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if strings.EqualFold(filepath.Ext(path), ".env") {
		b := []byte(formatEnvConfig(cfg))
		if err := os.WriteFile(path, b, 0o600); err != nil {
			return fmt.Errorf("write config %s: %w", path, err)
		}
		return nil
	}
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

func parseEnvConfig(raw string) (Config, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	boolValue := func(key string) bool {
		v, _ := strconv.ParseBool(values[key])
		return v
	}
	intValue := func(key string) int {
		v, _ := strconv.Atoi(values[key])
		return v
	}
	return Config{
		Username:      values["USERNAME"],
		Password:      values["PASSWORD"],
		Path:          values["PATH"],
		RSS:           values["RSS"],
		HTTPProxy:     values["HTTP_PROXY"],
		HTTPSProxy:    values["HTTPS_PROXY"],
		SocksProxy:    values["SOCKS_PROXY"],
		EnableProxy:   boolValue("ENABLE_PROXY"),
		MikanUsername: values["MIKAN_USERNAME"],
		MikanPassword: values["MIKAN_PASSWORD"],
		MySQLHost:     values["MYSQL_HOST"],
		MySQLPort:     intValue("MYSQL_PORT"),
		MySQLDatabase: values["MYSQL_DATABASE"],
		MySQLUsername: values["MYSQL_USERNAME"],
		MySQLPassword: values["MYSQL_PASSWORD"],
		MySQLDSN:      values["MYSQL_DSN"],
		RedisAddr:     values["REDIS_ADDR"],
		RedisPassword: values["REDIS_PASSWORD"],
		RedisDB:       intValue("REDIS_DB"),
	}, nil
}

func formatEnvConfig(c Config) string {
	lines := []string{
		"USERNAME=" + quoteEnv(c.Username),
		"PASSWORD=" + quoteEnv(c.Password),
		"PATH=" + quoteEnv(c.Path),
		"RSS=" + quoteEnv(c.RSS),
		"HTTP_PROXY=" + quoteEnv(c.HTTPProxy),
		"HTTPS_PROXY=" + quoteEnv(c.HTTPSProxy),
		"SOCKS_PROXY=" + quoteEnv(c.SocksProxy),
		"ENABLE_PROXY=" + strconv.FormatBool(c.EnableProxy),
		"MIKAN_USERNAME=" + quoteEnv(c.MikanUsername),
		"MIKAN_PASSWORD=" + quoteEnv(c.MikanPassword),
		"MYSQL_HOST=" + quoteEnv(c.MySQLHost),
		"MYSQL_PORT=" + strconv.Itoa(c.MySQLPort),
		"MYSQL_DATABASE=" + quoteEnv(c.MySQLDatabase),
		"MYSQL_USERNAME=" + quoteEnv(c.MySQLUsername),
		"MYSQL_PASSWORD=" + quoteEnv(c.MySQLPassword),
		"MYSQL_DSN=" + quoteEnv(c.MySQLDSN),
		"REDIS_ADDR=" + quoteEnv(c.RedisAddr),
		"REDIS_PASSWORD=" + quoteEnv(c.RedisPassword),
		"REDIS_DB=" + strconv.Itoa(c.RedisDB),
	}
	return strings.Join(lines, "\n") + "\n"
}

func quoteEnv(v string) string {
	if v == "" {
		return ""
	}
	escaped := strings.ReplaceAll(v, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
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
	if len(exampleFields) > 0 {
		return errors.New("config contains example values: " + strings.Join(exampleFields, ", "))
	}
	return nil
}
