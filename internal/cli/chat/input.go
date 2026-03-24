package chat

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// inputModel wraps a textarea for the chat input area.
type inputModel struct {
	textarea textarea.Model
}

func newInputModel() inputModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message... (Alt+Enter for newline)"
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
	m.textarea.SetWidth(width)
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
