package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/tui"
)

// channelColor returns a color for the channel badge.
func channelColor(channel string) lipgloss.Color {
	switch channel {
	case "telegram":
		return lipgloss.Color("#0088cc") // Telegram blue
	case "discord":
		return lipgloss.Color("#5865F2") // Discord blurple
	case "slack":
		return lipgloss.Color("#4A154B") // Slack aubergine
	default:
		return tui.Muted
	}
}

// renderChannelBlock renders a channel message in the transcript.
func renderChannelBlock(text, channel, senderName string, width int) string {
	w := max(width, 1)

	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Foreground).
		Background(channelColor(channel)).
		Padding(0, 1).
		Render(channel)

	sender := ""
	if senderName != "" {
		sender = lipgloss.NewStyle().
			Foreground(tui.Highlight).
			Bold(true).
			Render("@" + senderName)
	}

	// Sanitize external channel input: strip ANSI/OSC escape sequences
	// to prevent terminal control injection from remote users, then
	// collapse newlines for single-line display.
	safe := ansi.Strip(text)
	flat := strings.ReplaceAll(safe, "\n", " ")
	maxText := w - lipgloss.Width(badge) - lipgloss.Width(sender) - 6
	if maxText < 10 {
		maxText = 10
	}
	displayText := ansi.Truncate(flat, maxText, "…")

	content := fmt.Sprintf(" %s", badge)
	if sender != "" {
		content += fmt.Sprintf("  %s:", sender)
	}
	content += fmt.Sprintf(" %s", lipgloss.NewStyle().Foreground(tui.Foreground).Render(displayText))

	return content
}
