package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	CookieName = "gw_session"
	sessionTTL = 12 * time.Hour
)

type Manager struct {
	accessToken string
	mu          sync.RWMutex
	sessions    map[string]time.Time
}

func NewManager(accessToken string) *Manager {
	return &Manager{
		accessToken: accessToken,
		sessions:    make(map[string]time.Time),
	}
}

func (m *Manager) Enabled() bool {
	return m.accessToken != ""
}

func (m *Manager) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if subtle.ConstantTimeCompare([]byte(body.Token), []byte(m.accessToken)) != 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(struct{ Message string }{Message: "Ungültiger Token"})
		return
	}

	sessionId, err := m.createSession()
	if err != nil {
		slog.Error("failed to create session", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sessionId,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(struct{ Message string }{Message: "ok"})
}

func (m *Manager) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(CookieName); err == nil {
		m.mu.Lock()
		delete(m.sessions, cookie.Value)
		m.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (m *Manager) createSession() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	sessionId := hex.EncodeToString(buf)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expireSessionsLocked()
	m.sessions[sessionId] = time.Now().Add(sessionTTL)
	return sessionId, nil
}

func (m *Manager) expireSessionsLocked() {
	now := time.Now()
	for id, expires := range m.sessions {
		if now.After(expires) {
			delete(m.sessions, id)
		}
	}
}

func (m *Manager) validSession(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	expires, ok := m.sessions[id]
	return ok && time.Now().Before(expires)
}

func (m *Manager) Session(r *http.Request) bool {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return false
	}
	return m.validSession(cookie.Value)
}

func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.Enabled() {
			next.ServeHTTP(w, r)
			return
		}
		if m.Session(r) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(struct{ Message string }{Message: "Nicht angemeldet"})
	})
}
