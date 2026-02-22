// Package handshake implements ZK-enhanced peer authentication and session management.
package handshake

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Session represents an authenticated peer session.
type Session struct {
	PeerDID   string    `json:"peerDid"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	ZKVerified bool    `json:"zkVerified"`
}

// IsExpired reports whether the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionStore manages authenticated peer sessions with TTL eviction.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session // keyed by peer DID
	hmacKey  []byte
	ttl      time.Duration
}

// NewSessionStore creates a session store with the given TTL.
func NewSessionStore(ttl time.Duration) (*SessionStore, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate HMAC key: %w", err)
	}

	return &SessionStore{
		sessions: make(map[string]*Session),
		hmacKey:  key,
		ttl:      ttl,
	}, nil
}

// Create creates a new session for the given peer DID.
func (s *SessionStore) Create(peerDID string, zkVerified bool) (*Session, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	mac := hmac.New(sha256.New, s.hmacKey)
	mac.Write(tokenBytes)
	mac.Write([]byte(peerDID))
	token := hex.EncodeToString(mac.Sum(nil))

	now := time.Now()
	sess := &Session{
		PeerDID:    peerDID,
		Token:      token,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.ttl),
		ZKVerified: zkVerified,
	}

	s.mu.Lock()
	s.sessions[peerDID] = sess
	s.mu.Unlock()

	return sess, nil
}

// Validate checks if a session token is valid for the given peer DID.
func (s *SessionStore) Validate(peerDID, token string) bool {
	s.mu.RLock()
	sess, ok := s.sessions[peerDID]
	s.mu.RUnlock()

	if !ok || sess.IsExpired() {
		if ok {
			s.Remove(peerDID)
		}
		return false
	}

	return sess.Token == token
}

// Get returns the session for the given peer DID, or nil if not found/expired.
func (s *SessionStore) Get(peerDID string) *Session {
	s.mu.RLock()
	sess, ok := s.sessions[peerDID]
	s.mu.RUnlock()

	if !ok {
		return nil
	}
	if sess.IsExpired() {
		s.Remove(peerDID)
		return nil
	}
	return sess
}

// Remove deletes a session.
func (s *SessionStore) Remove(peerDID string) {
	s.mu.Lock()
	delete(s.sessions, peerDID)
	s.mu.Unlock()
}

// ActiveSessions returns all non-expired sessions.
func (s *SessionStore) ActiveSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var active []*Session
	for _, sess := range s.sessions {
		if !sess.IsExpired() {
			active = append(active, sess)
		}
	}
	return active
}

// Cleanup removes all expired sessions.
func (s *SessionStore) Cleanup() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	for did, sess := range s.sessions {
		if sess.IsExpired() {
			delete(s.sessions, did)
			removed++
		}
	}
	return removed
}
