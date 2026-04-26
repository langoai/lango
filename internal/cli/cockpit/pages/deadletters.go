package pages

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/postadjudicationstatus"
)

const (
	deadLettersTableMinRows   = 6
	deadLettersDetailMinLines = 8
	deadLettersTableGutterW   = 4
	deadLettersTxColW         = 20
	deadLettersAdjColW        = 14
	deadLettersAttemptColW    = 8
	deadLettersStatusColW     = 12
	deadLettersColumnGapW     = 1
	deadLetterTopSummaryLimit = 5
)

var deadLettersNow = func() time.Time {
	return time.Now().UTC()
}

type deadLettersLoadedMsg struct {
	items             []postadjudicationstatus.DeadLetterBacklogEntry
	err               error
	preserveSelection bool
	selectedID        string
}

type deadLetterDetailLoadedMsg struct {
	transactionID string
	status        postadjudicationstatus.TransactionStatus
	err           error
}

type deadLetterRetryResultMsg struct {
	transactionID string
	err           error
}

type deadLetterAdjudicationFilter string

const (
	deadLetterAdjudicationAll     deadLetterAdjudicationFilter = "all"
	deadLetterAdjudicationRelease deadLetterAdjudicationFilter = "release"
	deadLetterAdjudicationRefund  deadLetterAdjudicationFilter = "refund"
)

type deadLetterSubtypeFilter string

const (
	deadLetterSubtypeAll                  deadLetterSubtypeFilter = "all"
	deadLetterSubtypeRetryScheduled       deadLetterSubtypeFilter = "retry-scheduled"
	deadLetterSubtypeManualRetryRequested deadLetterSubtypeFilter = "manual-retry-requested"
	deadLetterSubtypeDeadLettered         deadLetterSubtypeFilter = "dead-lettered"
)

type deadLetterFamilyFilter string

const (
	deadLetterFamilyAll         deadLetterFamilyFilter = "all"
	deadLetterFamilyRetry       deadLetterFamilyFilter = "retry"
	deadLetterFamilyManualRetry deadLetterFamilyFilter = "manual-retry"
	deadLetterFamilyDeadLetter  deadLetterFamilyFilter = "dead-letter"
)

type deadLetterTextField int

const (
	deadLetterTextFieldQuery deadLetterTextField = iota
	deadLetterTextFieldManualReplayActor
	deadLetterTextFieldDeadLetteredAfter
	deadLetterTextFieldDeadLetteredBefore
	deadLetterTextFieldDeadLetterReasonQuery
	deadLetterTextFieldLatestDispatchReference
)

type DeadLetterListOptions struct {
	Query                     string
	Adjudication              string
	LatestStatusSubtype       string
	LatestStatusSubtypeFamily string
	AnyMatchFamily            string
	ManualReplayActor         string
	DeadLetteredAfter         string
	DeadLetteredBefore        string
	DeadLetterReasonQuery     string
	LatestDispatchReference   string
}

// DeadLetterListFn loads the current dead-letter backlog rows for the cockpit table.
type DeadLetterListFn func(ctx context.Context, opts DeadLetterListOptions) ([]postadjudicationstatus.DeadLetterBacklogEntry, error)

// DeadLetterDetailFn loads the selected transaction detail for the cockpit pane.
type DeadLetterDetailFn func(ctx context.Context, transactionReceiptID string) (postadjudicationstatus.TransactionStatus, error)
type DeadLetterRetryFn func(ctx context.Context, transactionReceiptID string) error

// DeadLettersPage renders a read-only master-detail surface for post-adjudication dead letters.
type DeadLettersPage struct {
	listFn   DeadLetterListFn
	detailFn DeadLetterDetailFn
	retryFn  DeadLetterRetryFn

	items                          []postadjudicationstatus.DeadLetterBacklogEntry
	cursor                         int
	selectedID                     string
	detail                         *postadjudicationstatus.TransactionStatus
	loadErr                        error
	detailErr                      error
	activeTextField                deadLetterTextField
	queryDraft                     string
	appliedQuery                   string
	manualReplayActorDraft         string
	appliedManualReplayActor       string
	deadLetteredAfterDraft         string
	appliedDeadLetteredAfter       string
	deadLetteredBeforeDraft        string
	appliedDeadLetteredBefore      string
	deadLetterReasonQueryDraft     string
	appliedDeadLetterReasonQuery   string
	latestDispatchReferenceDraft   string
	appliedLatestDispatchReference string
	adjudicationDraft              deadLetterAdjudicationFilter
	appliedAdjudication            deadLetterAdjudicationFilter
	subtypeDraft                   deadLetterSubtypeFilter
	appliedSubtype                 deadLetterSubtypeFilter
	familyDraft                    deadLetterFamilyFilter
	appliedFamily                  deadLetterFamilyFilter
	anyMatchFamilyDraft            deadLetterFamilyFilter
	appliedAnyMatchFamily          deadLetterFamilyFilter
	width, height                  int
	statusMsg                      string
	retryConfirmID                 string
	retryRunningID                 string
	retryFollowUpID                string
	retryFollowUpNote              string
	retryFollowUpPending           bool
}

type deadLetterSummary struct {
	total             int
	retryable         int
	release           int
	refund            int
	retryFamily       int
	manualRetryFamily int
	deadLetterFamily  int
	topReasons        []deadLetterReasonSummaryItem
	reasonFamilies    []deadLetterReasonFamilySummaryItem
	actorFamilies     []deadLetterActorFamilySummaryItem
	topActors         []deadLetterActorSummaryItem
	topDispatches     []deadLetterDispatchSummaryItem
	dispatchFamilies  []deadLetterDispatchFamilySummaryItem
	last24Hours       int
	previous24Hours   int
	last7Days         int
	previous7Days     int
	olderThan14Days   int
	undated           int
}

type deadLetterReasonSummaryItem struct {
	reason string
	count  int
}

type deadLetterReasonFamilySummaryItem struct {
	family string
	count  int
}

type deadLetterActorSummaryItem struct {
	actor string
	count int
}

type deadLetterActorFamilySummaryItem struct {
	family string
	count  int
}

type deadLetterDispatchSummaryItem struct {
	dispatchReference string
	count             int
}

type deadLetterDispatchFamilySummaryItem struct {
	family string
	count  int
}

func NewDeadLettersPage(listFn DeadLetterListFn, detailFn DeadLetterDetailFn, retryFns ...DeadLetterRetryFn) *DeadLettersPage {
	var retryFn DeadLetterRetryFn
	if len(retryFns) > 0 {
		retryFn = retryFns[0]
	}
	return &DeadLettersPage{
		listFn:                listFn,
		detailFn:              detailFn,
		retryFn:               retryFn,
		activeTextField:       deadLetterTextFieldQuery,
		adjudicationDraft:     deadLetterAdjudicationAll,
		appliedAdjudication:   deadLetterAdjudicationAll,
		subtypeDraft:          deadLetterSubtypeAll,
		appliedSubtype:        deadLetterSubtypeAll,
		familyDraft:           deadLetterFamilyAll,
		appliedFamily:         deadLetterFamilyAll,
		anyMatchFamilyDraft:   deadLetterFamilyAll,
		appliedAnyMatchFamily: deadLetterFamilyAll,
		appliedQuery:          "",
		queryDraft:            "",
	}
}

func (p *DeadLettersPage) Title() string { return "Dead Letters" }

func (p *DeadLettersPage) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "field")),
		key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "adj")),
		key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[/]", "subtype")),
		key.NewBinding(key.WithKeys(",", "."), key.WithHelp(",/.", "family")),
		key.NewBinding(key.WithKeys(";", "/"), key.WithHelp(";/", "any")),
		key.NewBinding(key.WithKeys("backspace"), key.WithHelp("⌫", "query")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "apply")),
		key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "reset")),
	}
	if p.canRetrySelected() {
		help := "retry"
		if p.retryConfirmActive() {
			help = "confirm"
		} else if p.retryRunningActive() {
			help = "running"
		}
		if !p.retryRunning() {
			bindings = append(bindings, key.NewBinding(key.WithKeys("r"), key.WithHelp("r", help)))
		}
	}
	return bindings
}

func (p *DeadLettersPage) Init() tea.Cmd { return nil }

func (p *DeadLettersPage) Activate() tea.Cmd {
	return p.loadBacklog()
}

func (p *DeadLettersPage) Deactivate() {}

func (p *DeadLettersPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	case deadLettersLoadedMsg:
		p.loadErr = msg.err
		if msg.err != nil {
			if p.retryFollowUpPending && strings.TrimSpace(p.retryFollowUpID) != "" {
				p.retryFollowUpPending = false
				p.retryFollowUpNote = fmt.Sprintf("refresh failed: %v", msg.err)
				p.statusMsg = p.retryAcceptedMessage(p.retryFollowUpID, p.retryFollowUpNote)
			}
			p.items = nil
			p.cursor = 0
			p.selectedID = ""
			p.detail = nil
			p.detailErr = nil
			p.retryConfirmID = ""
			p.retryRunningID = ""
			return p, nil
		}

		p.items = append([]postadjudicationstatus.DeadLetterBacklogEntry(nil), msg.items...)
		p.detailErr = nil
		p.retryConfirmID = ""
		if len(p.items) == 0 {
			p.cursor = 0
			p.selectedID = ""
			p.detail = nil
			p.retryRunningID = ""
			return p, nil
		}
		p.cursor = 0
		if msg.preserveSelection && msg.selectedID != "" {
			for i, item := range p.items {
				if item.TransactionReceiptID == msg.selectedID {
					p.cursor = i
					break
				}
			}
		}
		p.selectedID = p.items[p.cursor].TransactionReceiptID
		p.detail = nil
		if p.retryFollowUpPending && strings.TrimSpace(p.retryFollowUpID) != "" {
			retryID := strings.TrimSpace(p.retryFollowUpID)
			if !deadLetterBacklogContains(p.items, retryID) {
				p.retryFollowUpPending = false
				p.retryFollowUpNote = "no longer present in current dead-letter backlog"
				p.statusMsg = p.retryAcceptedMessage(retryID, p.retryFollowUpNote)
			} else if p.selectedID != retryID {
				p.retryFollowUpPending = false
				p.retryFollowUpNote = fmt.Sprintf(
					"%s is still in backlog; select it to inspect latest status",
					retryID,
				)
				p.statusMsg = p.retryAcceptedMessage(retryID, p.retryFollowUpNote)
			} else {
				p.retryFollowUpNote = "backlog refreshed; loading latest status"
				p.statusMsg = p.retryAcceptedMessage(retryID, p.retryFollowUpNote)
			}
		}
		return p, p.loadSelectedDetail()
	case deadLetterDetailLoadedMsg:
		if msg.transactionID != p.selectedID {
			return p, nil
		}
		p.detailErr = msg.err
		if msg.err != nil {
			p.detail = nil
			if p.retryFollowUpPending && p.retryFollowUpID == msg.transactionID {
				p.retryFollowUpPending = false
				p.retryFollowUpNote = fmt.Sprintf("latest status load failed: %v", msg.err)
				p.statusMsg = p.retryAcceptedMessage(msg.transactionID, p.retryFollowUpNote)
			}
			return p, nil
		}
		status := msg.status
		p.detail = &status
		if p.retryFollowUpPending && p.retryFollowUpID == msg.transactionID {
			p.retryFollowUpPending = false
			p.retryFollowUpNote = p.describeRetryFollowUp(status)
			p.statusMsg = p.retryAcceptedMessage(msg.transactionID, p.retryFollowUpNote)
		}
	case deadLetterRetryResultMsg:
		p.retryConfirmID = ""
		if p.retryRunningID == msg.transactionID {
			p.retryRunningID = ""
		}
		if msg.err != nil {
			p.retryFollowUpID = ""
			p.retryFollowUpNote = ""
			p.retryFollowUpPending = false
			p.statusMsg = p.retryFailureMessage(msg.transactionID, msg.err)
			return p, nil
		}
		p.retryFollowUpID = msg.transactionID
		p.retryFollowUpNote = ""
		p.retryFollowUpPending = true
		p.statusMsg = p.retrySuccessMessage(msg.transactionID)
		p.detailErr = nil
		return p, p.refreshAfterRetry()
	case tea.KeyMsg:
		p.statusMsg = ""
		if p.activeTextField != deadLetterTextFieldQuery &&
			isDeadLetterQueryInput(msg) &&
			msg.String() != "[" &&
			msg.String() != "]" &&
			msg.String() != "," &&
			msg.String() != "." &&
			msg.String() != ";" &&
			msg.String() != "/" {
			p.retryConfirmID = ""
			p.appendToActiveField(msg.String())
			return p, nil
		}
		switch msg.String() {
		case "enter":
			selectedID := p.selectedID
			p.retryConfirmID = ""
			p.appliedQuery = strings.TrimSpace(p.queryDraft)
			p.appliedManualReplayActor = strings.TrimSpace(p.manualReplayActorDraft)
			p.appliedDeadLetteredAfter = strings.TrimSpace(p.deadLetteredAfterDraft)
			p.appliedDeadLetteredBefore = strings.TrimSpace(p.deadLetteredBeforeDraft)
			p.appliedDeadLetterReasonQuery = strings.TrimSpace(p.deadLetterReasonQueryDraft)
			p.appliedLatestDispatchReference = strings.TrimSpace(p.latestDispatchReferenceDraft)
			p.appliedAdjudication = p.adjudicationDraft
			p.appliedSubtype = p.subtypeDraft
			p.appliedFamily = p.familyDraft
			p.appliedAnyMatchFamily = p.anyMatchFamilyDraft
			p.detail = nil
			p.detailErr = nil
			return p, p.reloadBacklogPreservingSelection(selectedID)
		case "ctrl+r":
			if p.retryRunning() {
				return p, nil
			}
			selectedID := p.selectedID
			p.resetFilters()
			p.detail = nil
			p.detailErr = nil
			return p, p.reloadBacklogPreservingSelection(selectedID)
		case "tab":
			p.retryConfirmID = ""
			p.activeTextField = p.activeTextField.next()
			return p, nil
		case "esc":
			p.retryConfirmID = ""
			return p, nil
		case "r":
			if p.retryRunning() {
				return p, nil
			}
			if p.retryConfirmActive() {
				p.retryConfirmID = ""
				p.retryRunningID = p.selectedID
				return p, p.retrySelected()
			}
			if !p.canRetrySelected() {
				return p, nil
			}
			p.retryConfirmID = p.selectedID
			return p, nil
		case "left":
			p.retryConfirmID = ""
			p.adjudicationDraft = p.adjudicationDraft.prev()
		case "right":
			p.retryConfirmID = ""
			p.adjudicationDraft = p.adjudicationDraft.next()
		case "[":
			p.retryConfirmID = ""
			p.subtypeDraft = p.subtypeDraft.prev()
		case "]":
			p.retryConfirmID = ""
			p.subtypeDraft = p.subtypeDraft.next()
		case ",":
			p.retryConfirmID = ""
			p.familyDraft = p.familyDraft.prev()
		case ".":
			p.retryConfirmID = ""
			p.familyDraft = p.familyDraft.next()
		case ";":
			p.retryConfirmID = ""
			p.anyMatchFamilyDraft = p.anyMatchFamilyDraft.prev()
		case "/":
			p.retryConfirmID = ""
			p.anyMatchFamilyDraft = p.anyMatchFamilyDraft.next()
		case "backspace":
			p.retryConfirmID = ""
			p.backspaceActiveField()
		case "up", "k":
			if p.cursor > 0 {
				p.retryConfirmID = ""
				p.cursor--
				p.selectedID = p.items[p.cursor].TransactionReceiptID
				p.detailErr = nil
				p.detail = nil
				return p, p.loadSelectedDetail()
			}
		case "down", "j":
			if p.cursor < len(p.items)-1 {
				p.retryConfirmID = ""
				p.cursor++
				p.selectedID = p.items[p.cursor].TransactionReceiptID
				p.detailErr = nil
				p.detail = nil
				return p, p.loadSelectedDetail()
			}
		default:
			if isDeadLetterQueryInput(msg) {
				p.retryConfirmID = ""
				p.appendToActiveField(msg.String())
			}
		}
	}
	return p, nil
}

func (p *DeadLettersPage) View() string {
	title := lipgloss.NewStyle().
		Foreground(theme.TextPrimary).
		Bold(true).
		Render("Post-Adjudication Dead Letters")

	divider := lipgloss.NewStyle().
		Foreground(theme.BorderDefault).
		Render(strings.Repeat("─", max(p.width-4, 48)))

	summary := p.renderSummaryStrip()
	filterBar := p.renderFilterBar()
	table := p.renderTable()
	detail := p.renderDetailPane()
	parts := []string{title, divider, "", summary, "", filterBar, "", table, "", detail}
	if strings.TrimSpace(p.statusMsg) != "" {
		status := lipgloss.NewStyle().Foreground(theme.Accent).Render(p.statusMsg)
		parts = append(parts, "", status)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (p *DeadLettersPage) renderSummaryStrip() string {
	summary := summarizeDeadLetters(p.items)
	chipStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
	overview := lipgloss.JoinHorizontal(
		lipgloss.Left,
		chipStyle.Render(fmt.Sprintf("dead letters: %d", summary.total)),
		chipStyle.Render("  •  "),
		chipStyle.Render(fmt.Sprintf("retryable: %d", summary.retryable)),
		chipStyle.Render("  •  "),
		chipStyle.Render(fmt.Sprintf("release/refund: %d/%d", summary.release, summary.refund)),
		chipStyle.Render("  •  "),
		chipStyle.Render(fmt.Sprintf("retry/manual/dead: %d/%d/%d", summary.retryFamily, summary.manualRetryFamily, summary.deadLetterFamily)),
	)
	if summary.total == 0 {
		return overview
	}

	lines := []string{overview}

	if len(summary.topReasons) > 0 {
		reasonParts := make([]string, 0, len(summary.topReasons))
		for _, item := range summary.topReasons {
			reasonParts = append(reasonParts, fmt.Sprintf("%s(%d)", item.reason, item.count))
		}

		reasonsLine := "reasons: " + strings.Join(reasonParts, ", ")
		if p.width > 0 {
			reasonsLine = ansi.Truncate(reasonsLine, max(p.width-8, 48), "…")
		}
		lines = append(lines, chipStyle.Render(reasonsLine))
	}

	if len(summary.reasonFamilies) > 0 {
		familyParts := make([]string, 0, len(summary.reasonFamilies))
		for _, item := range summary.reasonFamilies {
			familyParts = append(familyParts, fmt.Sprintf("%s(%d)", item.family, item.count))
		}
		familiesLine := "reason families: " + strings.Join(familyParts, ", ")
		if p.width > 0 {
			familiesLine = ansi.Truncate(familiesLine, max(p.width-8, 48), "…")
		}
		lines = append(lines, chipStyle.Render(familiesLine))
	}

	if len(summary.topActors) > 0 {
		actorParts := make([]string, 0, len(summary.topActors))
		for _, item := range summary.topActors {
			actorParts = append(actorParts, fmt.Sprintf("%s(%d)", item.actor, item.count))
		}
		actorsLine := "actors: " + strings.Join(actorParts, ", ")
		if p.width > 0 {
			actorsLine = ansi.Truncate(actorsLine, max(p.width-8, 48), "…")
		}
		lines = append(lines, chipStyle.Render(actorsLine))
	}

	if len(summary.actorFamilies) > 0 {
		familyParts := make([]string, 0, len(summary.actorFamilies))
		for _, item := range summary.actorFamilies {
			familyParts = append(familyParts, fmt.Sprintf("%s(%d)", item.family, item.count))
		}
		actorFamiliesLine := "actor families: " + strings.Join(familyParts, ", ")
		if p.width > 0 {
			actorFamiliesLine = ansi.Truncate(actorFamiliesLine, max(p.width-8, 48), "…")
		}
		lines = append(lines, chipStyle.Render(actorFamiliesLine))
	}

	if len(summary.dispatchFamilies) > 0 {
		familyParts := make([]string, 0, len(summary.dispatchFamilies))
		for _, item := range summary.dispatchFamilies {
			familyParts = append(familyParts, fmt.Sprintf("%s(%d)", item.family, item.count))
		}
		dispatchFamiliesLine := "dispatch families: " + strings.Join(familyParts, ", ")
		if p.width > 0 {
			dispatchFamiliesLine = ansi.Truncate(dispatchFamiliesLine, max(p.width-8, 48), "…")
		}
		lines = append(lines, chipStyle.Render(dispatchFamiliesLine))
	}

	if len(summary.topDispatches) > 0 {
		dispatchParts := make([]string, 0, len(summary.topDispatches))
		for _, item := range summary.topDispatches {
			dispatchParts = append(dispatchParts, fmt.Sprintf("%s(%d)", item.dispatchReference, item.count))
		}
		dispatchLine := "dispatch: " + strings.Join(dispatchParts, ", ")
		if p.width > 0 {
			dispatchLine = ansi.Truncate(dispatchLine, max(p.width-8, 48), "…")
		}
		lines = append(lines, chipStyle.Render(dispatchLine))
	}

	trendLine := fmt.Sprintf(
		"trend: 24h %d vs prev24h %d (%s), 7d %d vs prev7d %d (%s), older %d, undated %d",
		summary.last24Hours,
		summary.previous24Hours,
		formatDeadLetterDelta(summary.last24Hours-summary.previous24Hours),
		summary.last7Days,
		summary.previous7Days,
		formatDeadLetterDelta(summary.last7Days-summary.previous7Days),
		summary.olderThan14Days,
		summary.undated,
	)
	if p.width > 0 {
		trendLine = ansi.Truncate(trendLine, max(p.width-8, 48), "…")
	}
	lines = append(lines, chipStyle.Render(trendLine))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (p *DeadLettersPage) renderFilterBar() string {
	labelStyle := lipgloss.NewStyle().Foreground(theme.TextTertiary).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("Filters"),
		p.renderFilterLine("Query", p.queryDraft, p.activeTextField == deadLetterTextFieldQuery),
		p.renderFilterLine("Manual replay actor", p.manualReplayActorDraft, p.activeTextField == deadLetterTextFieldManualReplayActor),
		p.renderFilterLine("Dead-lettered after", p.deadLetteredAfterDraft, p.activeTextField == deadLetterTextFieldDeadLetteredAfter),
		p.renderFilterLine("Dead-lettered before", p.deadLetteredBeforeDraft, p.activeTextField == deadLetterTextFieldDeadLetteredBefore),
		p.renderFilterLine("Dead-letter reason", p.deadLetterReasonQueryDraft, p.activeTextField == deadLetterTextFieldDeadLetterReasonQuery),
		p.renderFilterLine("Dispatch reference", p.latestDispatchReferenceDraft, p.activeTextField == deadLetterTextFieldLatestDispatchReference),
		lipgloss.NewStyle().Foreground(theme.TextPrimary).Render(fmt.Sprintf("Adjudication: %s", p.adjudicationDraft)),
		lipgloss.NewStyle().Foreground(theme.TextPrimary).Render(fmt.Sprintf("Latest subtype: %s", p.subtypeDraft)),
		lipgloss.NewStyle().Foreground(theme.TextPrimary).Render(fmt.Sprintf("Latest family: %s", p.familyDraft)),
		lipgloss.NewStyle().Foreground(theme.TextPrimary).Render(fmt.Sprintf("Any-match family: %s", p.anyMatchFamilyDraft)),
		hintStyle.Render("Tab fields, type text, use ←/→ for adjudication, [/] for subtype, ,/. for family, ;/ for any-match, Enter to apply, Ctrl+R to reset"),
	)
}

func (p *DeadLettersPage) renderFilterLine(label string, value string, active bool) string {
	style := lipgloss.NewStyle().Foreground(theme.TextPrimary)
	prefix := "  "
	if active {
		style = style.Foreground(theme.Accent).Bold(true)
		prefix = "> "
	}
	displayValue := strings.TrimSpace(value)
	if displayValue == "" {
		displayValue = "all"
	}
	return style.Render(fmt.Sprintf("%s%s: %s", prefix, label, displayValue))
}

func (p *DeadLettersPage) renderTable() string {
	if p.loadErr != nil {
		return lipgloss.NewStyle().Foreground(theme.Error).Render(fmt.Sprintf("Failed to load dead letters: %v", p.loadErr))
	}
	if len(p.items) == 0 {
		if p.hasAppliedFilters() {
			return lipgloss.NewStyle().Foreground(theme.Muted).Render("No dead-letter backlog matches the current filters.")
		}
		return lipgloss.NewStyle().Foreground(theme.Muted).Render("No current dead-letter backlog.")
	}

	promptW := p.width - deadLettersTableGutterW - deadLettersTxColW - deadLettersAdjColW - deadLettersAttemptColW - deadLettersStatusColW - deadLettersColumnGapW*4 - 4
	if promptW < 12 {
		promptW = 12
	}
	format := fmt.Sprintf("%%-%ds %%-%ds %%-%ds %%-%ds %%s", deadLettersTxColW, promptW, deadLettersAdjColW, deadLettersAttemptColW)

	header := lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		Bold(true).
		PaddingLeft(deadLettersTableGutterW).
		Render(fmt.Sprintf(format, "Transaction", "Reason", "Adjudication", "Attempt", "Status"))

	separator := lipgloss.NewStyle().
		Foreground(theme.BorderSubtle).
		PaddingLeft(deadLettersTableGutterW).
		Render(strings.Repeat("─", max(p.width-8, 48)))

	rows := make([]string, 0, len(p.items)+2)
	rows = append(rows, header, separator)
	for i, entry := range p.items {
		line := fmt.Sprintf(
			format,
			tui.Truncate(entry.TransactionReceiptID, deadLettersTxColW),
			ansi.Truncate(entry.LatestDeadLetterReason, promptW, "…"),
			tui.Truncate(entry.Adjudication, deadLettersAdjColW),
			fmt.Sprintf("%d", entry.LatestRetryAttempt),
			tui.Truncate(deadLetterStatusLabel(entry), deadLettersStatusColW),
		)
		style := lipgloss.NewStyle().PaddingLeft(deadLettersTableGutterW).Foreground(theme.TextPrimary)
		prefix := "  "
		if i == p.cursor {
			style = style.Foreground(theme.Accent).Bold(true)
			prefix = "> "
		}
		rows = append(rows, style.Render(prefix+line))
	}

	if minRows := max(deadLettersTableMinRows, len(rows)); len(rows) < minRows {
		filler := lipgloss.NewStyle().PaddingLeft(deadLettersTableGutterW).Render("")
		for len(rows) < minRows {
			rows = append(rows, filler)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (p *DeadLettersPage) renderDetailPane() string {
	title := lipgloss.NewStyle().
		Foreground(theme.TextPrimary).
		Bold(true).
		Render("Selected Transaction")

	switch {
	case len(p.items) == 0:
		return lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.NewStyle().Foreground(theme.Muted).Render("Select a backlog row to inspect detail."))
	case p.detailErr != nil:
		return lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.NewStyle().Foreground(theme.Error).Render(fmt.Sprintf("Failed to load detail: %v", p.detailErr)))
	case p.detail == nil:
		return lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.NewStyle().Foreground(theme.Muted).Render("Loading selected transaction detail..."))
	}

	selected := p.items[p.cursor]
	detail := p.detail
	lines := []string{
		fmt.Sprintf("Transaction: %s", detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID),
		fmt.Sprintf("Submission: %s", detail.CanonicalSnapshot.SubmissionReceipt.SubmissionReceiptID),
		fmt.Sprintf("Adjudication: %s", detail.Adjudication),
		fmt.Sprintf("Dead-lettered: %t", detail.IsDeadLettered),
		fmt.Sprintf("Retryable: %t", detail.CanRetry),
		fmt.Sprintf("Latest reason: %s", fallback(detail.RetryDeadLetterSummary.LatestDeadLetterReason, selected.LatestDeadLetterReason, "n/a")),
		fmt.Sprintf("Latest retry attempt: %d", detail.RetryDeadLetterSummary.LatestRetryAttempt),
		fmt.Sprintf("Latest dispatch reference: %s", fallback(detail.RetryDeadLetterSummary.LatestDispatchReference, selected.LatestDispatchReference, "n/a")),
	}

	if detail.LatestBackgroundTask != nil {
		lines = append(lines,
			fmt.Sprintf("Background task: %s", detail.LatestBackgroundTask.TaskID),
			fmt.Sprintf("Task status: %s", detail.LatestBackgroundTask.Status),
			fmt.Sprintf("Task attempts: %d", detail.LatestBackgroundTask.AttemptCount),
			fmt.Sprintf("Next retry at: %s", fallback(detail.LatestBackgroundTask.NextRetryAt, "", "n/a")),
		)
	} else {
		lines = append(lines, "Background task: n/a")
	}
	retryState := p.retryActionLabel()
	lines = append(lines, fmt.Sprintf("Retry action: %s", retryState))
	if p.selectedID == p.retryFollowUpID && strings.TrimSpace(p.retryFollowUpNote) != "" {
		lines = append(lines, fmt.Sprintf("Retry follow-up: %s", p.retryFollowUpNote))
	}

	minLines := deadLettersDetailMinLines
	if len(lines) < minLines {
		for len(lines) < minLines {
			lines = append(lines, "")
		}
	}

	body := make([]string, 0, len(lines))
	for _, line := range lines {
		body = append(body, lipgloss.NewStyle().Foreground(theme.TextSecondary).Render(line))
	}
	return lipgloss.JoinVertical(lipgloss.Left, append([]string{title}, body...)...)
}

func (p *DeadLettersPage) loadBacklog() tea.Cmd {
	return p.loadBacklogWithSelection("", false)
}

func (p *DeadLettersPage) reloadBacklogPreservingSelection(selectedID string) tea.Cmd {
	if strings.TrimSpace(selectedID) == "" {
		return p.loadBacklog()
	}
	return p.loadBacklogWithSelection(selectedID, true)
}

func (p *DeadLettersPage) loadBacklogWithSelection(selectedID string, preserveSelection bool) tea.Cmd {
	listFn := p.listFn
	opts := DeadLetterListOptions{
		Query: p.appliedQuery,
	}
	if p.appliedAdjudication != deadLetterAdjudicationAll {
		opts.Adjudication = string(p.appliedAdjudication)
	}
	if p.appliedSubtype != deadLetterSubtypeAll {
		opts.LatestStatusSubtype = string(p.appliedSubtype)
	}
	if p.appliedFamily != deadLetterFamilyAll {
		opts.LatestStatusSubtypeFamily = string(p.appliedFamily)
	}
	if p.appliedAnyMatchFamily != deadLetterFamilyAll {
		opts.AnyMatchFamily = string(p.appliedAnyMatchFamily)
	}
	if p.appliedManualReplayActor != "" {
		opts.ManualReplayActor = p.appliedManualReplayActor
	}
	if p.appliedDeadLetteredAfter != "" {
		opts.DeadLetteredAfter = p.appliedDeadLetteredAfter
	}
	if p.appliedDeadLetteredBefore != "" {
		opts.DeadLetteredBefore = p.appliedDeadLetteredBefore
	}
	if p.appliedDeadLetterReasonQuery != "" {
		opts.DeadLetterReasonQuery = p.appliedDeadLetterReasonQuery
	}
	if p.appliedLatestDispatchReference != "" {
		opts.LatestDispatchReference = p.appliedLatestDispatchReference
	}
	return func() tea.Msg {
		if listFn == nil {
			return deadLettersLoadedMsg{err: fmt.Errorf("dead-letter list function not configured")}
		}
		items, err := listFn(context.Background(), opts)
		return deadLettersLoadedMsg{
			items:             items,
			err:               err,
			preserveSelection: preserveSelection,
			selectedID:        selectedID,
		}
	}
}

func (p *DeadLettersPage) loadSelectedDetail() tea.Cmd {
	detailFn := p.detailFn
	transactionID := p.selectedID
	return func() tea.Msg {
		if detailFn == nil {
			return deadLetterDetailLoadedMsg{
				transactionID: transactionID,
				err:           fmt.Errorf("dead-letter detail function not configured"),
			}
		}
		status, err := detailFn(context.Background(), transactionID)
		return deadLetterDetailLoadedMsg{
			transactionID: transactionID,
			status:        status,
			err:           err,
		}
	}
}

func (p *DeadLettersPage) retrySelected() tea.Cmd {
	if p.retryFn == nil || !p.canRetrySelected() {
		return nil
	}
	transactionID := p.selectedID
	retryFn := p.retryFn
	return func() tea.Msg {
		err := retryFn(context.Background(), transactionID)
		return deadLetterRetryResultMsg{transactionID: transactionID, err: err}
	}
}

func (p *DeadLettersPage) refreshAfterRetry() tea.Cmd {
	return p.reloadBacklogPreservingSelection(p.selectedID)
}

func (p *DeadLettersPage) retryActionLabel() string {
	if !p.canRetrySelected() {
		return "disabled"
	}
	if p.retryRunningActive() {
		return "requesting retry..."
	}
	if p.retryConfirmActive() {
		return "confirm request (press r again)"
	}
	return "ready (press r to request retry)"
}

func (p *DeadLettersPage) retrySuccessMessage(transactionID string) string {
	return fmt.Sprintf("Retry request accepted for %s. Refreshing backlog and detail.", transactionID)
}

func (p *DeadLettersPage) retryFailureMessage(transactionID string, err error) string {
	return fmt.Sprintf("Retry request failed for %s: %v", transactionID, err)
}

func (p *DeadLettersPage) retryAcceptedMessage(transactionID string, note string) string {
	trimmed := strings.TrimSpace(note)
	if trimmed == "" {
		return p.retrySuccessMessage(transactionID)
	}
	return fmt.Sprintf("Retry request accepted for %s. Follow-up: %s.", transactionID, trimmed)
}

func (p *DeadLettersPage) describeRetryFollowUp(
	status postadjudicationstatus.TransactionStatus,
) string {
	if !status.IsDeadLettered {
		return "cleared from dead-letter state"
	}

	if task := status.LatestBackgroundTask; task != nil {
		parts := []string{fmt.Sprintf("task %s", fallback(task.Status, "", "unknown"))}
		details := make([]string, 0, 2)
		if task.AttemptCount > 0 {
			details = append(details, fmt.Sprintf("attempt %d", task.AttemptCount))
		}
		if strings.TrimSpace(task.NextRetryAt) != "" {
			details = append(details, fmt.Sprintf("next %s", task.NextRetryAt))
		}
		if len(details) > 0 {
			parts[0] += " (" + strings.Join(details, ", ") + ")"
		}
		return strings.Join(parts, "")
	}

	parts := make([]string, 0, 3)
	if subtype := strings.TrimSpace(status.RetryDeadLetterSummary.LatestStatusSubtype); subtype != "" {
		parts = append(parts, fmt.Sprintf("subtype %s", subtype))
	}
	if family := strings.TrimSpace(status.RetryDeadLetterSummary.LatestStatusSubtypeFamily); family != "" {
		parts = append(parts, fmt.Sprintf("family %s", family))
	}
	if attempt := status.RetryDeadLetterSummary.LatestRetryAttempt; attempt > 0 {
		parts = append(parts, fmt.Sprintf("attempt %d", attempt))
	}
	if len(parts) == 0 {
		return "still present in dead-letter backlog"
	}
	return "still present in dead-letter backlog (" + strings.Join(parts, ", ") + ")"
}

func (f deadLetterAdjudicationFilter) next() deadLetterAdjudicationFilter {
	switch f {
	case deadLetterAdjudicationRelease:
		return deadLetterAdjudicationRefund
	case deadLetterAdjudicationRefund:
		return deadLetterAdjudicationAll
	default:
		return deadLetterAdjudicationRelease
	}
}

func (f deadLetterAdjudicationFilter) prev() deadLetterAdjudicationFilter {
	switch f {
	case deadLetterAdjudicationRefund:
		return deadLetterAdjudicationRelease
	case deadLetterAdjudicationRelease:
		return deadLetterAdjudicationAll
	default:
		return deadLetterAdjudicationRefund
	}
}

func (f deadLetterSubtypeFilter) next() deadLetterSubtypeFilter {
	switch f {
	case deadLetterSubtypeRetryScheduled:
		return deadLetterSubtypeManualRetryRequested
	case deadLetterSubtypeManualRetryRequested:
		return deadLetterSubtypeDeadLettered
	case deadLetterSubtypeDeadLettered:
		return deadLetterSubtypeAll
	default:
		return deadLetterSubtypeRetryScheduled
	}
}

func (f deadLetterSubtypeFilter) prev() deadLetterSubtypeFilter {
	switch f {
	case deadLetterSubtypeDeadLettered:
		return deadLetterSubtypeManualRetryRequested
	case deadLetterSubtypeManualRetryRequested:
		return deadLetterSubtypeRetryScheduled
	case deadLetterSubtypeRetryScheduled:
		return deadLetterSubtypeAll
	default:
		return deadLetterSubtypeDeadLettered
	}
}

func (f deadLetterFamilyFilter) next() deadLetterFamilyFilter {
	switch f {
	case deadLetterFamilyRetry:
		return deadLetterFamilyManualRetry
	case deadLetterFamilyManualRetry:
		return deadLetterFamilyDeadLetter
	case deadLetterFamilyDeadLetter:
		return deadLetterFamilyAll
	default:
		return deadLetterFamilyRetry
	}
}

func (f deadLetterFamilyFilter) prev() deadLetterFamilyFilter {
	switch f {
	case deadLetterFamilyDeadLetter:
		return deadLetterFamilyManualRetry
	case deadLetterFamilyManualRetry:
		return deadLetterFamilyRetry
	case deadLetterFamilyRetry:
		return deadLetterFamilyAll
	default:
		return deadLetterFamilyDeadLetter
	}
}

func (f deadLetterTextField) next() deadLetterTextField {
	switch f {
	case deadLetterTextFieldManualReplayActor:
		return deadLetterTextFieldDeadLetteredAfter
	case deadLetterTextFieldDeadLetteredAfter:
		return deadLetterTextFieldDeadLetteredBefore
	case deadLetterTextFieldDeadLetteredBefore:
		return deadLetterTextFieldDeadLetterReasonQuery
	case deadLetterTextFieldDeadLetterReasonQuery:
		return deadLetterTextFieldLatestDispatchReference
	case deadLetterTextFieldLatestDispatchReference:
		return deadLetterTextFieldQuery
	default:
		return deadLetterTextFieldManualReplayActor
	}
}

func isDeadLetterQueryInput(msg tea.KeyMsg) bool {
	if len(msg.Runes) == 0 {
		return false
	}
	for _, r := range msg.Runes {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func (p *DeadLettersPage) hasAppliedFilters() bool {
	return strings.TrimSpace(p.appliedQuery) != "" ||
		strings.TrimSpace(p.appliedManualReplayActor) != "" ||
		strings.TrimSpace(p.appliedDeadLetteredAfter) != "" ||
		strings.TrimSpace(p.appliedDeadLetteredBefore) != "" ||
		strings.TrimSpace(p.appliedDeadLetterReasonQuery) != "" ||
		strings.TrimSpace(p.appliedLatestDispatchReference) != "" ||
		p.appliedAdjudication != deadLetterAdjudicationAll ||
		p.appliedSubtype != deadLetterSubtypeAll ||
		p.appliedFamily != deadLetterFamilyAll ||
		p.appliedAnyMatchFamily != deadLetterFamilyAll
}

func (p *DeadLettersPage) appendToActiveField(value string) {
	switch p.activeTextField {
	case deadLetterTextFieldManualReplayActor:
		p.manualReplayActorDraft += value
	case deadLetterTextFieldDeadLetteredAfter:
		p.deadLetteredAfterDraft += value
	case deadLetterTextFieldDeadLetteredBefore:
		p.deadLetteredBeforeDraft += value
	case deadLetterTextFieldDeadLetterReasonQuery:
		p.deadLetterReasonQueryDraft += value
	case deadLetterTextFieldLatestDispatchReference:
		p.latestDispatchReferenceDraft += value
	default:
		p.queryDraft += value
	}
}

func (p *DeadLettersPage) backspaceActiveField() {
	switch p.activeTextField {
	case deadLetterTextFieldManualReplayActor:
		p.manualReplayActorDraft = trimLastByte(p.manualReplayActorDraft)
	case deadLetterTextFieldDeadLetteredAfter:
		p.deadLetteredAfterDraft = trimLastByte(p.deadLetteredAfterDraft)
	case deadLetterTextFieldDeadLetteredBefore:
		p.deadLetteredBeforeDraft = trimLastByte(p.deadLetteredBeforeDraft)
	case deadLetterTextFieldDeadLetterReasonQuery:
		p.deadLetterReasonQueryDraft = trimLastByte(p.deadLetterReasonQueryDraft)
	case deadLetterTextFieldLatestDispatchReference:
		p.latestDispatchReferenceDraft = trimLastByte(p.latestDispatchReferenceDraft)
	default:
		p.queryDraft = trimLastByte(p.queryDraft)
	}
}

func (p *DeadLettersPage) resetFilters() {
	p.queryDraft = ""
	p.appliedQuery = ""
	p.manualReplayActorDraft = ""
	p.appliedManualReplayActor = ""
	p.deadLetteredAfterDraft = ""
	p.appliedDeadLetteredAfter = ""
	p.deadLetteredBeforeDraft = ""
	p.appliedDeadLetteredBefore = ""
	p.deadLetterReasonQueryDraft = ""
	p.appliedDeadLetterReasonQuery = ""
	p.latestDispatchReferenceDraft = ""
	p.appliedLatestDispatchReference = ""
	p.adjudicationDraft = deadLetterAdjudicationAll
	p.appliedAdjudication = deadLetterAdjudicationAll
	p.subtypeDraft = deadLetterSubtypeAll
	p.appliedSubtype = deadLetterSubtypeAll
	p.familyDraft = deadLetterFamilyAll
	p.appliedFamily = deadLetterFamilyAll
	p.anyMatchFamilyDraft = deadLetterFamilyAll
	p.appliedAnyMatchFamily = deadLetterFamilyAll
	p.activeTextField = deadLetterTextFieldQuery
	p.retryConfirmID = ""
}

func trimLastByte(value string) string {
	if value == "" {
		return ""
	}
	return value[:len(value)-1]
}

func (p *DeadLettersPage) canRetrySelected() bool {
	return p.retryFn != nil && p.detail != nil && p.detail.CanRetry
}

func (p *DeadLettersPage) retryConfirmActive() bool {
	return p.canRetrySelected() && p.retryConfirmID != "" && p.retryConfirmID == p.selectedID
}

func (p *DeadLettersPage) retryRunning() bool {
	return p.retryRunningID != ""
}

func (p *DeadLettersPage) retryRunningActive() bool {
	return p.canRetrySelected() && p.retryRunningID != "" && p.retryRunningID == p.selectedID
}

func deadLetterStatusLabel(entry postadjudicationstatus.DeadLetterBacklogEntry) string {
	switch {
	case entry.CanRetry:
		return "retryable"
	case entry.IsDeadLettered:
		return "dead-lettered"
	default:
		return "inactive"
	}
}

func fallback(value string, fallbackValue string, empty string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	if strings.TrimSpace(fallbackValue) != "" {
		return fallbackValue
	}
	return empty
}

func summarizeDeadLetters(items []postadjudicationstatus.DeadLetterBacklogEntry) deadLetterSummary {
	var summary deadLetterSummary
	reasonCounts := make(map[string]int)
	reasonFamilyCounts := make(map[string]int)
	actorFamilyCounts := make(map[string]int)
	actorCounts := make(map[string]int)
	dispatchCounts := make(map[string]int)
	dispatchFamilyCounts := make(map[string]int)
	now := deadLettersNow()
	for _, item := range items {
		summary.total++
		if item.CanRetry {
			summary.retryable++
		}
		switch strings.ToLower(strings.TrimSpace(item.Adjudication)) {
		case "release":
			summary.release++
		case "refund":
			summary.refund++
		}
		switch strings.ToLower(strings.TrimSpace(item.LatestStatusSubtypeFamily)) {
		case "retry":
			summary.retryFamily++
		case "manual-retry":
			summary.manualRetryFamily++
		case "dead-letter":
			summary.deadLetterFamily++
		}
		reason := strings.TrimSpace(item.LatestDeadLetterReason)
		if reason != "" {
			reasonCounts[reason]++
		}
		reasonFamily := postadjudicationstatus.ClassifyDeadLetterReasonFamily(item.LatestDeadLetterReason)
		reasonFamilyCounts[reasonFamily]++
		actorFamily := postadjudicationstatus.ClassifyManualReplayActorFamily(item.LatestManualReplayActor)
		actorFamilyCounts[actorFamily]++
		actor := strings.TrimSpace(item.LatestManualReplayActor)
		if actor != "" {
			actorCounts[actor]++
		}
		dispatchReference := strings.TrimSpace(item.LatestDispatchReference)
		if dispatchReference != "" {
			dispatchCounts[dispatchReference]++
			dispatchFamilyCounts[postadjudicationstatus.ClassifyDispatchReferenceFamily(dispatchReference)]++
		}

		deadLetteredAt := parseDeadLetterTimestamp(strings.TrimSpace(item.LatestDeadLetteredAt))
		switch {
		case deadLetteredAt.IsZero():
			summary.undated++
		default:
			age := now.Sub(deadLetteredAt.UTC())
			switch {
			case age <= 24*time.Hour:
				summary.last24Hours++
				summary.last7Days++
			case age <= 48*time.Hour:
				summary.previous24Hours++
				summary.last7Days++
			case age <= 7*24*time.Hour:
				summary.last7Days++
			case age <= 14*24*time.Hour:
				summary.previous7Days++
			default:
				summary.olderThan14Days++
			}
		}
	}
	if len(reasonCounts) > 0 {
		reasons := make([]deadLetterReasonSummaryItem, 0, len(reasonCounts))
		for reason, count := range reasonCounts {
			reasons = append(reasons, deadLetterReasonSummaryItem{reason: reason, count: count})
		}
		sort.Slice(reasons, func(i, j int) bool {
			if reasons[i].count != reasons[j].count {
				return reasons[i].count > reasons[j].count
			}
			return reasons[i].reason < reasons[j].reason
		})
		if len(reasons) > deadLetterTopSummaryLimit {
			reasons = reasons[:deadLetterTopSummaryLimit]
		}
		summary.topReasons = reasons
	}
	if len(reasonFamilyCounts) > 0 {
		preferredOrder := []string{
			postadjudicationstatus.DeadLetterReasonFamilyRetryExhausted,
			postadjudicationstatus.DeadLetterReasonFamilyPolicyBlocked,
			postadjudicationstatus.DeadLetterReasonFamilyReceiptInvalid,
			postadjudicationstatus.DeadLetterReasonFamilyBackgroundFailed,
			postadjudicationstatus.DeadLetterReasonFamilyUnknown,
		}
		families := make([]deadLetterReasonFamilySummaryItem, 0, len(reasonFamilyCounts))
		for _, family := range preferredOrder {
			count := reasonFamilyCounts[family]
			if count == 0 {
				continue
			}
			families = append(families, deadLetterReasonFamilySummaryItem{
				family: family,
				count:  count,
			})
		}
		summary.reasonFamilies = families
	}
	if len(actorFamilyCounts) > 0 {
		preferredOrder := []string{
			postadjudicationstatus.ManualReplayActorFamilyOperator,
			postadjudicationstatus.ManualReplayActorFamilySystem,
			postadjudicationstatus.ManualReplayActorFamilyService,
			postadjudicationstatus.ManualReplayActorFamilyUnknown,
		}
		families := make([]deadLetterActorFamilySummaryItem, 0, len(actorFamilyCounts))
		for _, family := range preferredOrder {
			count := actorFamilyCounts[family]
			if count == 0 {
				continue
			}
			families = append(families, deadLetterActorFamilySummaryItem{
				family: family,
				count:  count,
			})
		}
		summary.actorFamilies = families
	}
	if len(actorCounts) > 0 {
		actors := make([]deadLetterActorSummaryItem, 0, len(actorCounts))
		for actor, count := range actorCounts {
			actors = append(actors, deadLetterActorSummaryItem{actor: actor, count: count})
		}
		sort.Slice(actors, func(i, j int) bool {
			if actors[i].count != actors[j].count {
				return actors[i].count > actors[j].count
			}
			return actors[i].actor < actors[j].actor
		})
		if len(actors) > deadLetterTopSummaryLimit {
			actors = actors[:deadLetterTopSummaryLimit]
		}
		summary.topActors = actors
	}
	if len(dispatchCounts) > 0 {
		dispatches := make([]deadLetterDispatchSummaryItem, 0, len(dispatchCounts))
		for dispatchReference, count := range dispatchCounts {
			dispatches = append(dispatches, deadLetterDispatchSummaryItem{dispatchReference: dispatchReference, count: count})
		}
		sort.Slice(dispatches, func(i, j int) bool {
			if dispatches[i].count != dispatches[j].count {
				return dispatches[i].count > dispatches[j].count
			}
			return dispatches[i].dispatchReference < dispatches[j].dispatchReference
		})
		if len(dispatches) > deadLetterTopSummaryLimit {
			dispatches = dispatches[:deadLetterTopSummaryLimit]
		}
		summary.topDispatches = dispatches
	}
	if len(dispatchFamilyCounts) > 0 {
		families := make([]deadLetterDispatchFamilySummaryItem, 0, len(dispatchFamilyCounts))
		for family, count := range dispatchFamilyCounts {
			families = append(families, deadLetterDispatchFamilySummaryItem{
				family: family,
				count:  count,
			})
		}
		sort.Slice(families, func(i, j int) bool {
			if families[i].count != families[j].count {
				return families[i].count > families[j].count
			}
			return families[i].family < families[j].family
		})
		if len(families) > deadLetterTopSummaryLimit {
			families = families[:deadLetterTopSummaryLimit]
		}
		summary.dispatchFamilies = families
	}
	return summary
}

func parseDeadLetterTimestamp(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func formatDeadLetterDelta(delta int) string {
	if delta >= 0 {
		return fmt.Sprintf("+%d", delta)
	}
	return fmt.Sprintf("%d", delta)
}

func deadLetterBacklogContains(
	items []postadjudicationstatus.DeadLetterBacklogEntry,
	transactionID string,
) bool {
	for _, item := range items {
		if item.TransactionReceiptID == transactionID {
			return true
		}
	}
	return false
}
