package session

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Store describes the operations used by handlers/middleware.
type Store interface {
	Create(userID uuid.UUID, ttl time.Duration) (string, error)
	Get(sessionID string) (uuid.UUID, bool)
	Delete(sessionID string)
}

// MemoryStore keeps sessions in-process. Suitable for local dev.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]sessionValue
}

type sessionValue struct {
	userID    uuid.UUID
	expiresAt time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]sessionValue)}
}

func (s *MemoryStore) Create(userID uuid.UUID, ttl time.Duration) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("user id is required")
	}
	id := uuid.NewString()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = sessionValue{userID: userID, expiresAt: time.Now().Add(ttl)}
	return id, nil
}

func (s *MemoryStore) Get(sessionID string) (uuid.UUID, bool) {
	s.mu.RLock()
	session, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	if !ok {
		return uuid.Nil, false
	}
	if time.Now().After(session.expiresAt) {
		s.Delete(sessionID)
		return uuid.Nil, false
	}
	return session.userID, true
}

func (s *MemoryStore) Delete(sessionID string) {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}
