package workbench

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/turnrunner"
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

type stubWorkbenchMissionService struct{}

func (stubWorkbenchMissionService) StartMission(context.Context, mission.StartMissionInput) (*mission.Mission, error) {
	return &mission.Mission{}, nil
}

func (stubWorkbenchMissionService) AcceptProposal(context.Context, mission.AcceptProposalInput) (*mission.Mission, error) {
	return &mission.Mission{}, nil
}

func newWorkbenchModelWithDefaults(deps cockpit.Deps) *Model {
	if deps.MissionService == nil {
		deps.MissionService = stubWorkbenchMissionService{}
	}
	return New(deps)
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

func TestDoneAppendsAssistantSummaryToActivity(t *testing.T) {
	activity := cockpit.NewMissionActivityBuffer()
	page := &stubPage{}
	model := newModel(&config.Config{}, page, &fakeChild{}, nil, nil, activity)

	model.Update(chat.DoneMsg{Result: turnrunner.Result{
		Outcome:      "success",
		ResponseText: "Answer completed",
		Summary:      "Answer completed",
	}})

	snapshot := activity.Snapshot()
	require.Len(t, snapshot, 1)
	assert.Equal(t, cockpit.MissionActivityAssistant, snapshot[0].Kind)
	assert.Contains(t, snapshot[0].Summary, "Assistant reply: Answer completed")
}

func TestDoneAppendsFailureSummaryToActivity(t *testing.T) {
	activity := cockpit.NewMissionActivityBuffer()
	page := &stubPage{}
	model := newModel(&config.Config{}, page, &fakeChild{}, nil, nil, activity)

	model.Update(chat.DoneMsg{Result: turnrunner.Result{
		Outcome:     "timeout",
		UserMessage: "Request timed out",
		Summary:     "request exceeded limit",
	}})

	snapshot := activity.Snapshot()
	require.Len(t, snapshot, 1)
	assert.Equal(t, cockpit.MissionActivityAssistant, snapshot[0].Kind)
	assert.Contains(t, snapshot[0].Summary, "Turn timeout: request exceeded limit")
}

func TestDoneAppendsCompactAssistantSummaryToActivity(t *testing.T) {
	activity := cockpit.NewMissionActivityBuffer()
	page := &stubPage{}
	model := newModel(&config.Config{}, page, &fakeChild{}, nil, nil, activity)

	model.Update(chat.DoneMsg{Result: turnrunner.Result{
		Outcome: "success",
		Summary: "First line\n\n" + strings.Repeat("very long summary ", 20),
	}})

	snapshot := activity.Snapshot()
	require.Len(t, snapshot, 1)
	assert.Equal(t, cockpit.MissionActivityAssistant, snapshot[0].Kind)
	assert.NotContains(t, snapshot[0].Summary, "\n")
	assert.Len(t, []rune(snapshot[0].Summary), 160)
	assert.True(t, strings.HasSuffix(snapshot[0].Summary, "..."))
}

func TestNewSubscribesMissionControlEventsAtProgramLifetime(t *testing.T) {
	bus := eventbus.New()
	learning := cockpit.NewLearningSuggestionBuffer(func() time.Time {
		return time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	})
	activity := cockpit.NewMissionActivityBuffer()

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
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
	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            &config.Config{Agent: config.AgentConfig{Provider: "test", Model: "model"}},
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-2",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	assert.Contains(t, model.View(), "Loading Mission Control...")

	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	assert.Contains(t, model.View(), "No active missions or pending decisions.")
	assert.NotContains(t, model.View(), "Settings")
}

func TestWorkbenchEmptyStateGuidesInitialSetupWhenConfigIncomplete(t *testing.T) {
	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            config.DefaultConfig(),
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-setup",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "lango onboard")
	assert.Contains(t, view, "lango settings")
	assert.Contains(t, view, "lango doctor")
	assert.Contains(t, view, "Setup first:")
	assert.Contains(t, view, "Model: Setup required")
}

func TestWorkbenchEmptyStateSkipsSetupGuidanceWhenConfigReady(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-ready",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)

	view := model.View()
	assert.NotContains(t, view, "lango onboard")
	assert.NotContains(t, view, "lango settings")
	assert.NotContains(t, view, "lango doctor")
	assert.Contains(t, view, "Quick start:")
	assert.Contains(t, view, "press `Enter` for `Summarize this repository`")
	assert.Contains(t, view, "Summarize this repository")
	assert.Contains(t, view, "Explain the current project structure")
	assert.Contains(t, view, "Review recent changes")
	assert.Contains(t, view, "Press `Enter` for `Summarize this repository`")
	assert.Contains(t, view, "Enter default starter")
	assert.Contains(t, view, "Model: anthropic / claude-sonnet-4-5-20250929")
}

func TestWorkbenchStarterPromptHotkeySeedsComposer(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-starter-hotkey",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "> Explain the current project structure")
	assert.NotContains(t, view, "Press `Enter` for the default starter prompt")
	assert.Contains(t, view, "Enter submits starter")
	assert.Contains(t, view, "1-3 replace starter")
	assert.Contains(t, view, "Starter ready: press `Enter` to run it, use `1-3` to replace it")
	assert.NotContains(t, view, "Quick start:")
}

func TestWorkbenchStarterPromptHotkeyCanReplaceArmedStarter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-starter-replace",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "> Explain the current project structure")
	assert.NotContains(t, view, "> Summarize this repository")
	assert.Contains(t, view, "Enter submits starter")
}

func TestWorkbenchEnterSeedsDefaultStarterPromptWhenReadyAndEmpty(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-enter-starter",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "> Summarize this repository")
	assert.NotContains(t, view, "Press `Enter` for the default starter prompt")
	assert.Contains(t, view, "Enter submits starter")
	assert.Contains(t, view, "1-3 replace starter")
	assert.Contains(t, view, "Starter ready: press `Enter` to run it, use `1-3` to replace it")
	assert.NotContains(t, view, "Quick start:")
}

func TestWorkbenchSubmittedStarterShowsRunningStateInsteadOfQuickStart(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-running-state",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Current request running: type the next prompt now")
	assert.Contains(t, view, "Request running  Type next prompt  Enter interrupts and runs it")
	assert.NotContains(t, view, "Quick start:")
	assert.NotContains(t, view, "Starter ready:")
}

func TestWorkbenchRunningStateShowsFollowUpReadyHintWhenDraftStaged(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-followup-ready",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("next step")})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Follow-up ready: press `Enter` to interrupt and run it, use `1-3` to replace it")
	assert.Contains(t, view, "Enter runs staged follow-up  1-3 replace follow-up")
	assert.Contains(t, view, "> next step")
	assert.NotContains(t, view, "Current request running...")
}

func TestWorkbenchRunningStateEnterQueuesAndReplaysFollowUp(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-followup-redirect",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("next step")})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Current request running")
	assert.NotContains(t, view, "Follow-up ready:")
	assert.NotContains(t, view, "> next step")

	updated, _ = model.Update(chat.DoneMsg{})
	model = updated.(*Model)
	view = model.View()
	assert.Contains(t, view, "User submitted: next step")
}

func TestWorkbenchRunningStateEnterQueuesDefaultFollowUpWhenComposerEmpty(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-followup-default",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Current request running")
	assert.NotContains(t, view, "> Summarize this repository")

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	view = model.View()
	assert.Contains(t, view, "Current request running")
	assert.NotContains(t, view, "Follow-up ready:")
	assert.NotContains(t, view, "> Summarize this repository")

	updated, _ = model.Update(chat.DoneMsg{})
	model = updated.(*Model)
	view = model.View()
	assert.Contains(t, view, "User submitted: Summarize this repository")
}

func TestWorkbenchCompletedTurnChangesEmptyStateDefaultStarter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-post-turn-default-copy",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(chat.DoneMsg{Result: turnrunner.Result{
		Outcome:      "success",
		ResponseText: "Repository summary complete.",
	}})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Last turn finished. Pick the next step.")
	assert.Contains(t, view, "Last result: Repository summary complete.")
	assert.NotContains(t, view, "Last result: Assistant reply:")
	assert.Contains(t, view, "Type the next prompt here, or use `lango chat` for focused chat.")
	assert.NotContains(t, view, "Type to chat here, or use `lango chat` for focused chat.")
	assert.Contains(t, view, "Next step:")
	assert.Contains(t, view, "> Next step: press `Enter` for `Explain the current project structure`")
	assert.Contains(t, view, "press `Enter` for `Explain the current project structure`")
	assert.NotContains(t, view, "press `Enter` for `Summarize this repository`")
	assert.Contains(t, view, "Enter next-step starter")
	assert.Contains(t, view, "Type next prompt here")
	assert.NotContains(t, view, "Enter next-step starter  1-3 starter prompts  Type to chat here")
	assert.NotContains(t, view, "Enter default starter")
	assert.NotContains(t, view, "No active missions or pending decisions.")
	assert.Less(t, strings.Index(view, "Next step:"), strings.Index(view, "Type the next prompt here"))
}

func TestWorkbenchEnterSeedsNextStepStarterAfterCompletedTurn(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-post-turn-default-enter",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(chat.DoneMsg{Result: turnrunner.Result{
		Outcome:      "success",
		ResponseText: "Repository summary complete.",
	}})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "> Explain the current project structure")
	assert.NotContains(t, view, "> Summarize this repository")
	assert.Contains(t, view, "Enter submits starter")
	assert.NotContains(t, view, "Quick start:")
}

func TestWorkbenchFailedTurnChangesCompletedTurnLead(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-post-turn-failure-copy",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(chat.DoneMsg{Result: turnrunner.Result{
		Outcome:     "timeout",
		UserMessage: "Request timed out",
		Summary:     "request exceeded limit",
	}})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Last turn needs attention. Pick the recovery step.")
	assert.Contains(t, view, "Last result: Turn timeout: request exceeded limit")
	assert.Contains(t, view, "Recovery step:")
	assert.Contains(t, view, "press `Enter` for `Review recent changes`")
	assert.NotContains(t, view, "press `Enter` for `Explain the current project structure`")
	assert.Contains(t, view, "Type the recovery prompt here, or use `lango chat` for focused chat.")
	assert.Contains(t, view, "Enter recovery starter")
	assert.Contains(t, view, "Type recovery prompt here")
	assert.NotContains(t, view, "Enter recovery starter  1-3 starter prompts  Type next prompt here")
	assert.NotContains(t, view, "Last turn finished. Pick the next step.")
	assert.NotContains(t, view, "Last turn needs attention. Pick the next step.")
	assert.NotContains(t, view, "Next step:")
}

func TestWorkbenchEnterSeedsRecoveryStarterAfterFailedTurn(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-post-turn-recovery-enter",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(chat.DoneMsg{Result: turnrunner.Result{
		Outcome:     "timeout",
		UserMessage: "Request timed out",
		Summary:     "request exceeded limit",
	}})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "> Review recent changes")
	assert.NotContains(t, view, "> Explain the current project structure")
	assert.Contains(t, view, "Enter submits starter")
	assert.NotContains(t, view, "Recovery step:")
}

func TestWorkbenchRunningStateFollowUpSubmitsFromDecisionsFocus(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-followup-decisions-submit",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("next step")})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(chat.DoneMsg{})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "User submitted: next step")
	assert.Contains(t, view, "Focus: Composer")
}

func TestWorkbenchRunningStateStarterHotkeyCanReplaceStagedFollowUp(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-followup-replace",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("next step")})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "> Explain the current project structure")
	assert.NotContains(t, view, "> next step")
	assert.Contains(t, view, "1-3 replace follow-up")
}

func TestWorkbenchRunningStateReplacedFollowUpRunsOnEnter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-followup-replace-run",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("next step")})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(chat.DoneMsg{})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "User submitted: Explain the current project structure")
	assert.NotContains(t, view, "User submitted: next step")
}

func TestWorkbenchFollowUpEditingKeyReturnsFocusToComposer(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-followup-editing",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("next")})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Focus: Composer")
	assert.Contains(t, view, "> nex")
}

func TestWorkbenchSeededStarterHintChangesWhenFocusLeavesComposer(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-seeded-focus",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Starter ready: press `Enter` to run it")
	assert.Contains(t, view, "Enter submits starter")
	assert.Contains(t, view, "1-3 replace starter")
}

func TestWorkbenchSeededStarterEditingKeyReturnsFocusToComposer(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-seeded-editing",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Focus: Composer")
	assert.Contains(t, view, "> Summarize this repositor")
}

func TestWorkbenchSeededStarterSubmitsFromDecisionsFocus(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-seeded-decisions-submit",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "User submitted: Summarize this repository")
	assert.Contains(t, view, "Focus: Composer")
}

func TestWorkbenchEnterSeedsChangedReviewPromptWhenRepoDirty(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	repoRoot := t.TempDir()
	runGit(t, repoRoot, "init")
	runGit(t, repoRoot, "config", "user.email", "test@example.com")
	runGit(t, repoRoot, "config", "user.name", "Lango Test")
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module example.com/demo\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "README.md"), []byte("hello\n"), 0o644))
	runGit(t, repoRoot, "add", "go.mod", "README.md")
	runGit(t, repoRoot, "commit", "-m", "init")
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "README.md"), []byte("changed\n"), 0o644))
	workDir := filepath.Join(repoRoot, "internal", "feature")
	require.NoError(t, os.MkdirAll(workDir, 0o755))

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		WorkDir:           repoRoot,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-enter-dirty",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "> Review the uncommitted changes on branch")
	assert.Contains(t, view, "`README.md`")
	assert.Contains(t, view, "Enter submits starter")
	assert.Contains(t, view, "1-3 replace starter")
	assert.Contains(t, view, "Starter ready: press `Enter` to run it, use `1-3` to replace it")
	assert.NotContains(t, view, "Quick start:")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(out))
}

func TestWorkbenchEnterDoesNotSeedStarterPromptWhenSetupIncomplete(t *testing.T) {
	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            config.DefaultConfig(),
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-enter-setup",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Setup first:")
	assert.NotContains(t, view, "> Summarize this repository")
}

func TestWorkbenchReadyEmptyStateUsesContextAwareStarterPrompts(t *testing.T) {
	repoRoot := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module example.com/demo\n"), 0o644))
	workDir := filepath.Join(repoRoot, "internal", "feature")
	require.NoError(t, os.MkdirAll(workDir, 0o755))

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-live"},
	}

	model := newWorkbenchModelWithDefaults(cockpit.Deps{
		Config:            cfg,
		WorkDir:           workDir,
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-contextual",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(*Model)
	updated, _ = model.Update(cockpit.MissionControlRefreshMsg{})
	model = updated.(*Model)

	view := model.View()
	assert.Contains(t, view, "Summarize the "+filepath.Base(repoRoot)+" repository and its current purpose")
	assert.Contains(t, view, "Explain the Go package layout in "+filepath.Base(repoRoot)+" and where to start editing")
	assert.Contains(t, view, "Review the likely active workstream in "+filepath.Base(repoRoot)+" and suggest the best next change")
	assert.Contains(t, view, "press `Enter` for `Summarize the "+filepath.Base(repoRoot)+" repository and its current purpose`")
}

func TestWorkbenchApprovalRefreshFlowsIntoMissionControlView(t *testing.T) {
	model := New(cockpit.Deps{
		Config:            &config.Config{Agent: config.AgentConfig{Provider: "test", Model: "model"}},
		BackgroundManager: &background.Manager{},
		SessionKey:        "workbench-3",
		PendingApprovals:  cockpit.NewPendingApprovalRegistry(),
		LearningBuffer:    cockpit.NewLearningSuggestionBuffer(nil),
		ActivityBuffer:    cockpit.NewMissionActivityBuffer(),
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
