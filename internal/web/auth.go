package web

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"bangumi-pikpak/internal/cache"
	"bangumi-pikpak/internal/config"
)

const authCookieName = "animex_session"
const sessionTTL = 7 * 24 * time.Hour

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]sessionRecord
	redis    *cache.RedisCache
}

type sessionRecord struct {
	Username  string
	Role      string
	ExpiresAt time.Time
}

func NewSessionStore(redisCache ...*cache.RedisCache) *SessionStore {
	store := &SessionStore{sessions: make(map[string]sessionRecord)}
	if len(redisCache) > 0 {
		store.redis = redisCache[0]
	}
	return store
}

func (s *SessionStore) SetRedis(redisCache *cache.RedisCache) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.redis = redisCache
	s.mu.Unlock()
}

func (s *SessionStore) Create(username, role string) (string, time.Time, error) {
	if s == nil {
		return "", time.Time{}, errors.New("session store is nil")
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(b)
	expires := time.Now().Add(sessionTTL)
	rec := sessionRecord{Username: username, Role: config.NormalizeRole(role), ExpiresAt: expires}
	s.mu.Lock()
	s.sessions[token] = rec
	redisCache := s.redis
	s.mu.Unlock()
	if redisCache != nil {
		if err := redisCache.SetJSON(context.Background(), sessionKey(token), rec, time.Until(expires)); err != nil {
			return "", time.Time{}, err
		}
	}
	return token, expires, nil
}

func (s *SessionStore) User(token string) (config.User, bool) {
	if s == nil || strings.TrimSpace(token) == "" {
		return config.User{}, false
	}
	s.mu.Lock()
	rec, ok := s.sessions[token]
	if !ok {
		s.mu.Unlock()
		return s.redisUser(token)
	}
	if time.Now().After(rec.ExpiresAt) {
		delete(s.sessions, token)
		s.mu.Unlock()
		return config.User{}, false
	}
	s.mu.Unlock()
	return config.User{Username: rec.Username, Role: config.NormalizeRole(rec.Role)}, true
}

func (s *SessionStore) redisUser(token string) (config.User, bool) {
	s.mu.Lock()
	redisCache := s.redis
	s.mu.Unlock()
	if redisCache == nil {
		return config.User{}, false
	}
	var rec sessionRecord
	ok, err := redisCache.GetJSON(context.Background(), sessionKey(token), &rec)
	if err != nil || !ok || time.Now().After(rec.ExpiresAt) {
		if ok {
			_ = redisCache.Delete(context.Background(), sessionKey(token))
		}
		return config.User{}, false
	}
	s.mu.Lock()
	s.sessions[token] = rec
	s.mu.Unlock()
	return config.User{Username: rec.Username, Role: config.NormalizeRole(rec.Role)}, true
}

func (s *SessionStore) Username(token string) (string, bool) {
	user, ok := s.User(token)
	return user.Username, ok
}

func (s *SessionStore) Delete(token string) {
	if s == nil || strings.TrimSpace(token) == "" {
		return
	}
	s.mu.Lock()
	delete(s.sessions, token)
	redisCache := s.redis
	s.mu.Unlock()
	if redisCache != nil {
		_ = redisCache.Delete(context.Background(), sessionKey(token))
	}
}

func sessionKey(token string) string {
	return "session:" + strings.TrimSpace(token)
}

func validAdminCredentials(cfg config.Config, username, password string) bool {
	expectedUser := strings.TrimSpace(cfg.AdminUsername)
	expectedPass := cfg.AdminPassword
	if expectedUser == "" || expectedPass == "" {
		return false
	}
	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(expectedUser)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(expectedPass)) == 1
	return userOK && passOK
}

func validUserCredentials(user config.User, password string) bool {
	if strings.TrimSpace(user.Username) == "" || user.Password == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(user.Password)) == 1
}

func setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
