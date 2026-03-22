// Package automation provides shared interfaces for the automation subsystem
// (cron, background, workflow). Domain packages import these interfaces instead
// of defining their own, eliminating duplication and ensuring contract consistency.
package automation

import (
	"context"
	"strings"

	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

// AgentRunner executes agent prompts. The app layer provides a concrete
// implementation that delegates to the orchestration pipeline; automation
// subsystems depend only on this interface to avoid import cycles.
type AgentRunner interface {
	Run(ctx context.Context, sessionKey string, prompt string) (string, error)
}

// ChannelSender sends results to communication channels (e.g. Telegram, Discord, Slack).
type ChannelSender interface {
	SendMessage(ctx context.Context, channel string, message string) error
}

// DetectChannelFromContext extracts the delivery target from the session key in context.
// Returns "channel:targetID" (e.g. "telegram:123456789") or "" if no known channel prefix is found.
func DetectChannelFromContext(ctx context.Context) string {
	sessionKey := session.SessionKeyFromContext(ctx)
	if sessionKey == "" {
		return ""
	}
	// Session key format: "channel:targetID:userID"
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 2 {
		return ""
	}
	ch := types.ChannelType(parts[0])
	if ch.Valid() {
		return parts[0] + ":" + parts[1]
	}
	return ""
}
