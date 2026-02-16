package approval

import (
	"context"

	"github.com/langowarny/lango/internal/logging"
)

// HeadlessProvider auto-approves all tool execution requests.
// Intended for headless (Docker) environments where no TTY or companion
// is available. Every approval is logged at WARN level for audit.
type HeadlessProvider struct{}

var _ Provider = (*HeadlessProvider)(nil)

// RequestApproval always approves and logs a warning for audit trail.
func (h *HeadlessProvider) RequestApproval(_ context.Context, req ApprovalRequest) (bool, error) {
	logging.App().Warnw("headless auto-approve",
		"tool", req.ToolName,
		"sessionKey", req.SessionKey,
		"requestID", req.ID,
		"summary", req.Summary,
	)
	return true, nil
}

// CanHandle always returns false. HeadlessProvider is used as a TTY
// fallback slot, not prefix-matched by session key.
func (h *HeadlessProvider) CanHandle(_ string) bool {
	return false
}
