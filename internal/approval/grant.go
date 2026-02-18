package approval

import (
	"strings"
	"sync"
)

// GrantStore tracks per-session, per-tool "always allow" grants in memory.
// Grants are cleared on application restart (no persistence).
type GrantStore struct {
	mu     sync.RWMutex
	grants map[string]struct{} // key = "sessionKey:toolName"
}

// NewGrantStore creates an empty GrantStore.
func NewGrantStore() *GrantStore {
	return &GrantStore{
		grants: make(map[string]struct{}),
	}
}

func grantKey(sessionKey, toolName string) string {
	return sessionKey + "\x00" + toolName
}

// Grant records a persistent approval for the given session and tool.
func (s *GrantStore) Grant(sessionKey, toolName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grants[grantKey(sessionKey, toolName)] = struct{}{}
}

// IsGranted reports whether the tool has been permanently approved for this session.
func (s *GrantStore) IsGranted(sessionKey, toolName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.grants[grantKey(sessionKey, toolName)]
	return ok
}

// Revoke removes a single tool grant for the given session.
func (s *GrantStore) Revoke(sessionKey, toolName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.grants, grantKey(sessionKey, toolName))
}

// RevokeSession removes all grants for the given session.
func (s *GrantStore) RevokeSession(sessionKey string) {
	prefix := sessionKey + "\x00"
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.grants {
		if strings.HasPrefix(k, prefix) {
			delete(s.grants, k)
		}
	}
}
