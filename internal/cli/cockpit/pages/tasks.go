package pages

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
)

// taskTickMsg triggers periodic task list refresh.
type taskTickMsg time.Time

func taskTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return taskTickMsg(t)
	})
}

// TaskInfo holds summary info about a background task for display.
type TaskInfo struct {
	ID      string
	Prompt  string
	Status  string
	Elapsed time.Duration
}

// TaskLister provides the list of background tasks.
type TaskLister interface {
	ListTasks() []TaskInfo
}

// TasksPage displays background tasks in a table view.
type TasksPage struct {
	lister       TaskLister
	tasks        []TaskInfo
	cursor       int
	tickActive   bool
	width, height int
}

// NewTasksPage creates a new TasksPage. lister may be nil.
func NewTasksPage(lister TaskLister) *TasksPage {
	return &TasksPage{lister: lister}
}

// Title returns the page tab label.
func (m *TasksPage) Title() string { return "Tasks" }

// ShortHelp returns key bindings for the help bar.
func (m *TasksPage) ShortHelp() []key.Binding { return nil }

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
		return m, taskTickCmd()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
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

	// Table header.
	header := lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		Bold(true).
		PaddingLeft(4).
		Render(fmt.Sprintf("%-10s %-30s %-10s %s", "ID", "Prompt", "Status", "Elapsed"))

	separator := lipgloss.NewStyle().
		Foreground(theme.BorderSubtle).
		PaddingLeft(4).
		Render(strings.Repeat("─", max(m.width-8, 40)))

	// Table rows.
	var rows []string
	for i, task := range m.tasks {
		id := task.ID
		if len(id) > 8 {
			id = id[:8]
		}
		prompt := task.Prompt
		if len(prompt) > 28 {
			prompt = prompt[:25] + "..."
		}
		elapsed := task.Elapsed.Round(time.Second).String()

		row := fmt.Sprintf("%-10s %-30s %-10s %s", id, prompt, task.Status, elapsed)

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
	return strings.Join(sections, "\n")
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
