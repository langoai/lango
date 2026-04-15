// Package chat implements an interactive TUI chat interface using bubbletea.
// It provides a coding-agent cockpit for conversing with the Lango agent,
// including streaming responses, inline tool approval, slash commands, and
// turn-state visibility.
package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turnrunner"
)

// Deps holds the dependencies injected into the chat model.
type Deps struct {
	TurnRunner        *turnrunner.Runner
	Config            *config.Config
	SessionKey        string
	SessionStore      session.Store       // optional; used by /mode to persist the session's mode
	EventBus          *eventbus.Bus       // optional; used by /mode to publish ModeChangedEvent
	BackgroundManager *background.Manager // optional, nil when background tasks unavailable
}

// cursorBlinkInterval is the period between cursor blink toggles.
const cursorBlinkInterval = 400 * time.Millisecond

// ChatParts holds the discrete rendered sections of the chat view.
type ChatParts struct {
	Header    string
	TurnStrip string
	Main      string
	Pending   string // empty when not in pending state
	TaskStrip string // empty when no background tasks
	Footer    string
	Approval  string // empty when no approval pending
}

// ChatModel is the root bubbletea model for the interactive TUI chat.
type ChatModel struct {
	turnRunner   *turnrunner.Runner
	cfg          *config.Config
	sessionKey   string
	sessionStore session.Store // optional; enables /mode to persist session mode
	eventBus     *eventbus.Bus // optional; enables /mode to publish ModeChangedEvent

	input    inputModel
	chatView chatViewModel

	state    chatState
	width    int
	height   int
	quitting bool

	runCtx   context.Context
	cancelFn context.CancelFunc

	approval approvalState

	program *tea.Program

	bgManager *background.Manager
	taskStrip taskStripModel
	pending   pendingIndicator

	lastCtrlC time.Time

	pendingRedirectInput string // queued input to submit after cancelled turn completes

	// Session cumulative usage (for /cost slash command).
	sessionInputTokens  int64
	sessionOutputTokens int64
	sessionCostUSD      float64

	cpr cprFilter
}

// New creates a new ChatModel with the given dependencies.
func New(deps Deps) *ChatModel {
	return &ChatModel{
		turnRunner:   deps.TurnRunner,
		cfg:          deps.Config,
		sessionKey:   deps.SessionKey,
		sessionStore: deps.SessionStore,
		eventBus:     deps.EventBus,
		bgManager:    deps.BackgroundManager,
		taskStrip:    newTaskStripModel(deps.BackgroundManager),
		input:        newInputModel(),
		chatView:     newChatViewModel(80, 20),
		state:        stateIdle,
	}
}

// currentModeName returns the active session mode name, or "" if none is set
// or the session store is unavailable.
func (m *ChatModel) currentModeName() string {
	if m.sessionStore == nil {
		return ""
	}
	s, err := m.sessionStore.Get(m.sessionKey)
	if err != nil || s == nil {
		return ""
	}
	return s.Mode()
}

// setSessionMode persists the given mode name on the active session and
// publishes a ModeChangedEvent so multi-channel subscribers can render the
// transition in their native format. An empty name clears the mode.
func (m *ChatModel) setSessionMode(name string) error {
	if m.sessionStore == nil {
		return fmt.Errorf("session store not available")
	}
	var oldMode string
	s, err := m.sessionStore.Get(m.sessionKey)
	if err != nil || s == nil {
		// Create the session record on the fly so /mode can be set before the
		// first turn (which would otherwise create the session lazily).
		s = &session.Session{Key: m.sessionKey}
		if createErr := m.sessionStore.Create(s); createErr != nil {
			return fmt.Errorf("create session: %w", createErr)
		}
	} else {
		oldMode = s.Mode()
	}
	s.SetMode(name)
	if err := m.sessionStore.Update(s); err != nil {
		return err
	}
	if m.eventBus != nil {
		m.eventBus.Publish(eventbus.ModeChangedEvent{
			SessionKey: m.sessionKey,
			OldMode:    oldMode,
			NewMode:    name,
		})
	}
	return nil
}

// SetProgram stores a reference to the tea.Program for sending messages from callbacks.
func (m *ChatModel) SetProgram(p *tea.Program) {
	m.program = p
}

// Init implements tea.Model.
func (m *ChatModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.SetWindowTitle("Lango Chat"),
		m.input.SetState(stateIdle),
	}
	if tick := m.taskStripTick(); tick != nil {
		cmds = append(cmds, tick)
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model.
func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcLayout()
		return m, nil

	case cprTimeoutMsg:
		if keys := m.cpr.HandleTimeout(); len(keys) > 0 {
			cmds = append(cmds, m.replayKeys(keys)...)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if m.state == stateFailed {
			if cmd := m.transitionTo(stateIdle); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		if m.inputAcceptsText() {
			filtered, filterCmds := m.cpr.Filter(msg)
			if len(filterCmds) > 0 {
				cmds = append(cmds, filterCmds...)
			}
			if !filtered && len(m.cpr.buf) > 0 {
				cmds = append(cmds, m.cprFlush()...)
			}
			if filtered {
				return m, tea.Batch(cmds...)
			}
		}

		cmd := m.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

	case ToolStartedMsg:
		m.dismissPending()
		m.chatView.appendToolStart(msg.CallID, msg.ToolName, msg.Params)
		return m, nil

	case ToolFinishedMsg:
		m.chatView.finalizeToolResult(msg.CallID, msg.Success, msg.Duration, msg.Output)
		return m, nil

	case ThinkingStartedMsg:
		m.dismissPending()
		m.chatView.appendThinking(msg.Summary)
		return m, nil

	case ThinkingFinishedMsg:
		m.chatView.finalizeThinking(msg.Summary, msg.Duration)
		return m, nil

	case TaskStripTickMsg:
		m.taskStrip.refresh()
		m.refreshView()
		return m, m.taskStripTick()

	case PendingIndicatorTickMsg:
		if m.pending.IsActive() {
			m.refreshView()
			return m, m.pending.TickCmd()
		}
		return m, nil

	case ChunkMsg:
		m.dismissPending()
		m.chatView.appendChunk(msg.Chunk)
		if !m.chatView.cursorTickActive {
			m.chatView.cursorTickActive = true
			m.chatView.showCursor = true
			cmds = append(cmds, tea.Tick(cursorBlinkInterval, func(t time.Time) tea.Msg {
				return CursorTickMsg(t)
			}))
		}
		return m, tea.Batch(cmds...)

	case CursorTickMsg:
		if m.state == stateStreaming {
			m.chatView.showCursor = !m.chatView.showCursor
			m.chatView.render()
			return m, tea.Tick(cursorBlinkInterval, func(t time.Time) tea.Msg {
				return CursorTickMsg(t)
			})
		}
		m.chatView.stopCursorBlink()
		return m, nil

	case DoneMsg:
		m.dismissPending()
		m.chatView.stopCursorBlink()

		// Short-circuit: if a redirect is queued, skip error display and
		// immediately submit the pending input as the next turn.
		if m.pendingRedirectInput != "" {
			redirect := m.pendingRedirectInput
			m.pendingRedirectInput = ""
			if m.chatView.streamBuf.Len() > 0 {
				m.chatView.finalizeStream()
			}
			m.chatView.appendUser(redirect)
			m.pending.Activate()
			return m, tea.Batch(
				m.transitionTo(stateStreaming),
				m.submitCmd(redirect),
				m.pending.TickCmd(),
			)
		}

		if m.chatView.streamBuf.Len() > 0 {
			m.chatView.finalizeStream()
		} else if strings.TrimSpace(msg.Result.ResponseText) != "" {
			m.chatView.appendAssistant(msg.Result.ResponseText)
		}

		nextState := stateIdle
		if msg.Result.Outcome != "success" {
			nextState = stateFailed
			text := strings.TrimSpace(msg.Result.UserMessage)
			if text == "" {
				text = strings.TrimSpace(msg.Result.ResponseText)
			}
			if text != "" && strings.TrimSpace(m.chatView.lastAssistantRaw()) != text {
				m.chatView.appendStatus(text, "error")
			}
		}

		if cmd := m.transitionTo(nextState); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case ErrorMsg:
		m.dismissPending()
		m.chatView.stopCursorBlink()
		if m.chatView.streamBuf.Len() > 0 {
			m.chatView.finalizeStream()
		}

		if errors.Is(msg.Err, context.Canceled) {
			m.chatView.appendStatus("Generation cancelled.", "warning")
			if cmd := m.transitionTo(stateIdle); cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		m.chatView.appendStatus(fmt.Sprintf("Error: %v", msg.Err), "error")
		if cmd := m.transitionTo(stateFailed); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case WarningMsg:
		m.chatView.appendStatus(
			fmt.Sprintf("Approaching timeout (%s / %s)", msg.Elapsed.Round(time.Second), msg.HardCeiling.Round(time.Second)),
			"warning",
		)
		return m, nil

	case ApprovalRequestMsg:
		m.dismissPending()
		m.approval.Reset(&msg)
		m.chatView.appendApprovalEvent(fmt.Sprintf("Approval requested for %s", msg.Request.ToolName), "requested")
		if cmd := m.transitionTo(stateApproving); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case DelegationMsg:
		m.dismissPending()
		m.chatView.appendDelegation(msg.From, msg.To, msg.Reason)
		return m, nil

	case BudgetWarningMsg:
		m.chatView.appendStatus(
			fmt.Sprintf("\u26A0 Delegation budget: %d/%d (%.0f%%)",
				msg.Used, msg.Max, float64(msg.Used)/float64(msg.Max)*100),
			"warning",
		)
		return m, nil

	case RecoveryMsg:
		m.chatView.appendRecovery(msg.Action, msg.CauseClass, msg.Attempt, msg.Backoff)
		return m, nil

	case TurnTokenUsageMsg:
		m.chatView.appendTokenSummaryWithCost(msg.InputTokens, msg.OutputTokens, msg.TotalTokens, msg.CacheTokens, msg.EstimatedCostUSD)
		// Track session cumulative totals for /cost summary.
		m.sessionInputTokens += msg.InputTokens
		m.sessionOutputTokens += msg.OutputTokens
		m.sessionCostUSD += msg.EstimatedCostUSD
		return m, nil

	case ChannelMessageMsg:
		m.chatView.appendChannel(msg.Channel, msg.SenderName, msg.Text, msg.SessionKey, msg.Metadata)
		return m, nil

	case SystemMsg:
		m.chatView.appendSystem(msg.Text)
		return m, nil
	}

	if m.inputAcceptsText() {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	var cmd tea.Cmd
	m.chatView, cmd = m.chatView.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
// RenderParts returns the discrete rendered sections of the chat view.
func (m *ChatModel) RenderParts() ChatParts {
	parts := ChatParts{
		Header:    renderHeader(m.cfg, truncateSessionKey(m.sessionKey), m.width),
		TurnStrip: renderTurnStrip(m.state, m.width),
		Main:      m.chatView.View(),
		TaskStrip: m.taskStrip.View(m.width),
		Footer:    renderFooter(m.input, m.state, m.width),
	}

	if m.pending.IsActive() {
		parts.Pending = renderPendingIndicator(m.pending.Elapsed())
	}

	if m.state == stateApproving && m.approval.HasPending() {
		parts.Approval = renderApproval(m.approval.pending, &m.approval, m.width, m.height)
	}

	return parts
}

// View implements tea.Model.
func (m *ChatModel) View() string {
	if m.quitting {
		return ""
	}
	if m.width == 0 || m.height == 0 {
		return "\n  Waiting for terminal size..."
	}

	p := m.RenderParts()

	sections := make([]string, 0, 7)
	sections = append(sections, p.Header, p.TurnStrip, p.Main)
	if p.Pending != "" {
		sections = append(sections, p.Pending)
	}
	if p.TaskStrip != "" {
		sections = append(sections, p.TaskStrip)
	}
	if p.Approval != "" {
		sections = append(sections, p.Approval)
	}
	sections = append(sections, p.Footer)

	return strings.Join(sections, "\n")
}

// dismissPending clears the pending indicator when the first content event arrives.
func (m *ChatModel) dismissPending() {
	if !m.pending.IsActive() {
		return
	}
	m.pending.Dismiss()
	if m.width > 0 && m.height > 0 {
		m.recalcLayout()
	}
}

// refreshView re-renders the chat view (used after task strip refresh).
func (m *ChatModel) refreshView() {
	m.chatView.render()
}

// taskStripTick returns a tick command for periodic task strip refresh.
func (m *ChatModel) taskStripTick() tea.Cmd {
	if m.bgManager == nil {
		return nil
	}
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return TaskStripTickMsg(t)
	})
}

func (m *ChatModel) inputAcceptsText() bool {
	return m.state == stateIdle || m.state == stateFailed || m.state == stateStreaming
}

func (m *ChatModel) transitionTo(state chatState) tea.Cmd {
	m.state = state
	cmd := m.input.SetState(state)
	if m.width > 0 && m.height > 0 {
		m.recalcLayout()
	}
	return cmd
}

func (m *ChatModel) handleKey(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+d"))) {
		m.quitting = true
		if m.cancelFn != nil {
			m.cancelFn()
		}
		return tea.Quit
	}

	switch m.state {
	case stateIdle, stateFailed:
		return m.handleIdleKey(msg)
	case stateStreaming:
		return m.handleStreamingKey(msg)
	case stateApproving:
		return m.handleApprovingKey(msg)
	case stateCancelling:
		return nil
	}
	return nil
}

func (m *ChatModel) handleIdleKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
		now := time.Now()
		if now.Sub(m.lastCtrlC) < 500*time.Millisecond {
			m.quitting = true
			return tea.Quit
		}
		m.lastCtrlC = now
		return func() tea.Msg {
			return SystemMsg{Text: "Press Ctrl+C again to quit, or Ctrl+D to quit immediately."}
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		input := strings.TrimSpace(m.input.Value())
		if input == "" {
			return nil
		}
		m.input.Reset()

		if handled, cmd := dispatchSlash(m, input); handled {
			return cmd
		}

		m.chatView.appendUser(input)
		// Set pending state before transition so recalcLayout accounts for the strip.
		m.pending.Activate()
		return tea.Batch(
			m.transitionTo(stateStreaming),
			m.submitCmd(input),
			m.pending.TickCmd(),
		)
	}

	return nil
}

func (m *ChatModel) handleStreamingKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
		if m.cancelFn != nil {
			m.cancelFn()
		}
		return m.transitionTo(stateCancelling)

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		input := strings.TrimSpace(m.input.Value())
		if input == "" {
			return nil
		}
		m.pendingRedirectInput = input
		m.input.Reset()
		if m.cancelFn != nil {
			m.cancelFn()
		}
		m.chatView.stopCursorBlink()
		if m.chatView.streamBuf.Len() > 0 {
			m.chatView.finalizeStream()
		}
		m.chatView.appendStatus("[interrupted]", "warning")
		return nil
	}
	return nil
}

func (m *ChatModel) handleApprovingKey(msg tea.KeyMsg) tea.Cmd {
	if !m.approval.HasPending() {
		return nil
	}

	// Expire confirm-pending state after 3 seconds.
	if m.approval.IsConfirmExpired() {
		m.approval.CancelConfirm()
	}

	req := m.approval.pending.Request
	respond := func(approved, alwaysAllow bool, outcome string, eventText string) tea.Cmd {
		resp := approval.ApprovalResponse{
			Approved:    approved,
			AlwaysAllow: alwaysAllow,
			Provider:    "tui",
		}
		ch := m.approval.pending.Response
		m.approval.Clear()
		m.chatView.appendApprovalEvent(eventText, outcome)
		return tea.Batch(
			m.transitionTo(stateStreaming),
			func() tea.Msg {
				ch <- resp
				return nil
			},
		)
	}

	// Tier 2 dialog key dispatch (scroll, diff toggle).
	if m.approval.pending.ViewModel.Tier == approval.TierFullscreen {
		if cmd := handleApprovalDialogKey(msg, &m.approval); cmd != nil {
			return cmd
		}
	}

	// If a confirm is already pending, the second press of the SAME key
	// completes the action. Any OTHER key resets the pending state.
	if m.approval.confirmPending {
		pressedKey := ""
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			pressedKey = "a"
		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			pressedKey = "s"
		}
		if pressedKey == m.approval.confirmAction {
			// Second press matches — execute the confirmed action.
			m.approval.CancelConfirm()
			if pressedKey == "s" {
				return respond(true, true, "session", fmt.Sprintf("Always allow enabled for %s", req.ToolName))
			}
			return respond(true, false, "approved", fmt.Sprintf("Approved %s", req.ToolName))
		}
		// Different key — reset pending state and fall through.
		m.approval.CancelConfirm()
	}

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
		// Double-press guardrail for critical-risk tools.
		if m.approval.pending.ViewModel.Risk.Level == "critical" {
			m.approval.StartConfirm("a")
			return nil
		}
		return respond(true, false, "approved", fmt.Sprintf("Approved %s", req.ToolName))
	case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
		// Double-press guardrail for critical-risk session grants.
		if m.approval.pending.ViewModel.Risk.Level == "critical" {
			m.approval.StartConfirm("s")
			return nil
		}
		return respond(true, true, "session", fmt.Sprintf("Always allow enabled for %s", req.ToolName))
	case key.Matches(msg, key.NewBinding(key.WithKeys("d", "esc"))):
		return respond(false, false, "denied", fmt.Sprintf("Denied %s", req.ToolName))
	}

	return nil
}

// submitCmd creates a tea.Cmd that runs a turn via the TurnRunner.
func (m *ChatModel) submitCmd(input string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.runCtx = ctx
	m.cancelFn = cancel

	program := m.program
	return func() tea.Msg {
		req := turnrunner.Request{
			SessionKey: m.sessionKey,
			Input:      input,
			Entrypoint: "tui",
			OnChunk: func(chunk string) {
				if program != nil {
					program.Send(ChunkMsg{Chunk: chunk})
				}
			},
			OnWarning: func(elapsed, hardCeiling time.Duration) {
				if program != nil {
					program.Send(WarningMsg{Elapsed: elapsed, HardCeiling: hardCeiling})
				}
			},
		}
		enrichRequest(program, &req)

		result, err := m.turnRunner.Run(ctx, req)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return DoneMsg{Result: result}
	}
}

func (m *ChatModel) recalcLayout() {
	m.input.SetWidth(m.width)

	fixedParts := []string{
		renderHeader(m.cfg, truncateSessionKey(m.sessionKey), m.width),
		renderTurnStrip(m.state, m.width),
	}
	if m.pending.IsActive() {
		fixedParts = append(fixedParts, renderPendingIndicator(m.pending.Elapsed()))
	}
	if ts := m.taskStrip.View(m.width); ts != "" {
		fixedParts = append(fixedParts, ts)
	}
	if m.state == stateApproving && m.approval.HasPending() {
		fixedParts = append(fixedParts, renderApproval(m.approval.pending, &m.approval, m.width, m.height))
	}
	fixedParts = append(fixedParts, renderFooter(m.input, m.state, m.width))

	fixedHeight := 0
	for _, part := range fixedParts {
		fixedHeight += lipgloss.Height(part)
	}

	separators := len(fixedParts)
	chatHeight := m.height - fixedHeight - separators
	if chatHeight < 3 {
		chatHeight = 3
	}
	m.chatView.setSize(m.width, chatHeight)
}

func generateSessionKey() string {
	return fmt.Sprintf("tui-%d", time.Now().UnixMilli())
}

func truncateSessionKey(key string) string {
	if len(key) > 20 {
		return key[:20] + "..."
	}
	return key
}

// cprFlush retrieves buffered keys from the CPR filter and replays them.
func (m *ChatModel) cprFlush() []tea.Cmd {
	return m.replayKeys(m.cpr.Flush())
}

// replayKeys feeds buffered key messages through handleKey / input.Update.
func (m *ChatModel) replayKeys(keys []tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd
	for _, k := range keys {
		cmd := m.handleKey(k)
		if cmd != nil {
			cmds = append(cmds, cmd)
			continue
		}
		if m.inputAcceptsText() {
			var inputCmd tea.Cmd
			m.input, inputCmd = m.input.Update(k)
			if inputCmd != nil {
				cmds = append(cmds, inputCmd)
			}
		}
	}
	return cmds
}
