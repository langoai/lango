package chat

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

type transcriptItemKind string

const (
	itemUser      transcriptItemKind = "user"
	itemAssistant transcriptItemKind = "assistant"
	itemSystem    transcriptItemKind = "system"
	itemStatus    transcriptItemKind = "status"
	itemApproval  transcriptItemKind = "approval"
	itemTool      transcriptItemKind = "tool"
	itemThinking  transcriptItemKind = "thinking"
	itemChannel    transcriptItemKind = "channel"
	itemDelegation transcriptItemKind = "delegation"
	itemRecovery   transcriptItemKind = "recovery"
)

type transcriptItem struct {
	kind       transcriptItemKind
	content    string
	rawContent string
	meta       map[string]string
}

// chatViewModel manages the scrollable chat transcript viewport.
type chatViewModel struct {
	viewport         viewport.Model
	entries          []transcriptItem
	streamBuf        strings.Builder
	width            int
	showCursor       bool
	cursorTickActive bool
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

func (m *chatViewModel) appendUser(content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	m.entries = append(m.entries, transcriptItem{
		kind:    itemUser,
		content: strings.TrimSpace(content),
	})
	m.render()
}

func (m *chatViewModel) appendAssistant(raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	m.entries = append(m.entries, transcriptItem{
		kind:       itemAssistant,
		content:    strings.TrimRight(renderMarkdown(raw, m.contentWidth()), "\n"),
		rawContent: raw,
	})
	m.render()
}

func (m *chatViewModel) appendSystem(content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	m.entries = append(m.entries, transcriptItem{
		kind:    itemSystem,
		content: strings.TrimSpace(content),
	})
	m.render()
}

func (m *chatViewModel) appendStatus(content string, tone string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	m.entries = append(m.entries, transcriptItem{
		kind:    itemStatus,
		content: strings.TrimSpace(content),
		meta:    map[string]string{"tone": tone},
	})
	m.render()
}

func (m *chatViewModel) appendApprovalEvent(content string, outcome string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	m.entries = append(m.entries, transcriptItem{
		kind:    itemApproval,
		content: strings.TrimSpace(content),
		meta:    map[string]string{"outcome": outcome},
	})
	m.render()
}

func (m *chatViewModel) appendToolStart(callID, toolName string, params map[string]any) {
	m.entries = append(m.entries, transcriptItem{
		kind:    itemTool,
		content: toolName,
		meta: map[string]string{
			"callID": callID,
			"state":  string(toolStateRunning),
		},
	})
	m.render()
}

func (m *chatViewModel) finalizeToolResult(callID string, success bool, duration time.Duration, output string) {
	for i := len(m.entries) - 1; i >= 0; i-- {
		e := &m.entries[i]
		if e.kind == itemTool && e.meta["callID"] == callID {
			if success {
				e.meta["state"] = string(toolStateSuccess)
			} else {
				e.meta["state"] = string(toolStateError)
			}
			e.meta["duration"] = duration.Round(time.Millisecond * 100).String()
			if output != "" {
				e.meta["output"] = output
			}
			break
		}
	}
	m.render()
}

func (m *chatViewModel) appendThinking(summary string) {
	m.entries = append(m.entries, transcriptItem{
		kind:    itemThinking,
		content: summary,
		meta:    map[string]string{"state": "active"},
	})
	m.render()
}

func (m *chatViewModel) finalizeThinking(summary string, duration time.Duration) {
	for i := len(m.entries) - 1; i >= 0; i-- {
		e := &m.entries[i]
		if e.kind == itemThinking && e.meta["state"] == "active" {
			e.meta["state"] = "done"
			e.meta["duration"] = duration.Round(time.Millisecond * 100).String()
			if summary != "" {
				e.content = summary
			}
			break
		}
	}
	m.render()
}

func (m *chatViewModel) appendChannel(channel, senderName, text, sessionKey string, metadata map[string]string) {
	meta := map[string]string{
		"channel":    channel,
		"sender":     senderName,
		"sessionKey": sessionKey,
	}
	// Preserve all metadata for future use (thread grouping, origin jump, etc.)
	for k, val := range metadata {
		meta[k] = val
	}
	m.entries = append(m.entries, transcriptItem{
		kind:       itemChannel,
		rawContent: text,
		meta:       meta,
	})
	m.render()
}

func (m *chatViewModel) appendDelegation(from, to, reason string) {
	m.entries = append(m.entries, transcriptItem{
		kind: itemDelegation,
		meta: map[string]string{
			"from":   from,
			"to":     to,
			"reason": reason,
		},
	})
	m.render()
}

func (m *chatViewModel) appendRecovery(action, causeClass string, attempt int, backoff time.Duration) {
	m.entries = append(m.entries, transcriptItem{
		kind: itemRecovery,
		meta: map[string]string{
			"action":     action,
			"causeClass": causeClass,
			"attempt":    strconv.Itoa(attempt),
			"backoff":    backoff.String(),
		},
	})
	m.render()
}

func (m *chatViewModel) appendTokenSummary(input, output, total, cache int64) {
	summary := fmt.Sprintf("\U0001F4CA Token usage: %s input, %s output, %s total",
		formatTokenCount(input), formatTokenCount(output), formatTokenCount(total))
	if cache > 0 {
		summary += fmt.Sprintf(" (%s cached)", formatTokenCount(cache))
	}
	m.appendStatus(summary, "")
}

// formatTokenCount formats a token count for display (e.g. 1500 → "1.5k").
func formatTokenCount(n int64) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return strconv.FormatInt(n, 10)
}

func (m *chatViewModel) appendChunk(chunk string) {
	m.streamBuf.WriteString(chunk)
	m.render()
}

func (m *chatViewModel) finalizeStream() {
	raw := m.streamBuf.String()
	m.streamBuf.Reset()
	m.appendAssistant(raw)
}

func (m *chatViewModel) lastAssistantRaw() string {
	for i := len(m.entries) - 1; i >= 0; i-- {
		if m.entries[i].kind == itemAssistant {
			return strings.TrimSpace(m.entries[i].rawContent)
		}
	}
	return ""
}

// stopCursorBlink resets both cursor-blink fields in one call.
func (m *chatViewModel) stopCursorBlink() {
	m.showCursor = false
	m.cursorTickActive = false
}

func (m *chatViewModel) clear() {
	m.entries = nil
	m.streamBuf.Reset()
	m.stopCursorBlink()
	m.viewport.SetContent("")
	m.viewport.GotoTop()
}

func (m *chatViewModel) setSize(width, height int) {
	prevWidth := m.width
	m.width = width
	m.viewport.Width = width
	m.viewport.Height = height
	if width != prevWidth {
		for i := range m.entries {
			if m.entries[i].kind == itemAssistant && m.entries[i].rawContent != "" {
				m.entries[i].content = strings.TrimRight(
					renderMarkdown(m.entries[i].rawContent, m.contentWidth()), "\n")
			}
		}
	}
	m.render()
}

func (m *chatViewModel) contentWidth() int {
	w := m.width - 2
	if w < 10 {
		w = 10
	}
	return w
}

func (m *chatViewModel) render() {
	var blocks []string

	for _, entry := range m.entries {
		switch entry.kind {
		case itemUser:
			blocks = append(blocks, renderTranscriptBlock("You", entry.content, tui.Highlight))
		case itemAssistant:
			blocks = append(blocks, renderTranscriptBlock("Lango", entry.content, tui.Primary))
		case itemSystem:
			blocks = append(blocks, renderSystemBlock(entry.content))
		case itemStatus:
			blocks = append(blocks, renderStatusBlock(entry.content, entry.meta["tone"]))
		case itemApproval:
			blocks = append(blocks, renderApprovalEventBlock(entry.content, entry.meta["outcome"]))
		case itemTool:
			blocks = append(blocks, renderToolBlock(
				entry.content,
				ToolItemState(entry.meta["state"]),
				entry.meta["duration"],
				entry.meta["output"],
				m.contentWidth(),
			))
		case itemThinking:
			blocks = append(blocks, renderThinkingBlock(
				entry.content,
				entry.meta["state"],
				entry.meta["duration"],
				m.contentWidth(),
			))
		case itemChannel:
			blocks = append(blocks, renderChannelBlock(
				entry.rawContent,
				entry.meta["channel"],
				entry.meta["sender"],
				m.contentWidth(),
			))

		case itemDelegation:
			blocks = append(blocks, renderDelegationBlock(
				entry.meta["from"],
				entry.meta["to"],
				entry.meta["reason"],
				m.contentWidth(),
			))

		case itemRecovery:
			attempt, _ := strconv.Atoi(entry.meta["attempt"])
			backoff, _ := time.ParseDuration(entry.meta["backoff"])
			blocks = append(blocks, renderRecoveryBlock(
				entry.meta["action"],
				entry.meta["causeClass"],
				attempt,
				backoff,
				m.contentWidth(),
			))
		}
	}

	if m.streamBuf.Len() > 0 {
		streamContent := m.streamBuf.String()
		if m.showCursor {
			streamContent += "\u258c" // "▌" block cursor
		}
		blocks = append(blocks, renderTranscriptBlock("Lango", streamContent, tui.Primary))
	}

	m.viewport.SetContent(strings.Join(blocks, "\n\n"))
	m.viewport.GotoBottom()
}

func renderTranscriptBlock(label, content string, color lipgloss.Color) string {
	labelText := lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Render(label)
	separatorWidth := min(16, max(lipgloss.Width(label)+6, 8))
	separator := lipgloss.NewStyle().
		Foreground(color).
		Render(strings.Repeat("─", separatorWidth))
	body := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(color).
		PaddingLeft(1).
		Render(strings.TrimRight(content, "\n"))
	return fmt.Sprintf(" %s  %s\n%s", labelText, separator, body)
}

func renderSystemBlock(content string) string {
	label := lipgloss.NewStyle().Bold(true).Foreground(tui.Muted).Render("System")
	body := lipgloss.NewStyle().PaddingLeft(1).Render(strings.TrimRight(content, "\n"))
	return fmt.Sprintf(" %s\n%s", label, body)
}

func renderStatusBlock(content, tone string) string {
	color := tui.Muted
	switch tone {
	case "success":
		color = tui.Success
	case "warning":
		color = tui.Warning
	case "error":
		color = tui.Error
	}
	label := lipgloss.NewStyle().Bold(true).Foreground(color).Render("Status")
	body := lipgloss.NewStyle().Foreground(color).Render(content)
	return fmt.Sprintf(" %s  %s", label, body)
}

func renderApprovalEventBlock(content, outcome string) string {
	color := tui.Warning
	switch outcome {
	case "approved", "session":
		color = tui.Success
	case "denied":
		color = tui.Error
	}
	label := lipgloss.NewStyle().Bold(true).Foreground(color).Render("Approval")
	body := lipgloss.NewStyle().Foreground(color).Render(content)
	return fmt.Sprintf(" %s  %s", label, body)
}
