package session

import (
	"fmt"
	"time"
)

// ChildSessionConfig configures how a child session behaves.
type ChildSessionConfig struct {
	// MaxMessages limits the child session's message history.
	// Zero means unlimited (inherits parent limit).
	MaxMessages int

	// InheritHistory copies the last N messages from parent.
	// Zero means start with empty history.
	InheritHistory int

	// SummarizeOnMerge applies a summarizer when merging back to parent.
	SummarizeOnMerge bool
}

// ChildSession represents an isolated sub-session forked from a parent.
// It follows "read parent, write child" semantics: the child session
// has its own message history that does not pollute the parent until
// explicitly merged.
type ChildSession struct {
	// Key is the unique identifier for this child session.
	Key string

	// ParentKey is the key of the parent session.
	ParentKey string

	// AgentName is the sub-agent that owns this child session.
	AgentName string

	// History contains messages added during this child session.
	History []Message

	// Config holds the child session settings.
	Config ChildSessionConfig

	// CreatedAt is when this child session was forked.
	CreatedAt time.Time

	// MergedAt is set when the child session is merged back to parent.
	// Zero value means not yet merged.
	MergedAt time.Time
}

// NewChildSession creates a new child session forked from a parent.
func NewChildSession(parentKey, agentName string, cfg ChildSessionConfig) *ChildSession {
	return &ChildSession{
		Key:       fmt.Sprintf("%s:child:%s:%d", parentKey, agentName, time.Now().UnixNano()),
		ParentKey: parentKey,
		AgentName: agentName,
		Config:    cfg,
		CreatedAt: time.Now(),
	}
}

// AppendMessage adds a message to the child session's history.
func (cs *ChildSession) AppendMessage(msg Message) {
	cs.History = append(cs.History, msg)

	// Enforce max messages limit if configured.
	if cs.Config.MaxMessages > 0 && len(cs.History) > cs.Config.MaxMessages {
		cs.History = cs.History[len(cs.History)-cs.Config.MaxMessages:]
	}
}

// IsMerged returns true if this child session has been merged back to parent.
func (cs *ChildSession) IsMerged() bool {
	return !cs.MergedAt.IsZero()
}

// ChildSessionStore extends Store with child session operations.
type ChildSessionStore interface {
	// ForkChild creates a new child session from a parent session.
	// If cfg.InheritHistory > 0, the last N messages are copied from parent.
	ForkChild(parentKey, agentName string, cfg ChildSessionConfig) (*ChildSession, error)

	// MergeChild merges a child session's messages back into the parent.
	// The summary parameter, if non-empty, replaces the full child history
	// with a single assistant message containing the summary.
	MergeChild(childKey string, summary string) error

	// DiscardChild removes a child session without merging.
	DiscardChild(childKey string) error

	// GetChild retrieves a child session by key.
	GetChild(childKey string) (*ChildSession, error)

	// ChildrenOf returns all child sessions for a parent.
	ChildrenOf(parentKey string) ([]*ChildSession, error)
}
