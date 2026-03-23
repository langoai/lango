package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// chatEntry represents a single message in the chat history.
type chatEntry struct {
	role    string // "user", "assistant", "system"
	content string
}

// chatViewModel manages the scrollable chat history viewport.
type chatViewModel struct {
	viewport viewport.Model
	entries  []chatEntry
	// streamBuf accumulates text chunks during streaming.
	streamBuf strings.Builder
	width     int
}

func newChatViewModel(width, height int) chatViewModel {
	vp := viewport.New(width, height)
	vp.SetContent("")
	return chatViewModel{
		viewport: vp,
		width:    width,
	}
}

func (m chatViewModel) Update(msg tea.Msg) (chatViewModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m chatViewModel) View() string {
	return m.viewport.View()
}

// appendUser adds a user message and re-renders the viewport.
func (m *chatViewModel) appendUser(content string) {
	m.entries = append(m.entries, chatEntry{role: "user", content: content})
	m.render()
}

// appendSystem adds a system/info message and re-renders the viewport.
func (m *chatViewModel) appendSystem(content string) {
	m.entries = append(m.entries, chatEntry{role: "system", content: content})
	m.render()
}

// appendChunk appends a streaming text chunk to the buffer and re-renders.
func (m *chatViewModel) appendChunk(chunk string) {
	m.streamBuf.WriteString(chunk)
	m.render()
}

// finalizeStream commits the streamed buffer as an assistant message, applying
// markdown rendering, and clears the streaming buffer.
func (m *chatViewModel) finalizeStream(width int) {
	raw := m.streamBuf.String()
	m.streamBuf.Reset()
	if strings.TrimSpace(raw) == "" {
		return
	}
	rendered := renderMarkdown(raw, width)
	m.entries = append(m.entries, chatEntry{role: "assistant", content: strings.TrimRight(rendered, "\n")})
	m.render()
}

// finalizeWithText adds an assistant message directly (for error/fallback).
func (m *chatViewModel) finalizeWithText(text string) {
	m.streamBuf.Reset()
	if strings.TrimSpace(text) == "" {
		return
	}
	m.entries = append(m.entries, chatEntry{role: "assistant", content: text})
	m.render()
}

// clear removes all entries and resets the viewport.
func (m *chatViewModel) clear() {
	m.entries = nil
	m.streamBuf.Reset()
	m.viewport.SetContent("")
	m.viewport.GotoTop()
}

// setSize updates viewport dimensions.
func (m *chatViewModel) setSize(width, height int) {
	m.width = width
	m.viewport.Width = width
	m.viewport.Height = height
	m.render()
}

// render rebuilds the viewport content from all entries plus any in-flight stream.
func (m *chatViewModel) render() {
	var b strings.Builder

	userLabel := lipgloss.NewStyle().Bold(true).Foreground(tui.Highlight).Render("You")
	assistantLabel := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render("Lango")
	systemLabel := lipgloss.NewStyle().Bold(true).Foreground(tui.Muted).Render("System")

	for _, entry := range m.entries {
		switch entry.role {
		case "user":
			b.WriteString(fmt.Sprintf("\n %s\n", userLabel))
			b.WriteString(fmt.Sprintf(" %s\n", entry.content))
		case "assistant":
			b.WriteString(fmt.Sprintf("\n %s\n", assistantLabel))
			b.WriteString(fmt.Sprintf(" %s\n", entry.content))
		case "system":
			b.WriteString(fmt.Sprintf("\n %s  %s\n", systemLabel, entry.content))
		}
	}

	// Render in-flight streaming content.
	if m.streamBuf.Len() > 0 {
		b.WriteString(fmt.Sprintf("\n %s\n", assistantLabel))
		b.WriteString(fmt.Sprintf(" %s", m.streamBuf.String()))
	}

	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}
