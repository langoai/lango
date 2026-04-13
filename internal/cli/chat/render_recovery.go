package chat

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// Pre-allocated styles for recovery block rendering.
var (
	recoveryHeaderStyle = lipgloss.NewStyle().Foreground(tui.Warning).Bold(true)
	recoveryDetailStyle = lipgloss.NewStyle().Foreground(tui.Muted)
)

// renderRecoveryBlock renders a recovery decision event in the transcript.
func renderRecoveryBlock(action, causeClass string, attempt int, backoff time.Duration, width int) string {
	_ = width // reserved for future layout constraints

	actionLabel := recoveryActionDisplayName(action)

	header := fmt.Sprintf(" \U0001F504 %s #%d", actionLabel, attempt)
	detail := fmt.Sprintf("(%s)", causeClass)
	if backoff > 0 {
		detail += fmt.Sprintf(" %s backoff", backoff.Truncate(time.Millisecond))
	}

	line := recoveryHeaderStyle.Render(header)
	line += "  " + recoveryDetailStyle.Render(detail)
	return line
}

// recoveryActionDisplayName maps RecoveryDecisionEvent.Action string to display name.
func recoveryActionDisplayName(action string) string {
	switch action {
	case "retry":
		return "Retry"
	case "retry_with_hint":
		return "Reroute"
	case "direct_answer":
		return "Direct Answer"
	case "escalate":
		return "Escalate"
	default:
		return "Recovery"
	}
}
