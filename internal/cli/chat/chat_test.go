package chat

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/turnrunner"
)

func newTestModel() *ChatModel {
	m := &ChatModel{
		cfg: &config.Config{
			Agent: config.AgentConfig{
				Provider: "openai",
				Model:    "gpt-test",
			},
		},
		sessionKey: "test-session",
		input:      newInputModel(),
		chatView:   newChatViewModel(80, 24),
		state:      stateIdle,
		width:      80,
		height:     24,
	}
	if cmd := m.input.SetState(stateIdle); cmd != nil {
		_ = cmd
	}
	m.recalcLayout()
	return m
}

func TestDoneMsg_StreamSuccess(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("streamed ")
	m.chatView.appendChunk("content")

	m.Update(DoneMsg{Result: turnrunner.Result{Outcome: "success"}})

	if len(m.chatView.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(m.chatView.entries))
	}
	e := m.chatView.entries[0]
	if e.kind != itemAssistant {
		t.Fatalf("want assistant item, got %q", e.kind)
	}
	if e.rawContent != "streamed content" {
		t.Fatalf("want raw content preserved, got %q", e.rawContent)
	}
	if m.state != stateIdle {
		t.Fatalf("want stateIdle, got %v", m.state)
	}
}

func TestDoneMsg_NonStreamingResponseText(t *testing.T) {
	m := newTestModel()

	m.Update(DoneMsg{Result: turnrunner.Result{
		Outcome:      "success",
		ResponseText: "non-streaming response",
	}})

	if len(m.chatView.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(m.chatView.entries))
	}
	if got := m.chatView.entries[0].rawContent; got != "non-streaming response" {
		t.Fatalf("want raw response text, got %q", got)
	}
}

func TestDoneMsg_FailurePreservesStreamAndStatus(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("partial response")

	m.Update(DoneMsg{Result: turnrunner.Result{
		Outcome:     "timeout",
		UserMessage: "Operation timed out",
	}})

	if len(m.chatView.entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(m.chatView.entries))
	}
	if m.chatView.entries[0].kind != itemAssistant {
		t.Fatalf("first item should be assistant, got %q", m.chatView.entries[0].kind)
	}
	if m.chatView.entries[1].kind != itemStatus {
		t.Fatalf("second item should be status, got %q", m.chatView.entries[1].kind)
	}
	if m.state != stateFailed {
		t.Fatalf("want stateFailed, got %v", m.state)
	}
}

func TestDoneMsg_DuplicateFailureStatusSkipped(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("same text")

	m.Update(DoneMsg{Result: turnrunner.Result{
		Outcome:      "model_error",
		ResponseText: "same text",
	}})

	if len(m.chatView.entries) != 1 {
		t.Fatalf("want only assistant entry, got %d", len(m.chatView.entries))
	}
}

func TestErrorMsg_PreservesPartialStream(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("partial ")

	m.Update(ErrorMsg{Err: fmt.Errorf("connection lost")})

	if len(m.chatView.entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(m.chatView.entries))
	}
	if m.chatView.entries[0].kind != itemAssistant {
		t.Fatalf("first item should be assistant, got %q", m.chatView.entries[0].kind)
	}
	if m.chatView.entries[1].kind != itemStatus {
		t.Fatalf("second item should be status, got %q", m.chatView.entries[1].kind)
	}
	if m.state != stateFailed {
		t.Fatalf("want stateFailed, got %v", m.state)
	}
}

func TestErrorMsg_CancelledReturnsIdle(t *testing.T) {
	m := newTestModel()
	m.state = stateCancelling

	m.Update(ErrorMsg{Err: context.Canceled})

	if m.state != stateIdle {
		t.Fatalf("want stateIdle, got %v", m.state)
	}
	if len(m.chatView.entries) != 1 || m.chatView.entries[0].kind != itemStatus {
		t.Fatalf("want one cancellation status entry, got %#v", m.chatView.entries)
	}
}

func TestFailedStateResetsOnNextKeyInteraction(t *testing.T) {
	m := newTestModel()
	m.state = stateFailed

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	updated := model.(*ChatModel)

	if updated.state != stateIdle {
		t.Fatalf("want failed state reset to idle on next key, got %v", updated.state)
	}
}

func TestInputWidth_DoesNotExceedTerminal(t *testing.T) {
	m := newTestModel()
	m.width = 60
	m.recalcLayout()

	view := m.input.View()
	for i, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > 60 {
			t.Fatalf("input line %d width exceeds terminal width", i)
		}
	}
}

func TestLayout_MinViewportHeight(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 15
	m.recalcLayout()

	if m.chatView.viewport.Height < 3 {
		t.Fatalf("viewport height should be at least 3, got %d", m.chatView.viewport.Height)
	}
}

func TestView_PartsBasedNoTripleNewlines(t *testing.T) {
	m := newTestModel()
	view := m.View()
	if strings.Contains(view, "\n\n\n") {
		t.Fatal("view should not contain triple newlines")
	}
}

func TestApprovalState_RecalcsLayoutAndHidesComposer(t *testing.T) {
	m := newTestModel()
	m.Update(ApprovalRequestMsg{
		Request:  approval.ApprovalRequest{ToolName: "exec", Summary: "Run command"},
		Response: make(chan approval.ApprovalResponse, 1),
	})

	if m.chatView.viewport.Height < 3 {
		t.Fatalf("viewport height should remain clamped, got %d", m.chatView.viewport.Height)
	}

	view := m.View()
	if !strings.Contains(view, "Tool Approval Required") {
		t.Fatal("approval card should be rendered")
	}
	if strings.Contains(view, defaultComposerPlaceholder) {
		t.Fatal("composer should be hidden while approval card is active")
	}
}

func TestRenderBars_MinimalWidth(t *testing.T) {
	cfg := &config.Config{Agent: config.AgentConfig{Provider: "openai", Model: "gpt-test"}}
	if renderHeader(cfg, "s", 20) == "" {
		t.Fatal("header should render at narrow width")
	}
	if renderTurnStrip(stateStreaming, 20) == "" {
		t.Fatal("turn strip should render at narrow width")
	}
}

func TestCPRFullSequenceDiscarded(t *testing.T) {
	m := newTestModel()
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

	if m.cpr.state != cprIdle {
		t.Fatalf("want cprIdle after full CPR, got %v", m.cpr.state)
	}
	if len(m.cpr.buf) != 0 {
		t.Fatalf("want empty cprBuf, got %d", len(m.cpr.buf))
	}
	if got := m.input.Value(); got != "" {
		t.Fatalf("CPR leaked into input: %q", got)
	}
}

func TestOSC11BELSequenceDiscarded(t *testing.T) {
	m := newTestModel()
	keys := []tea.KeyMsg{
		{Type: tea.KeyEscape},
		{Type: tea.KeyRunes, Runes: []rune{']'}},
		{Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune{';'}},
		{Type: tea.KeyRunes, Runes: []rune{'r'}},
		{Type: tea.KeyRunes, Runes: []rune{'g'}},
		{Type: tea.KeyRunes, Runes: []rune{'b'}},
		{Type: tea.KeyRunes, Runes: []rune{':'}},
		{Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyCtrlG},
	}
	for _, k := range keys {
		m.Update(k)
	}

	if m.cpr.state != cprIdle {
		t.Fatalf("want cprIdle after OSC BEL sequence, got %v", m.cpr.state)
	}
	if got := m.input.Value(); got != "" {
		t.Fatalf("OSC sequence leaked into input: %q", got)
	}
}

func TestOSCSTSequenceDiscarded(t *testing.T) {
	m := newTestModel()
	keys := []tea.KeyMsg{
		{Type: tea.KeyEscape},
		{Type: tea.KeyRunes, Runes: []rune{']'}},
		{Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune{';'}},
		{Type: tea.KeyRunes, Runes: []rune{'?'}},
		{Type: tea.KeyEscape},
		{Type: tea.KeyRunes, Runes: []rune{'\\'}},
	}
	for _, k := range keys {
		m.Update(k)
	}

	if m.cpr.state != cprIdle {
		t.Fatalf("want cprIdle after OSC ST sequence, got %v", m.cpr.state)
	}
	if got := m.input.Value(); got != "" {
		t.Fatalf("OSC ST sequence leaked into input: %q", got)
	}
}

func TestCPRFilterIgnoredDuringApproval(t *testing.T) {
	m := newTestModel()
	m.state = stateApproving

	m.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if m.cpr.state != cprIdle {
		t.Fatalf("CPR filter should remain idle outside composer, got %v", m.cpr.state)
	}
}

func TestAltSequencePreserved(t *testing.T) {
	m := newTestModel()
	m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true})

	if got := m.input.Value(); got == "" {
		t.Fatal("non-OSC/non-CPR buffered sequence should replay into composer input")
	}
}

func TestTranscriptBlocksUseVisualSeparators(t *testing.T) {
	cv := newChatViewModel(80, 24)
	cv.appendUser("hello")
	cv.appendAssistant("world")

	content := cv.viewport.View()
	if !strings.Contains(content, "─") {
		t.Fatal("transcript blocks should include visible separator lines")
	}
	if !strings.Contains(content, "│") {
		t.Fatal("transcript blocks should include a left accent border")
	}
}

func TestDoublePress_CriticalFirstPress(t *testing.T) {
	m := newTestModel()
	m.state = stateApproving
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ToolName:    "exec",
			SafetyLevel: "dangerous",
			Category:    "automation",
		},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"},
		},
		Response: make(chan approval.ApprovalResponse, 1),
	}

	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	m.Update(aKey)

	if !m.approval.confirmPending {
		t.Fatal("first 'a' on critical tool should set approvalConfirmPending=true")
	}
	if m.approval.pending == nil {
		t.Fatal("first 'a' on critical tool should not consume the approval")
	}
}

func TestDoublePress_CriticalSecondPress(t *testing.T) {
	m := newTestModel()
	m.state = stateApproving
	respCh := make(chan approval.ApprovalResponse, 1)
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ToolName:    "exec",
			SafetyLevel: "dangerous",
			Category:    "automation",
		},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"},
		},
		Response: respCh,
	}

	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	m.Update(aKey) // first press
	m.Update(aKey) // second press

	if m.approval.pending != nil {
		t.Fatal("second 'a' should consume the approval")
	}
	if m.approval.confirmPending {
		t.Fatal("confirm pending should be cleared after approval")
	}
}

func TestDoublePress_NonCriticalImmediateApproval(t *testing.T) {
	m := newTestModel()
	m.state = stateApproving
	respCh := make(chan approval.ApprovalResponse, 1)
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ToolName:    "browser_search",
			SafetyLevel: "moderate",
			Category:    "browser",
		},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "moderate", Label: "Reads data"},
		},
		Response: respCh,
	}

	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	m.Update(aKey)

	if m.approval.pending != nil {
		t.Fatal("non-critical tool should be approved immediately on first 'a'")
	}
	if m.approval.confirmPending {
		t.Fatal("confirm pending should not be set for non-critical tool")
	}
}

func TestDoublePress_OtherKeyResetsConfirm(t *testing.T) {
	m := newTestModel()
	m.state = stateApproving
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ToolName:    "exec",
			SafetyLevel: "dangerous",
			Category:    "automation",
		},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"},
		},
		Response: make(chan approval.ApprovalResponse, 1),
	}

	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	m.Update(aKey) // first press — sets confirmPending

	if !m.approval.confirmPending {
		t.Fatal("expected confirmPending=true after first 'a'")
	}

	// Press an unrelated key — should reset.
	xKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	m.Update(xKey)

	if m.approval.confirmPending {
		t.Fatal("unrelated key should reset approvalConfirmPending")
	}
	if m.approval.pending == nil {
		t.Fatal("unrelated key should not consume the approval")
	}
}

func TestDoublePress_DenyResetsConfirm(t *testing.T) {
	m := newTestModel()
	m.state = stateApproving
	respCh := make(chan approval.ApprovalResponse, 1)
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ToolName:    "exec",
			SafetyLevel: "dangerous",
			Category:    "automation",
		},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"},
		},
		Response: respCh,
	}

	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	m.Update(aKey) // first press — sets confirmPending

	dKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	m.Update(dKey) // deny

	if m.approval.confirmPending {
		t.Fatal("deny should reset approvalConfirmPending")
	}
	if m.approval.pending != nil {
		t.Fatal("deny should consume the approval")
	}
}

func TestCPRTimeoutFlushesEsc(t *testing.T) {
	m := newTestModel()
	m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.cpr.state != cprGotEsc {
		t.Fatalf("want cprGotEsc, got %v", m.cpr.state)
	}
	m.Update(cprTimeoutMsg{})
	if m.cpr.state != cprIdle {
		t.Fatalf("want cprIdle after timeout, got %v", m.cpr.state)
	}
}
