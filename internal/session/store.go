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
