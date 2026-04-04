package chat

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// taskStripModel displays a compact summary of background tasks.
type taskStripModel struct {
	manager BackgroundManagerI
	tasks   []BackgroundTaskInfo
}

func newTaskStripModel(manager BackgroundManagerI) taskStripModel {
	return taskStripModel{manager: manager}
}

func (m *taskStripModel) refresh() {
	if m.manager == nil {
		m.tasks = nil
		return
	}
	m.tasks = m.manager.List()
}

// View returns the task strip content. Empty string when no tasks or no manager.
func (m *taskStripModel) View(width int) string {
	if m.manager == nil || len(m.tasks) == 0 {
		return ""
	}

	var running, pending int
	var latestName, latestStatus string
	var latestElapsed string
	for _, t := range m.tasks {
		switch t.Status {
		case "running":
			running++
		case "pending":
			pending++
		}
		latestName = t.Prompt
		latestStatus = t.Status
		latestElapsed = t.Elapsed.Round(1e9).String() // round to seconds
	}

	// Truncate task name
	if len(latestName) > 30 {
		latestName = latestName[:27] + "..."
	}

	label := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Info).
		Render("Tasks:")

	summary := fmt.Sprintf("%d running", running)
	if pending > 0 {
		summary += fmt.Sprintf("  %d pending", pending)
	}

	latest := fmt.Sprintf("[%s] %s %s", latestName, latestStatus, latestElapsed)

	content := fmt.Sprintf(" %s %s | %s", label, summary, latest)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#0f1724")).
		Foreground(tui.Foreground).
		Width(max(width, 1)).
		Render(content)
}
