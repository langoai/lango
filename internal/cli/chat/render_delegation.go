package chat

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/tui"
)

// renderDelegationBlock renders an agent delegation event in the transcript.
func renderDelegationBlock(from, to, reason string, width int) string {
	w := max(width, 1)

	icon := lipgloss.NewStyle().Foreground(tui.Muted).Render("\U0001F500")
	fromLabel := lipgloss.NewStyle().Foreground(tui.Highlight).Bold(true).Render(from)
	arrow := lipgloss.NewStyle().Foreground(tui.Muted).Render("\u2192")
	toLabel := lipgloss.NewStyle().Foreground(tui.Highlight).Bold(true).Render(to)

	base := fmt.Sprintf(" %s %s %s %s", icon, fromLabel, arrow, toLabel)

	if reason != "" {
		safe := ansi.Strip(reason)
		maxReason := w - lipgloss.Width(base) - 4
		if maxReason < 10 {
			maxReason = 10
		}
		reasonText := ansi.Truncate(safe, maxReason, "\u2026")
		base += "  " + lipgloss.NewStyle().Foreground(tui.Muted).Italic(true).Render(reasonText)
	}

	return base
}
