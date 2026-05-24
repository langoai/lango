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
	if width < 10 {
		width = 10
	}

	icon := delegationIconStyle.Render("\U0001F500")
	fromLabel := delegationNameStyle.Render(sanitizeDisplayText(from))
	arrow := delegationArrowStyle.Render("\u2192")
	toLabel := delegationNameStyle.Render(sanitizeDisplayText(to))

	base := fmt.Sprintf(" %s %s %s %s", icon, fromLabel, arrow, toLabel)

	if reason != "" {
		safe := singleLineValue(ansi.Strip(reason))
		maxReason := width - lipgloss.Width(base) - 2
		if maxReason < 1 {
			maxReason = 1
		}
		reasonText := ansi.Truncate(safe, maxReason, "\u2026")
		base += "  " + delegationReasonStyle.Render(reasonText)
	}

	return ansi.Truncate(base, width, "…")
}
