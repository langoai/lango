package pages

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/cli/cockpit/theme"
	"github.com/langoai/lango/internal/cli/tui"
)

const missionControlTickInterval = 2 * time.Second

type missionControlTickMsg time.Time

type missionControlFocus int

const (
	missionControlFocusMissions missionControlFocus = iota
	missionControlFocusDecisions
	missionControlFocusComposer
)

type missionControlProjector interface {
	Project([]background.TaskSnapshot) cockpit.MissionControlSnapshot
}

type missionControlTaskSource interface {
	List() []background.TaskSnapshot
}

// MissionControlPage renders the Wave 1 Mission Control surface.
type MissionControlPage struct {
	projector  missionControlProjector
	taskSource missionControlTaskSource
	composer   *chat.ChatModel

	width      int
	height     int
	tickActive bool
	hasLoaded  bool

	snapshot       cockpit.MissionControlSnapshot
	focus          missionControlFocus
	missionCursor  int
	decisionCursor int
	activityOffset int
}

// NewMissionControlPage creates a Mission Control page backed by the shared projector.
func NewMissionControlPage(deps cockpit.Deps, composer *chat.ChatModel) *MissionControlPage {
	return newMissionControlPage(cockpit.NewMissionControlProjector(deps), deps.BackgroundManager, composer)
}

func newMissionControlPage(
	projector missionControlProjector,
	taskSource missionControlTaskSource,
	composer *chat.ChatModel,
) *MissionControlPage {
	if composer == nil {
		composer = chat.New(chat.Deps{})
	}
	return &MissionControlPage{
		projector:  projector,
		taskSource: taskSource,
		composer:   composer,
		focus:      missionControlFocusMissions,
	}
}

func (p *MissionControlPage) Title() string { return "Mission Control" }

func (p *MissionControlPage) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "focus")),
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	}
}

func (p *MissionControlPage) Init() tea.Cmd { return nil }

func (p *MissionControlPage) Activate() tea.Cmd {
	p.tickActive = true
	return missionControlTickCmd()
}

func (p *MissionControlPage) Deactivate() {
	p.tickActive = false
}

func (p *MissionControlPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		if p.composer != nil {
			updated, _ := p.composer.Update(msg)
			p.composer = updated.(*chat.ChatModel)
		}
		return p, nil
	case missionControlTickMsg:
		if !p.tickActive {
			return p, nil
		}
		p.refreshSnapshot()
		return p, missionControlTickCmd()
	case tea.KeyMsg:
		return p.handleKey(msg)
	case chat.DoneMsg, chat.ErrorMsg, chat.WarningMsg, chat.ToolStartedMsg, chat.ToolFinishedMsg,
		chat.ThinkingStartedMsg, chat.ThinkingFinishedMsg, chat.ApprovalRequestMsg, chat.DelegationMsg,
		chat.BudgetWarningMsg, chat.RecoveryMsg, chat.TurnTokenUsageMsg, chat.SystemMsg,
		chat.ChunkMsg, chat.ChannelMessageMsg:
		if p.composer == nil {
			return p, nil
		}
		updated, cmd := p.composer.Update(msg)
		p.composer = updated.(*chat.ChatModel)
		p.refreshSnapshot()
		return p, cmd
	}
	return p, nil
}

func (p *MissionControlPage) View() string {
	if p.width == 0 || p.height == 0 {
		return "\n  Waiting for terminal size..."
	}

	header := p.renderHeader()
	footer := p.renderFooter()

	if !p.hasLoaded {
		body := p.renderLoading()
		return joinNonEmpty(header, body, footer)
	}

	if p.isEmpty() {
		body := joinNonEmpty(
			p.renderEmpty(),
			p.renderActivityPane(false),
			p.renderComposerLine(),
		)
		return joinNonEmpty(header, body, footer)
	}

	if p.height < 24 {
		body := p.renderFocusedLane(true)
		return joinNonEmpty(header, body, footer)
	}

	switch {
	case p.width >= 120:
		top := p.renderWideTop()
		body := joinNonEmpty(top, p.renderActivityPane(true), p.renderComposerLine())
		return joinNonEmpty(header, body, footer)
	case p.width >= 80:
		body := joinNonEmpty(p.renderMissionPane(), p.renderDecisionPane(false), p.renderActivityPane(true), p.renderComposerLine())
		return joinNonEmpty(header, body, footer)
	default:
		body := p.renderFocusedLane(false)
		return joinNonEmpty(header, body, footer)
	}
}

func (p *MissionControlPage) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
		p.focus = (p.focus + 1) % 3
		return p, nil
	case p.focus == missionControlFocusDecisions:
		if cmd, handled := p.forwardDecisionKey(msg); handled {
			return p, cmd
		}
	case isMissionControlPrintableKey(msg):
		p.focus = missionControlFocusComposer
		return p.forwardComposerKey(msg)
	case p.focus == missionControlFocusComposer:
		return p.forwardComposerKey(msg)
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		p.moveCursor(-1)
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		p.moveCursor(1)
	}
	return p, nil
}

func (p *MissionControlPage) forwardDecisionKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if p.snapshot.Decision == nil || p.composer == nil {
		return nil, false
	}
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("a", "s", "d", "esc"))):
		return p.composer.HandlePendingApprovalKey(msg), true
	default:
		return nil, false
	}
}

func (p *MissionControlPage) forwardComposerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if p.composer == nil {
		return p, nil
	}
	updated, cmd := p.composer.Update(msg)
	p.composer = updated.(*chat.ChatModel)
	p.refreshSnapshot()
	return p, cmd
}

func (p *MissionControlPage) moveCursor(delta int) {
	switch p.focus {
	case missionControlFocusMissions:
		maxIdx := len(p.snapshot.Missions) - 1
		p.missionCursor = clamp(p.missionCursor+delta, 0, maxIdx)
	case missionControlFocusDecisions:
		if p.snapshot.Decision == nil {
			p.decisionCursor = 0
			return
		}
		p.decisionCursor = clamp(p.decisionCursor+delta, 0, 0)
	case missionControlFocusComposer:
		maxIdx := len(p.snapshot.Activities) - 1
		p.activityOffset = clamp(p.activityOffset+delta, 0, maxIdx)
	}
}

func (p *MissionControlPage) refreshSnapshot() {
	if p.projector == nil {
		p.snapshot = cockpit.MissionControlSnapshot{}
		p.hasLoaded = true
		return
	}

	var tasks []background.TaskSnapshot
	if p.taskSource != nil {
		tasks = p.taskSource.List()
	}
	p.snapshot = p.projector.Project(tasks)
	p.hasLoaded = true
}

func (p *MissionControlPage) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Render("Mission Control")

	lines := []string{title}
	if text := strings.TrimSpace(p.snapshot.Header.ActiveAgentSummary); text != "" {
		lines = append(lines, "Agents: "+text)
	}
	lines = append(lines, fmt.Sprintf("Pending decisions: %d", p.snapshot.Header.PendingDecisionCount))
	if text := strings.TrimSpace(p.snapshot.Header.ModelProviderSummary); text != "" {
		lines = append(lines, "Model: "+text)
	}
	if text := strings.TrimSpace(p.snapshot.Header.ContextSummary); text != "" {
		lines = append(lines, "Context: "+text)
	}
	if text := strings.TrimSpace(p.snapshot.Header.MetricsSummary); text != "" {
		lines = append(lines, "Metrics: "+text)
	}
	if text := strings.TrimSpace(p.snapshot.Header.DegradedNote); text != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Warning).Render("Degraded: "+text))
	}
	return strings.Join(lines, "\n")
}

func (p *MissionControlPage) renderLoading() string {
	return lipgloss.NewStyle().
		Foreground(theme.TextSecondary).
		Render("Loading Mission Control...")
}

func (p *MissionControlPage) renderEmpty() string {
	lines := []string{
		"No active missions or pending decisions.",
		"Type to chat here, or use `lango chat` for focused chat.",
	}
	if text := strings.TrimSpace(p.snapshot.Header.DegradedNote); text != "" {
		lines = append(lines, "Degraded: "+text)
	}
	return lipgloss.NewStyle().
		Foreground(theme.TextSecondary).
		Render(strings.Join(lines, "\n"))
}

func (p *MissionControlPage) renderWideTop() string {
	leftWidth := max(40, (p.width-3)*2/3)
	rightWidth := max(24, p.width-leftWidth-3)
	left := lipgloss.NewStyle().Width(leftWidth).Render(p.renderMissionPane())
	right := lipgloss.NewStyle().Width(rightWidth).Render(p.renderDecisionPane(true))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right)
}

func (p *MissionControlPage) renderFocusedLane(compact bool) string {
	switch p.focus {
	case missionControlFocusDecisions:
		return p.renderDecisionPane(true)
	case missionControlFocusComposer:
		showComposer := !compact || p.composerVisibleInCompact()
		return joinNonEmpty(p.renderActivityPane(showComposer), cond(showComposer, p.renderComposerLine()))
	default:
		return p.renderMissionPane()
	}
}

func (p *MissionControlPage) renderMissionPane() string {
	title := p.sectionTitle("Missions", p.focus == missionControlFocusMissions)
	if len(p.snapshot.Missions) == 0 {
		return joinNonEmpty(title, "No active or proposed missions.")
	}

	lines := []string{title}
	for idx, mission := range p.snapshot.Missions {
		prefix := "  "
		if p.focus == missionControlFocusMissions && idx == p.missionCursor {
			prefix = "> "
		}

		meta := strings.TrimSpace(strings.Join([]string{
			string(mission.Kind),
			string(mission.Status),
		}, " / "))
		if meta != "" {
			lines = append(lines, prefix+mission.Title+" ["+meta+"]")
		} else {
			lines = append(lines, prefix+mission.Title)
		}
		if detail := strings.TrimSpace(firstNonEmpty(mission.Detail, mission.NextAction, mission.BlockedReason)); detail != "" {
			lines = append(lines, "    "+detail)
		}
	}
	if extra := strings.TrimSpace(p.snapshot.MissionOverflowSummary); extra != "" {
		lines = append(lines, "  "+extra)
	}
	return strings.Join(lines, "\n")
}

func (p *MissionControlPage) renderDecisionPane(_ bool) string {
	title := p.sectionTitle("Decisions", p.focus == missionControlFocusDecisions)
	if p.snapshot.Decision == nil {
		return joinNonEmpty(title, "No pending decisions.")
	}

	decision := p.snapshot.Decision
	lines := []string{
		title,
		"> Action: " + firstNonEmpty(decision.Title, "Pending approval"),
		"  Reason: " + firstNonEmpty(decision.Reason, "—"),
		"  Effect: " + firstNonEmpty(decision.EffectText, "—"),
		"  Risk: " + firstNonEmpty(decision.RiskLabel, decision.RiskLevel, "—"),
	}
	return strings.Join(lines, "\n")
}

func (p *MissionControlPage) renderActivityPane(includeComposerHint bool) string {
	title := p.sectionTitle("Activity", p.focus == missionControlFocusComposer)
	if len(p.snapshot.Activities) == 0 {
		lines := []string{title, "No recent activity yet."}
		if includeComposerHint {
			lines = append(lines, "Type to chat here, or use `lango chat` for focused chat.")
		}
		return strings.Join(lines, "\n")
	}

	lines := []string{title}
	for _, item := range visibleActivities(p.snapshot.Activities, p.activityOffset, 6) {
		summary := item.Summary
		if ts := item.Timestamp; !ts.IsZero() {
			summary = fmt.Sprintf("%s  %s", tui.RelativeTime(time.Now(), ts), summary)
		}
		lines = append(lines, "- "+summary)
	}
	if extra := strings.TrimSpace(p.snapshot.ActivityOverflowSummary); extra != "" {
		lines = append(lines, "- "+extra)
	}
	return strings.Join(lines, "\n")
}

func (p *MissionControlPage) renderComposerLine() string {
	value := ""
	placeholder := "Type to chat here, or use `lango chat` for focused chat."
	if p.composer != nil {
		value = p.composer.ComposerValue()
		if text := strings.TrimSpace(p.composer.ComposerPlaceholder()); text != "" {
			placeholder = text
		}
	}

	content := value
	style := lipgloss.NewStyle().Foreground(theme.TextPrimary)
	if strings.TrimSpace(content) == "" {
		content = placeholder
		style = lipgloss.NewStyle().Foreground(theme.TextSecondary)
	}
	return style.Render("> " + singleLine(content))
}

func (p *MissionControlPage) renderFooter() string {
	lane := "Missions"
	switch p.focus {
	case missionControlFocusDecisions:
		lane = "Decisions"
	case missionControlFocusComposer:
		lane = "Composer"
	}

	pending := p.snapshot.Header.PendingDecisionCount
	pendingText := fmt.Sprintf("%d pending", pending)
	hint := "Tab lanes: Missions / Decisions / Composer  Type to chat here  `lango chat` fallback"
	return lipgloss.NewStyle().
		Foreground(theme.TextSecondary).
		Render(fmt.Sprintf("Focus: %s  %s  %s", lane, pendingText, hint))
}

func (p *MissionControlPage) composerVisibleInCompact() bool {
	if p.focus == missionControlFocusComposer {
		return true
	}
	if p.composer == nil {
		return false
	}
	return strings.TrimSpace(p.composer.ComposerValue()) != ""
}

func (p *MissionControlPage) sectionTitle(label string, focused bool) string {
	style := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	if focused {
		style = style.Foreground(theme.Accent)
	}
	return style.Render(label)
}

func (p *MissionControlPage) isEmpty() bool {
	return len(p.snapshot.Missions) == 0 && p.snapshot.Decision == nil
}

func missionControlTickCmd() tea.Cmd {
	return tea.Tick(missionControlTickInterval, func(t time.Time) tea.Msg {
		return missionControlTickMsg(t)
	})
}

func isMissionControlPrintableKey(msg tea.KeyMsg) bool {
	if msg.Type != tea.KeyRunes || len(msg.Runes) == 0 || msg.Alt {
		return false
	}
	for _, r := range msg.Runes {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func visibleActivities(items []cockpit.ActivityView, offset, limit int) []cockpit.ActivityView {
	if len(items) == 0 {
		return nil
	}
	if limit <= 0 {
		return items
	}
	start := clamp(offset, 0, len(items)-1)
	end := min(len(items), start+limit)
	return items[start:end]
}

func singleLine(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	return strings.Join(strings.Fields(text), " ")
}

func joinNonEmpty(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, "\n\n")
}

func firstNonEmpty(items ...string) string {
	for _, item := range items {
		if text := strings.TrimSpace(item); text != "" {
			return text
		}
	}
	return ""
}

func cond(ok bool, text string) string {
	if ok {
		return text
	}
	return ""
}

func clamp(value, minValue, maxValue int) int {
	if maxValue < minValue {
		return 0
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
