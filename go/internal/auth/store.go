package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var _ SessionStore = (*MemorySessionStore)(nil)

// MemorySessionStore is an in-memory implementation of SessionStore.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *MemorySessionStore) Create(_ context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.Token] = session
	return nil
}

func (s *MemorySessionStore) FindByToken(_ context.Context, token string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[token]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return sess, nil
}

func (s *MemorySessionStore) Revoke(_ context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
	return nil
}

func (s *MemorySessionStore) RevokeAllForUser(_ context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, sess := range s.sessions {
		if sess.UserID == userID {
			delete(s.sessions, token)
		}
	}
	return nil
}

// MemoryCodeStore is an in-memory implementation of CodeStore.
type MemoryCodeStore struct {
	mu    sync.Mutex
	codes map[string]*Code
}

func NewMemoryCodeStore() *MemoryCodeStore {
	return &MemoryCodeStore{
		codes: make(map[string]*Code),
	}
}

func (s *MemoryCodeStore) Store(_ context.Context, code *Code) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[code.Code] = code
	return nil
}

func (s *MemoryCodeStore) Consume(_ context.Context, code string) (*Code, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ac, ok := s.codes[code]
	if !ok {
		return nil, fmt.Errorf("authorization code not found")
	}
	delete(s.codes, code)
	if time.Now().After(ac.ExpiresAt) && !ac.ExpiresAt.IsZero() {
		return nil, fmt.Errorf("authorization code expired")
	}
	return ac, nil
}
