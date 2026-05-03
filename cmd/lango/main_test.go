package main

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/cli/cockpit/pages"
	"github.com/langoai/lango/internal/collabview"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/proposal"
)

type fakeServeApp struct {
	stopFn func(ctx context.Context) error
}

func (f *fakeServeApp) Start(ctx context.Context) error { return nil }

func (f *fakeServeApp) Stop(ctx context.Context) error {
	if f.stopFn != nil {
		return f.stopFn(ctx)
	}
	return nil
}

func TestWatchServeSignals_FirstSignalStartsGracefulShutdown(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopped := make(chan struct{})
	app := &fakeServeApp{
		stopFn: func(ctx context.Context) error {
			close(stopped)
			return nil
		},
	}

	sigChan := make(chan os.Signal, 2)
	forced := make(chan int, 1)

	go watchServeSignals(ctx, app, zap.NewNop().Sugar(), sigChan, time.Second, cancel, func(code int) {
		forced <- code
	})

	sigChan <- os.Interrupt

	select {
	case <-stopped:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected graceful shutdown to start")
	}

	select {
	case code := <-forced:
		t.Fatalf("unexpected forced exit with code %d", code)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWatchServeSignals_SecondSignalForcesExit(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release := make(chan struct{})
	app := &fakeServeApp{
		stopFn: func(ctx context.Context) error {
			<-release
			return nil
		},
	}

	sigChan := make(chan os.Signal, 2)
	forced := make(chan int, 1)
	var once sync.Once

	go watchServeSignals(ctx, app, zap.NewNop().Sugar(), sigChan, time.Second, cancel, func(code int) {
		once.Do(func() { forced <- code })
	})

	sigChan <- os.Interrupt
	sigChan <- os.Interrupt

	select {
	case code := <-forced:
		assert.Equal(t, 130, code)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected forced exit on second signal")
	}

	close(release)
}

func TestCockpitDeadLetterListOptions_MapsAllFields(t *testing.T) {
	t.Parallel()

	got := cockpitDeadLetterListOptions(pages.DeadLetterListOptions{
		Query:                     "tx-1",
		Adjudication:              "release",
		LatestStatusSubtype:       "dead-lettered",
		LatestStatusSubtypeFamily: "dead-letter",
		AnyMatchFamily:            "manual-retry",
		ManualReplayActor:         "operator:alice",
		DeadLetteredAfter:         "2026-04-27T10:00:00Z",
		DeadLetteredBefore:        "2026-04-27T11:00:00Z",
		DeadLetterReasonQuery:     "worker exhausted",
		LatestDispatchReference:   "dispatch-7",
	})

	assert.Equal(t, cockpit.DeadLetterListOptions{
		Query:                     "tx-1",
		Adjudication:              "release",
		LatestStatusSubtype:       "dead-lettered",
		LatestStatusSubtypeFamily: "dead-letter",
		AnyMatchFamily:            "manual-retry",
		ManualReplayActor:         "operator:alice",
		DeadLetteredAfter:         "2026-04-27T10:00:00Z",
		DeadLetteredBefore:        "2026-04-27T11:00:00Z",
		DeadLetterReasonQuery:     "worker exhausted",
		LatestDispatchReference:   "dispatch-7",
	}, got)
}

func TestRunCockpitBuildDepsCarriesMissionService(t *testing.T) {
	t.Parallel()

	svc := mission.NewService(nil)
	store := &stubMainMissionStore{}
	registry := proposal.NewRegistry(nil)
	psvc := proposal.NewService(registry, nil)
	inquiryReader := &stubMainLoopInquiryReader{}
	deadReader := &stubMainLoopDeadReader{}
	cronReader := &stubMainLoopCronReader{}
	collabMissionLinks := &stubMainCollabMissionLinks{}
	collabAgentRuns := &stubMainCollabAgentRuns{}
	collabDelegations := &stubMainCollabDelegations{}
	collabRuntime := &stubMainCollabRuntime{}
	application := &app.App{
		MissionService:                 svc,
		MissionStore:                   store,
		ProposalRegistry:               registry,
		ProposalService:                psvc,
		LoopInquiryReader:              inquiryReader,
		LoopDeadLetterReader:           deadReader,
		LoopCronReader:                 cronReader,
		CollaborationMissionLinkReader: collabMissionLinks,
		CollaborationAgentRunReader:    collabAgentRuns,
		CollaborationDelegationReader:  collabDelegations,
		CollaborationRuntimeReader:     collabRuntime,
	}
	cfg := &config.Config{}
	pending := cockpit.NewPendingApprovalRegistry()
	learning := cockpit.NewLearningSuggestionBuffer(nil)
	activity := cockpit.NewMissionActivityBuffer()

	deps := buildCockpitDeps(application, cfg, "sess-1", nil, "", nil, pending, learning, activity)

	assert.Same(t, svc, deps.MissionService)
	assert.Same(t, store, deps.MissionReader)
	assert.Same(t, registry, deps.ProposalReader)
	assert.Same(t, psvc, deps.ProposalService)
	assert.Same(t, inquiryReader, deps.LoopInquiryReader)
	assert.Same(t, deadReader, deps.LoopDeadReader)
	assert.Same(t, cronReader, deps.LoopCronReader)
	assert.Same(t, collabMissionLinks, deps.CollabMissionLinks)
	assert.Same(t, collabAgentRuns, deps.CollabAgentRuns)
	assert.Same(t, collabDelegations, deps.CollabDelegations)
	assert.Same(t, collabRuntime, deps.CollabRuntime)
	assert.Same(t, learning, deps.LearningBuffer)
	assert.Same(t, activity, deps.ActivityBuffer)
}

type stubMainMissionStore struct{}

func (*stubMainMissionStore) CreateMission(context.Context, mission.CreateMissionInput) (*mission.Mission, error) {
	return nil, nil
}
func (*stubMainMissionStore) GetMission(context.Context, string) (*mission.Mission, error) {
	return nil, nil
}
func (*stubMainMissionStore) ListMissionsBySession(context.Context, string, int) ([]*mission.Mission, error) {
	return nil, nil
}
func (*stubMainMissionStore) TransitionMission(context.Context, mission.TransitionMissionInput) (*mission.Mission, error) {
	return nil, nil
}
func (*stubMainMissionStore) AppendExecutionLink(context.Context, mission.AppendExecutionLinkInput) error {
	return nil
}
func (*stubMainMissionStore) ListExecutionLinks(context.Context, string) ([]*mission.ExecutionLink, error) {
	return nil, nil
}
func (*stubMainMissionStore) FindExecutionLinkByExecution(context.Context, mission.ExecutionKind, string) (*mission.ExecutionLink, error) {
	return nil, nil
}
func (*stubMainMissionStore) FindMissionByExecution(context.Context, mission.ExecutionKind, string) (*mission.Mission, error) {
	return nil, nil
}

type stubMainLoopInquiryReader struct{}

func (*stubMainLoopInquiryReader) ListPendingInquiries(context.Context, string, int) ([]librarian.Inquiry, error) {
	return nil, nil
}

type stubMainLoopDeadReader struct{}

func (*stubMainLoopDeadReader) ListCurrentDeadLetters(context.Context) ([]postadjudicationstatus.DeadLetterBacklogEntry, error) {
	return nil, nil
}

type stubMainLoopCronReader struct{}

func (*stubMainLoopCronReader) List(context.Context) ([]cron.Job, error) {
	return []cron.Job{{ID: uuid.NewString(), Name: "job"}}, nil
}

func (*stubMainLoopCronReader) ListHistory(context.Context, string, int) ([]cron.HistoryEntry, error) {
	return nil, nil
}

type stubMainCollabMissionLinks struct{}

func (*stubMainCollabMissionLinks) ListMissionExecutionLinks(context.Context, string) ([]collabview.CollaborationMissionExecutionLink, error) {
	return nil, nil
}

type stubMainCollabAgentRuns struct{}

func (*stubMainCollabAgentRuns) ListAgentRuns() []collabview.CollaborationAgentRunView { return nil }

type stubMainCollabDelegations struct{}

func (*stubMainCollabDelegations) ListDelegationsForSession(context.Context, string) ([]collabview.CollaborationDelegationRecord, error) {
	return nil, nil
}

type stubMainCollabRuntime struct{}

func (*stubMainCollabRuntime) ListBudgetSignals(string) []collabview.CollaborationBudgetRecord {
	return nil
}
func (*stubMainCollabRuntime) ListRecoverySignals(string) []collabview.CollaborationRecoveryRecord {
	return nil
}
