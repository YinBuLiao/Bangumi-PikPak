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

func SaveState(path string, state State) error {
	b, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal pikpak state: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("write pikpak state: %w", err)
	}
	return nil
}
