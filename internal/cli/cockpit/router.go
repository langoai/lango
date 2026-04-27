package cockpit

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/cockpit/sidebar"
	"github.com/langoai/lango/internal/cli/cockpit/theme"
)

// PageID identifies a cockpit page.
type PageID int

const (
	PageChat PageID = iota
	PageSettings
	PageTools
	PageStatus
	PageSessions
	PageTasks
	PageDeadLetters
	PageApprovals
)

// String returns the page name for sidebar matching.
func (p PageID) String() string {
	switch p {
	case PageChat:
		return "chat"
	case PageSettings:
		return "settings"
	case PageTools:
		return "tools"
	case PageStatus:
		return "status"
	case PageSessions:
		return "sessions"
	case PageTasks:
		return "tasks"
	case PageDeadLetters:
		return "dead-letters"
	case PageApprovals:
		return "approvals"
	default:
		return "unknown"
	}
}

// Page is the interface for cockpit pages (non-chat).
// ChatModel uses the separate childModel interface.
type Page interface {
	tea.Model

	// Title returns the page display name for the sidebar.
	Title() string

	// ShortHelp returns context-sensitive keybindings for the help bar.
	ShortHelp() []key.Binding

	// Activate is called when the page becomes active.
	// Returns a tea.Cmd to execute (e.g., start a tick timer).
	Activate() tea.Cmd

	// Deactivate is called when the page loses focus.
	// Used to stop timers or release resources.
	Deactivate()
}

// AllPageMetas returns the sidebar menu items for all known pages.
// The order matches the sidebar display order.
func AllPageMetas() []sidebar.MenuItem {
	return []sidebar.MenuItem{
		{ID: PageChat.String(), Icon: theme.IconChat, Label: "Chat"},
		{ID: PageSettings.String(), Icon: theme.IconSettings, Label: "Settings"},
		{ID: PageTools.String(), Icon: theme.IconTools, Label: "Tools"},
		{ID: PageStatus.String(), Icon: theme.IconStatus, Label: "Status"},
		{ID: PageSessions.String(), Icon: theme.IconSessions, Label: "Sessions"},
		{ID: PageTasks.String(), Icon: theme.IconStatus, Label: "Tasks"},
		{ID: PageDeadLetters.String(), Icon: theme.IconStatus, Label: "Dead Letters"},
		{ID: PageApprovals.String(), Icon: theme.IconApprovals, Label: "Approvals"},
	}
}

// PageIDFromString converts a sidebar item ID to a PageID.
// Returns PageChat for unknown IDs.
func PageIDFromString(id string) PageID {
	switch id {
	case "chat":
		return PageChat
	case "settings":
		return PageSettings
	case "tools":
		return PageTools
	case "status":
		return PageStatus
	case "sessions":
		return PageSessions
	case "tasks":
		return PageTasks
	case "dead-letters":
		return PageDeadLetters
	case "approvals":
		return PageApprovals
	default:
		return PageChat
	}
}
