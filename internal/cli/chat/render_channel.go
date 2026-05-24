package chat

import (
	"fmt"

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

// Pre-allocated styles for channel block rendering.
var (
	channelBadgeStyle  = lipgloss.NewStyle().Bold(true).Foreground(tui.Foreground).Padding(0, 1)
	channelSenderStyle = lipgloss.NewStyle().Foreground(tui.Highlight).Bold(true)
	channelTextStyle   = lipgloss.NewStyle().Foreground(tui.Foreground)
)

// renderChannelBlock renders a channel message in the transcript.
func renderChannelBlock(text, channel, senderName string, width int) string {
	if width < 10 {
		width = 10
	}

	safeChannel := sanitizeDisplayText(channel)
	badge := channelBadgeStyle.Background(channelColor(safeChannel)).Render(safeChannel)

	sender := ""
	if senderName != "" {
		safeSender := singleLineValue(ansi.Strip(senderName))
		sender = channelSenderStyle.Render("@" + safeSender)
	}

	// Sanitize external channel input: strip ANSI/OSC escape sequences
	// to prevent terminal control injection from remote users, then
	// collapse whitespace for single-line display.
	safe := ansi.Strip(text)
	flat := singleLineValue(safe)
	prefix := fmt.Sprintf(" %s", badge)
	if sender != "" {
		prefix += fmt.Sprintf("  %s:", sender)
	}
	prefix += " "
	maxText := width - lipgloss.Width(prefix)
	if maxText < 1 {
		maxText = 1
	}
	displayText := ansi.Truncate(flat, maxText, "…")
	content := prefix + channelTextStyle.Render(displayText)
	return ansi.Truncate(content, width, "…")
}
