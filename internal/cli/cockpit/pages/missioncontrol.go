package pages

import (
	"context"
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
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/proposal"
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

	sessionKey     string
	missionService cockpit.MissionLifecycleService
	proposalReader cockpit.ProposalReader
	proposalSvc    cockpit.ProposalMutationService
	learningBuffer *cockpit.LearningSuggestionBuffer

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
	page := newMissionControlPage(cockpit.NewMissionControlProjector(deps), deps.BackgroundManager, composer)
	page.sessionKey = deps.SessionKey
	page.missionService = deps.MissionService
	page.proposalReader = deps.ProposalReader
	page.proposalSvc = deps.ProposalService
	page.learningBuffer = deps.LearningBuffer
	return page
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
	case cockpit.MissionControlRefreshMsg:
		p.refreshSnapshot()
		return p, nil
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
			p.renderLoopPane(),
			p.renderActivityPane(false),
			p.renderComposerLine(),
		)
		return joinNonEmpty(header, body, footer)
	}

	if p.height < 24 {
		body := joinNonEmpty(p.renderFocusedLane(true), p.renderLoopPane())
		return joinNonEmpty(header, body, footer)
	}

	switch {
	case p.width >= 120:
		top := p.renderWideTop()
		body := joinNonEmpty(top, p.renderLoopPane(), p.renderActivityPane(true), p.renderComposerLine())
		return joinNonEmpty(header, body, footer)
	case p.width >= 80:
		body := joinNonEmpty(p.renderMissionPane(), p.renderDecisionPane(false), p.renderLoopPane(), p.renderActivityPane(true), p.renderComposerLine())
		return joinNonEmpty(header, body, footer)
	default:
		body := joinNonEmpty(p.renderFocusedLane(false), p.renderLoopPane())
		return joinNonEmpty(header, body, footer)
	}
}

func (p *MissionControlPage) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("tab"))) {
		p.focus = (p.focus + 1) % 3
		return p, nil
	}

	if p.focus == missionControlFocusDecisions {
		if cmd, handled := p.forwardDecisionKey(msg); handled {
			return p, cmd
		}
	}

	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if p.focus == missionControlFocusMissions {
			if cmd, handled := p.acceptSelectedProposal(); handled {
				return p, cmd
			}
		}
		if p.focus == missionControlFocusComposer {
			if cmd, handled := p.submitComposerFromMissionControl(); handled {
				return p, cmd
			}
		}
	}
	if p.focus == missionControlFocusMissions {
		if key.Matches(msg, key.NewBinding(key.WithKeys("d"))) {
			if cmd, handled := p.dismissSelectedProposal(); handled {
				return p, cmd
			}
		}
	}

	switch {
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

func (p *MissionControlPage) acceptSelectedProposal() (tea.Cmd, bool) {
	missionView := p.selectedMission()
	if missionView == nil || missionView.Kind != cockpit.MissionKindProposed || p.missionService == nil {
		return nil, false
	}

	description := ""
	if p.proposalReader != nil && p.proposalSvc != nil {
		proposalID := strings.TrimSpace(missionView.ID)
		proposalRow, ok := p.proposalReader.GetByID(proposalID)
		if !ok {
			return func() tea.Msg {
				return chat.SystemMsg{Text: fmt.Sprintf("Proposal %q is no longer available", proposalID)}
			}, true
		}
		if proposalRow.PreparedBrief != nil {
			description = compactPreparedBrief(*proposalRow.PreparedBrief)
		}
		if description == "" {
			description = strings.TrimSpace(firstNonEmpty(proposalRow.Summary, proposalRow.Reason))
		}
		if _, err := p.proposalSvc.Accept(context.Background(), proposalID); err != nil {
			return func() tea.Msg {
				return chat.SystemMsg{Text: fmt.Sprintf("Proposal acceptance failed: %v", err)}
			}, true
		}
		if _, err := p.missionService.AcceptProposal(context.Background(), mission.AcceptProposalInput{
			SessionKey:  strings.TrimSpace(p.sessionKey),
			SourceKind:  strings.TrimSpace(proposalRow.Source.Kind),
			SourceRef:   strings.TrimSpace(proposalRow.Source.Ref),
			Title:       strings.TrimSpace(firstNonEmpty(proposalRow.Title, missionView.Title)),
			Description: description,
		}); err != nil {
			if _, restoreErr := p.proposalSvc.RestorePrepared(context.Background(), proposalID); restoreErr != nil {
				return func() tea.Msg {
					return chat.SystemMsg{Text: fmt.Sprintf("Mission proposal acceptance failed: %v (restore failed: %v)", err, restoreErr)}
				}, true
			}
			return func() tea.Msg {
				return chat.SystemMsg{Text: fmt.Sprintf("Mission proposal acceptance failed: %v", err)}
			}, true
		}
		p.refreshSnapshot()
		return nil, true
	}

	sourceKind := strings.TrimSpace(missionView.SourceKind)
	sourceRef := strings.TrimSpace(missionView.SourceRef)
	if sourceKind == "" || sourceRef == "" {
		// Backward compatibility for stale snapshots created before explicit
		// proposal metadata was carried on MissionView.
		if sourceKind == "" {
			sourceKind = "proposed_learning"
		}
		if sourceRef == "" && strings.HasPrefix(missionView.ID, "learn:") {
			sourceRef = strings.TrimSpace(strings.TrimPrefix(missionView.ID, "learn:"))
		}
	}

	title := strings.TrimSpace(missionView.Title)
	description = strings.TrimSpace(missionView.Detail)
	if item := p.lookupLearningSuggestion(sourceRef); item != nil && description == "" {
		description = strings.TrimSpace(item.Rationale)
	}

	if _, err := p.missionService.AcceptProposal(context.Background(), mission.AcceptProposalInput{
		SessionKey:  strings.TrimSpace(p.sessionKey),
		SourceKind:  sourceKind,
		SourceRef:   sourceRef,
		Title:       title,
		Description: description,
	}); err != nil {
		return func() tea.Msg {
			return chat.SystemMsg{Text: fmt.Sprintf("Mission proposal acceptance failed: %v", err)}
		}, true
	}

	if p.learningBuffer != nil && sourceRef != "" {
		p.learningBuffer.Dismiss(sourceRef)
	}
	p.refreshSnapshot()
	return nil, true
}

func (p *MissionControlPage) dismissSelectedProposal() (tea.Cmd, bool) {
	missionView := p.selectedMission()
	if missionView == nil || missionView.Kind != cockpit.MissionKindProposed {
		return nil, false
	}
	if p.proposalSvc == nil {
		return nil, false
	}
	if _, err := p.proposalSvc.Dismiss(context.Background(), strings.TrimSpace(missionView.ID)); err != nil {
		return func() tea.Msg {
			return chat.SystemMsg{Text: fmt.Sprintf("Proposal dismiss failed: %v", err)}
		}, true
	}
	p.refreshSnapshot()
	return nil, true
}

func (p *MissionControlPage) forwardDecisionKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if p.snapshot.Decision == nil || p.composer == nil {
		return nil, false
	}
	if !p.composer.CanHandlePendingApprovalKey(msg) {
		return nil, false
	}
	cmd := p.composer.HandlePendingApprovalKey(msg)
	p.refreshSnapshot()
	return cmd, true
}

func (p *MissionControlPage) forwardComposerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if p.composer == nil {
		return p, nil
	}
	if p.composer.HasPendingApproval() && isMissionControlComposerEditingKey(msg) {
		cmd := p.composer.HandleComposerEditingKey(msg)
		p.refreshSnapshot()
		return p, cmd
	}
	updated, cmd := p.composer.Update(msg)
	p.composer = updated.(*chat.ChatModel)
	p.refreshSnapshot()
	return p, cmd
}

func (p *MissionControlPage) submitComposerFromMissionControl() (tea.Cmd, bool) {
	if p.composer == nil {
		return nil, false
	}
	input := strings.TrimSpace(p.composer.ComposerValue())
	if input == "" {
		return nil, true
	}
	if !p.composer.CanStartTurnFromComposer() {
		return nil, false
	}
	if strings.HasPrefix(input, "/") || p.missionService == nil {
		cmd := p.composer.SubmitComposerWithParent(context.Background())
		p.refreshSnapshot()
		return cmd, true
	}

	row, err := p.missionService.StartMission(context.Background(), mission.StartMissionInput{
		SessionKey:  strings.TrimSpace(p.sessionKey),
		Title:       input,
		Description: "",
		SourceKind:  "user",
		StartActive: true,
	})
	if err != nil {
		return func() tea.Msg {
			return chat.SystemMsg{Text: fmt.Sprintf("Mission start failed: %v", err)}
		}, true
	}

	cmd := p.composer.SubmitComposerWithParent(ctxkeys.WithMissionID(context.Background(), row.ID.String()))
	p.refreshSnapshot()
	return cmd, true
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

func (p *MissionControlPage) selectedMission() *cockpit.MissionView {
	if len(p.snapshot.Missions) == 0 {
		return nil
	}
	idx := clamp(p.missionCursor, 0, len(p.snapshot.Missions)-1)
	return &p.snapshot.Missions[idx]
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

func (p *MissionControlPage) lookupLearningSuggestion(id string) *eventbus.LearningSuggestionEvent {
	if p.learningBuffer == nil || id == "" {
		return nil
	}
	return p.learningBuffer.Find(id)
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
		if summary := compactCollaborationSummary(mission.Collaboration); summary != "" {
			maxWidth := max(24, p.width-8)
			lines = append(lines, "    "+tui.Truncate(summary, maxWidth))
		}
		if mission.Kind == cockpit.MissionKindProposed {
			if sourceLabel := strings.TrimSpace(compactProposalSourceLabel(mission)); sourceLabel != "" {
				lines = append(lines, "    source: "+sourceLabel)
			}
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

func (p *MissionControlPage) renderLoopPane() string {
	if len(p.snapshot.Loops) == 0 {
		return ""
	}
	lines := []string{p.sectionTitle("Agenda", false)}
	lines = append(lines, fmt.Sprintf("Open loops: %d", p.snapshot.OpenLoopCount))
	for _, loop := range p.snapshot.Loops {
		meta := strings.TrimSpace(strings.Join([]string{string(loop.Kind), string(loop.Status)}, " / "))
		line := "- " + loop.Title
		if meta != "" {
			line += " [" + meta + "]"
		}
		lines = append(lines, line)
		if detail := strings.TrimSpace(firstNonEmpty(loop.Summary, loop.NextAction)); detail != "" {
			lines = append(lines, "  "+detail)
		}
	}
	if extra := strings.TrimSpace(p.snapshot.LoopOverflowSummary); extra != "" {
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
	return len(p.snapshot.Missions) == 0 && p.snapshot.Decision == nil && len(p.snapshot.Loops) == 0
}

func missionControlTickCmd() tea.Cmd {
	return tea.Tick(missionControlTickInterval, func(t time.Time) tea.Msg {
		return missionControlTickMsg(t)
	})
}

func compactProposalSourceLabel(mission cockpit.MissionView) string {
	parts := make([]string, 0, 2)
	if kind := strings.TrimSpace(mission.SourceKind); kind != "" {
		parts = append(parts, kind)
	}
	if ref := strings.TrimSpace(mission.SourceRef); ref != "" {
		parts = append(parts, ref)
	}
	return strings.Join(parts, " / ")
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

func isMissionControlComposerEditingKey(msg tea.KeyMsg) bool {
	if isMissionControlPrintableKey(msg) {
		return true
	}
	switch msg.Type {
	case tea.KeyBackspace, tea.KeyDelete, tea.KeyLeft, tea.KeyRight, tea.KeyHome, tea.KeyEnd:
		return true
	default:
		return false
	}
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

func compactPreparedBrief(brief proposal.PreparedBrief) string {
	parts := make([]string, 0, 3)
	if text := strings.TrimSpace(brief.SourceSummary); text != "" {
		parts = append(parts, text)
	}
	if text := strings.TrimSpace(brief.Reason); text != "" {
		parts = append(parts, text)
	}
	if text := strings.TrimSpace(brief.SuggestedAcceptanceEffect); text != "" {
		parts = append(parts, text)
	}
	for _, evidence := range brief.SupportingEvidence {
		if text := strings.TrimSpace(evidence); text != "" {
			parts = append(parts, "Evidence: "+text)
		}
	}
	return strings.Join(parts, "\n")
}

func compactCollaborationSummary(view cockpit.CollaborationView) string {
	parts := make([]string, 0, 5)
	if text := strings.TrimSpace(view.ParticipantSummary); text != "" {
		parts = append(parts, "people: "+text)
	}
	if text := strings.TrimSpace(view.StateHint); text != "" {
		parts = append(parts, text)
	}
	if text := strings.TrimSpace(view.HandoffSummary); text != "" {
		parts = append(parts, text)
	}
	if text := strings.TrimSpace(view.BudgetHint); text != "" {
		parts = append(parts, text)
	}
	if text := strings.TrimSpace(view.RecoveryHint); text != "" {
		parts = append(parts, text)
	}
	return strings.Join(parts, " | ")
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
