package cockpit

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
	"github.com/langoai/lango/internal/observability"
)

// contextTickMsg triggers a periodic refresh of the context panel.
type contextTickMsg time.Time

// toolEntry pairs a tool name with its invocation count for sorted display.
type toolEntry struct {
	name  string
	count int64
}

// channelStatus represents the state of a single communication channel.
type channelStatus struct {
	Name         string
	Connected    bool
	MessageCount int
	LastActivity time.Time
}

// ContextPanel is a standalone tea.Model that displays live system metrics
// in a right-side panel. It is NOT a Page — it uses Start()/Stop() lifecycle
// managed by the cockpit toggle (Ctrl+P).
type ContextPanel struct {
	collector       *observability.MetricsCollector
	snapshot        observability.SystemSnapshot
	tickActive      bool
	visible         bool
	width           int
	height          int
	sortedTools      []toolEntry     // cached sorted tool stats
	sortedToolsDirty bool            // true when snapshot updated, cleared after sort
	channelStatuses  []channelStatus // live channel status for display
	runtimeStat      runtimeStatus   // live runtime status for display
}

// NewContextPanel creates a ContextPanel backed by the given collector.
// If collector is nil, the panel renders placeholder text.
func NewContextPanel(collector *observability.MetricsCollector) *ContextPanel {
	return &ContextPanel{
		collector: collector,
		width:     theme.ContextPanelWidth,
	}
}

// Init implements tea.Model. No initial command — ticks start on Start().
func (p *ContextPanel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (p *ContextPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	case contextTickMsg:
		if !p.tickActive {
			return p, nil
		}
		p.refreshSnapshot()
		return p, contextTickCmd()
	}
	return p, nil
}

// Pre-allocated styles for the context panel.
var (
	cpTitleStyle = lipgloss.NewStyle().
			Foreground(theme.TextPrimary).
			Bold(true)

	cpLabelStyle = lipgloss.NewStyle().
			Foreground(theme.TextSecondary)

	cpValueStyle = lipgloss.NewStyle().
			Foreground(theme.TextPrimary)

	cpMutedStyle = lipgloss.NewStyle().
			Foreground(theme.Muted)

	cpBorderStyle = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeftForeground(theme.BorderDefault).
			Background(theme.Surface0)

	cpDividerStyle = lipgloss.NewStyle().
			Foreground(theme.BorderDefault)
)

// View implements tea.Model.
func (p *ContextPanel) View() string {
	if !p.visible {
		return ""
	}

	contentWidth := p.width - 3 // account for border + padding
	if contentWidth < 8 {
		contentWidth = 8
	}
	divider := cpDividerStyle.Render(strings.Repeat("─", contentWidth))

	sections := []string{
		p.renderTokenUsage(contentWidth, divider),
		p.renderToolStats(contentWidth, divider),
	}
	// Runtime section — only shown when a turn is active
	if runtimeSection := p.renderRuntimeStatus(contentWidth, divider); runtimeSection != "" {
		sections = append(sections, runtimeSection)
	}
	if channelSection := p.renderChannelStatus(contentWidth, divider); channelSection != "" {
		sections = append(sections, channelSection)
	}
	sections = append(sections, p.renderSystem(contentWidth, divider))

	content := strings.Join(sections, "\n")

	// Fill remaining height.
	lines := strings.Split(content, "\n")
	for len(lines) < p.height {
		lines = append(lines, strings.Repeat(" ", contentWidth))
	}
	if len(lines) > p.height && p.height > 0 {
		lines = lines[:p.height]
	}

	rendered := lipgloss.NewStyle().
		Padding(0, 1).
		Width(p.width).
		Background(theme.Surface0).
		Render(strings.Join(lines, "\n"))

	return cpBorderStyle.Render(rendered)
}

// Start begins the 5-second tick cycle for auto-refresh.
func (p *ContextPanel) Start() tea.Cmd {
	p.tickActive = true
	p.refreshSnapshot()
	return contextTickCmd()
}

// Stop halts the tick cycle.
func (p *ContextPanel) Stop() {
	p.tickActive = false
}

// SetHeight updates the available render height.
func (p *ContextPanel) SetHeight(h int) {
	p.height = h
}

// SetVisible controls whether the panel renders content.
func (p *ContextPanel) SetVisible(v bool) {
	p.visible = v
}

// SetChannelStatuses updates the channel status display data.
func (p *ContextPanel) SetChannelStatuses(statuses []channelStatus) {
	p.channelStatuses = make([]channelStatus, len(statuses))
	copy(p.channelStatuses, statuses)
}

// SetRuntimeStatus updates the runtime status display data.
// Rendering is handled by Unit 3A; this setter allows the cockpit
// to push snapshots from RuntimeTracker.
func (p *ContextPanel) SetRuntimeStatus(status runtimeStatus) {
	p.runtimeStat = status
}

// --- rendering helpers ---

func (p *ContextPanel) renderTokenUsage(width int, divider string) string {
	var b strings.Builder
	b.WriteString(cpTitleStyle.Render("Token Usage"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	t := p.snapshot.TokenUsageTotal
	rows := []struct {
		label string
		value int64
	}{
		{"Input:", t.InputTokens},
		{"Output:", t.OutputTokens},
		{"Total:", t.TotalTokens},
		{"Cache:", t.CacheTokens},
	}

	labelW := 8
	for _, r := range rows {
		label := cpLabelStyle.Width(labelW).Render(r.label)
		val := cpValueStyle.Render(formatCompact(r.value))
		b.WriteString(label + val)
		b.WriteByte('\n')
	}
	return b.String()
}

func (p *ContextPanel) renderToolStats(width int, divider string) string {
	var b strings.Builder
	b.WriteString(cpTitleStyle.Render("Tool Stats"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	if len(p.snapshot.ToolBreakdown) == 0 {
		b.WriteString(cpMutedStyle.Render("No tool executions"))
		b.WriteByte('\n')
		return b.String()
	}

	// Re-sort only when snapshot has been refreshed.
	if p.sortedToolsDirty || p.sortedTools == nil {
		entries := make([]toolEntry, 0, len(p.snapshot.ToolBreakdown))
		for name, tm := range p.snapshot.ToolBreakdown {
			entries = append(entries, toolEntry{name: name, count: tm.Count})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].count > entries[j].count
		})
		if len(entries) > 5 {
			entries = entries[:5]
		}
		p.sortedTools = entries
		p.sortedToolsDirty = false
	}
	entries := p.sortedTools

	nameW := width - 6 // space for count column
	if nameW < 8 {
		nameW = 8
	}
	for _, e := range entries {
		name := truncateName(e.name, nameW)
		line := fmt.Sprintf("%-*s %4d", nameW, name, e.count)
		b.WriteString(cpLabelStyle.Render(line))
		b.WriteByte('\n')
	}
	return b.String()
}

func (p *ContextPanel) renderSystem(_ int, divider string) string {
	var b strings.Builder
	b.WriteString(cpTitleStyle.Render("System"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	uptime := formatUptime(p.snapshot.Uptime)
	b.WriteString(cpLabelStyle.Render("Uptime:  "))
	b.WriteString(cpValueStyle.Render(uptime))
	b.WriteByte('\n')

	return b.String()
}

func (p *ContextPanel) renderRuntimeStatus(_ int, divider string) string {
	if !p.runtimeStat.IsRunning {
		return "" // graceful degradation — no section when idle
	}

	var b strings.Builder
	b.WriteString(cpTitleStyle.Render("Runtime"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	// Active agent indicator
	statusIcon := lipgloss.NewStyle().Foreground(theme.Success).Render("●")
	label := "Running"
	if p.runtimeStat.ActiveAgent != "" {
		label += "  " + lipgloss.NewStyle().Foreground(theme.TextPrimary).Render(p.runtimeStat.ActiveAgent)
	}
	b.WriteString(fmt.Sprintf("  %s %s", statusIcon, cpLabelStyle.Render(label)))
	b.WriteByte('\n')

	// Delegation count (only show if > 0)
	if p.runtimeStat.DelegationCount > 0 {
		b.WriteString(fmt.Sprintf("  🔀 %s",
			cpValueStyle.Render(fmt.Sprintf("%d delegations", p.runtimeStat.DelegationCount))))
		b.WriteByte('\n')
	}

	// Token usage (only show if > 0)
	if p.runtimeStat.TurnTokens > 0 {
		b.WriteString(fmt.Sprintf("  📊 %s",
			cpValueStyle.Render(fmt.Sprintf("%s tokens", formatCompact(p.runtimeStat.TurnTokens)))))
		b.WriteByte('\n')
	}

	return b.String()
}

func (p *ContextPanel) renderChannelStatus(_ int, divider string) string {
	if len(p.channelStatuses) == 0 {
		return "" // graceful degradation — no section when no channels
	}

	var b strings.Builder
	b.WriteString(cpTitleStyle.Render("Channels"))
	b.WriteByte('\n')
	b.WriteString(divider)
	b.WriteByte('\n')

	for _, ch := range p.channelStatuses {
		status := "●"
		statusColor := theme.Success
		if !ch.Connected {
			status = "○"
			statusColor = theme.Error
		}

		statusIcon := lipgloss.NewStyle().Foreground(statusColor).Render(status)
		name := lipgloss.NewStyle().Foreground(theme.TextPrimary).Render(ch.Name)
		count := lipgloss.NewStyle().Foreground(theme.TextTertiary).Render(fmt.Sprintf("%d msgs", ch.MessageCount))

		b.WriteString(fmt.Sprintf("  %s %s  %s", statusIcon, name, count))
		b.WriteByte('\n')
	}
	return b.String()
}

func (p *ContextPanel) refreshSnapshot() {
	if p.collector != nil {
		prev := p.snapshot
		p.snapshot = p.collector.Snapshot()
		if len(p.snapshot.ToolBreakdown) != len(prev.ToolBreakdown) || toolCountSum(p.snapshot) != toolCountSum(prev) {
			p.sortedToolsDirty = true
		}
	}
}

// --- utility functions ---

func contextTickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return contextTickMsg(t)
	})
}

// formatCompact renders a number in compact form (e.g. 12345 -> "12,345").
func formatCompact(n int64) string {
	if n < 0 {
		return "-" + formatCompact(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var buf strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		buf.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if buf.Len() > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(s[i : i+3])
	}
	return buf.String()
}

// formatUptime renders a duration as a human-friendly string.
func formatUptime(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	totalSec := int(d.Seconds())
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	s := totalSec % 60

	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	default:
		return fmt.Sprintf("%dm %ds", m, s)
	}
}

// toolCountSum returns the total invocation count across all tools in a snapshot.
func toolCountSum(snap observability.SystemSnapshot) int64 {
	var sum int64
	for _, tm := range snap.ToolBreakdown {
		sum += tm.Count
	}
	return sum
}

// truncateName shortens a tool name if it exceeds maxLen.
func truncateName(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
