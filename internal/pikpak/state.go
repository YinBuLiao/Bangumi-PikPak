package pikpak

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type State struct {
	Username        string    `json:"username"`
	LastLoginTime   time.Time `json:"last_login_time"`
	LastRefreshTime time.Time `json:"last_refresh_time"`
	Client          string    `json:"client"`
	AccessToken     string    `json:"access_token,omitempty"`
	RefreshToken    string    `json:"refresh_token,omitempty"`
}

func LoadState(path string) (State, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return State{}, fmt.Errorf("read pikpak state: %w", err)
	}
	var state State
	if err := json.Unmarshal(b, &state); err != nil {
		return State{}, fmt.Errorf("parse pikpak state: %w", err)
	}
	return state, nil
}
