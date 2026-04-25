package pages

import (
	"context"
	"fmt"
	"strings"
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
)

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
}

type deadLetterSummary struct {
	total             int
	retryable         int
	release           int
	refund            int
	retryFamily       int
	manualRetryFamily int
	deadLetterFamily  int
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
		return p, p.loadSelectedDetail()
	case deadLetterDetailLoadedMsg:
		if msg.transactionID != p.selectedID {
			return p, nil
		}
		p.detailErr = msg.err
		if msg.err != nil {
			p.detail = nil
			return p, nil
		}
		status := msg.status
		p.detail = &status
	case deadLetterRetryResultMsg:
		p.retryConfirmID = ""
		if p.retryRunningID == msg.transactionID {
			p.retryRunningID = ""
		}
		if msg.err != nil {
			p.statusMsg = p.retryFailureMessage(msg.transactionID, msg.err)
			return p, nil
		}
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
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		chipStyle.Render(fmt.Sprintf("dead letters: %d", summary.total)),
		chipStyle.Render("  •  "),
		chipStyle.Render(fmt.Sprintf("retryable: %d", summary.retryable)),
		chipStyle.Render("  •  "),
		chipStyle.Render(fmt.Sprintf("release/refund: %d/%d", summary.release, summary.refund)),
		chipStyle.Render("  •  "),
		chipStyle.Render(fmt.Sprintf("retry/manual/dead: %d/%d/%d", summary.retryFamily, summary.manualRetryFamily, summary.deadLetterFamily)),
	)
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
	}
	return summary
}
