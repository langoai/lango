package chat

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/tui"
)

// Pre-allocated styles for thinking block rendering.
var (
	thinkingLabelStyle   = lipgloss.NewStyle().Bold(true).Foreground(tui.Muted)
	thinkingPreviewStyle = lipgloss.NewStyle().Foreground(tui.Muted).Italic(true)
	thinkingDoneStyle    = lipgloss.NewStyle().Foreground(tui.Muted)
	pendingLabelStyle    = lipgloss.NewStyle().Bold(true).Foreground(tui.Muted)
)

// renderThinkingBlock renders a thinking/reasoning transcript item.
// Active state shows a spinner; done state shows duration in a compact line.
func renderThinkingBlock(content, state, duration string, width int) string {
	if width < 10 {
		width = 10
	}

	switch state {
	case "active":
		label := thinkingLabelStyle.Render("\U0001F4AD Thinking...")
		line := " " + label
		if content != "" {
			maxPreview := width - lipgloss.Width(line) - 2
			if maxPreview < 1 {
				maxPreview = 1
			}
			preview := ansi.Truncate(sanitizeDisplayText(content), maxPreview, "…")
			line += "  " + thinkingPreviewStyle.Render(preview)
		}
		return ansi.Truncate(line, width, "…")

	case "done":
		label := thinkingDoneStyle.Render(fmt.Sprintf("\U0001F4AD Thinking (%s)", duration))
		return ansi.Truncate(" "+label, width, "…")

	default:
		line := fmt.Sprintf(" \U0001F4AD %s", sanitizeDisplayText(content))
		return ansi.Truncate(line, width, "…")
	}
}

// renderPendingIndicator renders the submit-to-first-event waiting indicator.
func renderPendingIndicator(elapsed string, width int) string {
	if width < 10 {
		width = 10
	}
	label := pendingLabelStyle.Render(fmt.Sprintf("\u23F3 Working... (%s)", elapsed))
	return ansi.Truncate(fmt.Sprintf(" %s", label), width, "…")
}
