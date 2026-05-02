package chat

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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

func newTestModelWithSharedPending(shared PendingApprovalStore) *ChatModel {
	m := newTestModel()
	m.sharedPending = shared
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

func TestRedirect_DuringStreamingQueuesAndCancels(t *testing.T) {
	m := newTestModel()
	m.state = stateStreaming
	cancelled := false
	m.cancelFn = func() { cancelled = true }
	m.chatView.appendChunk("partial response")

	// Directly set the field and call handleStreamingKey to simulate
	// Enter with content, bypassing textarea's own Enter→newline behavior.
	m.input.textarea.SetValue("new question")
	m.handleStreamingKey(tea.KeyMsg{Type: tea.KeyEnter})

	if m.pendingRedirectInput != "new question" {
		t.Fatalf("want pendingRedirectInput='new question', got %q", m.pendingRedirectInput)
	}
	if !cancelled {
		t.Fatal("cancelFn should have been called")
	}
}

func TestRedirect_EmptyInputDuringStreamingIgnored(t *testing.T) {
	m := newTestModel()
	m.state = stateStreaming
	cancelled := false
	m.cancelFn = func() { cancelled = true }

	m.input.textarea.SetValue("")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.pendingRedirectInput != "" {
		t.Fatalf("want empty pendingRedirectInput, got %q", m.pendingRedirectInput)
	}
	if cancelled {
		t.Fatal("cancelFn should not be called for empty input")
	}
}

func TestRedirect_DoneMsgConsumesAndResubmits(t *testing.T) {
	m := newTestModel()
	m.pendingRedirectInput = "follow-up question"

	// DoneMsg short-circuit path modifies m in place (pointer receiver).
	m.Update(DoneMsg{Result: turnrunner.Result{Outcome: "timeout"}})

	if m.pendingRedirectInput != "" {
		t.Fatalf("pendingRedirectInput should be cleared, got %q", m.pendingRedirectInput)
	}
	// The user input should appear as a user entry in the chatView.
	var found bool
	for _, e := range m.chatView.entries {
		if e.kind == itemUser && e.content == "follow-up question" {
			found = true
			break
		}
	}
	if !found {
		kinds := make([]string, len(m.chatView.entries))
		for i, e := range m.chatView.entries {
			kinds[i] = fmt.Sprintf("%s:content=%q", e.kind, e.content)
		}
		t.Fatalf("redirect input should be appended as user entry, got entries: %v", kinds)
	}
	if m.state != stateStreaming {
		t.Fatalf("want stateStreaming after redirect, got %v", m.state)
	}
}

func TestRedirect_DoneMsgWithoutRedirectPreservesExistingBehavior(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("partial")

	m.Update(DoneMsg{Result: turnrunner.Result{
		Outcome:     "timeout",
		UserMessage: "Request timed out",
	}})

	if m.state != stateFailed {
		t.Fatalf("want stateFailed without redirect, got %v", m.state)
	}
}

func TestApprovalSharedRenderUsesRegistry(t *testing.T) {
	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-1",
				ToolName: "exec",
				Summary:  "Run command",
			},
			ViewModel: approval.ApprovalViewModel{
				Risk: approval.RiskIndicator{Level: "moderate", Label: "Runs command"},
			},
			Response: make(chan approval.ApprovalResponse, 1),
		},
	}
	m := newTestModelWithSharedPending(shared)
	m.state = stateApproving

	view := m.View()
	if !strings.Contains(view, "Tool Approval Required") {
		t.Fatal("shared pending approval should be rendered")
	}
}

func TestCockpitApprovalSharedResolveUsesRegistry(t *testing.T) {
	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-1",
				ToolName: "browser_search",
			},
			ViewModel: approval.ApprovalViewModel{
				Risk: approval.RiskIndicator{Level: "moderate", Label: "Reads data"},
			},
			Response: make(chan approval.ApprovalResponse, 1),
		},
		resolveOK: true,
	}
	m := newTestModelWithSharedPending(shared)
	m.state = stateApproving

	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	m.Update(aKey)

	if shared.resolveCount != 1 {
		t.Fatalf("want shared registry resolve count 1, got %d", shared.resolveCount)
	}
	if shared.lastResolvedID != "apr-1" {
		t.Fatalf("want resolve id apr-1, got %q", shared.lastResolvedID)
	}
	if m.state != stateStreaming {
		t.Fatalf("want stateStreaming after shared approval, got %v", m.state)
	}
}

func TestCockpitActivityUserSubmissionCallback(t *testing.T) {
	var gotSession string
	var gotInput string
	m := newTestModel()
	m.onUserSubmission = func(sessionKey, input string) {
		gotSession = sessionKey
		gotInput = input
	}

	m.input.textarea.SetValue("ship it")
	m.handleIdleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if gotSession != "test-session" {
		t.Fatalf("want session test-session, got %q", gotSession)
	}
	if gotInput != "ship it" {
		t.Fatalf("want input ship it, got %q", gotInput)
	}
}

func TestCockpitActivityTurnSummaryCallback(t *testing.T) {
	var gotSession string
	var got TurnTokenUsageMsg
	m := newTestModel()
	m.onTurnSummary = func(sessionKey string, msg TurnTokenUsageMsg) {
		gotSession = sessionKey
		got = msg
	}

	m.Update(TurnTokenUsageMsg{
		InputTokens:      12,
		OutputTokens:     18,
		TotalTokens:      30,
		CacheTokens:      4,
		EstimatedCostUSD: 0.42,
	})

	if gotSession != "test-session" {
		t.Fatalf("want session test-session, got %q", gotSession)
	}
	if got.TotalTokens != 30 {
		t.Fatalf("want total tokens 30, got %d", got.TotalTokens)
	}
}

type stubSharedPendingStore struct {
	latest         *ApprovalRequestMsg
	resolveCount   int
	lastResolvedID string
	lastResponse   approval.ApprovalResponse
	resolveOK      bool
}

func (s *stubSharedPendingStore) Latest() *ApprovalRequestMsg {
	return s.latest
}

func (s *stubSharedPendingStore) HasPending() bool {
	return s.latest != nil
}

func (s *stubSharedPendingStore) Resolve(id string, resp approval.ApprovalResponse) bool {
	s.resolveCount++
	s.lastResolvedID = id
	s.lastResponse = resp
	if s.resolveOK {
		s.latest = nil
	}
	return s.resolveOK
}

func (s *stubSharedPendingStore) Register(_ ApprovalRequestMsg) {}

func (s *stubSharedPendingStore) CurrentTime() time.Time {
	return time.Now()
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
