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
	role       string // "user", "assistant", "system"
	content    string // rendered content (display cache)
	rawContent string // original markdown (assistant only, for resize reflow)
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

// appendAssistant adds an assistant message with raw markdown preserved for
// resize reflow. All assistant messages must go through this helper.
func (m *chatViewModel) appendAssistant(raw string) {
	if strings.TrimSpace(raw) == "" {
		return
	}
	rendered := renderMarkdown(raw, m.contentWidth())
	m.entries = append(m.entries, chatEntry{
		role:       "assistant",
		content:    strings.TrimRight(rendered, "\n"),
		rawContent: raw,
	})
	m.render()
}

// finalizeStream commits the streamed buffer as an assistant message and
// clears the streaming buffer.
func (m *chatViewModel) finalizeStream() {
	raw := m.streamBuf.String()
	m.streamBuf.Reset()
	m.appendAssistant(raw)
}

// contentWidth returns the width available for assistant markdown rendering.
func (m *chatViewModel) contentWidth() int {
	w := m.width - 2 // left indent + safety margin
	if w < 10 {
		w = 10
	}
	return w
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
// Uses block-join to avoid accumulating leading blank lines.
func (m *chatViewModel) render() {
	userLabel := lipgloss.NewStyle().Bold(true).Foreground(tui.Highlight).Render("You")
	assistantLabel := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render("Lango")
	systemLabel := lipgloss.NewStyle().Bold(true).Foreground(tui.Muted).Render("System")

	var blocks []string
	for _, entry := range m.entries {
		switch entry.role {
		case "user":
			blocks = append(blocks, fmt.Sprintf(" %s\n %s", userLabel, entry.content))
		case "assistant":
			content := entry.content
			if entry.rawContent != "" {
				content = strings.TrimRight(renderMarkdown(entry.rawContent, m.contentWidth()), "\n")
			}
			blocks = append(blocks, fmt.Sprintf(" %s\n %s", assistantLabel, content))
		case "system":
			blocks = append(blocks, fmt.Sprintf(" %s  %s", systemLabel, entry.content))
		}
	}

	// Render in-flight streaming content.
	if m.streamBuf.Len() > 0 {
		blocks = append(blocks, fmt.Sprintf(" %s\n %s", assistantLabel, m.streamBuf.String()))
	}

	m.viewport.SetContent(strings.Join(blocks, "\n\n"))
	m.viewport.GotoBottom()
}
