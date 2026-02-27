// Package tui provides shared TUI components for Lango CLI commands.
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette for consistent theming
var (
	Primary    = lipgloss.Color("#7C3AED") // Purple
	Success    = lipgloss.Color("#10B981") // Green
	Warning    = lipgloss.Color("#F59E0B") // Amber
	Error      = lipgloss.Color("#EF4444") // Red
	Muted      = lipgloss.Color("#6B7280") // Gray
	Foreground = lipgloss.Color("#F9FAFB") // White
	Background = lipgloss.Color("#1F2937") // Dark gray
	Highlight  = lipgloss.Color("#3B82F6") // Blue
	Accent     = lipgloss.Color("#04B575") // Green (selection/focus)
	Dim        = lipgloss.Color("#626262") // Dim gray (descriptions)
	Separator  = lipgloss.Color("#374151") // Dark gray (dividers)
)

// Base styles for TUI components
var (
	// TitleStyle for main headers
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	// SubtitleStyle for secondary headers
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginBottom(1)

	// SuccessStyle for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	// WarningStyle for warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error)

	// MutedStyle for less important text
	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// HighlightStyle for emphasized text
	HighlightStyle = lipgloss.NewStyle().
			Foreground(Highlight).
			Bold(true)

	// BoxStyle for bordered containers
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted).
			Padding(1, 2)

	// ListItemStyle for list items
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// SelectedItemStyle for selected list items
	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(Primary).
				Bold(true)

	// SectionHeaderStyle for menu section titles
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(Highlight).
				Bold(true).
				PaddingLeft(2)

	// SeparatorLineStyle for section dividers
	SeparatorLineStyle = lipgloss.NewStyle().
				Foreground(Separator)

	// CursorStyle for the selection arrow
	CursorStyle = lipgloss.NewStyle().
			Foreground(Accent)

	// ActiveItemStyle for highlighted/selected items
	ActiveItemStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	// SearchBarStyle for the search input container
	SearchBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	// FormTitleBarStyle for form titles
	FormTitleBarStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Primary).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(Primary).
				MarginBottom(1)

	// FieldDescStyle for field description/help text
	FieldDescStyle = lipgloss.NewStyle().
				Foreground(Dim).
				Italic(true).
				PaddingLeft(2)
)

// Check result indicators
const (
	CheckPass = "✓"
	CheckWarn = "⚠"
	CheckFail = "✗"
	Spinner   = "◌"
)

// FormatPass formats a passing check message
func FormatPass(msg string) string {
	return SuccessStyle.Render(CheckPass) + " " + msg
}

// FormatWarn formats a warning message
func FormatWarn(msg string) string {
	return WarningStyle.Render(CheckWarn) + " " + msg
}

// FormatFail formats a failing check message
func FormatFail(msg string) string {
	return ErrorStyle.Render(CheckFail) + " " + msg
}

// FormatMuted formats muted/hint text
func FormatMuted(msg string) string {
	return MutedStyle.Render(msg)
}

// KeyBadge renders a keyboard shortcut as a styled badge.
func KeyBadge(key string) string {
	badge := lipgloss.NewStyle().
		Foreground(Foreground).
		Background(Separator).
		Bold(true).
		Padding(0, 1)
	return badge.Render(key)
}

// HelpEntry renders a single help entry: key badge + label.
func HelpEntry(key, label string) string {
	return KeyBadge(key) + " " + lipgloss.NewStyle().Foreground(Dim).Render(label)
}

// HelpBar renders a full help footer from HelpEntry results.
func HelpBar(entries ...string) string {
	return lipgloss.NewStyle().Foreground(Dim).Render(strings.Join(entries, "  "))
}

// Breadcrumb renders a navigation path like "Settings > Agent Configuration".
func Breadcrumb(segments ...string) string {
	if len(segments) == 0 {
		return ""
	}
	sep := lipgloss.NewStyle().Foreground(Dim).Render(" > ")
	var parts []string
	for i, s := range segments {
		if i == len(segments)-1 {
			parts = append(parts, lipgloss.NewStyle().Foreground(Primary).Bold(true).Render(s))
		} else {
			parts = append(parts, lipgloss.NewStyle().Foreground(Muted).Render(s))
		}
	}
	return strings.Join(parts, sep)
}
