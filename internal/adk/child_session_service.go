package adk

import (
	"context"

	"github.com/langoai/lango/internal/session"
)

// childSessionCtxKey is used to store child session info in context.
type childSessionCtxKey struct{}

// ChildSessionInfo holds child session metadata in context.
type ChildSessionInfo struct {
	ChildKey  string
	ParentKey string
	AgentName string
}

// WithChildSession stores child session info in context.
func WithChildSession(ctx context.Context, info ChildSessionInfo) context.Context {
	return context.WithValue(ctx, childSessionCtxKey{}, info)
}

// ChildSessionFromContext retrieves child session info from context.
func ChildSessionFromContext(ctx context.Context) (ChildSessionInfo, bool) {
	info, ok := ctx.Value(childSessionCtxKey{}).(ChildSessionInfo)
	return info, ok
}

// ChildSessionServiceAdapter wraps a ChildSessionStore to provide
// fork/merge/discard operations integrated with ADK's session management.
type ChildSessionServiceAdapter struct {
	childStore session.ChildSessionStore
	summarizer Summarizer
}

// NewChildSessionServiceAdapter creates a new adapter.
func NewChildSessionServiceAdapter(childStore session.ChildSessionStore, summarizer Summarizer) *ChildSessionServiceAdapter {
	if summarizer == nil {
		summarizer = &StructuredSummarizer{}
	}
	return &ChildSessionServiceAdapter{
		childStore: childStore,
		summarizer: summarizer,
	}
}

// Fork creates a child session for a sub-agent.
func (a *ChildSessionServiceAdapter) Fork(parentKey, agentName string, cfg session.ChildSessionConfig) (*session.ChildSession, error) {
	return a.childStore.ForkChild(parentKey, agentName, cfg)
}

// MergeWithSummary merges a child session using the configured summarizer.
func (a *ChildSessionServiceAdapter) MergeWithSummary(childKey string) error {
	child, err := a.childStore.GetChild(childKey)
	if err != nil {
		return err
	}

	summary, err := a.summarizer.Summarize(child.History)
	if err != nil {
		return err
	}

	return a.childStore.MergeChild(childKey, summary)
}

// Discard removes a child session without merging.
func (a *ChildSessionServiceAdapter) Discard(childKey string) error {
	return a.childStore.DiscardChild(childKey)
}
