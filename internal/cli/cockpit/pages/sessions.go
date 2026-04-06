package pages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/session"
)

// sessionsKeyMap holds the key bindings for the sessions page.
type sessionsKeyMap struct {
	Up   key.Binding
	Down key.Binding
}

func defaultSessionsKeyMap() sessionsKeyMap {
	return sessionsKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "navigate up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "navigate down"),
		),
	}
}

// sessionsLoadedMsg carries the result of an async session list fetch.
type sessionsLoadedMsg struct {
	sessions []session.SessionSummary
	err      error
}

// SessionsPage displays sessions from Store.ListSessions() with
// cursor navigation. Each entry shows the session key and relative
// time since last update.
type SessionsPage struct {
	listFn   func(ctx context.Context) ([]session.SessionSummary, error)
	sessions []session.SessionSummary
	cursor   int
	keymap   sessionsKeyMap
	loadErr  error
	width    int
	height   int
}

// Reusable styles for the sessions page — allocated once, not per-render.
var (
	sessionsTitleStyle  = lipgloss.NewStyle().Foreground(theme.TextPrimary).Bold(true)
	sessionsDivStyle    = lipgloss.NewStyle().Foreground(theme.BorderDefault)
	sessionsErrStyle    = lipgloss.NewStyle().Foreground(theme.Error)
	sessionsEmptyStyle  = lipgloss.NewStyle().Foreground(theme.Muted)
	sessionsActiveStyle = lipgloss.NewStyle().Foreground(theme.TextPrimary).Bold(true)
	sessionsInactStyle  = lipgloss.NewStyle().Foreground(theme.Muted)
	sessionsTimeStyle   = lipgloss.NewStyle().Foreground(theme.TextSecondary)
	sessionsCurStyle    = lipgloss.NewStyle().Foreground(theme.Primary)
	sessionsPadStyle    = lipgloss.NewStyle().Padding(1, 2)
)

// NewSessionsPage creates a SessionsPage. listFn is called on Activate
// to populate the session list.
func NewSessionsPage(
	listFn func(ctx context.Context) ([]session.SessionSummary, error),
) *SessionsPage {
	return &SessionsPage{
		listFn: listFn,
		keymap: defaultSessionsKeyMap(),
	}
}

// Title returns the page tab label.
func (p *SessionsPage) Title() string { return "Sessions" }

// ShortHelp returns context-sensitive keybindings for the help bar.
func (p *SessionsPage) ShortHelp() []key.Binding {
	return []key.Binding{p.keymap.Up, p.keymap.Down}
}

// Init satisfies tea.Model. No initial command is needed.
func (p *SessionsPage) Init() tea.Cmd { return nil }

// Activate is called when the page becomes active.
// It fires an async command to load sessions.
func (p *SessionsPage) Activate() tea.Cmd {
	return p.loadSessions()
}

// Deactivate is called when the page loses focus.
func (p *SessionsPage) Deactivate() {}

// Update handles messages.
func (p *SessionsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height

	case sessionsLoadedMsg:
		if msg.err != nil {
			p.loadErr = msg.err
			p.sessions = nil
		} else {
			p.loadErr = nil
			p.sessions = msg.sessions
		}
		if p.cursor >= len(p.sessions) {
			p.cursor = max(0, len(p.sessions)-1)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keymap.Up):
			if p.cursor > 0 {
				p.cursor--
			}
		case key.Matches(msg, p.keymap.Down):
			if p.cursor < len(p.sessions)-1 {
				p.cursor++
			}
		}
	}
	return p, nil
}

// View renders the session list.
func (p *SessionsPage) View() string {
	title := sessionsTitleStyle.Render("Sessions")
	divider := sessionsDivStyle.Render(strings.Repeat("-", 40))

	var body string
	switch {
	case p.loadErr != nil:
		body = sessionsErrStyle.Render(fmt.Sprintf("  Error: %v", p.loadErr))
	case len(p.sessions) == 0:
		body = sessionsEmptyStyle.Render("  No sessions found.")
	default:
		body = p.renderList()
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, divider, "", body)
	return sessionsPadStyle.Render(content)
}

// renderList builds the cursor-navigable session list.
func (p *SessionsPage) renderList() string {
	keyWidth := p.width - 20
	if keyWidth < 16 {
		keyWidth = 16
	}

	lines := make([]string, 0, len(p.sessions))
	for i, s := range p.sessions {
		keyText := tui.Truncate(s.Key, keyWidth)
		relTime := tui.RelativeTimeHuman(time.Now(), s.UpdatedAt)

		var line string
		if i == p.cursor {
			line = sessionsCurStyle.Render("> ") +
				sessionsActiveStyle.Render(keyText) +
				"  " +
				sessionsTimeStyle.Render(relTime)
		} else {
			line = "  " +
				sessionsInactStyle.Render(keyText) +
				"  " +
				sessionsTimeStyle.Render(relTime)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// loadSessions returns a tea.Cmd that fetches sessions from the list function.
func (p *SessionsPage) loadSessions() tea.Cmd {
	listFn := p.listFn
	return func() tea.Msg {
		if listFn == nil {
			return sessionsLoadedMsg{err: fmt.Errorf("session list function not configured")}
		}
		sessions, err := listFn(context.Background())
		return sessionsLoadedMsg{sessions: sessions, err: err}
	}
}
