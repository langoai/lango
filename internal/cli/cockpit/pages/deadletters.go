package pages

import (
	"context"
	"fmt"
	"reflect"
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
	items []postadjudicationstatus.DeadLetterBacklogEntry
	err   error
}

type deadLetterDetailLoadedMsg struct {
	transactionID string
	status        postadjudicationstatus.TransactionStatus
	err           error
}

type deadLetterAdjudicationFilter string

const (
	deadLetterAdjudicationAll     deadLetterAdjudicationFilter = "all"
	deadLetterAdjudicationRelease deadLetterAdjudicationFilter = "release"
	deadLetterAdjudicationRefund  deadLetterAdjudicationFilter = "refund"
)

type DeadLetterListOptions struct {
	Query        string
	Adjudication string
}

// DeadLetterListFn loads the current dead-letter backlog rows for the cockpit table.
// DeadLetterDetailFn loads the selected transaction detail for the cockpit pane.
type DeadLetterDetailFn func(ctx context.Context, transactionReceiptID string) (postadjudicationstatus.TransactionStatus, error)

// DeadLettersPage renders a read-only master-detail surface for post-adjudication dead letters.
type DeadLettersPage struct {
	listFn   any
	detailFn DeadLetterDetailFn

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
	width, height       int
}

func NewDeadLettersPage(listFn any, detailFn DeadLetterDetailFn) *DeadLettersPage {
	return &DeadLettersPage{
		listFn:              listFn,
		detailFn:            detailFn,
		adjudicationDraft:   deadLetterAdjudicationAll,
		appliedAdjudication: deadLetterAdjudicationAll,
		appliedQuery:        "",
		queryDraft:          "",
	}
}

func (p *DeadLettersPage) Title() string { return "Dead Letters" }

func (p *DeadLettersPage) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "adj")),
		key.NewBinding(key.WithKeys("backspace"), key.WithHelp("⌫", "query")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "apply")),
	}
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
			return p, nil
		}

		p.items = append([]postadjudicationstatus.DeadLetterBacklogEntry(nil), msg.items...)
		p.cursor = 0
		p.detailErr = nil
		if len(p.items) == 0 {
			p.selectedID = ""
			p.detail = nil
			return p, nil
		}
		p.selectedID = p.items[0].TransactionReceiptID
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
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			p.appliedQuery = strings.TrimSpace(p.queryDraft)
			p.appliedAdjudication = p.adjudicationDraft
			p.detail = nil
			p.detailErr = nil
			return p, p.loadBacklog()
		case "left":
			p.adjudicationDraft = p.adjudicationDraft.prev()
		case "right":
			p.adjudicationDraft = p.adjudicationDraft.next()
		case "backspace":
			if p.queryDraft != "" {
				p.queryDraft = p.queryDraft[:len(p.queryDraft)-1]
			}
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
				p.selectedID = p.items[p.cursor].TransactionReceiptID
				p.detailErr = nil
				p.detail = nil
				return p, p.loadSelectedDetail()
			}
		case "down", "j":
			if p.cursor < len(p.items)-1 {
				p.cursor++
				p.selectedID = p.items[p.cursor].TransactionReceiptID
				p.detailErr = nil
				p.detail = nil
				return p, p.loadSelectedDetail()
			}
		default:
			if isDeadLetterQueryInput(msg) {
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
	content := lipgloss.JoinVertical(lipgloss.Left, title, divider, "", filterBar, "", table, "", detail)
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
		hintStyle.Render("Type query, use ←/→ to change adjudication, Enter to apply"),
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
	listFn := p.listFn
	opts := DeadLetterListOptions{
		Query: p.appliedQuery,
	}
	if p.appliedAdjudication != deadLetterAdjudicationAll {
		opts.Adjudication = string(p.appliedAdjudication)
	}
	return func() tea.Msg {
		items, err := invokeDeadLetterList(listFn, context.Background(), opts)
		return deadLettersLoadedMsg{items: items, err: err}
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
	return strings.TrimSpace(p.appliedQuery) != "" || p.appliedAdjudication != deadLetterAdjudicationAll
}

func invokeDeadLetterList(listFn any, ctx context.Context, opts DeadLetterListOptions) ([]postadjudicationstatus.DeadLetterBacklogEntry, error) {
	if listFn == nil {
		return nil, fmt.Errorf("dead-letter list function not configured")
	}
	fnValue := reflect.ValueOf(listFn)
	if fnValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("dead-letter list function is invalid")
	}
	fnType := fnValue.Type()
	if fnType.NumIn() != 2 || fnType.NumOut() != 2 {
		return nil, fmt.Errorf("dead-letter list function signature is invalid")
	}

	optsValue := reflect.New(fnType.In(1)).Elem()
	if field := optsValue.FieldByName("Query"); field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
		field.SetString(opts.Query)
	}
	if field := optsValue.FieldByName("Adjudication"); field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
		field.SetString(opts.Adjudication)
	}

	results := fnValue.Call([]reflect.Value{reflect.ValueOf(ctx), optsValue})
	if errValue := results[1]; !errValue.IsNil() {
		if err, ok := errValue.Interface().(error); ok {
			return nil, err
		}
		return nil, fmt.Errorf("dead-letter list function returned invalid error")
	}
	items, ok := results[0].Interface().([]postadjudicationstatus.DeadLetterBacklogEntry)
	if !ok {
		return nil, fmt.Errorf("dead-letter list function returned invalid items")
	}
	return items, nil
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
