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

// InvalidationReason describes why a session was invalidated.
type InvalidationReason string

const (
	ReasonLogout           InvalidationReason = "logout"
	ReasonReputationDrop   InvalidationReason = "reputation_drop"
	ReasonRepeatedFailures InvalidationReason = "repeated_failures"
	ReasonManualRevoke     InvalidationReason = "manual_revoke"
	ReasonSecurityEvent    InvalidationReason = "security_event"
)

// InvalidationRecord stores details about a session invalidation.
type InvalidationRecord struct {
	PeerDID       string             `json:"peerDid"`
	Reason        InvalidationReason `json:"reason"`
	InvalidatedAt time.Time          `json:"invalidatedAt"`
}

// Session represents an authenticated peer session.
type Session struct {
	PeerDID           string             `json:"peerDid"`
	Token             string             `json:"token"`
	CreatedAt         time.Time          `json:"createdAt"`
	ExpiresAt         time.Time          `json:"expiresAt"`
	ZKVerified        bool               `json:"zkVerified"`
	Invalidated       bool               `json:"invalidated"`
	InvalidatedReason InvalidationReason `json:"invalidatedReason,omitempty"`
}

// IsExpired reports whether the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionStore manages authenticated peer sessions with TTL eviction.
type SessionStore struct {
	mu                  sync.RWMutex
	sessions            map[string]*Session // keyed by peer DID
	hmacKey             []byte
	ttl                 time.Duration
	invalidationHistory []InvalidationRecord
	onInvalidate        func(peerDID string, reason InvalidationReason)
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

	if !ok || sess.IsExpired() || sess.Invalidated {
		if ok {
			s.Remove(peerDID)
		}
		return false
	}

	return sess.Token == token
}

// Get returns the session for the given peer DID, or nil if not found/expired/invalidated.
func (s *SessionStore) Get(peerDID string) *Session {
	s.mu.RLock()
	sess, ok := s.sessions[peerDID]
	s.mu.RUnlock()

	if !ok {
		return nil
	}
	if sess.IsExpired() || sess.Invalidated {
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

// ActiveSessions returns all non-expired, non-invalidated sessions.
func (s *SessionStore) ActiveSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var active []*Session
	for _, sess := range s.sessions {
		if !sess.IsExpired() && !sess.Invalidated {
			active = append(active, sess)
		}
	}
	return active
}

// Cleanup removes all expired and invalidated sessions.
func (s *SessionStore) Cleanup() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	for did, sess := range s.sessions {
		if sess.IsExpired() || sess.Invalidated {
			delete(s.sessions, did)
			removed++
		}
	}
	return removed
}

// Invalidate marks a session as invalidated, removes it from active sessions,
// records the invalidation, and fires the onInvalidate callback.
func (s *SessionStore) Invalidate(peerDID string, reason InvalidationReason) {
	s.mu.Lock()
	sess, ok := s.sessions[peerDID]
	if ok {
		sess.Invalidated = true
		sess.InvalidatedReason = reason
		delete(s.sessions, peerDID)
	}
	s.invalidationHistory = append(s.invalidationHistory, InvalidationRecord{
		PeerDID:       peerDID,
		Reason:        reason,
		InvalidatedAt: time.Now(),
	})
	cb := s.onInvalidate
	s.mu.Unlock()

	if cb != nil {
		cb(peerDID, reason)
	}
}

// InvalidateAll invalidates all active sessions with the given reason.
func (s *SessionStore) InvalidateAll(reason InvalidationReason) {
	s.mu.Lock()
	now := time.Now()
	var dids []string
	for did, sess := range s.sessions {
		sess.Invalidated = true
		sess.InvalidatedReason = reason
		dids = append(dids, did)
		s.invalidationHistory = append(s.invalidationHistory, InvalidationRecord{
			PeerDID:       did,
			Reason:        reason,
			InvalidatedAt: now,
		})
	}
	for _, did := range dids {
		delete(s.sessions, did)
	}
	cb := s.onInvalidate
	s.mu.Unlock()

	if cb != nil {
		for _, did := range dids {
			cb(did, reason)
		}
	}
}

// InvalidateByCondition invalidates sessions matching the predicate.
func (s *SessionStore) InvalidateByCondition(reason InvalidationReason, predicate func(*Session) bool) {
	s.mu.Lock()
	now := time.Now()
	var dids []string
	for did, sess := range s.sessions {
		if predicate(sess) {
			sess.Invalidated = true
			sess.InvalidatedReason = reason
			dids = append(dids, did)
			s.invalidationHistory = append(s.invalidationHistory, InvalidationRecord{
				PeerDID:       did,
				Reason:        reason,
				InvalidatedAt: now,
			})
		}
	}
	for _, did := range dids {
		delete(s.sessions, did)
	}
	cb := s.onInvalidate
	s.mu.Unlock()

	if cb != nil {
		for _, did := range dids {
			cb(did, reason)
		}
	}
}

// InvalidationHistory returns all recorded invalidation events.
func (s *SessionStore) InvalidationHistory() []InvalidationRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]InvalidationRecord, len(s.invalidationHistory))
	copy(history, s.invalidationHistory)
	return history
}

// SetInvalidationCallback sets a function to be called when a session is invalidated.
func (s *SessionStore) SetInvalidationCallback(fn func(peerDID string, reason InvalidationReason)) {
	s.mu.Lock()
	s.onInvalidate = fn
	s.mu.Unlock()
}
