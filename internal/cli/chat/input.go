package chat

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

const defaultComposerPlaceholder = "Ask Lango to inspect code, explain behavior, or run a task. /help for commands, Alt+Enter for newline."

// inputModel wraps a textarea for the chat composer.
type inputModel struct {
	textarea textarea.Model
}

func newInputModel() inputModel {
	ta := textarea.New()
	ta.Placeholder = defaultComposerPlaceholder
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Primary).
		Padding(0, 1)
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Muted).
		Padding(0, 1)
	ta.SetHeight(3)
	ta.Focus()

	return inputModel{textarea: ta}
}

func (m inputModel) Update(msg tea.Msg) (inputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	return m.textarea.View()
}

func (m *inputModel) SetWidth(width int) {
	w := width - 2
	if w < 10 {
		w = 10
	}
	m.textarea.SetWidth(w)
}

func (m *inputModel) SetState(state chatState) tea.Cmd {
	switch state {
	case stateIdle:
		m.textarea.Placeholder = defaultComposerPlaceholder
		return m.textarea.Focus()
	case stateStreaming:
		m.textarea.Placeholder = "Lango is responding. Ctrl+C cancels the current turn."
		m.textarea.Blur()
	case stateApproving:
		m.textarea.Placeholder = "Approval is required below before the turn can continue."
		m.textarea.Blur()
	case stateCancelling:
		m.textarea.Placeholder = "Cancelling the current turn..."
		m.textarea.Blur()
	case stateFailed:
		m.textarea.Placeholder = "The last turn failed. Type a retry, adjust the request, or use /help."
		return m.textarea.Focus()
	}
	return nil
}

func (m *inputModel) Value() string {
	return m.textarea.Value()
}

func (m *inputModel) Reset() {
	m.textarea.Reset()
}

func (m *inputModel) Focus() tea.Cmd {
	return m.textarea.Focus()
}

func (m *inputModel) Blur() {
	m.textarea.Blur()
}

func renderFooter(input inputModel, state chatState, width int) string {
	parts := make([]string, 0, 2)
	if state != stateApproving {
		parts = append(parts, input.View())
	}
	parts = append(parts, renderHelpBar(state, width))
	return strings.Join(parts, "\n")
}
