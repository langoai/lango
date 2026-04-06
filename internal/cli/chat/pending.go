package chat

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// pendingIndicator tracks the "waiting for first content event" state
// between user submission and the first response event (chunk, tool, thinking).
type pendingIndicator struct {
	active bool
	start  time.Time
}

// Activate marks the indicator as active and records the current time.
func (p *pendingIndicator) Activate() {
	p.active = true
	p.start = time.Now()
}

// Dismiss clears the active state.
func (p *pendingIndicator) Dismiss() {
	p.active = false
}

// IsActive reports whether the indicator is currently showing.
func (p *pendingIndicator) IsActive() bool {
	return p.active
}

// Elapsed returns the human-readable duration since activation, rounded to seconds.
func (p *pendingIndicator) Elapsed() string {
	return time.Since(p.start).Round(time.Second).String()
}

// TickCmd returns a tea.Cmd that fires a PendingIndicatorTickMsg after 500ms.
func (p *pendingIndicator) TickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return PendingIndicatorTickMsg(t)
	})
}
