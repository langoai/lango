// Package approval provides a unified interface for tool execution approval
// across multiple channels (Gateway WebSocket, Telegram, Discord, Slack, TTY).
package approval

import (
	"context"
	"time"
)

// ApprovalRequest represents a request for tool execution approval.
type ApprovalRequest struct {
	ID         string
	ToolName   string
	SessionKey string
	Params     map[string]interface{}
	Summary    string // Human-readable description of what the tool will do
	CreatedAt  time.Time
}

// ApprovalResponse carries the result of an approval request.
type ApprovalResponse struct {
	Approved    bool
	AlwaysAllow bool
}

// Provider defines the interface for approval request handling.
type Provider interface {
	// RequestApproval sends an approval request and blocks until approved/denied or context is cancelled.
	RequestApproval(ctx context.Context, req ApprovalRequest) (ApprovalResponse, error)

	// CanHandle reports whether this provider can handle the given session key.
	CanHandle(sessionKey string) bool
}
