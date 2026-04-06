package pages

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/types"
)

// tickMsg triggers a periodic metrics refresh.
type tickMsg time.Time

// StatusPage displays feature status and system metrics.
type StatusPage struct {
	featureStatuses  []types.FeatureStatus
	statusProvider   func() []types.FeatureStatus
	metricsCollector *observability.MetricsCollector
	snapshot         observability.SystemSnapshot
	cfg              *config.Config
	tickActive       bool
	width, height    int
}

// NewStatusPage creates a new StatusPage.
//
// statusProvider is called periodically to fetch current feature statuses,
// avoiding a direct import of internal/app.
func NewStatusPage(
	statusProvider func() []types.FeatureStatus,
	collector *observability.MetricsCollector,
	cfg *config.Config,
) *StatusPage {
	return &StatusPage{
		statusProvider:   statusProvider,
		metricsCollector: collector,
		cfg:              cfg,
	}
}

// Title returns the page tab label.
func (m *StatusPage) Title() string { return "Status" }

// ShortHelp returns key bindings for the help bar (none — read-only page).
func (m *StatusPage) ShortHelp() []key.Binding { return nil }

// Init satisfies tea.Model but does nothing; ticks start on Activate.
func (m *StatusPage) Init() tea.Cmd { return nil }

// Activate starts periodic metric collection.
func (m *StatusPage) Activate() tea.Cmd {
	m.tickActive = true
	m.refreshData()
	return tickCmd()
}

// Deactivate stops the tick loop.
func (m *StatusPage) Deactivate() {
	m.tickActive = false
}

// Update handles messages.
func (m *StatusPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		if !m.tickActive {
			return m, nil
		}
		m.refreshData()
		return m, tickCmd()
	}
	return m, nil
}

// View renders the status dashboard.
func (m *StatusPage) View() string {
	sectionTitle := lipgloss.NewStyle().
		Foreground(theme.TextPrimary).
		Bold(true)

	separator := lipgloss.NewStyle().
		Foreground(theme.BorderDefault)

	divider := separator.Render(strings.Repeat("─", 30))

	sections := []string{
		m.renderFeatureFlags(sectionTitle, divider),
		m.renderTokenUsage(sectionTitle, divider),
		m.renderToolExecution(sectionTitle, divider),
		m.renderSystemInfo(sectionTitle, divider),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

// --- sections ---

func (m *StatusPage) renderFeatureFlags(
	titleStyle lipgloss.Style, divider string,
) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Feature Status"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	enabledStyle := lipgloss.NewStyle().Foreground(theme.Success)
	disabledStyle := lipgloss.NewStyle().Foreground(theme.Muted)
	labelStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary).Width(20)
	reasonStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	for _, fs := range m.featureStatuses {
		var indicator string
		var statusText string
		if fs.Enabled {
			indicator = enabledStyle.Render("●")
			statusText = enabledStyle.Render("enabled")
		} else {
			indicator = disabledStyle.Render("○")
			statusText = disabledStyle.Render("disabled")
		}
		line := fmt.Sprintf("%s %s%s", indicator, labelStyle.Render(fs.Name), statusText)
		if !fs.Enabled && fs.Reason != "" {
			line += reasonStyle.Render(" (" + fs.Reason + ")")
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func (m *StatusPage) renderTokenUsage(
	titleStyle lipgloss.Style, divider string,
) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Token Usage"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	labelStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary).Width(12)
	valueStyle := lipgloss.NewStyle().Foreground(theme.TextPrimary).Align(lipgloss.Right).Width(12)

	t := m.snapshot.TokenUsageTotal
	rows := []struct {
		label string
		value int64
	}{
		{"Input:", t.InputTokens},
		{"Output:", t.OutputTokens},
		{"Total:", t.TotalTokens},
		{"Cache:", t.CacheTokens},
	}
	for _, r := range rows {
		b.WriteString(labelStyle.Render(r.label))
		b.WriteString(valueStyle.Render(tui.FormatNumber(r.value)))
		b.WriteByte('\n')
	}
	return b.String()
}

func (m *StatusPage) renderToolExecution(
	titleStyle lipgloss.Style, divider string,
) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Tool Execution"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	labelStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
	valueStyle := lipgloss.NewStyle().Foreground(theme.TextPrimary)

	b.WriteString(labelStyle.Render("Total executions:  "))
	b.WriteString(valueStyle.Render(tui.FormatNumber(m.snapshot.ToolExecutions)))
	b.WriteByte('\n')

	// Sort tools by call count descending for stable output.
	tools := make([]observability.ToolMetric, 0, len(m.snapshot.ToolBreakdown))
	for _, tm := range m.snapshot.ToolBreakdown {
		tools = append(tools, tm)
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Count > tools[j].Count
	})

	nameStyle := lipgloss.NewStyle().Foreground(theme.TextPrimary).Width(22)
	detailStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)

	for _, tm := range tools {
		avg := tui.FormatDuration(tm.AvgDuration)
		detail := fmt.Sprintf("%d calls  avg %s  %d errors",
			tm.Count, avg, tm.Errors)
		b.WriteString(nameStyle.Render(tm.Name))
		b.WriteString(detailStyle.Render(detail))
		b.WriteByte('\n')
	}
	return b.String()
}

func (m *StatusPage) renderSystemInfo(
	titleStyle lipgloss.Style, divider string,
) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("System"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	labelStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary).Width(12)
	valueStyle := lipgloss.NewStyle().Foreground(theme.TextPrimary)

	provider := ""
	model := ""
	if m.cfg != nil {
		provider = m.cfg.Agent.Provider
		model = m.cfg.Agent.Model
	}

	rows := []struct {
		label string
		value string
	}{
		{"Provider:", provider},
		{"Model:", model},
		{"Uptime:", tui.FormatDuration(m.snapshot.Uptime)},
	}
	for _, r := range rows {
		b.WriteString(labelStyle.Render(r.label))
		b.WriteString(valueStyle.Render(r.value))
		b.WriteByte('\n')
	}
	return b.String()
}

// --- helpers ---

func (m *StatusPage) refreshData() {
	if m.metricsCollector != nil {
		m.snapshot = m.metricsCollector.Snapshot()
	}
	if m.statusProvider != nil {
		m.featureStatuses = m.statusProvider()
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
