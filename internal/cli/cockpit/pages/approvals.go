package pages

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/cockpit/theme"
)

// Column layout constants for the approvals table.
const (
	apprGutterW     = 6  // "> " or "  " prefix (2) + PaddingLeft(4)
	apprColTimeW    = 10 // Time column width
	apprColToolW    = 14 // Tool column width
	apprColOutcomeW = 12 // Outcome column width
	apprColProvW    = 10 // Provider column width
	apprColGapW     = 1  // space between columns

	grantColSessionW = 16 // Session column width
	grantColToolW    = 14 // Tool column width
	grantColTimeW    = 10 // Granted column width
)

// approvalTickMsg triggers periodic refresh of approval data.
type approvalTickMsg time.Time

func approvalTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return approvalTickMsg(t)
	})
}

// ApprovalsPage displays approval history and active grants.
type ApprovalsPage struct {
	history *approval.HistoryStore
	grants  *approval.GrantStore

	histEntries []approval.HistoryEntry
	grantList   []approval.GrantInfo

	section     int // 0 = history, 1 = grants
	cursor      int // history cursor
	grantCursor int // grants cursor (independent, preserved on tab switch)

	tickActive    bool
	width, height int

	nowFn func() time.Time // for testing; defaults to time.Now
}

// NewApprovalsPage creates a new ApprovalsPage. Both stores may be nil.
func NewApprovalsPage(history *approval.HistoryStore, grants *approval.GrantStore) *ApprovalsPage {
	return &ApprovalsPage{
		history: history,
		grants:  grants,
		nowFn:   time.Now,
	}
}

// Title returns the page tab label.
func (m *ApprovalsPage) Title() string { return "Approvals" }

// ShortHelp returns context-sensitive key bindings for the help bar.
func (m *ApprovalsPage) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "switch")),
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	}
	if m.section == 1 && m.grants != nil {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "revoke")),
			key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "revoke all")),
		)
	}
	return bindings
}

// Init satisfies tea.Model.
func (m *ApprovalsPage) Init() tea.Cmd { return nil }

// Activate starts periodic data refresh.
func (m *ApprovalsPage) Activate() tea.Cmd {
	m.tickActive = true
	m.refreshData()
	return approvalTickCmd()
}

// Deactivate stops the tick loop.
func (m *ApprovalsPage) Deactivate() {
	m.tickActive = false
}

// Update handles messages.
func (m *ApprovalsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case approvalTickMsg:
		if !m.tickActive {
			return m, nil
		}
		m.refreshData()
		return m, approvalTickCmd()
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *ApprovalsPage) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
		m.section = 1 - m.section
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.section == 0 {
			if m.cursor > 0 {
				m.cursor--
			}
		} else {
			if m.grantCursor > 0 {
				m.grantCursor--
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.section == 0 {
			if m.cursor < len(m.histEntries)-1 {
				m.cursor++
			}
		} else {
			if m.grantCursor < len(m.grantList)-1 {
				m.grantCursor++
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		if m.section == 1 && m.grants != nil && m.grantCursor < len(m.grantList) {
			g := m.grantList[m.grantCursor]
			m.grants.Revoke(g.SessionKey, g.ToolName)
			m.refreshData()
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("R"))):
		if m.section == 1 && m.grants != nil && m.grantCursor < len(m.grantList) {
			g := m.grantList[m.grantCursor]
			m.grants.RevokeSession(g.SessionKey)
			m.refreshData()
		}
	}
	return m, nil
}

// View renders the approvals page with history and grants sections.
func (m *ApprovalsPage) View() string {
	if m.history == nil && m.grants == nil {
		return lipgloss.NewStyle().
			Foreground(theme.TextSecondary).
			PaddingLeft(2).
			PaddingTop(1).
			Render("No approval history yet.")
	}

	if len(m.histEntries) == 0 && len(m.grantList) == 0 {
		return lipgloss.NewStyle().
			Foreground(theme.TextSecondary).
			PaddingLeft(2).
			PaddingTop(1).
			Render("No approval history yet.")
	}

	var sections []string
	sections = append(sections, m.viewHistory()...)
	sections = append(sections, "") // separator between sections
	sections = append(sections, m.viewGrants()...)
	sections = append(sections, m.viewFooter())

	return strings.Join(sections, "\n")
}

func (m *ApprovalsPage) viewHistory() []string {
	titleText := fmt.Sprintf("Approval History (%d events)", len(m.histEntries))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).PaddingLeft(2)
	if m.section == 0 {
		titleStyle = titleStyle.Foreground(theme.Accent)
	}
	title := titleStyle.Render(titleText)

	separator := lipgloss.NewStyle().
		Foreground(theme.BorderSubtle).
		PaddingLeft(4).
		Render(strings.Repeat("─", max(m.width-8, 40)))

	// Compute dynamic summary column width.
	fixedW := apprGutterW + apprColTimeW + apprColToolW + apprColOutcomeW + apprColProvW + apprColGapW*4
	summaryW := max(m.width-fixedW, 8)

	// Header.
	fmtStr := fmt.Sprintf("%%-%ds %%-%ds %%-%ds %%-%ds %%s", apprColTimeW, apprColToolW, summaryW, apprColOutcomeW)
	headerText := fmt.Sprintf(fmtStr, "Time", "Tool", "Summary", "Outcome", "Provider")
	header := lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		Bold(true).
		PaddingLeft(4).
		Render(headerText)

	result := []string{title, "", header, separator}

	if len(m.histEntries) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(theme.TextSecondary).
			PaddingLeft(4).
			Render("  No history entries")
		return append(result, empty)
	}

	now := m.nowFn()
	for i, entry := range m.histEntries {
		timeStr := relativeTime(now, entry.Timestamp)
		timeStr = truncate(timeStr, apprColTimeW-2)
		toolStr := truncate(entry.ToolName, apprColToolW-2)
		summaryStr := ansi.Truncate(entry.Summary, summaryW, "…")
		outcomeStr := truncate(entry.Outcome, apprColOutcomeW-2)
		provStr := truncate(entry.Provider, apprColProvW)

		row := fmt.Sprintf(fmtStr, timeStr, toolStr, summaryStr, outcomeStr, provStr)

		style := lipgloss.NewStyle().PaddingLeft(4)
		if m.section == 0 && i == m.cursor {
			style = style.Foreground(theme.Accent).Bold(true)
			row = "> " + row
		} else {
			style = style.Foreground(theme.TextPrimary)
			row = "  " + row
		}
		result = append(result, style.Render(row))
	}
	return result
}

func (m *ApprovalsPage) viewGrants() []string {
	titleText := fmt.Sprintf("Active Grants (%d)", len(m.grantList))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).PaddingLeft(2)
	if m.section == 1 {
		titleStyle = titleStyle.Foreground(theme.Accent)
	}
	title := titleStyle.Render(titleText)

	separator := lipgloss.NewStyle().
		Foreground(theme.BorderSubtle).
		PaddingLeft(4).
		Render(strings.Repeat("─", max(m.width-8, 40)))

	fmtStr := fmt.Sprintf("%%-%ds %%-%ds %%s", grantColSessionW, grantColToolW)
	headerText := fmt.Sprintf(fmtStr, "Session", "Tool", "Granted")
	header := lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		Bold(true).
		PaddingLeft(4).
		Render(headerText)

	result := []string{title, "", header, separator}

	if len(m.grantList) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(theme.TextSecondary).
			PaddingLeft(4).
			Render("  No active grants")
		return append(result, empty)
	}

	now := m.nowFn()
	for i, grant := range m.grantList {
		sessionStr := truncate(grant.SessionKey, grantColSessionW-2)
		toolStr := truncate(grant.ToolName, grantColToolW-2)
		timeStr := relativeTime(now, grant.GrantedAt)

		row := fmt.Sprintf(fmtStr, sessionStr, toolStr, timeStr)

		style := lipgloss.NewStyle().PaddingLeft(4)
		if m.section == 1 && i == m.grantCursor {
			style = style.Foreground(theme.Accent).Bold(true)
			row = "> " + row
		} else {
			style = style.Foreground(theme.TextPrimary)
			row = "  " + row
		}
		result = append(result, style.Render(row))
	}
	return result
}

func (m *ApprovalsPage) viewFooter() string {
	help := " [/] switch  [↑/↓] navigate"
	if m.section == 1 && len(m.grantList) > 0 {
		help += "  [r] revoke  [R] revoke session"
	}
	return lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		PaddingLeft(2).
		PaddingTop(1).
		Render(help)
}

func (m *ApprovalsPage) refreshData() {
	if m.history != nil {
		m.histEntries = m.history.List()
	} else {
		m.histEntries = nil
	}
	if m.grants != nil {
		m.grantList = m.grants.List()
	} else {
		m.grantList = nil
	}
	// Clamp cursors.
	if m.cursor >= len(m.histEntries) {
		m.cursor = max(len(m.histEntries)-1, 0)
	}
	if m.grantCursor >= len(m.grantList) {
		m.grantCursor = max(len(m.grantList)-1, 0)
	}
}

// relativeTime formats a timestamp as a human-readable relative duration.
func relativeTime(now, t time.Time) string {
	d := now.Sub(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
