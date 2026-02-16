package session

import (
	"time"
)

// Message represents a single message in conversation history
type Message struct {
	Role      string     `json:"role"` // "user", "assistant", "tool"
	Content   string     `json:"content"`
	Timestamp time.Time  `json:"timestamp"`
	ToolCalls []ToolCall `json:"toolCalls,omitempty"`
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output,omitempty"`
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
	// Close closes the store
	Close() error

	// GetSalt retrieves the encryption salt for LocalCryptoProvider
	GetSalt(name string) ([]byte, error)
	// SetSalt stores the encryption salt for LocalCryptoProvider
	SetSalt(name string, salt []byte) error
}

