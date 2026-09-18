package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"dnd-backend/internal/domain"
)

// Session is the identity bound to an issued token.
type Session struct {
	Pseudo   string
	CharName string
	Class    string
	Created  time.Time
}

// SessionStore keeps issued tokens in memory. Tokens are opaque random values;
// sessions expire after a TTL so stale tokens cannot linger forever.
type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]Session
	ttl      time.Duration
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{
		sessions: make(map[string]Session),
		ttl:      ttl,
	}
}

func (s *SessionStore) Create(pseudo, charName, class string) string {
	token := newToken()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = Session{
		Pseudo:   pseudo,
		CharName: charName,
		Class:    class,
		Created:  time.Now(),
	}
	return token
}

func (s *SessionStore) Get(token string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if token == "" {
		return Session{}, false
	}
	session, ok := s.sessions[token]
	if !ok {
		return Session{}, false
	}
	if time.Since(session.Created) > s.ttl {
		delete(s.sessions, token)
		return Session{}, false
	}
	return session, true
}

func (s *SessionStore) Revoke(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func newToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

type loginRequest struct {
	Pseudo    string `json:"pseudo"`
	CharName  string `json:"charName"`
	CharClass string `json:"charClass"`
}

type loginResponse struct {
	Token  string `json:"token"`
	Pseudo string `json:"pseudo"`
}

// LoginHandler issues a token for an authenticated (pseudo-based) identity.
// Optional: accepts a sha256 of a shared secret when LOGIN_SECRET is set.
func LoginHandler(sessions *SessionStore, loginLimiter *RateLimiter, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ip := ClientIP(r)
		if !loginLimiter.Allow(ip) {
			logger.Warn("login rate limited", "ip", ip)
			http.Error(w, "too many attempts", http.StatusTooManyRequests)
			return
		}

		var req loginRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		pseudo := domain.SanitizeString(req.Pseudo, 24)
		charName := domain.SanitizeString(req.CharName, 40)
		class := domain.SanitizeString(req.CharClass, 20)
		if !domain.ValidPseudo(pseudo) || !domain.ValidCharName(charName) || class == "" {
			logger.Warn("login rejected", "ip", ip, "pseudo", pseudo)
			http.Error(w, "invalid identity", http.StatusBadRequest)
			return
		}

		token := sessions.Create(pseudo, charName, class)
		logger.Info("login", "pseudo", pseudo)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(loginResponse{Token: token, Pseudo: pseudo})
	}
}

// LogoutHandler revokes the token passed as the "token" query parameter.
func LogoutHandler(sessions *SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions.Revoke(r.URL.Query().Get("token"))
		w.WriteHeader(http.StatusNoContent)
	}
}

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := indexByte(xff, ','); i >= 0 {
			return trimSpace(xff[:i])
		}
		return trimSpace(xff)
	}
	host := r.RemoteAddr
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}