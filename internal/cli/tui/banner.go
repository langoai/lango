package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Package-level version info, set by main.go via SetVersionInfo.
var (
	_version   = "dev"
	_buildTime = "unknown"
	_profile   = "default"
)

// SetVersionInfo injects version and build time from main.go.
func SetVersionInfo(version, buildTime string) {
	_version = version
	_buildTime = buildTime
}

// SetProfile injects the active profile name.
func SetProfile(name string) {
	_profile = name
}

// squirrelFace returns the squirrel mascot ASCII art lines.
func squirrelFace() string {
	return " ▄▀▄▄▄▀▄\n ▜ ●.● ▛\n  ▜▄▄▄▛"
}

// Banner returns the squirrel mascot with brand info side-by-side.
func Banner() string {
	artStyle := lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)

	infoLines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(Foreground).Render(fmt.Sprintf("Lango v%s", _version)),
		MutedStyle.Render("Fast AI Agent in Go"),
		MutedStyle.Render(fmt.Sprintf("profile: %s", _profile)),
	}

	art := artStyle.Render(squirrelFace())
	info := strings.Join(infoLines, "\n")

	// Add padding between art and info
	infoBlock := lipgloss.NewStyle().PaddingLeft(4).Render(info)

	return lipgloss.JoinHorizontal(lipgloss.Top, art, infoBlock)
}

// BannerBox wraps the Banner in a rounded border box (for settings welcome).
func BannerBox() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(1, 3)

	return box.Render(Banner())
}

// ServeBanner returns a banner for the serve command with a separator line.
func ServeBanner() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(Banner())
	b.WriteString("\n")

	sep := lipgloss.NewStyle().Foreground(Separator).Render(strings.Repeat("─", 48))
	b.WriteString(sep)
	b.WriteString("\n\n")

	return b.String()
}
