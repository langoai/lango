package browser

import (
	"fmt"
	"sync"
)

// SessionManager provides implicit session management.
// It auto-creates a session on first use and reuses it for subsequent calls.
type SessionManager struct {
	tool      *Tool
	sessionID string
	mu        sync.Mutex
}

// NewSessionManager creates a SessionManager that wraps the given Tool.
func NewSessionManager(tool *Tool) *SessionManager {
	return &SessionManager{tool: tool}
}

// EnsureSession returns the active session ID, creating one if none exists.
func (sm *SessionManager) EnsureSession() (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.sessionID != "" && sm.tool.HasSession(sm.sessionID) {
		return sm.sessionID, nil
	}

	id, err := sm.tool.NewSession()
	if err != nil {
		return "", fmt.Errorf("create browser session: %w", err)
	}
	sm.sessionID = id
	return id, nil
}

// Tool returns the underlying browser Tool.
func (sm *SessionManager) Tool() *Tool {
	return sm.tool
}

// Close closes the underlying browser tool and clears the session.
func (sm *SessionManager) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessionID = ""
	return sm.tool.Close()
}
