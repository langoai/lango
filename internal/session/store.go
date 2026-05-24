package session

import (
	"context"
	"time"

	"github.com/langoai/lango/internal/types"
)

// SessionSummary is a lightweight view of a session for listing purposes.
type SessionSummary struct {
	Key       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Message represents a single message in conversation history
type Message struct {
	Role      types.MessageRole `json:"role"` // "user", "assistant", "tool"
	Content   string            `json:"content"`
	Timestamp time.Time         `json:"timestamp"`
	ToolCalls []ToolCall        `json:"toolCalls,omitempty"`
	Author    string            `json:"author,omitempty"` // ADK agent name for multi-agent routing
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Input            string `json:"input"`
	Output           string `json:"output,omitempty"`
	Thought          bool   `json:"thought,omitempty"`
	ThoughtSignature []byte `json:"thoughtSignature,omitempty"`
}

// Session represents a conversation session
type Session struct {
	Key         string            `json:"key"`
	AgentID     string            `json:"agentId,omitempty"`
	ChannelType string            `json:"channelType,omitempty"`
	ChannelID   string            `json:"channelId,omitempty"`
	History     []Message         `json:"history"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Model       string            `json:"model,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// MetadataKeyMode is the reserved key used to persist the session's active
// mode name inside Session.Metadata.
const MetadataKeyMode = "lango.mode"

// MetadataKeyEndPending marks a session that has a pending session-end
// processor invocation (recall indexing). Soft-end (channel idle) sets this
// flag synchronously and defers processing to the next session-open sweep.
// Hard-end (TUI quit) invokes the processor directly but still marks the
// flag first so a crash mid-processing can be recovered on next open.
const MetadataKeyEndPending = "lango.session_end_pending"

// MetadataValueTrue is the canonical string representation of boolean true
// in Session.Metadata (stored as map[string]string).
const MetadataValueTrue = "true"

// EndPending reports whether the session is marked for pending session-end
// processing (soft-end or crashed hard-end).
func (s *Session) EndPending() bool {
	if s == nil || s.Metadata == nil {
		return false
	}
	return s.Metadata[MetadataKeyEndPending] == MetadataValueTrue
}

// Mode returns the active mode name persisted in the session's metadata.
// Returns an empty string if no mode is set.
func (s *Session) Mode() string {
	if s == nil || s.Metadata == nil {
		return ""
	}
	return s.Metadata[MetadataKeyMode]
}

// SetMode persists the given mode name in the session's metadata.
// Passing an empty string clears the mode.
func (s *Session) SetMode(name string) {
	if s == nil {
		return
	}
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}
	if name == "" {
		delete(s.Metadata, MetadataKeyMode)
		return
	}
	s.Metadata[MetadataKeyMode] = name
}

// Store defines the interface for session storage
type Store interface {
	// Create creates a new session
	Create(session *Session) error
	// Get retrieves a session by key
	Get(key string) (*Session, error)
	// Update updates an existing session
	Update(session *Session) error
	// Delete removes a session
	Delete(key string) error
	// AppendMessage adds a message to session history
	AppendMessage(key string, msg Message) error
	// AnnotateTimeout appends a synthetic assistant message to indicate that the
	// previous turn was interrupted by a timeout. This prevents incomplete history
	// from leaking into subsequent turns.
	// partial is any partial response text accumulated before the timeout.
	AnnotateTimeout(key string, partial string) error
	// End marks the session as ended. The metadata key
	// MetadataKeyEndPending is set to MetadataValueTrue; the configured
	// session-end processor (if any) is invoked with the concrete store's
	// own timeout semantics. Calling End on an already-ended session is a
	// no-op. Calling End on an unknown session returns an error.
	End(key string) error
	// Close closes the store
	Close() error

	// ListSessions returns lightweight summaries of all sessions,
	// ordered by most recent update first.
	ListSessions(ctx context.Context) ([]SessionSummary, error)

	// GetSalt retrieves the encryption salt for LocalCryptoProvider
	GetSalt(name string) ([]byte, error)
	// SetSalt stores the encryption salt for LocalCryptoProvider
	SetSalt(name string, salt []byte) error
}
