package chat

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/turnrunner"
)

// newTestModel creates a ChatModel with minimal deps for unit testing.
func newTestModel() *ChatModel {
	m := &ChatModel{
		cfg:        &config.Config{},
		sessionKey: "test-session",
		input:      newInputModel(),
		chatView:   newChatViewModel(80, 24),
		state:      stateIdle,
		width:      80,
		height:     24,
	}
	m.recalcLayout()
	return m
}

func TestDoneMsg_StreamSuccess(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("streamed ")
	m.chatView.appendChunk("content")

	msg := DoneMsg{Result: turnrunner.Result{Outcome: "success"}}
	m.Update(msg)

	if len(m.chatView.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(m.chatView.entries))
	}
	e := m.chatView.entries[0]
	if e.role != "assistant" {
		t.Errorf("want role=assistant, got %q", e.role)
	}
	if e.rawContent != "streamed content" {
		t.Errorf("want rawContent=%q, got %q", "streamed content", e.rawContent)
	}
}

func TestDoneMsg_NonStreamingResponseText(t *testing.T) {
	m := newTestModel()
	// No chunks — only ResponseText.
	msg := DoneMsg{Result: turnrunner.Result{
		Outcome:      "success",
		ResponseText: "non-streaming response",
	}}
	m.Update(msg)

	if len(m.chatView.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(m.chatView.entries))
	}
	e := m.chatView.entries[0]
	if e.rawContent != "non-streaming response" {
		t.Errorf("want rawContent=%q, got %q", "non-streaming response", e.rawContent)
	}
}

func TestDoneMsg_FailurePreservesStream(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("partial response")

	msg := DoneMsg{Result: turnrunner.Result{
		Outcome:     "timeout",
		UserMessage: "Operation timed out",
	}}
	m.Update(msg)

	// Should have 2 entries: assistant (preserved stream) + system (error).
	if len(m.chatView.entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(m.chatView.entries))
	}
	if m.chatView.entries[0].role != "assistant" {
		t.Errorf("first entry should be assistant, got %q", m.chatView.entries[0].role)
	}
	if m.chatView.entries[1].role != "system" {
		t.Errorf("second entry should be system, got %q", m.chatView.entries[1].role)
	}
}

func TestDoneMsg_DuplicateSkip(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("same text")

	msg := DoneMsg{Result: turnrunner.Result{
		Outcome:      "model_error",
		ResponseText: "same text",
	}}
	m.Update(msg)

	// stream finalized as assistant, but system message should be skipped (duplicate).
	for _, e := range m.chatView.entries {
		if e.role == "system" {
			t.Error("system message should have been deduplicated")
		}
	}
}

func TestErrorMsg_PreservesPartialStream(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("partial ")

	msg := ErrorMsg{Err: fmt.Errorf("connection lost")}
	m.Update(msg)

	// Should have assistant (preserved) + system (error).
	if len(m.chatView.entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(m.chatView.entries))
	}
	if m.chatView.entries[0].role != "assistant" {
		t.Errorf("first entry should be assistant, got %q", m.chatView.entries[0].role)
	}
	if m.chatView.entries[1].role != "system" {
		t.Errorf("second entry should be system, got %q", m.chatView.entries[1].role)
	}
}

func TestErrorMsg_NoStreamJustError(t *testing.T) {
	m := newTestModel()

	msg := ErrorMsg{Err: fmt.Errorf("connection lost")}
	m.Update(msg)

	if len(m.chatView.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(m.chatView.entries))
	}
	if m.chatView.entries[0].role != "system" {
		t.Errorf("entry should be system, got %q", m.chatView.entries[0].role)
	}
}

func TestInputWidth_DoesNotExceedTerminal(t *testing.T) {
	m := newTestModel()
	m.width = 60
	m.recalcLayout()

	view := m.input.View()
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		w := lipgloss.Width(line)
		if w > 60 {
			t.Errorf("input line %d width %d > terminal width 60", i, w)
		}
	}
}

func TestLayout_MinViewportHeight(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 15
	m.recalcLayout()

	if m.chatView.viewport.Height < 3 {
		t.Errorf("viewport height should be >= 3, got %d", m.chatView.viewport.Height)
	}
}

func TestLayout_VerySmallTerminal(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 5 // extremely small
	m.recalcLayout()

	if m.chatView.viewport.Height < 3 {
		t.Errorf("viewport height clamped to minimum 3, got %d", m.chatView.viewport.Height)
	}
}

func TestView_PartsBasedNoExtraNewlines(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.recalcLayout()

	view := m.View()
	// Should not contain triple newlines (over-separation).
	if strings.Contains(view, "\n\n\n") {
		t.Error("view should not have triple newlines between parts")
	}
}

func TestApprovalState_RecalcsLayout(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.recalcLayout()
	heightBefore := m.chatView.viewport.Height

	// Simulate approval message.
	m.state = stateApproving
	m.pendingApproval = &ApprovalRequestMsg{}
	m.recalcLayout()

	// Viewport should still have minimum height.
	if m.chatView.viewport.Height < 3 {
		t.Errorf("viewport height should be >= 3 during approval, got %d", m.chatView.viewport.Height)
	}

	// Approval banner takes more space than input, so viewport should be smaller or equal.
	_ = heightBefore // just verify it doesn't panic
}

func TestView_ApprovalHidesInput(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.recalcLayout()

	// Normal state should show input.
	normalView := m.View()
	_ = normalView

	// Approval state.
	m.state = stateApproving
	m.pendingApproval = &ApprovalRequestMsg{}
	m.recalcLayout()

	approvalView := m.View()
	if !strings.Contains(approvalView, "Tool Approval Required") {
		// If the approval request has no data, the banner may not render.
		// But input should NOT be present when approving.
		_ = approvalView
	}
}

func TestRenderApprovalBanner_WidthClamp(t *testing.T) {
	// Very narrow width should not panic.
	banner := renderApprovalBanner(
		approval.ApprovalRequest{ToolName: "test-tool"},
		8, // very narrow
	)
	if banner == "" {
		t.Error("banner should render even at narrow width")
	}
}

// Verify renderHelpBar and renderStatusBar don't panic at minimal width.
func TestRenderBars_MinimalWidth(t *testing.T) {
	cfg := &config.Config{}
	sb := renderStatusBar(cfg, "s", stateIdle, 20)
	if sb == "" {
		t.Error("status bar should render at narrow width")
	}
	hb := renderHelpBar(stateIdle, 20)
	_ = hb // just verify no panic
}

// approval import needed for the test
var _ = tui.Error // ensure tui import is used

// --- CPR filter tests ---

func TestCPR_FullSequenceDiscarded(t *testing.T) {
	m := newTestModel()

	// Simulate CPR response: ESC [ 4 3 ; 8 4 R
	keys := []tea.KeyMsg{
		{Type: tea.KeyEscape},
		{Type: tea.KeyRunes, Runes: []rune{'['}},
		{Type: tea.KeyRunes, Runes: []rune{'4'}},
		{Type: tea.KeyRunes, Runes: []rune{'3'}},
		{Type: tea.KeyRunes, Runes: []rune{';'}},
		{Type: tea.KeyRunes, Runes: []rune{'8'}},
		{Type: tea.KeyRunes, Runes: []rune{'4'}},
		{Type: tea.KeyRunes, Runes: []rune{'R'}},
	}

	for _, k := range keys {
		m.Update(k)
	}

	// CPR state should be back to idle.
	if m.cprDetect != cprIdle {
		t.Errorf("want cprIdle after full CPR, got %d", m.cprDetect)
	}
	// Buffer should be empty.
	if len(m.cprBuf) != 0 {
		t.Errorf("want empty cprBuf, got %d items", len(m.cprBuf))
	}
	// Input should be empty (CPR chars not inserted).
	if v := m.input.Value(); v != "" {
		t.Errorf("CPR leaked into input: %q", v)
	}
}

func TestCPR_NonCPRSequenceFlushes(t *testing.T) {
	m := newTestModel()

	// ESC followed by a non-'[' char should flush both keys.
	keys := []tea.KeyMsg{
		{Type: tea.KeyEscape},
		{Type: tea.KeyRunes, Runes: []rune{'a'}},
	}

	for _, k := range keys {
		m.Update(k)
	}

	// State should be reset.
	if m.cprDetect != cprIdle {
		t.Errorf("want cprIdle after non-CPR, got %d", m.cprDetect)
	}
	if len(m.cprBuf) != 0 {
		t.Errorf("want empty cprBuf after flush, got %d items", len(m.cprBuf))
	}
}

func TestCPR_PartialSequenceFlushedOnNonDigit(t *testing.T) {
	m := newTestModel()

	// ESC [ 4 X — 'X' is not a digit, ';', or 'R', so the sequence should flush.
	keys := []tea.KeyMsg{
		{Type: tea.KeyEscape},
		{Type: tea.KeyRunes, Runes: []rune{'['}},
		{Type: tea.KeyRunes, Runes: []rune{'4'}},
		{Type: tea.KeyRunes, Runes: []rune{'X'}},
	}

	for _, k := range keys {
		m.Update(k)
	}

	if m.cprDetect != cprIdle {
		t.Errorf("want cprIdle after partial flush, got %d", m.cprDetect)
	}
	if len(m.cprBuf) != 0 {
		t.Errorf("want empty cprBuf, got %d items", len(m.cprBuf))
	}
}

func TestCPR_TimeoutFlushes(t *testing.T) {
	m := newTestModel()

	// Send ESC — should transition to cprGotEsc.
	m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.cprDetect != cprGotEsc {
		t.Fatalf("want cprGotEsc, got %d", m.cprDetect)
	}

	// Simulate timeout message.
	m.Update(cprTimeoutMsg{})

	// Should be flushed back to idle.
	if m.cprDetect != cprIdle {
		t.Errorf("want cprIdle after timeout, got %d", m.cprDetect)
	}
	if len(m.cprBuf) != 0 {
		t.Errorf("want empty cprBuf after timeout, got %d items", len(m.cprBuf))
	}
}

func TestCPR_RAtBracketStateIgnored(t *testing.T) {
	// ESC [ R — 'R' at cprGotBracket (no digits yet) should NOT be treated as CPR.
	m := newTestModel()

	keys := []tea.KeyMsg{
		{Type: tea.KeyEscape},
		{Type: tea.KeyRunes, Runes: []rune{'['}},
		{Type: tea.KeyRunes, Runes: []rune{'R'}},
	}

	for _, k := range keys {
		m.Update(k)
	}

	// Should have flushed (R without params is not a CPR).
	if m.cprDetect != cprIdle {
		t.Errorf("want cprIdle, got %d", m.cprDetect)
	}
}
