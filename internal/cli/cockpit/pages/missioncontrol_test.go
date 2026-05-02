package pages

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turnrunner"
)

type stubMissionControlProjector struct {
	snapshot cockpit.MissionControlSnapshot
	calls    int
	tasks    []background.TaskSnapshot
}

func (s *stubMissionControlProjector) Project(tasks []background.TaskSnapshot) cockpit.MissionControlSnapshot {
	s.calls++
	s.tasks = append([]background.TaskSnapshot(nil), tasks...)
	return s.snapshot
}

type stubMissionTaskSource struct {
	tasks []background.TaskSnapshot
}

func (s stubMissionTaskSource) List() []background.TaskSnapshot {
	return append([]background.TaskSnapshot(nil), s.tasks...)
}

type stubMissionExecutor struct {
	calls     int
	sessionID string
	input     string
	response  string
}

type stubMissionSharedPendingStore struct {
	latest         *chat.ApprovalRequestMsg
	resolveCount   int
	lastResolvedID string
	lastResponse   approval.ApprovalResponse
	resolveOK      bool
}

func (s *stubMissionSharedPendingStore) Latest() *chat.ApprovalRequestMsg {
	return s.latest
}

func (s *stubMissionSharedPendingStore) HasPending() bool {
	return s.latest != nil
}

func (s *stubMissionSharedPendingStore) Resolve(id string, resp approval.ApprovalResponse) bool {
	s.resolveCount++
	s.lastResolvedID = id
	s.lastResponse = resp
	if s.resolveOK {
		s.latest = nil
	}
	return s.resolveOK
}

func (s *stubMissionExecutor) RunStreamingDetailed(
	_ context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	_ ...adk.RunOption,
) (adk.RunReport, error) {
	s.calls++
	s.sessionID = sessionID
	s.input = input
	if onChunk != nil && s.response != "" {
		onChunk(s.response)
	}
	return adk.RunReport{Response: s.response}, nil
}

type stubMissionSessionStore struct{}

func (s *stubMissionSessionStore) Create(*session.Session) error               { return nil }
func (s *stubMissionSessionStore) Get(string) (*session.Session, error)        { return nil, nil }
func (s *stubMissionSessionStore) Update(*session.Session) error               { return nil }
func (s *stubMissionSessionStore) Delete(string) error                         { return nil }
func (s *stubMissionSessionStore) AppendMessage(string, session.Message) error { return nil }
func (s *stubMissionSessionStore) AnnotateTimeout(string, string) error        { return nil }
func (s *stubMissionSessionStore) End(string) error                            { return nil }
func (s *stubMissionSessionStore) Close() error                                { return nil }
func (s *stubMissionSessionStore) ListSessions(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}
func (s *stubMissionSessionStore) GetSalt(string) ([]byte, error) { return nil, nil }
func (s *stubMissionSessionStore) SetSalt(string, []byte) error   { return nil }

func newMissionComposerWithExecutor(t *testing.T, activity *cockpit.MissionActivityBuffer) (*chat.ChatModel, *stubMissionExecutor) {
	t.Helper()
	return newMissionComposerWithDeps(t, activity, nil)
}

func newMissionComposerWithDeps(
	t *testing.T,
	activity *cockpit.MissionActivityBuffer,
	shared chat.PendingApprovalStore,
) (*chat.ChatModel, *stubMissionExecutor) {
	t.Helper()

	executor := &stubMissionExecutor{response: "acknowledged"}
	runner := turnrunner.New(turnrunner.Config{}, executor, &stubMissionSessionStore{}, nil)
	composer := chat.New(chat.Deps{
		TurnRunner: runner,
		Config: &config.Config{
			Agent: config.AgentConfig{
				Provider: "openai",
				Model:    "gpt-5",
			},
		},
		SessionKey:    "mission-session",
		SharedPending: shared,
		OnUserSubmission: func(sessionKey, input string) {
			if activity != nil {
				activity.Append(cockpit.MissionActivityItem{
					Kind:       cockpit.MissionActivityUser,
					SessionKey: sessionKey,
					Summary:    "User submitted: " + input,
					Timestamp:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
				})
			}
		},
	})

	updated, _ := composer.Update(tea.WindowSizeMsg{Width: 100, Height: 28})
	return updated.(*chat.ChatModel), executor
}

func loadedMissionControlPage(t *testing.T, snapshot cockpit.MissionControlSnapshot) *MissionControlPage {
	t.Helper()

	projector := &stubMissionControlProjector{snapshot: snapshot}
	page := newMissionControlPage(projector, stubMissionTaskSource{}, newTestMissionComposer(t, nil))
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	return updated.(*MissionControlPage)
}

func newTestMissionComposer(t *testing.T, activity *cockpit.MissionActivityBuffer) *chat.ChatModel {
	t.Helper()
	composer, _ := newMissionComposerWithExecutor(t, activity)
	return composer
}

func TestMissionControlFirstScreenChatFallbackCopy(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{})
	view := page.View()

	assert.Contains(t, view, "Type to chat here")
	assert.Contains(t, view, "lango chat")
}

func TestMissionControlHeaderContextRendering(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Header: cockpit.HeaderView{
			ActiveAgentSummary:   "worker-c active",
			ModelProviderSummary: "openai / gpt-5",
			ContextSummary:       "mission-control-wave-one",
			MetricsSummary:       "150 tokens across 1 requests",
			DegradedNote:         "RunLedger unavailable",
		},
	})

	view := page.View()
	assert.Contains(t, view, "worker-c active")
	assert.Contains(t, view, "Pending decisions: 0")
	assert.Contains(t, view, "openai / gpt-5")
	assert.Contains(t, view, "mission-control-wave-one")
	assert.Contains(t, view, "150 tokens across 1 requests")
	assert.Contains(t, view, "RunLedger unavailable")
}

func TestMissionControlFooterDiscoverabilityHint(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Header: cockpit.HeaderView{PendingDecisionCount: 2},
	})
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 72, Height: 24})
	page = updated.(*MissionControlPage)

	view := page.View()
	assert.Contains(t, view, "Tab")
	assert.Contains(t, view, "Composer")
	assert.Contains(t, view, "2 pending")
	assert.Contains(t, view, "Pending decisions: 2")
}

func TestMissionControlFocusCycling(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{})
	require.Equal(t, missionControlFocusMissions, page.focus)

	updated, _ := page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = updated.(*MissionControlPage)
	assert.Equal(t, missionControlFocusDecisions, page.focus)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = updated.(*MissionControlPage)
	assert.Equal(t, missionControlFocusComposer, page.focus)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = updated.(*MissionControlPage)
	assert.Equal(t, missionControlFocusMissions, page.focus)
}

func TestMissionControlPrintableKeyHandoffToComposer(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{})
	require.Equal(t, missionControlFocusMissions, page.focus)

	updated, _ := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	page = updated.(*MissionControlPage)

	assert.Equal(t, missionControlFocusComposer, page.focus)
	assert.Equal(t, "h", page.composer.ComposerValue())
}

func TestMissionControlSlashBehavior(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{})

	updated, _ := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	page = updated.(*MissionControlPage)
	assert.Equal(t, missionControlFocusComposer, page.focus)
	assert.Equal(t, "/", page.composer.ComposerValue())

	page.composer.SetComposerValue("mode")
	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	page = updated.(*MissionControlPage)
	assert.Equal(t, "mode/", page.composer.ComposerValue())
}

func TestMissionControlComposerSubmitReusesTurnPath(t *testing.T) {
	t.Parallel()

	activity := cockpit.NewMissionActivityBuffer()
	composer, executor := newMissionComposerWithExecutor(t, activity)
	projector := cockpit.NewMissionControlProjector(cockpit.Deps{
		Config:         &config.Config{Agent: config.AgentConfig{Provider: "openai", Model: "gpt-5"}},
		SessionKey:     "mission-session",
		ActivityBuffer: activity,
	})
	page := newMissionControlPage(projector, stubMissionTaskSource{}, composer)
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)

	page.composer.SetComposerValue("ship mission control")
	page.focus = missionControlFocusComposer
	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*MissionControlPage)
	require.NotNil(t, cmd)

	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = page.Update(msg)
		page = updated.(*MissionControlPage)
	}

	assert.Equal(t, 1, executor.calls)
	assert.Equal(t, "mission-session", executor.sessionID)
	assert.Equal(t, "ship mission control", executor.input)
}

func TestMissionControlComposerSubmitEchoesUserTextIntoActivity(t *testing.T) {
	t.Parallel()

	activity := cockpit.NewMissionActivityBuffer()
	composer, _ := newMissionComposerWithExecutor(t, activity)
	projector := cockpit.NewMissionControlProjector(cockpit.Deps{
		Config:         &config.Config{Agent: config.AgentConfig{Provider: "openai", Model: "gpt-5"}},
		SessionKey:     "mission-session",
		ActivityBuffer: activity,
	})
	page := newMissionControlPage(projector, stubMissionTaskSource{}, composer)
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)

	page.composer.SetComposerValue("echo this")
	page.focus = missionControlFocusComposer
	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*MissionControlPage)

	assert.Contains(t, page.View(), "User submitted: echo this")
}

func TestMissionControlLoadingEmptyDegradedRendering(t *testing.T) {
	t.Parallel()

	projector := &stubMissionControlProjector{
		snapshot: cockpit.MissionControlSnapshot{
			Degraded: true,
			Header: cockpit.HeaderView{
				DegradedNote: "Agent runtime unavailable",
			},
		},
	}
	page := newMissionControlPage(projector, stubMissionTaskSource{}, newTestMissionComposer(t, nil))
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 90, Height: 28})
	page = updated.(*MissionControlPage)

	assert.Contains(t, page.View(), "Loading Mission Control")

	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)
	view := page.View()
	assert.Contains(t, view, "No active missions")
	assert.Contains(t, view, "Degraded")
	assert.Contains(t, view, "Agent runtime unavailable")
}

func TestMissionControlDecisionsFocusRoutesApprovalKeys(t *testing.T) {
	t.Parallel()

	shared := &stubMissionSharedPendingStore{
		latest: &chat.ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-1",
				ToolName: "fs_write",
			},
			ViewModel: approval.ApprovalViewModel{
				Risk: approval.RiskIndicator{Level: "moderate", Label: "Writes files"},
			},
			Response: make(chan approval.ApprovalResponse, 1),
		},
		resolveOK: true,
	}
	composer, _ := newMissionComposerWithDeps(t, nil, shared)
	page := newMissionControlPage(&stubMissionControlProjector{
		snapshot: cockpit.MissionControlSnapshot{
			Decision: &cockpit.DecisionView{
				ID:         "apr-1",
				Title:      "Approve fs_write",
				Reason:     "Filesystem writes require approval.",
				EffectText: "Update mission control copy.",
				RiskLabel:  "Writes files",
			},
		},
	}, stubMissionTaskSource{}, composer)
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)
	page.focus = missionControlFocusDecisions

	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	page = updated.(*MissionControlPage)

	assert.Equal(t, 1, shared.resolveCount)
	assert.Equal(t, "apr-1", shared.lastResolvedID)
	assert.Equal(t, missionControlFocusDecisions, page.focus)
	assert.Empty(t, page.composer.ComposerValue())
	require.NotNil(t, cmd)
}

func TestMissionControlDecisionsFocusKeepsFullscreenApprovalKeysOutOfComposer(t *testing.T) {
	t.Parallel()

	shared := &stubMissionSharedPendingStore{
		latest: &chat.ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-2",
				ToolName: "fs_write",
			},
			ViewModel: approval.ApprovalViewModel{
				Tier: approval.TierFullscreen,
				Risk: approval.RiskIndicator{Level: "critical", Label: "Writes files"},
			},
			Response: make(chan approval.ApprovalResponse, 1),
		},
	}
	composer, _ := newMissionComposerWithDeps(t, nil, shared)
	page := newMissionControlPage(&stubMissionControlProjector{
		snapshot: cockpit.MissionControlSnapshot{
			Decision: &cockpit.DecisionView{
				ID:         "apr-2",
				Title:      "Approve fs_write",
				Reason:     "Filesystem writes require approval.",
				EffectText: "Update mission control copy.",
				RiskLabel:  "Writes files",
			},
		},
	}, stubMissionTaskSource{}, composer)
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)
	page.focus = missionControlFocusDecisions

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	page = updated.(*MissionControlPage)
	assert.Equal(t, missionControlFocusDecisions, page.focus)
	assert.Empty(t, page.composer.ComposerValue())

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	page = updated.(*MissionControlPage)
	assert.Equal(t, missionControlFocusDecisions, page.focus)
	assert.Empty(t, page.composer.ComposerValue())

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	page = updated.(*MissionControlPage)
	assert.Equal(t, missionControlFocusComposer, page.focus)
	assert.Equal(t, "x", page.composer.ComposerValue())
}

func TestMissionControlDecisionShowsRiskReasonEffectInAllModes(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Decision: &cockpit.DecisionView{
			ID:         "apr-1",
			Title:      "Approve fs_write",
			Reason:     "Filesystem writes require approval.",
			EffectText: "Update mission control copy.",
			RiskLabel:  "Writes files",
		},
	})

	view := page.View()
	assert.Contains(t, view, "Action: Approve fs_write")
	assert.Contains(t, view, "Reason: Filesystem writes require approval.")
	assert.Contains(t, view, "Effect: Update mission control copy.")
	assert.Contains(t, view, "Risk: Writes files")

	updated, _ := page.Update(tea.WindowSizeMsg{Width: 70, Height: 28})
	page = updated.(*MissionControlPage)
	page.focus = missionControlFocusDecisions
	view = page.View()
	assert.Contains(t, view, "Action: Approve fs_write")
	assert.Contains(t, view, "Reason: Filesystem writes require approval.")
	assert.Contains(t, view, "Effect: Update mission control copy.")
	assert.Contains(t, view, "Risk: Writes files")
}

func TestMissionControlEmptyStateIgnoresStandaloneActivitiesAndKeepsComposerVisible(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Activities: []cockpit.ActivityView{{
			Summary:   "User submitted: keep history",
			Timestamp: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
		}},
	})

	view := page.View()
	assert.Contains(t, view, "No active missions or pending decisions.")
	assert.Contains(t, view, page.composer.ComposerPlaceholder())
}

func TestMissionControlDecisionResolutionRefreshesSnapshotImmediately(t *testing.T) {
	t.Parallel()

	shared := &stubMissionSharedPendingStore{
		latest: &chat.ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-3",
				ToolName: "fs_write",
			},
			ViewModel: approval.ApprovalViewModel{
				Risk: approval.RiskIndicator{Level: "moderate", Label: "Writes files"},
			},
			Response: make(chan approval.ApprovalResponse, 1),
		},
		resolveOK: true,
	}
	projector := &stubMissionControlProjector{}
	projector.snapshot = cockpit.MissionControlSnapshot{
		Decision: &cockpit.DecisionView{
			ID:         "apr-3",
			Title:      "Approve fs_write",
			Reason:     "Filesystem writes require approval.",
			EffectText: "Update mission control copy.",
			RiskLabel:  "Writes files",
		},
	}
	composer, _ := newMissionComposerWithDeps(t, nil, shared)
	page := newMissionControlPage(projector, stubMissionTaskSource{}, composer)
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)
	page.focus = missionControlFocusDecisions

	projector.snapshot = cockpit.MissionControlSnapshot{}
	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	page = updated.(*MissionControlPage)

	view := page.View()
	assert.NotContains(t, view, "Action: Approve fs_write")
	assert.Contains(t, view, "No active missions or pending decisions.")
}

func TestMissionControlNarrowAndShortLayouts(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Missions: []cockpit.MissionView{{
			ID:     "bg:1",
			Title:  "Ship layout",
			Status: cockpit.MissionStatusRunning,
		}},
		Decision: &cockpit.DecisionView{
			ID:    "apr-1",
			Title: "Approve fs_write",
		},
	})

	updated, _ := page.Update(tea.WindowSizeMsg{Width: 70, Height: 20})
	page = updated.(*MissionControlPage)
	view := page.View()
	assert.Contains(t, view, "Missions")
	assert.NotContains(t, view, "\nDecisions\n")
	assert.NotContains(t, view, page.composer.ComposerPlaceholder())

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	page = updated.(*MissionControlPage)
	view = page.View()
	assert.Contains(t, view, "> x")
}

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
