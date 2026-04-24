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

type DeadLetterListOptions struct {
	Query               string
	Adjudication        string
	LatestStatusSubtype string
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

	items               []postadjudicationstatus.DeadLetterBacklogEntry
	cursor              int
	selectedID          string
	detail              *postadjudicationstatus.TransactionStatus
	loadErr             error
	detailErr           error
	queryDraft          string
	appliedQuery        string
	adjudicationDraft   deadLetterAdjudicationFilter
	appliedAdjudication deadLetterAdjudicationFilter
	subtypeDraft        deadLetterSubtypeFilter
	appliedSubtype      deadLetterSubtypeFilter
	width, height       int
	statusMsg           string
	retryConfirmID      string
}

func NewDeadLettersPage(listFn DeadLetterListFn, detailFn DeadLetterDetailFn, retryFns ...DeadLetterRetryFn) *DeadLettersPage {
	var retryFn DeadLetterRetryFn
	if len(retryFns) > 0 {
		retryFn = retryFns[0]
	}
	return &DeadLettersPage{
		listFn:              listFn,
		detailFn:            detailFn,
		retryFn:             retryFn,
		adjudicationDraft:   deadLetterAdjudicationAll,
		appliedAdjudication: deadLetterAdjudicationAll,
		subtypeDraft:        deadLetterSubtypeAll,
		appliedSubtype:      deadLetterSubtypeAll,
		appliedQuery:        "",
		queryDraft:          "",
	}
}

func (p *DeadLettersPage) Title() string { return "Dead Letters" }

func (p *DeadLettersPage) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "adj")),
		key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[/]", "subtype")),
		key.NewBinding(key.WithKeys("backspace"), key.WithHelp("⌫", "query")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "apply")),
	}
	if p.canRetrySelected() {
		help := "retry"
		if p.retryConfirmActive() {
			help = "confirm"
		}
		bindings = append(bindings, key.NewBinding(key.WithKeys("r"), key.WithHelp("r", help)))
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
			return p, nil
		}

		p.items = append([]postadjudicationstatus.DeadLetterBacklogEntry(nil), msg.items...)
		p.detailErr = nil
		p.retryConfirmID = ""
		if len(p.items) == 0 {
			p.cursor = 0
			p.selectedID = ""
			p.detail = nil
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
		if msg.err != nil {
			p.statusMsg = "Retry failed: " + msg.err.Error()
			return p, nil
		}
		p.statusMsg = "Retry requested: " + msg.transactionID
		p.detailErr = nil
		return p, p.refreshAfterRetry(msg.transactionID)
	case tea.KeyMsg:
		p.statusMsg = ""
		switch msg.String() {
		case "enter":
			p.retryConfirmID = ""
			p.appliedQuery = strings.TrimSpace(p.queryDraft)
			p.appliedAdjudication = p.adjudicationDraft
			p.appliedSubtype = p.subtypeDraft
			p.detail = nil
			p.detailErr = nil
			return p, p.loadBacklog()
		case "esc":
			p.retryConfirmID = ""
			return p, nil
		case "r":
			if p.retryConfirmActive() {
				p.retryConfirmID = ""
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
		case "backspace":
			p.retryConfirmID = ""
			if p.queryDraft != "" {
				p.queryDraft = p.queryDraft[:len(p.queryDraft)-1]
			}
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
				p.queryDraft += msg.String()
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

	filterBar := p.renderFilterBar()
	table := p.renderTable()
	detail := p.renderDetailPane()
	parts := []string{title, divider, "", filterBar, "", table, "", detail}
	if strings.TrimSpace(p.statusMsg) != "" {
		status := lipgloss.NewStyle().Foreground(theme.Accent).Render(p.statusMsg)
		parts = append(parts, "", status)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (p *DeadLettersPage) renderFilterBar() string {
	labelStyle := lipgloss.NewStyle().Foreground(theme.TextTertiary).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(theme.TextPrimary)
	hintStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	query := p.queryDraft
	if strings.TrimSpace(query) == "" {
		query = "all"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("Filters"),
		valueStyle.Render(fmt.Sprintf("Query: %s", query)),
		valueStyle.Render(fmt.Sprintf("Adjudication: %s", p.adjudicationDraft)),
		valueStyle.Render(fmt.Sprintf("Latest subtype: %s", p.subtypeDraft)),
		hintStyle.Render("Type query, use ←/→ for adjudication, [/] for subtype, Enter to apply"),
	)
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
	retryState := "disabled"
	if p.canRetrySelected() {
		retryState = "enabled (press r)"
		if p.retryConfirmActive() {
			retryState = "confirm (press r again)"
		}
	}
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

func (p *DeadLettersPage) refreshAfterRetry(transactionID string) tea.Cmd {
	return p.loadBacklogWithSelection(transactionID, true)
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
		p.appliedAdjudication != deadLetterAdjudicationAll ||
		p.appliedSubtype != deadLetterSubtypeAll
}

func (p *DeadLettersPage) canRetrySelected() bool {
	return p.retryFn != nil && p.detail != nil && p.detail.CanRetry
}

func (p *DeadLettersPage) retryConfirmActive() bool {
	return p.canRetrySelected() && p.retryConfirmID != "" && p.retryConfirmID == p.selectedID
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
