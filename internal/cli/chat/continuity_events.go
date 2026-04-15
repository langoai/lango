package chat

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/eventbus"
)

// CompactionCompletedTeaMsg is the tea.Msg mirror of
// eventbus.CompactionCompletedEvent so renderers can handle it through the
// standard Update loop.
type CompactionCompletedTeaMsg struct {
	SessionKey      string
	ReclaimedTokens int
}

// CompactionSlowTeaMsg mirrors eventbus.CompactionSlowEvent.
type CompactionSlowTeaMsg struct {
	SessionKey string
}

// LearningSuggestionTeaMsg mirrors eventbus.LearningSuggestionEvent.
type LearningSuggestionTeaMsg struct {
	SessionKey   string
	SuggestionID string
	Pattern      string
	ProposedRule string
	Confidence   float64
	Rationale    string
}

// subscribeContinuityEvents wires TUI handlers for Phase 3 events.
// Events are forwarded to the bubbletea program via Send so they flow
// through the Update loop and appear as transient status entries. When no
// program is set yet (pre-SetProgram), the subscription is deferred until
// the first call to SetProgram — here we hold on to the bus and defer via
// a simple guard.
func (m *ChatModel) subscribeContinuityEvents() {
	if m.eventBus == nil {
		return
	}
	sessionKey := m.sessionKey
	bus := m.eventBus

	eventbus.SubscribeTyped(bus, func(e eventbus.CompactionCompletedEvent) {
		if e.SessionKey != sessionKey {
			return
		}
		m.sendSafe(CompactionCompletedTeaMsg{
			SessionKey:      e.SessionKey,
			ReclaimedTokens: e.ReclaimedTokens,
		})
	})
	eventbus.SubscribeTyped(bus, func(e eventbus.CompactionSlowEvent) {
		if e.SessionKey != sessionKey {
			return
		}
		m.sendSafe(CompactionSlowTeaMsg{SessionKey: e.SessionKey})
	})
	eventbus.SubscribeTyped(bus, func(e eventbus.LearningSuggestionEvent) {
		if e.SessionKey != sessionKey {
			return
		}
		m.sendSafe(LearningSuggestionTeaMsg{
			SessionKey:   e.SessionKey,
			SuggestionID: e.SuggestionID,
			Pattern:      e.Pattern,
			ProposedRule: e.ProposedRule,
			Confidence:   e.Confidence,
			Rationale:    e.Rationale,
		})
	})
}

// sendSafe forwards a message through the tea.Program if one is wired.
// A nil program (pre-SetProgram) drops the message silently — subscribers
// are set up eagerly in New, and may fire before the program is ready
// during early-startup channel activity.
func (m *ChatModel) sendSafe(msg tea.Msg) {
	if m.program == nil {
		return
	}
	m.program.Send(msg)
}

// handleCompactionCompleted appends a status line summarizing the
// reclaimed tokens.
func (m *ChatModel) handleCompactionCompleted(msg CompactionCompletedTeaMsg) (*ChatModel, tea.Cmd) {
	m.chatView.appendStatus(
		fmt.Sprintf("context compacted (reclaimed %d tokens)", msg.ReclaimedTokens),
		"",
	)
	return m, nil
}

// handleCompactionSlow appends a warn-styled status line.
func (m *ChatModel) handleCompactionSlow(_ CompactionSlowTeaMsg) (*ChatModel, tea.Cmd) {
	m.chatView.appendStatus(
		"compaction still running — proceeded with current context",
		"warning",
	)
	return m, nil
}

// handleLearningSuggestion appends a status line describing the proposed
// rule. Phase 3 scope: informational only. Approval-gated persistence is a
// follow-up task (see design.md Open Questions).
func (m *ChatModel) handleLearningSuggestion(msg LearningSuggestionTeaMsg) (*ChatModel, tea.Cmd) {
	m.chatView.appendStatus(
		fmt.Sprintf("learning suggestion (%.0f%%): %s", msg.Confidence*100, msg.ProposedRule),
		"",
	)
	return m, nil
}
