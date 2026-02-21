package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SignedCookieStore は HMAC でセッション情報をクッキーにエンコードする。
type SignedCookieStore struct {
	secret  []byte
	now     func() time.Time
	mu      sync.Mutex
	revoked map[string]int64
}

// NewSignedCookieStore はマルチインスタンス運用向けのストアを返す。
func NewSignedCookieStore(secret string) (*SignedCookieStore, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, errors.New("session secret is required")
	}
	return &SignedCookieStore{
		secret: []byte(secret),
		now: func() time.Time {
			return time.Now().UTC()
		},
		revoked: make(map[string]int64),
	}, nil
}

func (s *SignedCookieStore) Create(userID uuid.UUID, ttl time.Duration) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("user id is required")
	}
	expiresAt := s.now().Add(ttl).Unix()
	payload := userID.String() + "|" + strconv.FormatInt(expiresAt, 10)
	sig := s.sign(payload)
	token := payload + "|" + sig
	sessionID := base64.RawURLEncoding.EncodeToString([]byte(token))
	s.mu.Lock()
	s.cleanupExpiredLocked()
	s.mu.Unlock()
	return sessionID, nil
}

func (s *SignedCookieStore) Get(sessionID string) (uuid.UUID, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(sessionID)
	if err != nil {
		return uuid.Nil, false
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 {
		return uuid.Nil, false
	}
	payload := strings.Join(parts[:2], "|")
	if !s.verify(payload, parts[2]) {
		return uuid.Nil, false
	}
	expiresUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return uuid.Nil, false
	}
	if s.now().Unix() > expiresUnix {
		return uuid.Nil, false
	}
	if s.isRevoked(sessionID, expiresUnix) {
		return uuid.Nil, false
	}
	userID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, false
	}
	return userID, true
}

func (s *SignedCookieStore) Delete(sessionID string) {
	if sessionID == "" {
		return
	}
	raw, err := base64.RawURLEncoding.DecodeString(sessionID)
	if err != nil {
		return
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 {
		return
	}
	expiresUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.cleanupExpiredLocked()
	s.revoked[sessionID] = expiresUnix
	s.mu.Unlock()
}

func (s *SignedCookieStore) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *SignedCookieStore) verify(payload, sig string) bool {
	expected, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return hmac.Equal(mac.Sum(nil), expected)
}

func (s *SignedCookieStore) isRevoked(sessionID string, expiresUnix int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked()
	exp, ok := s.revoked[sessionID]
	if !ok {
		return false
	}
	if exp <= s.now().Unix() {
		delete(s.revoked, sessionID)
		return false
	}
	return true
}

func (s *SignedCookieStore) cleanupExpiredLocked() {
	now := s.now().Unix()
	for token, exp := range s.revoked {
		if exp <= now {
			delete(s.revoked, token)
		}
	}
}
