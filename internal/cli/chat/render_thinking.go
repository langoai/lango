package chat

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// renderThinkingBlock renders a thinking/reasoning transcript item.
// Active state shows a spinner; done state shows duration in a compact line.
func renderThinkingBlock(content, state, duration string, width int) string {
	switch state {
	case "active":
		label := lipgloss.NewStyle().
			Bold(true).
			Foreground(tui.Muted).
			Render("\U0001F4AD Thinking...")
		return fmt.Sprintf(" %s", label)

	case "done":
		label := lipgloss.NewStyle().
			Foreground(tui.Muted).
			Render(fmt.Sprintf("\U0001F4AD Thinking (%s)", duration))
		return fmt.Sprintf(" %s", label)

	default:
		return fmt.Sprintf(" \U0001F4AD %s", content)
	}
}

// renderPendingIndicator renders the submit-to-first-event waiting indicator.
func renderPendingIndicator(elapsed string) string {
	label := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Muted).
		Render(fmt.Sprintf("\u23F3 Working... (%s)", elapsed))
	return fmt.Sprintf(" %s", label)
}
