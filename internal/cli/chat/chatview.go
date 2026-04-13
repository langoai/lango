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
	kind        transcriptItemKind
	content     string
	rawContent  string
	meta        map[string]string
	cachedBlock string // memoized rendered block, "" = needs re-render
}

// maxTranscriptEntries caps the number of transcript entries to prevent
// unbounded memory growth during long sessions.
const maxTranscriptEntries = 2000

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
	m.appendEntry(transcriptItem{
		kind:    itemUser,
		content: strings.TrimSpace(content),
	})
}

func (m *chatViewModel) appendAssistant(raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	m.appendEntry(transcriptItem{
		kind:       itemAssistant,
		content:    strings.TrimRight(renderMarkdown(raw, m.contentWidth()), "\n"),
		rawContent: raw,
	})
}

func (m *chatViewModel) appendSystem(content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	m.appendEntry(transcriptItem{
		kind:    itemSystem,
		content: strings.TrimSpace(content),
	})
}

func (m *chatViewModel) appendStatus(content string, tone string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	m.appendEntry(transcriptItem{
		kind:    itemStatus,
		content: strings.TrimSpace(content),
		meta:    map[string]string{"tone": tone},
	})
}

func (m *chatViewModel) appendApprovalEvent(content string, outcome string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	m.appendEntry(transcriptItem{
		kind:    itemApproval,
		content: strings.TrimSpace(content),
		meta:    map[string]string{"outcome": outcome},
	})
}

func (m *chatViewModel) appendToolStart(callID, toolName string, params map[string]any) {
	m.appendEntry(transcriptItem{
		kind:    itemTool,
		content: toolName,
		meta: map[string]string{
			"callID": callID,
			"state":  string(toolStateRunning),
		},
	})
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
			e.cachedBlock = "" // invalidate cache — state changed
			break
		}
	}
	m.render()
}

func (m *chatViewModel) appendThinking(summary string) {
	m.appendEntry(transcriptItem{
		kind:    itemThinking,
		content: summary,
		meta:    map[string]string{"state": "active"},
	})
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
			e.cachedBlock = "" // invalidate cache — state changed
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
	m.appendEntry(transcriptItem{
		kind:       itemChannel,
		rawContent: text,
		meta:       meta,
	})
}

func (m *chatViewModel) appendDelegation(from, to, reason string) {
	m.appendEntry(transcriptItem{
		kind: itemDelegation,
		meta: map[string]string{
			"from":   from,
			"to":     to,
			"reason": reason,
		},
	})
}

func (m *chatViewModel) appendRecovery(action, causeClass string, attempt int, backoff time.Duration) {
	m.appendEntry(transcriptItem{
		kind: itemRecovery,
		meta: map[string]string{
			"action":     action,
			"causeClass": causeClass,
			"attempt":    strconv.Itoa(attempt),
			"backoff":    backoff.String(),
		},
	})
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
			// Re-render markdown for assistant entries at the new width.
			if m.entries[i].kind == itemAssistant && m.entries[i].rawContent != "" {
				m.entries[i].content = strings.TrimRight(
					renderMarkdown(m.entries[i].rawContent, m.contentWidth()), "\n")
			}
			// Invalidate all cached blocks — width affects rendering.
			m.entries[i].cachedBlock = ""
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

// appendEntry is the single entry point for adding transcript items.
// It appends the entry, enforces the max cap with tombstone trimming,
// and triggers a re-render.
func (m *chatViewModel) appendEntry(entry transcriptItem) {
	m.entries = append(m.entries, entry)
	if len(m.entries) > maxTranscriptEntries {
		trimCount := maxTranscriptEntries / 4 // trim 500 at a time
		// Accumulate tombstone count across repeated trims.
		prevTrimmed := 0
		hasTombstone := m.entries[0].kind == itemSystem && strings.HasPrefix(m.entries[0].content, "---")
		if hasTombstone {
			_, _ = fmt.Sscanf(m.entries[0].content, "--- %d older messages trimmed ---", &prevTrimmed)
		}
		// Determine base trim boundary.
		boundary := trimCount
		if !hasTombstone {
			boundary = trimCount - 1 // reserve one slot for new tombstone
		}
		// Collect in-flight tool/thinking entries from the trim range so they survive.
		var preserved []transcriptItem
		for _, e := range m.entries[:boundary] {
			if (e.kind == itemTool && e.meta["state"] == "running") ||
				(e.kind == itemThinking && e.meta["state"] == "active") {
				preserved = append(preserved, e)
			}
		}
		// Build a new slice: tombstone + preserved active entries + kept entries.
		// Using a new backing array ensures the old one is eligible for GC.
		kept := m.entries[boundary:]
		newEntries := make([]transcriptItem, 0, 1+len(preserved)+len(kept))
		newEntries = append(newEntries, transcriptItem{
			kind:    itemSystem,
			content: fmt.Sprintf("--- %d older messages trimmed ---", prevTrimmed+trimCount),
		})
		newEntries = append(newEntries, preserved...)
		newEntries = append(newEntries, kept...)
		m.entries = newEntries
	}
	m.render()
}

// renderEntry renders a single transcript entry into a display block string.
func (m *chatViewModel) renderEntry(entry transcriptItem) string {
	switch entry.kind {
	case itemUser:
		return renderTranscriptBlock("You", entry.content, tui.Highlight)
	case itemAssistant:
		return renderTranscriptBlock("Lango", entry.content, tui.Primary)
	case itemSystem:
		return renderSystemBlock(entry.content)
	case itemStatus:
		return renderStatusBlock(entry.content, entry.meta["tone"])
	case itemApproval:
		return renderApprovalEventBlock(entry.content, entry.meta["outcome"])
	case itemTool:
		return renderToolBlock(
			entry.content,
			ToolItemState(entry.meta["state"]),
			entry.meta["duration"],
			entry.meta["output"],
			m.contentWidth(),
		)
	case itemThinking:
		return renderThinkingBlock(
			entry.content,
			entry.meta["state"],
			entry.meta["duration"],
			m.contentWidth(),
		)
	case itemChannel:
		return renderChannelBlock(
			entry.rawContent,
			entry.meta["channel"],
			entry.meta["sender"],
			m.contentWidth(),
		)
	case itemDelegation:
		return renderDelegationBlock(
			entry.meta["from"],
			entry.meta["to"],
			entry.meta["reason"],
			m.contentWidth(),
		)
	case itemRecovery:
		attempt, _ := strconv.Atoi(entry.meta["attempt"])
		backoff, _ := time.ParseDuration(entry.meta["backoff"])
		return renderRecoveryBlock(
			entry.meta["action"],
			entry.meta["causeClass"],
			attempt,
			backoff,
			m.contentWidth(),
		)
	default:
		return ""
	}
}

func (m *chatViewModel) render() {
	blocks := make([]string, 0, len(m.entries)+1)

	for i, entry := range m.entries {
		if entry.cachedBlock != "" {
			blocks = append(blocks, entry.cachedBlock)
			continue
		}
		block := m.renderEntry(entry)
		m.entries[i].cachedBlock = block
		blocks = append(blocks, block)
	}

	// Streaming entry is always re-rendered (no cache).
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

// Pre-allocated styles for transcript rendering.
// lipgloss.Style is an immutable value type — per-call .Foreground(color)
// returns a stack copy with only the color set, avoiding heap allocation.
var (
	transcriptLabelStyle = lipgloss.NewStyle().Bold(true)
	transcriptSepStyle   = lipgloss.NewStyle()
	transcriptBodyStyle  = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				PaddingLeft(1)
	systemLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(tui.Muted)
	systemBodyStyle  = lipgloss.NewStyle().PaddingLeft(1)
	statusLabelStyle = lipgloss.NewStyle().Bold(true)
	statusBodyStyle  = lipgloss.NewStyle()
)

func renderTranscriptBlock(label, content string, color lipgloss.Color) string {
	labelText := transcriptLabelStyle.Foreground(color).Render(label)
	separatorWidth := min(16, max(lipgloss.Width(label)+6, 8))
	separator := transcriptSepStyle.Foreground(color).Render(strings.Repeat("─", separatorWidth))
	body := transcriptBodyStyle.BorderForeground(color).Render(strings.TrimRight(content, "\n"))
	return fmt.Sprintf(" %s  %s\n%s", labelText, separator, body)
}

func renderSystemBlock(content string) string {
	label := systemLabelStyle.Render("System")
	body := systemBodyStyle.Render(strings.TrimRight(content, "\n"))
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
	label := statusLabelStyle.Foreground(color).Render("Status")
	body := statusBodyStyle.Foreground(color).Render(content)
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
	label := statusLabelStyle.Foreground(color).Render("Approval")
	body := statusBodyStyle.Foreground(color).Render(content)
	return fmt.Sprintf(" %s  %s", label, body)
}
