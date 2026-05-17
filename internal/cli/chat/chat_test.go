package chat

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turnrunner"
)

type submitCaptureExecutor struct {
	ctx       context.Context
	sessionID string
	input     string
}

func (s *submitCaptureExecutor) RunStreamingDetailed(
	ctx context.Context,
	sessionID, input string,
	_ adk.ChunkCallback,
	_ ...adk.RunOption,
) (adk.RunReport, error) {
	s.ctx = ctx
	s.sessionID = sessionID
	s.input = input
	return adk.RunReport{Response: "ok"}, nil
}

type submitTestSessionStore struct{}

func (submitTestSessionStore) Create(*session.Session) error               { return nil }
func (submitTestSessionStore) Get(string) (*session.Session, error)        { return nil, nil }
func (submitTestSessionStore) Update(*session.Session) error               { return nil }
func (submitTestSessionStore) Delete(string) error                         { return nil }
func (submitTestSessionStore) AppendMessage(string, session.Message) error { return nil }
func (submitTestSessionStore) AnnotateTimeout(string, string) error        { return nil }
func (submitTestSessionStore) End(string) error                            { return nil }
func (submitTestSessionStore) Close() error                                { return nil }
func (submitTestSessionStore) ListSessions(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}
func (submitTestSessionStore) GetSalt(string) ([]byte, error) { return nil, nil }
func (submitTestSessionStore) SetSalt(string, []byte) error   { return nil }

func newTestModel() *ChatModel {
	m := &ChatModel{
		cfg:        readyRemoteConfig(),
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

func incompleteSetupConfig() *config.Config {
	return config.DefaultConfig()
}

func readyRemoteConfig() *config.Config {
	return &config.Config{
		Agent: config.AgentConfig{
			Provider: "openai",
			Model:    "gpt-test",
		},
		Providers: map[string]config.ProviderConfig{
			"openai": {Type: "openai", APIKey: "sk-test"},
		},
	}
}

func readyOllamaConfig() *config.Config {
	return &config.Config{
		Agent: config.AgentConfig{
			Provider: "local-ollama",
			Model:    "llama3.1",
		},
		Providers: map[string]config.ProviderConfig{
			"local-ollama": {Type: "ollama"},
		},
	}
}

func newTestModelWithSharedPending(shared PendingApprovalStore) *ChatModel {
	m := newTestModel()
	m.sharedPending = shared
	return m
}

func runStatusCommandText(t *testing.T, m *ChatModel) string {
	t.Helper()
	cmd := cmdStatus(m, "")
	require.NotNil(t, cmd)
	msg := cmd()
	sys, ok := msg.(SystemMsg)
	require.True(t, ok, "expected SystemMsg, got %T", msg)
	return sys.Text
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

func TestStatusCommandReportsActiveMCPRuntime(t *testing.T) {
	m := newTestModel()
	m.cfg.MCP.Enabled = true
	m.runtimeFeatures = RuntimeFeatures{
		MCPActive:      true,
		MCPServerCount: 2,
		MCPToolCount:   5,
	}

	out := ansi.Strip(runStatusCommandText(t, m))

	assert.Contains(t, out, "MCP")
	assert.Contains(t, out, "active in TUI mode")
	assert.Contains(t, out, "2 servers")
	assert.Contains(t, out, "5 tools")
	assert.NotContains(t, out, "MCP configured but not active in TUI mode")
}

func TestStatusCommandKeepsConfiguredOnlyMCPDistinct(t *testing.T) {
	m := newTestModel()
	m.cfg.MCP.Enabled = true

	out := ansi.Strip(runStatusCommandText(t, m))

	assert.Contains(t, out, "MCP")
	assert.Contains(t, out, "configured but no active MCP runtime")
	assert.NotContains(t, out, "active in TUI mode")
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

func TestDoneMsg_DuplicateFailureStatusSkippedAfterSanitization(t *testing.T) {
	m := newTestModel()
	m.chatView.appendChunk("\x1b[31msame text\x1b[0m")

	m.Update(DoneMsg{Result: turnrunner.Result{
		Outcome:      "model_error",
		ResponseText: "\x1b[31msame text\x1b[0m",
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

func TestSetupReadiness_RenderPartsShowsSetupRequiredState(t *testing.T) {
	m := New(Deps{
		Config:     incompleteSetupConfig(),
		SessionKey: "setup-test",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(*ChatModel)

	parts := m.RenderParts()
	out := ansi.Strip(strings.Join([]string{
		parts.Header,
		parts.TurnStrip,
		parts.Footer,
	}, "\n"))

	assert.Contains(t, out, "Setup Required")
	assert.Contains(t, out, "lango onboard")
	assert.Contains(t, out, "lango settings")
	assert.Contains(t, out, "lango doctor")
	assert.NotContains(t, out, "Ready")
	assert.NotContains(t, out, "Enter send")
}

func TestSetupReadiness_BlocksNormalSubmissionAndKeepsDraft(t *testing.T) {
	executor := &submitCaptureExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil)
	m := New(Deps{
		TurnRunner: runner,
		Config:     incompleteSetupConfig(),
		SessionKey: "test-session",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(*ChatModel)
	m.SetComposerValue("explain this code")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ChatModel)
	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = m.Update(msg)
		m = updated.(*ChatModel)
	}

	assert.Empty(t, executor.input)
	assert.Equal(t, "explain this code", m.ComposerValue())
	assert.Equal(t, stateIdle, m.state)
	require.NotEmpty(t, m.chatView.entries)
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemStatus, last.kind)
	assert.Contains(t, last.content, "lango onboard")
	assert.Contains(t, last.content, "lango settings")
	assert.Contains(t, last.content, "lango doctor")
}

func TestSetupReadiness_SlashCommandsStillDispatchBeforeSetup(t *testing.T) {
	executor := &submitCaptureExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil)
	m := New(Deps{
		TurnRunner: runner,
		Config:     incompleteSetupConfig(),
		SessionKey: "test-session",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(*ChatModel)
	m.SetComposerValue("/help")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ChatModel)
	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = m.Update(msg)
		m = updated.(*ChatModel)
	}

	assert.Empty(t, executor.input)
	require.NotEmpty(t, m.chatView.entries)
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemSystem, last.kind)
	assert.Contains(t, last.content, "Commands")
}

func TestSetupReadiness_ReadyRemoteProviderSubmitsNormally(t *testing.T) {
	executor := &submitCaptureExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil)
	m := New(Deps{
		TurnRunner: runner,
		Config:     readyRemoteConfig(),
		SessionKey: "test-session",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(*ChatModel)
	m.SetComposerValue("ready remote")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ChatModel)
	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = m.Update(msg)
		m = updated.(*ChatModel)
	}

	assert.Equal(t, "ready remote", executor.input)
}

func TestSetupReadiness_ReadyOllamaProviderSubmitsNormally(t *testing.T) {
	executor := &submitCaptureExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil)
	m := New(Deps{
		TurnRunner: runner,
		Config:     readyOllamaConfig(),
		SessionKey: "test-session",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(*ChatModel)
	m.SetComposerValue("ready ollama")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ChatModel)
	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = m.Update(msg)
		m = updated.(*ChatModel)
	}

	assert.Equal(t, "ready ollama", executor.input)
}

func TestHelpCommandMentionsFailedStateDoublePressQuit(t *testing.T) {
	m := newTestModel()
	cmd := cmdHelp(m, "")
	require.NotNil(t, cmd)

	msg := cmd()
	sys, ok := msg.(SystemMsg)
	require.True(t, ok)
	assert.Contains(t, sys.Text, "double-press to quit when idle or failed")
	assert.Contains(t, sys.Text, "Scroll transcript")
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

func TestApprovalRequestMarksLatestToolRowAwaitingApproval(t *testing.T) {
	m := newTestModel()
	m.chatView.appendToolStart("call1", "exec", map[string]any{"path": "/tmp/x"})

	m.Update(ApprovalRequestMsg{
		Request: approval.ApprovalRequest{ToolName: "exec"},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"},
		},
		Response: make(chan approval.ApprovalResponse, 1),
	})

	require.Len(t, m.chatView.entries, 2)
	assert.Equal(t, string(toolStateAwaitingApproval), m.chatView.entries[0].meta["state"])
}

func TestApprovalDeniedMarksToolRowCanceled(t *testing.T) {
	m := newTestModel()
	respCh := make(chan approval.ApprovalResponse, 1)
	m.chatView.appendToolStart("call1", "exec", map[string]any{"path": "/tmp/x"})
	m.state = stateApproving
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ToolName: "exec",
		},
		Response: respCh,
	}
	m.chatView.transitionLatestToolState("exec", []ToolItemState{toolStateRunning}, toolStateAwaitingApproval)

	dKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	m.Update(dKey)

	require.Len(t, m.chatView.entries, 2)
	assert.Equal(t, string(toolStateCanceled), m.chatView.entries[0].meta["state"])
}

func TestApprovalGrantRestoresToolRowRunning(t *testing.T) {
	m := newTestModel()
	respCh := make(chan approval.ApprovalResponse, 1)
	m.chatView.appendToolStart("call1", "browser_search", map[string]any{"q": "hello"})
	m.state = stateApproving
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ToolName: "browser_search",
		},
		Response: respCh,
	}
	m.chatView.transitionLatestToolState("browser_search", []ToolItemState{toolStateRunning}, toolStateAwaitingApproval)

	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	m.Update(aKey)

	require.Len(t, m.chatView.entries, 2)
	assert.Equal(t, string(toolStateRunning), m.chatView.entries[0].meta["state"])
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
				Summary:  "Reads data",
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
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Contains(t, last.content, "Reads data")
	assert.Contains(t, last.content, "[apr-1]")
}

func TestCockpitApprovalSharedResolveKeepsApprovingForNextPending(t *testing.T) {
	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-1",
				ToolName: "browser_search",
				Summary:  "Reads data",
			},
			ViewModel: approval.ApprovalViewModel{
				Risk: approval.RiskIndicator{Level: "moderate", Label: "Reads data"},
			},
			Response: make(chan approval.ApprovalResponse, 1),
		},
		next: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-2",
				ToolName: "fs_write",
				Summary:  "Writes data",
			},
			ViewModel: approval.ApprovalViewModel{
				Risk: approval.RiskIndicator{Level: "high", Label: "Writes data"},
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
	if m.state != stateApproving {
		t.Fatalf("want stateApproving when another pending request is promoted, got %v", m.state)
	}
	if pending := m.currentPendingApproval(); pending == nil || pending.Request.ID != "apr-2" {
		t.Fatalf("want next pending approval apr-2, got %#v", pending)
	}
}

func TestCockpitApprovalSharedResolveFailureDoesNotAppendFalseSuccess(t *testing.T) {
	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-1",
				ToolName: "browser_search",
				Summary:  "Reads data",
			},
			ViewModel: approval.ApprovalViewModel{
				Risk: approval.RiskIndicator{Level: "moderate", Label: "Reads data"},
			},
			Response: make(chan approval.ApprovalResponse, 1),
		},
		resolveOK: false,
	}
	m := newTestModelWithSharedPending(shared)
	m.state = stateApproving

	before := len(m.chatView.entries)
	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	m.Update(aKey)

	if m.state != stateApproving {
		t.Fatalf("want stateApproving after shared resolve failure, got %v", m.state)
	}
	if shared.resolveCount != 1 {
		t.Fatalf("want shared registry resolve count 1, got %d", shared.resolveCount)
	}
	if m.currentPendingApproval() == nil {
		t.Fatal("pending approval should remain after shared resolve failure")
	}
	if len(m.chatView.entries) != before+1 {
		t.Fatalf("want one new error status entry, got %d new entries", len(m.chatView.entries)-before)
	}
	last := m.chatView.entries[len(m.chatView.entries)-1]
	if last.kind != itemStatus {
		t.Fatalf("want status entry on resolve failure, got %q", last.kind)
	}
	for _, entry := range m.chatView.entries {
		if entry.kind == itemApproval && strings.Contains(entry.content, "Approved") {
			t.Fatalf("unexpected approval success event after resolve failure: %q", entry.content)
		}
	}
}

func TestApprovalRequestEventIncludesRequestID(t *testing.T) {
	m := newTestModel()

	m.Update(ApprovalRequestMsg{
		Request: approval.ApprovalRequest{ID: "apr-42", ToolName: "exec", Summary: "Run command"},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"},
		},
		Response: make(chan approval.ApprovalResponse, 1),
	})

	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemApproval, last.kind)
	assert.Contains(t, last.content, "Run command")
	assert.Contains(t, last.content, "[apr-42]")
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

func TestSubmitComposerWithParentUsesProvidedMissionContext(t *testing.T) {
	executor := &submitCaptureExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil)
	m := New(Deps{
		TurnRunner: runner,
		Config:     readyRemoteConfig(),
		SessionKey: "test-session",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(*ChatModel)
	m.SetComposerValue("start mission")

	cmd := m.SubmitComposerWithParent(ctxkeys.WithMissionID(context.Background(), "mission-123"))
	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = m.Update(msg)
		m = updated.(*ChatModel)
	}

	if executor.input != "start mission" {
		t.Fatalf("want captured input start mission, got %q", executor.input)
	}
	if got := ctxkeys.MissionIDFromContext(executor.ctx); got != "mission-123" {
		t.Fatalf("want mission id mission-123, got %q", got)
	}
}

func TestComposerEnterDoesNotImplicitlyBindMission(t *testing.T) {
	executor := &submitCaptureExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil)
	m := New(Deps{
		TurnRunner: runner,
		Config:     readyRemoteConfig(),
		SessionKey: "test-session",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(*ChatModel)
	m.SetComposerValue("plain chat")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ChatModel)
	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = m.Update(msg)
		m = updated.(*ChatModel)
	}

	if executor.input != "plain chat" {
		t.Fatalf("want captured input plain chat, got %q", executor.input)
	}
	if got := ctxkeys.MissionIDFromContext(executor.ctx); got != "" {
		t.Fatalf("want no mission binding for normal chat submit, got %q", got)
	}
}

type stubSharedPendingStore struct {
	latest         *ApprovalRequestMsg
	next           *ApprovalRequestMsg
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
		s.latest = s.next
		s.next = nil
	}
	return s.resolveOK
}

func (s *stubSharedPendingStore) Register(_ ApprovalRequestMsg) {}

func collectImmediateMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	ch := make(chan tea.Msg, 1)
	go func() {
		ch <- cmd()
	}()

	select {
	case msg := <-ch:
		switch msg := msg.(type) {
		case nil:
			return nil
		case tea.BatchMsg:
			var out []tea.Msg
			for _, child := range msg {
				out = append(out, collectImmediateMsgs(child)...)
			}
			return out
		default:
			return []tea.Msg{msg}
		}
	case <-time.After(25 * time.Millisecond):
		return nil
	}
}

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
