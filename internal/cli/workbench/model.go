package workbench

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/cli/cockpit/pages"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/provider"
)

type childModel interface {
	SetProgram(*tea.Program)
}

type missionPage interface {
	tea.Model
	Activate() tea.Cmd
}

var _ childModel = (*chat.ChatModel)(nil)
var _ missionPage = (*pages.MissionControlPage)(nil)

// Model is the standalone mission workbench root shell.
// It mounts Mission Control directly while reusing the shared chat model and
// shared pending/activity registries.
type Model struct {
	cfg              *config.Config
	page             missionPage
	child            childModel
	runtimeTracker   *cockpit.RuntimeTracker
	pendingApprovals *cockpit.PendingApprovalRegistry
	learningBuffer   *cockpit.LearningSuggestionBuffer
	activityBuffer   *cockpit.MissionActivityBuffer
}

// New creates a standalone workbench shell backed by the existing Mission
// Control page and shared chat model.
func New(deps cockpit.Deps) *Model {
	chatModel := chat.New(chat.Deps{
		TurnRunner:        deps.TurnRunner,
		Config:            deps.Config,
		SessionKey:        deps.SessionKey,
		SessionStore:      deps.SessionStore,
		EventBus:          deps.EventBus,
		BackgroundManager: deps.BackgroundManager,
		RuntimeFeatures:   deps.RuntimeFeatures,
		SharedPending:     deps.PendingApprovals,
		OnUserSubmission: func(sessionKey, input string) {
			if deps.ActivityBuffer != nil {
				deps.ActivityBuffer.Append(cockpit.MissionActivityItem{
					Kind:       cockpit.MissionActivityUser,
					SessionKey: sessionKey,
					Summary:    "User submitted: " + input,
					Timestamp:  time.Now(),
				})
			}
		},
		OnTurnSummary: func(sessionKey string, msg chat.TurnTokenUsageMsg) {
			if deps.ActivityBuffer != nil {
				deps.ActivityBuffer.Append(cockpit.MissionActivityItem{
					Kind:       cockpit.MissionActivityTurn,
					SessionKey: sessionKey,
					Summary:    turnSummaryText(msg),
					Timestamp:  time.Now(),
				})
			}
		},
	})

	cockpit.SubscribeMissionControlEvents(
		deps.EventBus,
		deps.SessionKey,
		deps.LearningBuffer,
		deps.ActivityBuffer,
	)

	return newModel(
		deps.Config,
		pages.NewWorkbenchMissionControlPage(deps, chatModel),
		chatModel,
		deps.PendingApprovals,
		deps.LearningBuffer,
		deps.ActivityBuffer,
	)
}

func newModel(
	cfg *config.Config,
	page missionPage,
	child childModel,
	pending *cockpit.PendingApprovalRegistry,
	learning *cockpit.LearningSuggestionBuffer,
	activity *cockpit.MissionActivityBuffer,
) *Model {
	return &Model{
		cfg:              cfg,
		page:             page,
		child:            child,
		pendingApprovals: pending,
		learningBuffer:   learning,
		activityBuffer:   activity,
	}
}

// SetProgram delegates to the shared chat model so approval callbacks can send
// messages back into the Bubble Tea program.
func (m *Model) SetProgram(p *tea.Program) {
	if m == nil || m.child == nil {
		return
	}
	m.child.SetProgram(p)
}

// SetRuntimeTracker installs the shared runtime tracker used to preserve turn
// token flush and recovery signaling semantics that cockpit currently owns.
func (m *Model) SetRuntimeTracker(tracker *cockpit.RuntimeTracker) {
	if m == nil {
		return
	}
	m.runtimeTracker = tracker
}

// Init activates Mission Control immediately.
func (m *Model) Init() tea.Cmd {
	if m == nil || m.page == nil {
		return nil
	}
	return m.page.Activate()
}

// Update routes messages through the standalone Mission Control workbench while
// preserving the shared approval/activity/runtime semantics needed by the
// existing Mission Control projection.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case chat.ToolStartedMsg, chat.ThinkingStartedMsg, chat.ChunkMsg:
		m.markTurnStarted()
	}

	switch typed := msg.(type) {
	case chat.ChannelMessageMsg:
		if m.activityBuffer != nil {
			m.activityBuffer.Append(cockpit.MissionActivityItem{
				Kind:       cockpit.MissionActivityRuntime,
				SessionKey: typed.SessionKey,
				Summary:    typed.Channel + " message from " + typed.SenderName + ": " + typed.Text,
				Timestamp:  typed.Timestamp,
			})
		}
		return m.forward(msg)
	case chat.ApprovalRequestMsg:
		if m.pendingApprovals != nil {
			m.pendingApprovals.Register(typed)
		}
		return m.forward(msg)
	case chat.DelegationMsg:
		if m.activityBuffer != nil {
			m.activityBuffer.Append(cockpit.MissionActivityItem{
				Kind:      cockpit.MissionActivityRuntime,
				Summary:   "Delegated from " + typed.From + " to " + typed.To + " (" + typed.Reason + ")",
				Timestamp: time.Now(),
			})
		}
		if m.runtimeTracker != nil {
			if typed.To == "lango-orchestrator" {
				m.runtimeTracker.SetActiveAgent(typed.To)
			} else {
				m.runtimeTracker.RecordDelegation(typed.To)
			}
		}
		return m.forward(msg)
	case chat.BudgetWarningMsg:
		if m.activityBuffer != nil {
			m.activityBuffer.Append(cockpit.MissionActivityItem{
				Kind:      cockpit.MissionActivityRuntime,
				Summary:   budgetSummaryText(typed),
				Timestamp: time.Now(),
			})
		}
		return m.forward(msg)
	case chat.RecoveryMsg:
		if m.activityBuffer != nil {
			m.activityBuffer.Append(cockpit.MissionActivityItem{
				Kind:      cockpit.MissionActivityRuntime,
				Summary:   recoverySummaryText(typed),
				Timestamp: time.Now(),
			})
		}
		return m.forward(msg)
	case chat.DoneMsg:
		return m.handleDone(typed)
	default:
		return m.forward(msg)
	}
}

// View renders the mounted Mission Control page directly, without cockpit
// chrome.
func (m *Model) View() string {
	if m == nil || m.page == nil {
		return ""
	}
	return m.page.View()
}

func (m *Model) markTurnStarted() {
	if m != nil && m.runtimeTracker != nil {
		m.runtimeTracker.StartTurn()
	}
}

func (m *Model) handleDone(msg chat.DoneMsg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, 2)
	if m.activityBuffer != nil {
		if item, ok := cockpit.NewAssistantSummaryActivity("", msg, time.Now()); ok {
			m.activityBuffer.Append(item)
		}
	}
	if _, cmd := m.forward(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if m.runtimeTracker != nil {
		snap := m.runtimeTracker.FlushTurnTokens()
		if snap.TotalTokens > 0 {
			var costUSD float64
			if m.cfg != nil {
				costUSD = provider.EstimateCostUSD(
					m.cfg.Agent.Model,
					int(snap.InputTokens),
					int(snap.OutputTokens),
				)
			}
			if _, cmd := m.forward(chat.TurnTokenUsageMsg{
				InputTokens:      snap.InputTokens,
				OutputTokens:     snap.OutputTokens,
				TotalTokens:      snap.TotalTokens,
				CacheTokens:      snap.CacheTokens,
				EstimatedCostUSD: costUSD,
			}); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		m.runtimeTracker.ResetTurn()
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) forward(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m == nil || m.page == nil {
		return m, nil
	}
	updated, cmd := m.page.Update(msg)
	if updatedPage, ok := updated.(missionPage); ok {
		m.page = updatedPage
	}
	return m, cmd
}

func budgetSummaryText(msg chat.BudgetWarningMsg) string {
	return "Delegation budget warning " + itoa(msg.Used) + "/" + itoa(msg.Max)
}

func recoverySummaryText(msg chat.RecoveryMsg) string {
	return "Recovery " + msg.Action + " after " + msg.CauseClass + " (attempt " + itoa(msg.Attempt) + ")"
}

func turnSummaryText(msg chat.TurnTokenUsageMsg) string {
	return "Turn summary: " + itoa64(msg.TotalTokens) + " total tokens (" + itoa64(msg.InputTokens) + " in / " + itoa64(msg.OutputTokens) + " out)"
}

func itoa(v int) string {
	return itoa64(int64(v))
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}
