package cockpit

import (
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit/sidebar"
	"github.com/langoai/lango/internal/cli/cockpit/theme"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/provider"
)

// childModel is the minimal contract for the wrapped child.
// ChatModel satisfies this. Tests can use a mock.
type childModel interface {
	tea.Model
	SetProgram(p *tea.Program)
}

// Compile-time interface check.
var _ childModel = (*chat.ChatModel)(nil)

// Model is the root cockpit tea.Model.
type Model struct {
	child           childModel
	cfg             *config.Config
	pages           map[PageID]Page
	activePage      PageID
	sidebar         sidebar.Model
	contextPanel    *ContextPanel
	channelTracker  *ChannelTracker
	runtimeTracker  *RuntimeTracker
	keymap          keyMap
	sidebarVisible  bool
	sidebarFocused  bool
	contextVisible  bool
	width           int
	height          int
}

// New creates a cockpit Model wrapping a ChatModel.
func New(deps Deps) *Model {
	chatModel := chat.New(chat.Deps{
		TurnRunner:        deps.TurnRunner,
		Config:            deps.Config,
		SessionKey:        deps.SessionKey,
		SessionStore:      deps.SessionStore,
		EventBus:          deps.EventBus,
		BackgroundManager: deps.BackgroundManager,
	})

	return &Model{
		child:          chatModel,
		cfg:            deps.Config,
		pages:          make(map[PageID]Page),
		activePage:     PageChat,
		sidebar:        sidebar.New(AllPageMetas()),
		contextPanel:   NewContextPanel(deps.MetricsCollector),
		keymap:         defaultKeyMap(),
		sidebarVisible: true,
	}
}

// RegisterPage adds a page to the cockpit.
func (m *Model) RegisterPage(id PageID, page Page) {
	m.pages[id] = page
}

// SetProgram delegates to the wrapped child.
func (m *Model) SetProgram(p *tea.Program) {
	m.child.SetProgram(p)
}

// SetChannelTracker sets the channel tracker for live channel status updates.
// The tracker's snapshots are pushed to the context panel on each tick.
func (m *Model) SetChannelTracker(tracker *ChannelTracker) {
	m.channelTracker = tracker
}

// SetRuntimeTracker sets the runtime tracker for live metrics.
// The tracker aggregates token usage, delegation counts, and recovery events.
func (m *Model) SetRuntimeTracker(tracker *RuntimeTracker) {
	m.runtimeTracker = tracker
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return m.child.Init()
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case sidebar.PageSelectedMsg:
		target := PageIDFromString(msg.ID)
		cmd := m.switchPage(target)
		return m, cmd
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	case contextTickMsg:
		return m.handleContextTick(msg)
	case chat.ChannelMessageMsg:
		return m.handleChannelMessage(msg)
	case chat.ApprovalRequestMsg:
		return m.handleApprovalRequest(msg)
	case chat.DelegationMsg:
		return m.handleDelegation(msg)
	case chat.BudgetWarningMsg:
		return m.handleBudgetWarning(msg)
	case chat.RecoveryMsg:
		return m.handleRecovery(msg)
	case chat.DoneMsg:
		return m.handleDone(msg)
	}

	// Mark turn started on first content event.
	switch msg.(type) {
	case chat.ToolStartedMsg, chat.ThinkingStartedMsg, chat.ChunkMsg:
		m.markTurnStarted()
	}

	return m.forwardToActive(msg)
}

// handleContextTick refreshes channel/runtime tracker snapshots on the context panel.
func (m *Model) handleContextTick(msg contextTickMsg) (*Model, tea.Cmd) {
	if m.channelTracker != nil {
		m.contextPanel.SetChannelStatuses(m.channelTracker.Snapshot())
	}
	if m.runtimeTracker != nil {
		m.contextPanel.SetRuntimeStatus(m.runtimeTracker.Snapshot())
	}
	up, cmd := m.contextPanel.Update(msg)
	m.contextPanel = up.(*ContextPanel)
	return m, cmd
}

// handleChannelMessage always forwards to the chat child regardless of active page.
// Channel messages must always reach the chat model, even when another page
// (Settings, Status, etc.) is active. Otherwise traffic arriving while the
// user browses non-chat pages is lost.
func (m *Model) handleChannelMessage(msg chat.ChannelMessageMsg) (*Model, tea.Cmd) {
	up, cmd := m.child.Update(msg)
	m.child = up.(childModel)
	return m, cmd
}

// handleApprovalRequest switches to the chat page and forwards to the chat child.
// Approval requests must always reach the chat model AND switch to the chat
// page so the user can see and respond to the prompt. Without the page switch,
// approvals raised by background tasks retried from the Tasks page would
// remain invisible and time out.
func (m *Model) handleApprovalRequest(msg chat.ApprovalRequestMsg) (*Model, tea.Cmd) {
	switchCmd := m.switchPage(PageChat)
	up, childCmd := m.child.Update(msg)
	m.child = up.(childModel)
	return m, tea.Batch(switchCmd, childCmd)
}

// markTurnStarted calls runtimeTracker.StartTurn() on the first content event
// (ToolStartedMsg, ThinkingStartedMsg, ChunkMsg) so that the RuntimeTracker
// accumulates tokens and the context panel shows the Runtime section even for
// single-agent (no-delegation) turns.
func (m *Model) markTurnStarted() {
	if m.runtimeTracker != nil {
		m.runtimeTracker.StartTurn()
	}
}

// handleDelegation updates the runtime tracker and forwards to the chat child.
// Outward hops increment the delegation counter; orchestrator return hops only
// update the active-agent label (no counter bump), so the context panel shows
// the correct agent during the final phase.
func (m *Model) handleDelegation(msg chat.DelegationMsg) (*Model, tea.Cmd) {
	if m.runtimeTracker != nil {
		if msg.To == "lango-orchestrator" {
			m.runtimeTracker.SetActiveAgent(msg.To)
		} else {
			m.runtimeTracker.RecordDelegation(msg.To)
		}
		if m.contextPanel != nil {
			m.contextPanel.SetRuntimeStatus(m.runtimeTracker.Snapshot())
		}
	}
	up, cmd := m.child.Update(msg)
	m.child = up.(childModel)
	return m, cmd
}

// handleBudgetWarning always forwards to the chat child from any page.
func (m *Model) handleBudgetWarning(msg chat.BudgetWarningMsg) (*Model, tea.Cmd) {
	up, cmd := m.child.Update(msg)
	m.child = up.(childModel)
	return m, cmd
}

// handleRecovery always forwards to the chat child from any page.
func (m *Model) handleRecovery(msg chat.RecoveryMsg) (*Model, tea.Cmd) {
	up, cmd := m.child.Update(msg)
	m.child = up.(childModel)
	return m, cmd
}

// handleDone forwards DoneMsg to the chat child FIRST so the assistant response
// is appended, then flushes tokens so the summary appears AFTER the response,
// and finally resets the turn.
func (m *Model) handleDone(msg chat.DoneMsg) (*Model, tea.Cmd) {
	var cmds []tea.Cmd
	// 1. Forward DoneMsg to chat child first (appends assistant response).
	up, cmd := m.child.Update(msg)
	m.child = up.(childModel)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	// 2. Flush tokens and append summary AFTER the response.
	if m.runtimeTracker != nil {
		snap := m.runtimeTracker.FlushTurnTokens()
		if snap.TotalTokens > 0 {
			var costUSD float64
			if m.cfg != nil {
				costUSD = provider.EstimateCostUSD(m.cfg.Agent.Model, int(snap.InputTokens), int(snap.OutputTokens))
			}
			up2, cmd2 := m.child.Update(chat.TurnTokenUsageMsg{
				InputTokens:      snap.InputTokens,
				OutputTokens:     snap.OutputTokens,
				TotalTokens:      snap.TotalTokens,
				CacheTokens:      snap.CacheTokens,
				EstimatedCostUSD: costUSD,
			})
			m.child = up2.(childModel)
			if cmd2 != nil {
				cmds = append(cmds, cmd2)
			}
		}
		m.runtimeTracker.ResetTurn()
		if m.contextPanel != nil {
			m.contextPanel.SetRuntimeStatus(runtimeStatus{IsRunning: false})
		}
	}
	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m *Model) View() string {
	var mainView string
	if m.activePage == PageChat {
		mainView = m.child.View()
	} else if page, ok := m.pages[m.activePage]; ok {
		mainView = page.View()
	}

	panels := make([]string, 0, 3)
	if m.sidebarVisible {
		panels = append(panels, m.sidebar.View())
	}
	panels = append(panels, mainView)
	if m.contextVisible {
		panels = append(panels, m.contextPanel.View())
	}

	if len(panels) == 1 {
		return panels[0]
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (*Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	sw := m.sidebarWidth()
	cpw := m.contextPanelWidth()
	m.sidebar.SetHeight(msg.Height)
	m.contextPanel.SetHeight(msg.Height)

	childSize := tea.WindowSizeMsg{
		Width:  msg.Width - sw - cpw,
		Height: msg.Height,
	}

	// Forward to chat child.
	updated, cmd := m.child.Update(childSize)
	m.child = updated.(childModel)
	cmds := []tea.Cmd{cmd}

	// Forward to all registered pages.
	for id, page := range m.pages {
		up, c := page.Update(childSize)
		m.pages[id] = up.(Page)
		if c != nil {
			cmds = append(cmds, c)
		}
	}

	// Forward to context panel.
	up, c := m.contextPanel.Update(tea.WindowSizeMsg{
		Width:  cpw,
		Height: msg.Height,
	})
	m.contextPanel = up.(*ContextPanel)
	if c != nil {
		cmds = append(cmds, c)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleMouse(msg tea.MouseMsg) (*Model, tea.Cmd) {
	// Route to sidebar if click lands in sidebar region.
	if m.sidebarVisible && msg.X < m.sidebarWidth() {
		up, cmd := m.sidebar.Update(msg)
		m.sidebar = up.(sidebar.Model)
		return m, cmd
	}

	// Forward to active content area.
	if m.activePage == PageChat {
		updated, cmd := m.child.Update(msg)
		m.child = updated.(childModel)
		return m, cmd
	}
	if page, ok := m.pages[m.activePage]; ok {
		up, cmd := page.Update(msg)
		m.pages[m.activePage] = up.(Page)
		return m, cmd
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	// Global keys — always consumed regardless of focus.
	switch {
	case key.Matches(msg, m.keymap.ToggleSidebar):
		return m.toggleSidebar()
	case key.Matches(msg, m.keymap.ToggleContext):
		return m.toggleContext()
	case key.Matches(msg, m.keymap.CopyClipboard):
		return m.copyToClipboard()
	case key.Matches(msg, m.keymap.FocusToggle):
		m.sidebarFocused = !m.sidebarFocused
		m.sidebar.SetFocused(m.sidebarFocused)
		return m, nil
	case key.Matches(msg, m.keymap.Page1):
		return m, m.switchPage(PageChat)
	case key.Matches(msg, m.keymap.Page2):
		return m, m.switchPage(PageSettings)
	case key.Matches(msg, m.keymap.Page3):
		return m, m.switchPage(PageTools)
	case key.Matches(msg, m.keymap.Page4):
		return m, m.switchPage(PageStatus)
	case key.Matches(msg, m.keymap.Page5):
		return m, m.switchPage(PageTasks)
	case key.Matches(msg, m.keymap.Page6):
		return m, m.switchPage(PageApprovals)
	}

	// Focus-dependent routing.
	if m.sidebarFocused {
		up, cmd := m.sidebar.Update(msg)
		m.sidebar = up.(sidebar.Model)
		return m, cmd
	}

	return m.forwardToActive(msg)
}

func (m *Model) toggleSidebar() (*Model, tea.Cmd) {
	m.sidebarVisible = !m.sidebarVisible
	return m, m.propagateResize()
}

func (m *Model) toggleContext() (*Model, tea.Cmd) {
	m.contextVisible = !m.contextVisible
	m.contextPanel.SetVisible(m.contextVisible)

	var startCmd tea.Cmd
	if m.contextVisible {
		startCmd = m.contextPanel.Start()
		// Send correct width to context panel — it may still hold width=0
		// from when it was hidden during the initial WindowSizeMsg.
		cpw := m.contextPanelWidth()
		up, c := m.contextPanel.Update(tea.WindowSizeMsg{
			Width:  cpw,
			Height: m.height,
		})
		m.contextPanel = up.(*ContextPanel)
		if c != nil {
			startCmd = tea.Batch(startCmd, c)
		}
	} else {
		m.contextPanel.Stop()
	}

	return m, tea.Batch(startCmd, m.propagateResize())
}

// propagateResize sends a synthetic WindowSizeMsg to child and all pages
// with the current effective content width.
func (m *Model) propagateResize() tea.Cmd {
	newSize := tea.WindowSizeMsg{
		Width:  m.width - m.sidebarWidth() - m.contextPanelWidth(),
		Height: m.height,
	}
	updated, cmd := m.child.Update(newSize)
	m.child = updated.(childModel)
	cmds := []tea.Cmd{cmd}
	for id, page := range m.pages {
		up, c := page.Update(newSize)
		m.pages[id] = up.(Page)
		if c != nil {
			cmds = append(cmds, c)
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) copyToClipboard() (*Model, tea.Cmd) {
	var content string
	if m.activePage == PageChat {
		content = m.child.View()
	} else if page, ok := m.pages[m.activePage]; ok {
		content = page.View()
	}
	_ = clipboard.WriteAll(content)
	return m, nil
}

func (m *Model) forwardToActive(msg tea.Msg) (*Model, tea.Cmd) {
	if m.activePage == PageChat {
		updated, cmd := m.child.Update(msg)
		m.child = updated.(childModel)
		return m, cmd
	}
	if page, ok := m.pages[m.activePage]; ok {
		up, cmd := page.Update(msg)
		m.pages[m.activePage] = up.(Page)
		return m, cmd
	}
	return m, nil
}

func (m *Model) sidebarWidth() int {
	if !m.sidebarVisible {
		return 0
	}
	return theme.SidebarFullWidth
}

func (m *Model) contextPanelWidth() int {
	if !m.contextVisible {
		return 0
	}
	return theme.ContextPanelWidth
}

func (m *Model) switchPage(target PageID) tea.Cmd {
	if target == m.activePage {
		return nil
	}
	// Deactivate old page.
	if m.activePage != PageChat {
		if old, ok := m.pages[m.activePage]; ok {
			old.Deactivate()
		}
	}
	m.activePage = target
	m.sidebar.SetActive(target.String())
	m.sidebarFocused = false
	m.sidebar.SetFocused(false)
	// Activate new page.
	if target != PageChat {
		if page, ok := m.pages[target]; ok {
			return page.Activate()
		}
	}
	return nil
}
