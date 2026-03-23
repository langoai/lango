// Package chat implements an interactive TUI chat interface using bubbletea.
// It provides a Claude Code-like terminal experience for conversing with the
// Lango agent, including streaming responses, inline tool approval, and
// slash commands.
package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/turnrunner"
)

// Deps holds the dependencies injected into the chat model.
type Deps struct {
	TurnRunner *turnrunner.Runner
	Config     *config.Config
	SessionKey string
}

// ChatModel is the root bubbletea model for the interactive TUI chat.
type ChatModel struct {
	// Dependencies
	turnRunner *turnrunner.Runner
	cfg        *config.Config
	sessionKey string

	// UI components
	input    inputModel
	chatView chatViewModel

	// State
	state    chatState
	width    int
	height   int
	quitting bool

	// Streaming context
	runCtx    context.Context
	cancelFn  context.CancelFunc

	// Approval state
	pendingApproval *ApprovalRequestMsg

	// Program reference for sending messages from callbacks
	program *tea.Program

	// Track Ctrl+C double-tap for quit
	lastCtrlC time.Time
}

// New creates a new ChatModel with the given dependencies.
func New(deps Deps) *ChatModel {
	return &ChatModel{
		turnRunner: deps.TurnRunner,
		cfg:        deps.Config,
		sessionKey: deps.SessionKey,
		input:      newInputModel(),
		chatView:   newChatViewModel(80, 20),
		state:      stateIdle,
	}
}

// SetProgram stores a reference to the tea.Program for sending messages
// from goroutines (e.g., streaming callbacks).
func (m *ChatModel) SetProgram(p *tea.Program) {
	m.program = p
}

// Init implements tea.Model.
func (m *ChatModel) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("Lango Chat"),
		m.input.Focus(),
	)
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

	case tea.KeyMsg:
		cmd := m.handleKey(msg)
		if cmd != nil {
			return m, cmd
		}

	case ChunkMsg:
		m.chatView.appendChunk(msg.Chunk)
		return m, nil

	case DoneMsg:
		m.state = stateIdle
		if msg.Result.Outcome == "success" || msg.Result.ResponseText != "" {
			m.chatView.finalizeStream(m.width)
		} else {
			// Non-success: show the error/user message.
			m.chatView.streamBuf.Reset()
			text := msg.Result.UserMessage
			if text == "" {
				text = msg.Result.ResponseText
			}
			if text != "" {
				m.chatView.finalizeWithText(lipgloss.NewStyle().Foreground(tui.Error).Render(text))
			}
		}
		cmds = append(cmds, m.input.Focus())
		return m, tea.Batch(cmds...)

	case ErrorMsg:
		m.state = stateIdle
		m.chatView.streamBuf.Reset()
		m.chatView.finalizeWithText(lipgloss.NewStyle().Foreground(tui.Error).Render(fmt.Sprintf("Error: %v", msg.Err)))
		cmds = append(cmds, m.input.Focus())
		return m, tea.Batch(cmds...)

	case WarningMsg:
		m.chatView.appendSystem(
			lipgloss.NewStyle().Foreground(tui.Warning).Render(
				fmt.Sprintf("Approaching timeout (%s / %s)", msg.Elapsed.Round(time.Second), msg.HardCeiling.Round(time.Second)),
			),
		)
		return m, nil

	case ApprovalRequestMsg:
		m.state = stateApproving
		m.pendingApproval = &msg
		m.input.Blur()
		return m, nil

	case SystemMsg:
		m.chatView.appendSystem(msg.Text)
		return m, nil
	}

	// Delegate to input component when idle.
	if m.state == stateIdle {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Delegate to viewport for scrolling.
	var cmd tea.Cmd
	m.chatView, cmd = m.chatView.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m *ChatModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Status bar (top)
	b.WriteString(renderStatusBar(m.cfg, truncateSessionKey(m.sessionKey), m.state, m.width))
	b.WriteString("\n")

	// Chat viewport (middle)
	b.WriteString(m.chatView.View())
	b.WriteString("\n")

	// Approval banner (if approving)
	if m.state == stateApproving && m.pendingApproval != nil {
		b.WriteString(renderApprovalBanner(m.pendingApproval.Request, m.width))
		b.WriteString("\n")
	}

	// Input area (bottom, hidden during approval)
	if m.state != stateApproving {
		b.WriteString(m.input.View(m.width))
	}

	// Help bar (bottom)
	b.WriteString("\n")
	b.WriteString(renderHelpBar(m.state, m.width))

	return b.String()
}

// handleKey processes key events based on current state.
func (m *ChatModel) handleKey(msg tea.KeyMsg) tea.Cmd {
	// Ctrl+D: immediate quit from any state.
	if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+d"))) {
		m.quitting = true
		if m.cancelFn != nil {
			m.cancelFn()
		}
		return tea.Quit
	}

	switch m.state {
	case stateIdle:
		return m.handleIdleKey(msg)
	case stateStreaming:
		return m.handleStreamingKey(msg)
	case stateApproving:
		return m.handleApprovingKey(msg)
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
			return SystemMsg{Text: lipgloss.NewStyle().Foreground(tui.Muted).Render("Press Ctrl+C again to quit, or Ctrl+D to quit immediately.")}
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		input := strings.TrimSpace(m.input.Value())
		if input == "" {
			return nil
		}
		m.input.Reset()

		// Check for slash commands.
		if handled, cmd := dispatchSlash(m, input); handled {
			return cmd
		}

		// Submit user message.
		m.chatView.appendUser(input)
		m.state = stateStreaming
		m.input.Blur()
		return m.submitCmd(input)
	}

	return nil
}

func (m *ChatModel) handleStreamingKey(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
		if m.cancelFn != nil {
			m.cancelFn()
		}
		m.state = stateIdle
		m.chatView.finalizeWithText(lipgloss.NewStyle().Foreground(tui.Muted).Render("(cancelled)"))
		return m.input.Focus()
	}
	return nil
}

func (m *ChatModel) handleApprovingKey(msg tea.KeyMsg) tea.Cmd {
	if m.pendingApproval == nil {
		return nil
	}

	respond := func(approved, alwaysAllow bool) tea.Cmd {
		resp := approval.ApprovalResponse{
			Approved:    approved,
			AlwaysAllow: alwaysAllow,
			Provider:    "tui",
		}
		ch := m.pendingApproval.Response
		m.pendingApproval = nil
		m.state = stateStreaming
		return func() tea.Msg {
			ch <- resp
			return nil
		}
	}

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
		return respond(true, false)
	case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
		return respond(true, true)
	case key.Matches(msg, key.NewBinding(key.WithKeys("d", "esc"))):
		return respond(false, false)
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
		result, err := m.turnRunner.Run(ctx, turnrunner.Request{
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
		})
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return DoneMsg{Result: result}
	}
}

// recalcLayout adjusts component sizes after a window resize.
func (m *ChatModel) recalcLayout() {
	// Layout: statusbar(1) + chatView(dynamic) + input(5) + helpbar(1) + margins(2)
	inputHeight := 5
	if m.state == stateApproving {
		inputHeight = 8 // extra room for approval banner
	}
	chatHeight := m.height - 1 - inputHeight - 1 - 2
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
