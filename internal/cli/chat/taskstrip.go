package chat

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/tui"
)

// taskStripModel displays a compact summary of background tasks.
type taskStripModel struct {
	manager   *background.Manager
	snapshots []background.TaskSnapshot
}

func newTaskStripModel(manager *background.Manager) taskStripModel {
	return taskStripModel{manager: manager}
}

func (m *taskStripModel) refresh() {
	if m.manager == nil {
		m.snapshots = nil
		return
	}
	m.snapshots = m.manager.List()
}

// View returns the task strip content. Empty string when no tasks or no manager.
func (m *taskStripModel) View(width int) string {
	if m.manager == nil || len(m.snapshots) == 0 {
		return ""
	}

	// Sort by StartedAt descending for stable "latest" selection.
	sort.Slice(m.snapshots, func(i, j int) bool {
		return m.snapshots[i].StartedAt.After(m.snapshots[j].StartedAt)
	})

	var running, pending int
	for _, t := range m.snapshots {
		switch t.Status {
		case background.Running:
			running++
		case background.Pending:
			pending++
		}
	}

	latest := m.snapshots[0]
	latestName := latest.Prompt
	latestStatus := latest.StatusText
	var latestElapsed time.Duration
	if !latest.StartedAt.IsZero() {
		if !latest.CompletedAt.IsZero() {
			latestElapsed = latest.CompletedAt.Sub(latest.StartedAt) // terminal: freeze
		} else {
			latestElapsed = time.Since(latest.StartedAt) // running: wall-clock
		}
	}

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

	latestInfo := fmt.Sprintf("[%s] %s %s", latestName, latestStatus, latestElapsed.Round(time.Second))

	content := fmt.Sprintf(" %s %s | %s", label, summary, latestInfo)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#0f1724")).
		Foreground(tui.Foreground).
		Width(max(width, 1)).
		Render(content)
}
