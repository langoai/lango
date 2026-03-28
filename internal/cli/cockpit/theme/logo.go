package theme

import "github.com/charmbracelet/lipgloss"

// RenderLogo returns the squirrel mascot ASCII art with color styling.
func RenderLogo() string {
	body := lipgloss.NewStyle().Foreground(Primary)
	eyes := lipgloss.NewStyle().Foreground(TextPrimary)

	return body.Render("▄▀▄▄▄▀▄") + "\n" +
		body.Render("▜ ") + eyes.Render("●.●") + body.Render(" ▛") + "\n" +
		body.Render(" ▜▄▄▄▛")
}
