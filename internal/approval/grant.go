package approval

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// GrantInfo represents a single active grant for display purposes.
type GrantInfo struct {
	SessionKey string
	ToolName   string
	GrantedAt  time.Time
}

// grantEntry tracks when a grant was created for TTL expiration.
type grantEntry struct {
	grantedAt time.Time
}

// GrantStore tracks per-session, per-tool "always allow" grants in memory.
// Grants are cleared on application restart (no persistence).
// An optional TTL causes grants to expire automatically.
type GrantStore struct {
	mu     sync.RWMutex
	grants map[string]grantEntry // key = "sessionKey\x00toolName"
	ttl    time.Duration         // 0 = no expiry (backward compatible default)
	nowFn  func() time.Time      // for testing; defaults to time.Now
}

// NewGrantStore creates an empty GrantStore with no TTL.
func NewGrantStore() *GrantStore {
	return &GrantStore{
		grants: make(map[string]grantEntry),
		nowFn:  time.Now,
	}
}

// SetTTL sets the time-to-live for grants. Zero disables expiry.
func (s *GrantStore) SetTTL(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ttl = ttl
}

func grantKey(sessionKey, toolName string) string {
	return sessionKey + "\x00" + toolName
}

// Grant records an approval for the given session and tool.
func (s *GrantStore) Grant(sessionKey, toolName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grants[grantKey(sessionKey, toolName)] = grantEntry{grantedAt: s.nowFn()}
}

// IsGranted reports whether the tool has a valid (non-expired) grant.
func (s *GrantStore) IsGranted(sessionKey, toolName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.grants[grantKey(sessionKey, toolName)]
	if !ok {
		return false
	}
	if s.ttl > 0 && s.nowFn().Sub(entry.grantedAt) > s.ttl {
		return false
	}
	return true
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

// cleanExpiredLocked removes all grants that have exceeded the TTL.
// Returns the number of entries removed. No-op when TTL is zero.
// Caller must hold s.mu (write lock).
func (s *GrantStore) cleanExpiredLocked() int {
	if s.ttl == 0 {
		return 0
	}

	now := s.nowFn()
	removed := 0
	for k, entry := range s.grants {
		if now.Sub(entry.grantedAt) > s.ttl {
			delete(s.grants, k)
			removed++
		}
	}
	return removed
}

// CleanExpired removes all grants that have exceeded the TTL.
// Returns the number of entries removed. No-op when TTL is zero.
func (s *GrantStore) CleanExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cleanExpiredLocked()
}

// List returns all active (non-expired) grants sorted by session key then tool name.
// Expired grants are lazily cleaned before listing.
func (s *GrantStore) List() []GrantInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanExpiredLocked()

	result := make([]GrantInfo, 0, len(s.grants))
	for key, entry := range s.grants {
		parts := strings.SplitN(key, "\x00", 2)
		if len(parts) != 2 {
			continue
		}
		result = append(result, GrantInfo{
			SessionKey: parts[0],
			ToolName:   parts[1],
			GrantedAt:  entry.grantedAt,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SessionKey != result[j].SessionKey {
			return result[i].SessionKey < result[j].SessionKey
		}
		return result[i].ToolName < result[j].ToolName
	})
	return result
}
