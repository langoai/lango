package workbench

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
)

type stubPage struct {
	view          string
	activateCalls int
	msgs          []tea.Msg
	onUpdate      func(tea.Msg)
}

func (p *stubPage) Init() tea.Cmd { return nil }

func (p *stubPage) Activate() tea.Cmd {
	p.activateCalls++
	return nil
}

func (p *stubPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p.msgs = append(p.msgs, msg)
	if p.onUpdate != nil {
		p.onUpdate(msg)
	}
	return p, nil
}

func (p *stubPage) View() string { return p.view }

type fakeChild struct {
	program  *tea.Program
	setCalls int
}

func (f *fakeChild) SetProgram(p *tea.Program) {
	f.program = p
	f.setCalls++
}

func TestModelInitActivatesMissionControlImmediately(t *testing.T) {
	page := &stubPage{view: "mission-control"}
	model := newModel(nil, page, &fakeChild{}, nil, nil, nil)

	_ = model.Init()

	assert.Equal(t, 1, page.activateCalls)
	assert.Equal(t, "mission-control", model.View())
}

func TestModelSetProgramDelegatesToSharedChatModel(t *testing.T) {
	page := &stubPage{}
	child := &fakeChild{}
	model := newModel(nil, page, child, nil, nil, nil)

	program := &tea.Program{}
	model.SetProgram(program)

	assert.Equal(t, 1, child.setCalls)
	assert.Same(t, program, child.program)
}

func TestApprovalRequestRegistersPendingBeforeForwarding(t *testing.T) {
	registry := cockpit.NewPendingApprovalRegistry()
	page := &stubPage{
		onUpdate: func(msg tea.Msg) {
			_, ok := msg.(chat.ApprovalRequestMsg)
			if !ok {
				return
			}
			require.NotNil(t, registry.Latest())
			assert.Equal(t, "approve-1", registry.Latest().Request.ID)
		},
	}
	model := newModel(nil, page, &fakeChild{}, registry, nil, cockpit.NewMissionActivityBuffer())

	responseCh := make(chan approval.ApprovalResponse, 1)
	_, cmd := model.Update(chat.ApprovalRequestMsg{
		Request:  approval.ApprovalRequest{ID: "approve-1", CreatedAt: time.Now()},
		Response: responseCh,
	})

	require.Nil(t, cmd)
	require.Len(t, page.msgs, 1)
	_, ok := page.msgs[0].(chat.ApprovalRequestMsg)
	require.True(t, ok)
}

func TestDelegationBudgetAndRecoveryAppendActivityBeforeForwarding(t *testing.T) {
	activity := cockpit.NewMissionActivityBuffer()
	page := &stubPage{}
	model := newModel(nil, page, &fakeChild{}, nil, nil, activity)

	model.Update(chat.DelegationMsg{From: "planner", To: "reviewer", Reason: "handoff"})
	model.Update(chat.BudgetWarningMsg{Used: 9, Max: 12})
	model.Update(chat.RecoveryMsg{CauseClass: "timeout", Action: "retry_with_hint", Attempt: 2})

	snapshot := activity.Snapshot()
	require.Len(t, snapshot, 3)
	assert.Contains(t, snapshot[0].Summary, "Delegated from planner to reviewer")
	assert.Contains(t, snapshot[1].Summary, "9/12")
	assert.Contains(t, snapshot[2].Summary, "retry_with_hint")
	require.Len(t, page.msgs, 3)
	_, ok := page.msgs[0].(chat.DelegationMsg)
	require.True(t, ok)
}

func TestDoneFlushesTurnTokenSummaryAfterDone(t *testing.T) {
	page := &stubPage{}
	model := newModel(&config.Config{}, page, &fakeChild{}, nil, nil, nil)

	bus := eventbus.New()
	tracker := cockpit.NewRuntimeTracker(bus, nil, "workbench-1")
	model.SetRuntimeTracker(tracker)

	model.Update(chat.ThinkingStartedMsg{AgentName: "planner"})
	bus.Publish(eventbus.TokenUsageEvent{
		InputTokens:  3,
		OutputTokens: 5,
		TotalTokens:  8,
		CacheTokens:  1,
	})

	model.Update(chat.DoneMsg{})

	require.GreaterOrEqual(t, len(page.msgs), 3)
	_, thinkingOK := page.msgs[0].(chat.ThinkingStartedMsg)
	require.True(t, thinkingOK)
	_, doneOK := page.msgs[1].(chat.DoneMsg)
	require.True(t, doneOK)
	turnMsg, ok := page.msgs[2].(chat.TurnTokenUsageMsg)
	require.True(t, ok)
	assert.Equal(t, int64(3), turnMsg.InputTokens)
	assert.Equal(t, int64(5), turnMsg.OutputTokens)
	assert.Equal(t, int64(8), turnMsg.TotalTokens)
	assert.Equal(t, int64(1), turnMsg.CacheTokens)
}

func TestNewSubscribesMissionControlEventsAtProgramLifetime(t *testing.T) {
	bus := eventbus.New()
	learning := cockpit.NewLearningSuggestionBuffer(func() time.Time {
		return time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	})
	activity := cockpit.NewMissionActivityBuffer()

	model := New(cockpit.Deps{
		SessionKey:       "workbench-1",
		EventBus:         bus,
		PendingApprovals: cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:   learning,
		ActivityBuffer:   activity,
	})

	require.NotNil(t, model)
	require.NotNil(t, model.learningBuffer)

	bus.Publish(eventbus.LearningSuggestionEvent{
		SessionKey:   "workbench-1",
		SuggestionID: "learn-1",
		ProposedRule: "Prefer direct workbench routing",
		Confidence:   0.91,
		Timestamp:    time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC),
	})

	items := learning.Snapshot()
	require.Len(t, items, 1)
	assert.Equal(t, "learn-1", items[0].SuggestionID)
	activities := activity.Snapshot()
	require.Len(t, activities, 1)
	assert.Contains(t, activities[0].Summary, "Prefer direct workbench routing")
}

func TestWorkbenchRendersLoadingThenEmptyWithoutCockpitChrome(t *testing.T) {
	model := New(cockpit.Deps{
		Config: &config.Config{Agent: config.AgentConfig{Provider: "test", Model: "model"}},
		BackgroundManager: &background.Manager{},
		SessionKey:       "workbench-2",
		PendingApprovals: cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:   cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:   cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	assert.Contains(t, model.View(), "Loading Mission Control...")

	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	assert.Contains(t, model.View(), "No active missions or pending decisions.")
	assert.NotContains(t, model.View(), "Settings")
}

func TestWorkbenchApprovalRefreshFlowsIntoMissionControlView(t *testing.T) {
	model := New(cockpit.Deps{
		Config: &config.Config{Agent: config.AgentConfig{Provider: "test", Model: "model"}},
		BackgroundManager: &background.Manager{},
		SessionKey:       "workbench-3",
		PendingApprovals: cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:   cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:   cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)

	responseCh := make(chan approval.ApprovalResponse, 1)
	req := approval.ApprovalRequest{
		ID:        "approve-2",
		ToolName:  "fs_read",
		Summary:   "Read file",
		CreatedAt: time.Now(),
	}
	updated, _ = model.Update(chat.ApprovalRequestMsg{
		Request:   req,
		ViewModel: approval.NewViewModel(req),
		Response:  responseCh,
	})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Decisions")
	assert.Contains(t, view, "fs_read: Read file")
	assert.Contains(t, view, "Read-only operation")
}
