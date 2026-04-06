package pages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
)

// Column layout constants for the tasks table.
const (
	taskGutterW = 6  // "  " or "> " prefix (2) + PaddingLeft(4)
	taskColIDW  = 10 // ID column width
	taskColStatW = 10 // Status column width
	taskColElW  = 12 // Elapsed column width
	taskColGapW = 1  // space between columns (from format string)
	taskNarrowThreshold = 50 // width below which elapsed column is hidden

	tableMinHeight  = 6 // minimum rows reserved for the table
	detailMinHeight = 8 // minimum rows reserved for the detail panel
	statusMsgTTL    = 3 * time.Second // how long status messages are shown
)

// taskActionResultMsg carries the outcome of an async cancel/retry action.
type taskActionResultMsg struct {
	msg string
	err error
}

// taskTickMsg triggers periodic task list refresh.
type taskTickMsg time.Time

func taskTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return taskTickMsg(t)
	})
}

// TaskInfo holds summary info about a background task for display.
type TaskInfo struct {
	ID            string
	Prompt        string
	Status        string
	Elapsed       time.Duration
	Result        string        // completion result text
	Error         string        // error message if failed
	OriginChannel string        // originating channel (e.g. "telegram", "slack")
	TokensUsed    int           // token count for this task
}

// TaskLister provides the list of background tasks.
type TaskLister interface {
	ListTasks() []TaskInfo
}

// TaskActioner provides cancel and retry operations for background tasks.
type TaskActioner interface {
	CancelTask(id string) error
	RetryTask(ctx context.Context, id string) error
}

// TasksPage displays background tasks in a table view.
type TasksPage struct {
	lister       TaskLister
	actioner     TaskActioner // optional, nil when unavailable
	tasks        []TaskInfo
	cursor       int
	tickActive   bool
	width, height int

	detailMode   bool      // true when detail panel is expanded
	detailScroll int       // scroll offset within detail content
	statusMsg    string    // transient action feedback
	statusTime   time.Time // when statusMsg was set
}

// NewTasksPage creates a new TasksPage. lister and actioner may be nil.
func NewTasksPage(lister TaskLister, actioner TaskActioner) *TasksPage {
	return &TasksPage{lister: lister, actioner: actioner}
}

// Title returns the page tab label.
func (m *TasksPage) Title() string { return "Tasks" }

// ShortHelp returns key bindings for the help bar.
func (m *TasksPage) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "details")),
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	}
	if m.detailMode {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")))
	}
	if m.actioner != nil {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cancel")),
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		)
	}
	return bindings
}

// Init satisfies tea.Model.
func (m *TasksPage) Init() tea.Cmd { return nil }

// Activate starts periodic task list refresh.
func (m *TasksPage) Activate() tea.Cmd {
	m.tickActive = true
	m.refreshData()
	return taskTickCmd()
}

// Deactivate stops the tick loop.
func (m *TasksPage) Deactivate() {
	m.tickActive = false
}

// Update handles messages.
func (m *TasksPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case taskTickMsg:
		if !m.tickActive {
			return m, nil
		}
		m.refreshData()
		// Clear expired status message.
		if m.statusMsg != "" && time.Since(m.statusTime) > statusMsgTTL {
			m.statusMsg = ""
		}
		return m, taskTickCmd()
	case taskActionResultMsg:
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
		} else {
			m.statusMsg = msg.msg
		}
		m.statusTime = time.Now()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey processes key messages and returns the updated model and optional command.
func (m *TasksPage) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if len(m.tasks) == 0 {
			return m, nil
		}
		m.detailMode = !m.detailMode
		if m.detailMode {
			m.detailScroll = 0
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.detailMode = false
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.detailMode {
			if m.detailScroll > 0 {
				m.detailScroll--
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.detailMode {
			m.detailScroll++
		} else {
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		return m, m.cancelSelectedTask()
	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		return m, m.retrySelectedTask()
	}
	return m, nil
}

// cancelSelectedTask returns a tea.Cmd that cancels the currently selected task,
// or nil if the action is not applicable.
func (m *TasksPage) cancelSelectedTask() tea.Cmd {
	if m.actioner == nil || len(m.tasks) == 0 {
		return nil
	}
	task := m.tasks[m.cursor]
	if task.Status != "running" && task.Status != "pending" {
		return nil
	}
	id := task.ID
	actioner := m.actioner
	return func() tea.Msg {
		if err := actioner.CancelTask(id); err != nil {
			return taskActionResultMsg{err: err}
		}
		return taskActionResultMsg{msg: "Cancelled: " + id}
	}
}

// retrySelectedTask returns a tea.Cmd that retries the currently selected task,
// or nil if the action is not applicable.
func (m *TasksPage) retrySelectedTask() tea.Cmd {
	if m.actioner == nil || len(m.tasks) == 0 {
		return nil
	}
	task := m.tasks[m.cursor]
	if task.Status != "failed" && task.Status != "cancelled" {
		return nil
	}
	id := task.ID
	actioner := m.actioner
	return func() tea.Msg {
		if err := actioner.RetryTask(context.Background(), id); err != nil {
			return taskActionResultMsg{err: err}
		}
		return taskActionResultMsg{msg: "Retried: " + id}
	}
}

// View renders the tasks table.
func (m *TasksPage) View() string {
	if m.lister == nil {
		return lipgloss.NewStyle().
			Foreground(theme.TextSecondary).
			PaddingLeft(2).
			PaddingTop(1).
			Render("No background tasks available")
	}

	if len(m.tasks) == 0 {
		return lipgloss.NewStyle().
			Foreground(theme.TextSecondary).
			PaddingLeft(2).
			PaddingTop(1).
			Render("No active tasks")
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		PaddingLeft(2).
		Render("Background Tasks")

	// Compute dynamic prompt column width based on available space.
	wide := m.width >= taskNarrowThreshold
	var promptW int
	if wide {
		// Wide: ID | Prompt | Status | Elapsed
		fixedW := taskGutterW + taskColIDW + taskColStatW + taskColElW + taskColGapW*3
		promptW = max(m.width-fixedW, 8)
	} else {
		// Narrow: ID | Prompt | Status (no elapsed)
		fixedW := taskGutterW + taskColIDW + taskColStatW + taskColGapW*2
		promptW = max(m.width-fixedW, 8)
	}

	// Table header.
	var headerText string
	if wide {
		fmtStr := fmt.Sprintf("%%-%ds %%-%ds %%-%ds %%s", taskColIDW, promptW, taskColStatW)
		headerText = fmt.Sprintf(fmtStr, "ID", "Prompt", "Status", "Elapsed")
	} else {
		fmtStr := fmt.Sprintf("%%-%ds %%-%ds %%-%ds", taskColIDW, promptW, taskColStatW)
		headerText = fmt.Sprintf(fmtStr, "ID", "Prompt", "Status")
	}
	header := lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		Bold(true).
		PaddingLeft(4).
		Render(headerText)

	separator := lipgloss.NewStyle().
		Foreground(theme.BorderSubtle).
		PaddingLeft(4).
		Render(strings.Repeat("─", max(m.width-8, 40)))

	// Table rows.
	var rows []string
	for i, task := range m.tasks {
		id := truncate(task.ID, taskColIDW-2) // leave padding room
		promptStr := ansi.Truncate(task.Prompt, promptW, "…")
		status := truncate(task.Status, taskColStatW)

		var row string
		if wide {
			elapsed := task.Elapsed.Round(time.Second).String()
			fmtStr := fmt.Sprintf("%%-%ds %%-%ds %%-%ds %%s", taskColIDW, promptW, taskColStatW)
			row = fmt.Sprintf(fmtStr, id, promptStr, status, elapsed)
		} else {
			fmtStr := fmt.Sprintf("%%-%ds %%-%ds %%-%ds", taskColIDW, promptW, taskColStatW)
			row = fmt.Sprintf(fmtStr, id, promptStr, status)
		}

		style := lipgloss.NewStyle().PaddingLeft(4)
		if i == m.cursor {
			style = style.
				Foreground(theme.Accent).
				Bold(true)
			row = "> " + row
		} else {
			style = style.Foreground(theme.TextPrimary)
			row = "  " + row
		}
		rows = append(rows, style.Render(row))
	}

	sections := []string{title, "", header, separator}
	sections = append(sections, rows...)

	// Status message bar.
	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(theme.Warning).
			PaddingLeft(4)
		sections = append(sections, "", statusStyle.Render(m.statusMsg))
	}

	// Detail panel.
	if m.detailMode && m.cursor >= 0 && m.cursor < len(m.tasks) {
		detail := m.renderDetail(m.tasks[m.cursor])
		sections = append(sections, "", detail)
	}

	return strings.Join(sections, "\n")
}

// renderDetail builds the detail panel for the given task.
func (m *TasksPage) renderDetail(task TaskInfo) string {
	contentW := max(m.width-8, 30) // inner content width
	labelW := 10                   // "  Status:  " prefix width

	sep := lipgloss.NewStyle().
		Foreground(theme.BorderSubtle).
		PaddingLeft(2).
		Render("─── Task Detail " + strings.Repeat("─", max(contentW-16, 4)))

	valW := max(contentW-labelW, 10)

	statusLine := fmt.Sprintf("  Status:  %s (%s elapsed)",
		ansi.Truncate(task.Status, valW/2, "…"),
		task.Elapsed.Round(time.Second))

	originVal := task.OriginChannel
	if originVal == "" {
		originVal = "(none)"
	}
	originLine := fmt.Sprintf("  Origin:  %s", ansi.Truncate(originVal, valW, "…"))

	tokensLine := fmt.Sprintf("  Tokens:  %s", formatTokens(task.TokensUsed))

	promptWrapped := wordWrap(task.Prompt, valW)
	if promptWrapped == "" {
		promptWrapped = "(none)"
	}

	resultWrapped := wordWrap(task.Result, valW)
	if resultWrapped == "" {
		resultWrapped = "(none)"
	}

	errorVal := task.Error
	if errorVal == "" {
		errorVal = "(none)"
	}
	errorWrapped := wordWrap(errorVal, valW)

	// Build the full detail content as lines.
	var lines []string
	lines = append(lines, sep)
	lines = append(lines, statusLine)
	lines = append(lines, originLine)
	lines = append(lines, tokensLine)
	lines = append(lines, "")
	lines = append(lines, "  Prompt:")
	for _, l := range strings.Split(promptWrapped, "\n") {
		lines = append(lines, "    "+l)
	}
	lines = append(lines, "")
	lines = append(lines, "  Result:")
	for _, l := range strings.Split(resultWrapped, "\n") {
		lines = append(lines, "    "+l)
	}
	lines = append(lines, "")
	lines = append(lines, "  Error:")
	for _, l := range strings.Split(errorWrapped, "\n") {
		lines = append(lines, "    "+l)
	}

	// Apply scroll offset.
	if m.detailScroll > 0 {
		offset := m.detailScroll
		if offset >= len(lines) {
			offset = max(len(lines)-1, 0)
		}
		lines = lines[offset:]
	}

	// Compute available height for detail panel.
	detailH := m.detailHeight()
	if detailH > 0 && len(lines) > detailH {
		lines = lines[:detailH]
	}

	style := lipgloss.NewStyle().
		Foreground(theme.TextSecondary).
		PaddingLeft(2)
	return style.Render(strings.Join(lines, "\n"))
}

// detailHeight computes the maximum number of lines for the detail panel.
func (m *TasksPage) detailHeight() int {
	if m.height <= 0 {
		return 0 // unlimited when height is not set
	}
	if m.height < tableMinHeight+detailMinHeight {
		return max(m.height-tableMinHeight, detailMinHeight)
	}
	return m.height * 60 / 100
}

// formatTokens returns a human-readable token count with comma separators.
func formatTokens(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	offset := len(s) % 3
	if offset > 0 {
		b.WriteString(s[:offset])
	}
	for i := offset; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// wordWrap wraps text to the given width, breaking on spaces.
func wordWrap(text string, width int) string {
	if width <= 0 || text == "" {
		return text
	}
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		var cur strings.Builder
		for _, w := range words {
			if cur.Len() == 0 {
				cur.WriteString(w)
			} else if cur.Len()+1+len(w) > width {
				lines = append(lines, cur.String())
				cur.Reset()
				cur.WriteString(w)
			} else {
				cur.WriteByte(' ')
				cur.WriteString(w)
			}
		}
		if cur.Len() > 0 {
			lines = append(lines, cur.String())
		}
	}
	return strings.Join(lines, "\n")
}

func (m *TasksPage) refreshData() {
	if m.lister == nil {
		m.tasks = nil
		return
	}
	m.tasks = m.lister.ListTasks()
	if m.cursor >= len(m.tasks) {
		m.cursor = max(len(m.tasks)-1, 0)
	}
}
