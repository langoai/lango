package chat

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/tui"
)

// Pre-allocated styles for delegation block rendering.
var (
	delegationIconStyle   = lipgloss.NewStyle().Foreground(tui.Muted)
	delegationNameStyle   = lipgloss.NewStyle().Foreground(tui.Highlight).Bold(true)
	delegationArrowStyle  = lipgloss.NewStyle().Foreground(tui.Muted)
	delegationReasonStyle = lipgloss.NewStyle().Foreground(tui.Muted).Italic(true)
)

// renderDelegationBlock renders an agent delegation event in the transcript.
func renderDelegationBlock(from, to, reason string, width int) string {
	w := max(width, 1)

	icon := delegationIconStyle.Render("\U0001F500")
	fromLabel := delegationNameStyle.Render(from)
	arrow := delegationArrowStyle.Render("\u2192")
	toLabel := delegationNameStyle.Render(to)

	base := fmt.Sprintf(" %s %s %s %s", icon, fromLabel, arrow, toLabel)

	if reason != "" {
		safe := ansi.Strip(reason)
		maxReason := w - lipgloss.Width(base) - 4
		if maxReason < 10 {
			maxReason = 10
		}
		reasonText := ansi.Truncate(safe, maxReason, "\u2026")
		base += "  " + delegationReasonStyle.Render(reasonText)
	}

	return base
}
