// Package theme provides the color palette, icons, and logo for the
// Lango Cockpit TUI.
package theme

import "github.com/charmbracelet/lipgloss"

// Surface colors (dark mode, deepest to highest).
var (
	Surface0 = lipgloss.Color("#111827")
	Surface1 = lipgloss.Color("#1F2937")
	Surface2 = lipgloss.Color("#374151")
	Surface3 = lipgloss.Color("#4B5563")
)

// Text colors.
var (
	TextPrimary   = lipgloss.Color("#F9FAFB")
	TextSecondary = lipgloss.Color("#9CA3AF")
	TextTertiary  = lipgloss.Color("#6B7280")
)

// Border colors.
var (
	BorderFocused = lipgloss.Color("#7C3AED")
	BorderDefault = lipgloss.Color("#374151")
	BorderSubtle  = lipgloss.Color("#1F2937")
)

// Brand colors (match existing tui/styles.go).
var (
	Primary = lipgloss.Color("#7C3AED")
	Success = lipgloss.Color("#10B981")
	Warning = lipgloss.Color("#F59E0B")
	Error   = lipgloss.Color("#EF4444")
	Accent  = lipgloss.Color("#04B575")
	Muted   = lipgloss.Color("#6B7280")
)

// Semantic color aliases for state-driven rendering.
var (
	Danger    = Error     // destructive/high-risk operations
	Info      = lipgloss.Color("#3B82F6") // informational highlights
	Selection = Accent    // user selection/focus indicator
)

// Sidebar width constants.
const (
	SidebarFullWidth      = 20
	SidebarCollapsedWidth = 3
)

// Context panel width constant.
const ContextPanelWidth = 28
