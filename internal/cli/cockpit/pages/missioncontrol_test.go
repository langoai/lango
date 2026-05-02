package pages

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/proposal"
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
	ctx       context.Context
	sessionID string
	input     string
	response  string
}

type stubMissionLifecycleService struct {
	startCalls      int
	acceptCalls     int
	lastStartInput  mission.StartMissionInput
	lastAcceptInput mission.AcceptProposalInput
	startResult     *mission.Mission
	acceptResult    *mission.Mission
	startErr        error
	acceptErr       error
}

type stubMissionProposalReader struct {
	items map[string]proposal.Proposal
}

func (s *stubMissionProposalReader) ListBySession(sessionKey string) []proposal.Proposal {
	var out []proposal.Proposal
	for _, item := range s.items {
		if item.SessionKey == sessionKey {
			out = append(out, item)
		}
	}
	return out
}

func (s *stubMissionProposalReader) GetByID(proposalID string) (proposal.Proposal, bool) {
	item, ok := s.items[proposalID]
	return item, ok
}

type stubMissionProposalService struct {
	acceptCalls    int
	dismissCalls   int
	restoreCalls   int
	lastAcceptID   string
	lastDismissID  string
	lastRestoreID  string
	accepted       map[string]struct{}
	dismissed      map[string]struct{}
	acceptErr      error
	dismissErr     error
	restoreErr     error
	proposalReader *stubMissionProposalReader
}

func (s *stubMissionProposalService) Accept(_ context.Context, proposalID string) (*proposal.Proposal, error) {
	s.acceptCalls++
	s.lastAcceptID = proposalID
	if s.acceptErr != nil {
		return nil, s.acceptErr
	}
	if s.accepted == nil {
		s.accepted = make(map[string]struct{})
	}
	s.accepted[proposalID] = struct{}{}
	if s.proposalReader != nil {
		delete(s.proposalReader.items, proposalID)
	}
	item := proposal.Proposal{ProposalID: proposalID, Status: proposal.ProposalStatusAccepted}
	return &item, nil
}

func (s *stubMissionProposalService) Dismiss(_ context.Context, proposalID string) (*proposal.Proposal, error) {
	s.dismissCalls++
	s.lastDismissID = proposalID
	if s.dismissErr != nil {
		return nil, s.dismissErr
	}
	if s.dismissed == nil {
		s.dismissed = make(map[string]struct{})
	}
	s.dismissed[proposalID] = struct{}{}
	if s.proposalReader != nil {
		delete(s.proposalReader.items, proposalID)
	}
	item := proposal.Proposal{ProposalID: proposalID, Status: proposal.ProposalStatusDismissed}
	return &item, nil
}

func (s *stubMissionProposalService) RestorePrepared(_ context.Context, proposalID string) (*proposal.Proposal, error) {
	s.restoreCalls++
	s.lastRestoreID = proposalID
	if s.restoreErr != nil {
		return nil, s.restoreErr
	}
	if s.proposalReader != nil {
		s.proposalReader.items[proposalID] = proposal.Proposal{
			ProposalID: proposalID,
			SessionKey: "mission-session",
			Status:     proposal.ProposalStatusPrepared,
		}
	}
	item := proposal.Proposal{ProposalID: proposalID, Status: proposal.ProposalStatusPrepared}
	return &item, nil
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
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	_ ...adk.RunOption,
) (adk.RunReport, error) {
	s.calls++
	s.ctx = ctx
	s.sessionID = sessionID
	s.input = input
	if onChunk != nil && s.response != "" {
		onChunk(s.response)
	}
	return adk.RunReport{Response: s.response}, nil
}

func (s *stubMissionLifecycleService) StartMission(_ context.Context, in mission.StartMissionInput) (*mission.Mission, error) {
	s.startCalls++
	s.lastStartInput = in
	if s.startErr != nil {
		return nil, s.startErr
	}
	if s.startResult != nil {
		return s.startResult, nil
	}
	return &mission.Mission{ID: uuid.New()}, nil
}

func (s *stubMissionLifecycleService) AcceptProposal(_ context.Context, in mission.AcceptProposalInput) (*mission.Mission, error) {
	s.acceptCalls++
	s.lastAcceptInput = in
	if s.acceptErr != nil {
		return nil, s.acceptErr
	}
	if s.acceptResult != nil {
		return s.acceptResult, nil
	}
	return &mission.Mission{ID: uuid.New()}, nil
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

func TestMissionControlComposerSubmitCreatesMissionBeforeTurnDispatch(t *testing.T) {
	t.Parallel()

	activity := cockpit.NewMissionActivityBuffer()
	composer, executor := newMissionComposerWithExecutor(t, activity)
	svc := &stubMissionLifecycleService{
		startResult: &mission.Mission{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111")},
	}
	page := newMissionControlPage(&stubMissionControlProjector{}, stubMissionTaskSource{}, composer)
	page.missionService = svc
	page.sessionKey = "mission-session"
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()

	page.composer.SetComposerValue("ship mission control")
	page.focus = missionControlFocusComposer
	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*MissionControlPage)

	require.NotNil(t, cmd)
	assert.Equal(t, 1, svc.startCalls)
	assert.Equal(t, "mission-session", svc.lastStartInput.SessionKey)
	assert.Equal(t, "ship mission control", svc.lastStartInput.Title)
	assert.True(t, svc.lastStartInput.StartActive)

	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = page.Update(msg)
		page = updated.(*MissionControlPage)
	}

	assert.Equal(t, 1, executor.calls)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", ctxkeys.MissionIDFromContext(executor.ctx))
}

func TestMissionControlAcceptProposedMissionCreatesDurableRowAndRemovesOverlay(t *testing.T) {
	t.Parallel()

	svc := &stubMissionLifecycleService{
		acceptResult: &mission.Mission{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222")},
	}
	proposalReader := &stubMissionProposalReader{
		items: map[string]proposal.Proposal{
			"proposal-1": {
				ProposalID: "proposal-1",
				SessionKey: "mission-session",
				Source: proposal.ProposalSource{
					Kind: "proposed_learning",
					Ref:  "s-1",
				},
				Title:  "Apply bounded retry guidance",
				Status: proposal.ProposalStatusPrepared,
				PreparedBrief: &proposal.PreparedBrief{
					SourceSummary:             "Learning suggestion: repeated timeout retries.",
					Reason:                    "Repeated timeout failures benefited from bounded retry.",
					SuggestedAcceptanceEffect: "Create a mission to apply bounded retry guidance.",
					SupportingEvidence: []string{
						"Pattern: timeout retries",
						"Confidence: 0.75",
					},
				},
			},
		},
	}
	proposalSvc := &stubMissionProposalService{proposalReader: proposalReader}
	deps := cockpit.Deps{
		Config:         &config.Config{Agent: config.AgentConfig{Provider: "openai", Model: "gpt-5"}},
		SessionKey:     "mission-session",
		ProposalReader: proposalReader,
	}
	projector := cockpit.NewMissionControlProjector(deps)
	page := newMissionControlPage(projector, stubMissionTaskSource{}, newTestMissionComposer(t, nil))
	page.missionService = svc
	page.proposalReader = proposalReader
	page.proposalSvc = proposalSvc
	page.sessionKey = "mission-session"
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)

	require.Len(t, page.snapshot.Missions, 1)
	require.Equal(t, cockpit.MissionKindProposed, page.snapshot.Missions[0].Kind)
	page.focus = missionControlFocusMissions

	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*MissionControlPage)

	assert.Nil(t, cmd)
	assert.Equal(t, 1, proposalSvc.acceptCalls)
	assert.Equal(t, "proposal-1", proposalSvc.lastAcceptID)
	assert.Equal(t, 1, svc.acceptCalls)
	assert.Equal(t, "mission-session", svc.lastAcceptInput.SessionKey)
	assert.Equal(t, "proposed_learning", svc.lastAcceptInput.SourceKind)
	assert.Equal(t, "s-1", svc.lastAcceptInput.SourceRef)
	assert.Contains(t, svc.lastAcceptInput.Description, "Learning suggestion: repeated timeout retries.")
	assert.Contains(t, svc.lastAcceptInput.Description, "Repeated timeout failures benefited from bounded retry.")
	assert.Contains(t, svc.lastAcceptInput.Description, "Create a mission to apply bounded retry guidance.")
	assert.Contains(t, svc.lastAcceptInput.Description, "Evidence: Pattern: timeout retries")
	assert.Contains(t, svc.lastAcceptInput.Description, "Evidence: Confidence: 0.75")
	assert.Empty(t, page.snapshot.Missions)
}

func TestMissionControlAcceptProposedMissionRestoresProposalIfMissionCreateFails(t *testing.T) {
	t.Parallel()

	svc := &stubMissionLifecycleService{
		acceptErr: assert.AnError,
	}
	proposalReader := &stubMissionProposalReader{
		items: map[string]proposal.Proposal{
			"proposal-1": {
				ProposalID: "proposal-1",
				SessionKey: "mission-session",
				Source: proposal.ProposalSource{
					Kind: "proposed_learning",
					Ref:  "s-1",
				},
				Title:  "Apply bounded retry guidance",
				Status: proposal.ProposalStatusPrepared,
				PreparedBrief: &proposal.PreparedBrief{
					SourceSummary: "Prepared summary",
				},
			},
		},
	}
	proposalSvc := &stubMissionProposalService{proposalReader: proposalReader}
	deps := cockpit.Deps{
		Config:         &config.Config{Agent: config.AgentConfig{Provider: "openai", Model: "gpt-5"}},
		SessionKey:     "mission-session",
		ProposalReader: proposalReader,
	}
	projector := cockpit.NewMissionControlProjector(deps)
	page := newMissionControlPage(projector, stubMissionTaskSource{}, newTestMissionComposer(t, nil))
	page.missionService = svc
	page.proposalReader = proposalReader
	page.proposalSvc = proposalSvc
	page.sessionKey = "mission-session"
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)
	require.Len(t, page.snapshot.Missions, 1)

	page.focus = missionControlFocusMissions
	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*MissionControlPage)

	require.NotNil(t, cmd)
	assert.Equal(t, 1, proposalSvc.acceptCalls)
	assert.Equal(t, 1, proposalSvc.restoreCalls)
	assert.Equal(t, "proposal-1", proposalSvc.lastRestoreID)
	assert.Equal(t, 1, svc.acceptCalls)

	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = page.Update(msg)
		page = updated.(*MissionControlPage)
	}

	page.refreshSnapshot()
	require.Len(t, page.snapshot.Missions, 1)
	assert.Equal(t, "proposal-1", page.snapshot.Missions[0].ID)
}

func TestMissionControlDismissProposedMissionUsesProposalServiceAndRemovesOverlay(t *testing.T) {
	t.Parallel()

	proposalReader := &stubMissionProposalReader{
		items: map[string]proposal.Proposal{
			"proposal-1": {
				ProposalID: "proposal-1",
				SessionKey: "mission-session",
				Title:      "Dismiss me",
				Status:     proposal.ProposalStatusPrepared,
			},
		},
	}
	proposalSvc := &stubMissionProposalService{proposalReader: proposalReader}
	deps := cockpit.Deps{
		Config:         &config.Config{Agent: config.AgentConfig{Provider: "openai", Model: "gpt-5"}},
		SessionKey:     "mission-session",
		ProposalReader: proposalReader,
	}
	projector := cockpit.NewMissionControlProjector(deps)
	page := newMissionControlPage(projector, stubMissionTaskSource{}, newTestMissionComposer(t, nil))
	page.proposalReader = proposalReader
	page.proposalSvc = proposalSvc
	page.sessionKey = "mission-session"
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)
	require.Len(t, page.snapshot.Missions, 1)

	page.focus = missionControlFocusMissions
	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	page = updated.(*MissionControlPage)

	assert.Nil(t, cmd)
	assert.Equal(t, 1, proposalSvc.dismissCalls)
	assert.Equal(t, "proposal-1", proposalSvc.lastDismissID)
	assert.Empty(t, page.snapshot.Missions)
}

func TestMissionControlPageFlowDoesNotCreateDurableMissionBeforeAcceptance(t *testing.T) {
	t.Parallel()

	svc := &stubMissionLifecycleService{}
	proposalReader := &stubMissionProposalReader{
		items: map[string]proposal.Proposal{
			"proposal-1": {
				ProposalID: "proposal-1",
				SessionKey: "mission-session",
				Title:      "Prepared proposal",
				Status:     proposal.ProposalStatusPrepared,
			},
		},
	}
	deps := cockpit.Deps{
		Config:         &config.Config{Agent: config.AgentConfig{Provider: "openai", Model: "gpt-5"}},
		SessionKey:     "mission-session",
		ProposalReader: proposalReader,
	}
	projector := cockpit.NewMissionControlProjector(deps)
	page := newMissionControlPage(projector, stubMissionTaskSource{}, newTestMissionComposer(t, nil))
	page.missionService = svc
	page.proposalReader = proposalReader
	page.proposalSvc = &stubMissionProposalService{proposalReader: proposalReader}
	page.sessionKey = "mission-session"
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	page = updated.(*MissionControlPage)
	page.Activate()
	updated, _ = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)

	require.Len(t, page.snapshot.Missions, 1)
	assert.Equal(t, 0, svc.acceptCalls)
}

func TestMissionControlSharedComposerDoesNotReplayChannelMessages(t *testing.T) {
	t.Parallel()

	deps := cockpit.Deps{
		Config:           &config.Config{Agent: config.AgentConfig{Provider: "openai", Model: "gpt-5"}},
		PendingApprovals: cockpit.NewPendingApprovalRegistry(),
	}
	model := cockpit.New(deps)
	page := newMissionControlPage(cockpit.NewMissionControlProjector(deps), stubMissionTaskSource{}, model.ChatModel())
	model.RegisterPage(cockpit.PageMissionControl, page)

	model.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	model.Update(chat.ChannelMessageMsg{
		Channel:    "telegram",
		SessionKey: "telegram:1:2",
		SenderName: "alice",
		Text:       "hello from channel",
		Timestamp:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	})

	main := ansi.Strip(model.ChatModel().RenderParts().Main)
	assert.Equal(t, 1, strings.Count(main, "hello from channel"))
	assert.Equal(t, 1, strings.Count(ansi.Strip(page.composer.RenderParts().Main), "hello from channel"))
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
