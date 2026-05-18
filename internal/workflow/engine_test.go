package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockAgentRunner struct {
	mu       sync.Mutex
	result   string
	err      error
	delay    time.Duration
	sessions []string
}

func (m *mockAgentRunner) Run(ctx context.Context, sessionKey string, _ string) (string, error) {
	m.mu.Lock()
	m.sessions = append(m.sessions, sessionKey)
	m.mu.Unlock()
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return m.result, m.err
}

type blockingAgentRunner struct {
	result      string
	started     chan struct{}
	release     chan struct{}
	startOnce   sync.Once
	releaseOnce sync.Once
}

func newBlockingAgentRunner(result string) *blockingAgentRunner {
	return &blockingAgentRunner{
		result:  result,
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (r *blockingAgentRunner) Run(ctx context.Context, _ string, _ string) (string, error) {
	r.startOnce.Do(func() {
		close(r.started)
	})
	select {
	case <-r.release:
		return r.result, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (r *blockingAgentRunner) releaseAll() {
	r.releaseOnce.Do(func() {
		close(r.release)
	})
}

type workflowStepUpdate struct {
	stepID string
	status string
	result string
	errMsg string
}

type memoryRunStore struct {
	mu             sync.Mutex
	nextRunID      string
	createRunErr   error
	updateRunErr   error
	createStepErr  error
	completeRunErr error
	getStatusErr   error
	getResultsErr  error
	listRunsErr    error
	runs           map[string]*RunStatus
	stepResults    map[string]map[string]string
	stepCreated    map[string]map[string]bool
	runStatuses    []string
	completions    []RunStatus
	completed      chan struct{}
	stepCreates    []string
	stepUpdates    []workflowStepUpdate
}

func newMemoryRunStore() *memoryRunStore {
	return &memoryRunStore{
		nextRunID:   "run-1",
		runs:        make(map[string]*RunStatus),
		stepResults: make(map[string]map[string]string),
		stepCreated: make(map[string]map[string]bool),
		completed:   make(chan struct{}, 10),
	}
}

func (s *memoryRunStore) CreateRun(_ context.Context, w *Workflow) (string, error) {
	if s.createRunErr != nil {
		return "", s.createRunErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	runID := s.nextRunID
	s.runs[runID] = &RunStatus{
		RunID:        runID,
		WorkflowName: w.Name,
		Status:       "pending",
		TotalSteps:   len(w.Steps),
		StartedAt:    time.Now(),
	}
	s.stepResults[runID] = make(map[string]string)
	return runID, nil
}

func (s *memoryRunStore) UpdateRunStatus(_ context.Context, runID string, status string) error {
	if s.updateRunErr != nil {
		return s.updateRunErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runStatuses = append(s.runStatuses, status)
	if run := s.runs[runID]; run != nil {
		run.Status = status
	}
	return nil
}

func (s *memoryRunStore) CompleteRun(
	_ context.Context,
	runID string,
	status string,
	errMsg string,
) error {
	if s.completeRunErr != nil {
		return s.completeRunErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	run := s.runs[runID]
	if run == nil {
		run = &RunStatus{RunID: runID}
		s.runs[runID] = run
	}
	run.Status = status
	run.CompletedSteps = len(s.stepResults[runID])
	completion := *run
	for _, update := range s.stepUpdates {
		if update.errMsg != "" {
			completion.StepStatuses = append(completion.StepStatuses, StepStatus{
				StepID: update.stepID,
				Status: update.status,
				Error:  update.errMsg,
			})
		}
	}
	s.completions = append(s.completions, completion)
	select {
	case s.completed <- struct{}{}:
	default:
	}
	return nil
}

func (s *memoryRunStore) CreateStepRun(
	_ context.Context,
	runID string,
	step Step,
	_ string,
) error {
	if s.createStepErr != nil {
		return s.createStepErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stepCreated[runID] == nil {
		s.stepCreated[runID] = make(map[string]bool)
	}
	if s.stepCreated[runID][step.ID] {
		return nil
	}
	s.stepCreated[runID][step.ID] = true
	s.stepCreates = append(s.stepCreates, step.ID)
	if run := s.runs[runID]; run != nil {
		run.StepStatuses = append(run.StepStatuses, StepStatus{
			StepID: step.ID,
			Agent:  step.Agent,
			Status: "pending",
		})
	}
	return nil
}

func (s *memoryRunStore) stepCreateSnapshot() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.stepCreates))
	copy(out, s.stepCreates)
	return out
}

func (s *memoryRunStore) waitForCompletion(t *testing.T) {
	t.Helper()
	select {
	case <-s.completed:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for workflow completion")
	}
}

func (s *memoryRunStore) UpdateStepStatus(
	_ context.Context,
	runID string,
	stepID string,
	status string,
	result string,
	errMsg string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stepUpdates = append(s.stepUpdates, workflowStepUpdate{
		stepID: stepID,
		status: status,
		result: result,
		errMsg: errMsg,
	})
	if status == "completed" {
		s.stepResults[runID][stepID] = result
	}
	if run := s.runs[runID]; run != nil {
		for i := range run.StepStatuses {
			if run.StepStatuses[i].StepID == stepID {
				run.StepStatuses[i].Status = status
				run.StepStatuses[i].Error = errMsg
				return nil
			}
		}
		run.StepStatuses = append(run.StepStatuses, StepStatus{
			StepID: stepID,
			Status: status,
			Error:  errMsg,
		})
	}
	return nil
}

func (s *memoryRunStore) GetRunStatus(_ context.Context, runID string) (*RunStatus, error) {
	if s.getStatusErr != nil {
		return nil, s.getStatusErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	run := s.runs[runID]
	if run == nil {
		return nil, fmt.Errorf("missing run %q", runID)
	}
	cp := *run
	cp.StepStatuses = append([]StepStatus(nil), run.StepStatuses...)
	return &cp, nil
}

func (s *memoryRunStore) GetStepResults(
	_ context.Context,
	runID string,
) (map[string]string, error) {
	if s.getResultsErr != nil {
		return nil, s.getResultsErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	results := make(map[string]string, len(s.stepResults[runID]))
	for stepID, result := range s.stepResults[runID] {
		results[stepID] = result
	}
	return results, nil
}

func (s *memoryRunStore) ListRuns(_ context.Context, _ int) ([]RunStatus, error) {
	if s.listRunsErr != nil {
		return nil, s.listRunsErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	runs := make([]RunStatus, 0, len(s.runs))
	for _, run := range s.runs {
		runs = append(runs, *run)
	}
	return runs, nil
}

type promptCapturingRunner struct {
	mu      sync.Mutex
	results map[string]string
	errs    map[string]error
	prompts []string
}

func (r *promptCapturingRunner) Run(
	_ context.Context,
	sessionKey string,
	prompt string,
) (string, error) {
	stepID := sessionKey[strings.LastIndex(sessionKey, ":")+1:]
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prompts = append(r.prompts, prompt)
	if err := r.errs[stepID]; err != nil {
		return "", err
	}
	return r.results[stepID], nil
}

type recordingChannelSender struct {
	mu       sync.Mutex
	messages map[string][]string
}

func (s *recordingChannelSender) SendMessage(
	_ context.Context,
	channel string,
	message string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.messages == nil {
		s.messages = make(map[string][]string)
	}
	s.messages[channel] = append(s.messages[channel], message)
	return nil
}

func TestEngine_ExecuteStep_ChecksCancellation(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{result: "ok"}
	logger := zap.NewNop().Sugar()

	e := &Engine{
		runner:         runner,
		maxConcurrent:  4,
		defaultTimeout: 5 * time.Minute,
		logger:         logger,
		cancels:        make(map[string]context.CancelFunc),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	step := &Step{ID: "step-1", Prompt: "do something"}
	_, err := e.executeStep(ctx, "run-1", "wf", step, nil)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	runner.mu.Lock()
	assert.Empty(t, runner.sessions)
	runner.mu.Unlock()
}

func TestEngine_SessionKeyFormat(t *testing.T) {
	t.Parallel()

	key1 := fmt.Sprintf("workflow:%s:%s:%s", "my-wf", "run-1", "step-a")
	key2 := fmt.Sprintf("workflow:%s:%s:%s", "my-wf", "run-2", "step-a")

	assert.Equal(t, "workflow:my-wf:run-1:step-a", key1)
	assert.Equal(t, "workflow:my-wf:run-2:step-a", key2)
	assert.NotEqual(t, key1, key2, "different runIDs must produce different session keys")
	assert.True(t, strings.Contains(key1, "run-1"))
	assert.True(t, strings.Contains(key2, "run-2"))
}

func TestEngine_ExecuteStep_RunnerError(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{err: fmt.Errorf("agent failed")}
	logger := zap.NewNop().Sugar()

	e := &Engine{
		runner:         runner,
		maxConcurrent:  4,
		defaultTimeout: 5 * time.Minute,
		logger:         logger,
		cancels:        make(map[string]context.CancelFunc),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	step := &Step{ID: "step-1", Prompt: "fail"}
	_, err := e.executeStep(ctx, "run-1", "wf", step, nil)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestEngineShutdownHonorsContextDeadline(t *testing.T) {
	t.Parallel()

	engine := NewEngine(nil, nil, nil, 1, time.Minute, zap.NewNop().Sugar())

	release := make(chan struct{})
	engine.wg.Add(1)
	go func() {
		defer engine.wg.Done()
		<-release
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := engine.Shutdown(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	close(release)
	require.NoError(t, engine.Shutdown(context.Background()))
}

func TestEngine_Run_ExecutesDependenciesAndDeliversResults(t *testing.T) {
	t.Parallel()

	state := newMemoryRunStore()
	runner := &promptCapturingRunner{
		results: map[string]string{
			"gather": "alpha",
			"final":  "omega",
		},
		errs: make(map[string]error),
	}
	sender := &recordingChannelSender{}
	engine := NewEngine(runner, state, sender, 1, time.Minute, zap.NewNop().Sugar())
	workflow := &Workflow{
		Name:      "demo",
		DeliverTo: []string{"ops"},
		Steps: []Step{
			{ID: "gather", Prompt: "collect data", DeliverTo: []string{"step-log"}},
			{ID: "final", Prompt: "summarize {{gather.result}}", DependsOn: []string{"gather"}},
		},
	}

	result, err := engine.Run(context.Background(), workflow)
	require.NoError(t, err)

	assert.Equal(t, "run-1", result.RunID)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, map[string]string{"gather": "alpha", "final": "omega"}, result.StepResults)
	assert.Equal(t, []string{"running"}, state.runStatuses)
	assert.ElementsMatch(t, []string{"gather", "final"}, state.stepCreates)
	assert.Len(t, state.completions, 1)
	assert.Equal(t, "completed", state.completions[0].Status)

	runner.mu.Lock()
	require.Len(t, runner.prompts, 2)
	assert.Contains(t, runner.prompts[0], automationPrefix)
	assert.Contains(t, runner.prompts[1], "summarize alpha")
	runner.mu.Unlock()

	sender.mu.Lock()
	require.Len(t, sender.messages["step-log"], 1)
	assert.Contains(t, sender.messages["step-log"][0], "[demo/gather] alpha")
	require.Len(t, sender.messages["ops"], 1)
	assert.Contains(t, sender.messages["ops"][0], "Workflow 'demo' completed.")
	assert.Contains(t, sender.messages["ops"][0], "omega")
	sender.mu.Unlock()
}

func TestEngine_Run_FailedStepMarksRunFailed(t *testing.T) {
	t.Parallel()

	state := newMemoryRunStore()
	runner := &promptCapturingRunner{
		results: make(map[string]string),
		errs:    map[string]error{"first": fmt.Errorf("agent exploded")},
	}
	engine := NewEngine(runner, state, nil, 1, time.Minute, zap.NewNop().Sugar())
	workflow := &Workflow{
		Name:  "failure",
		Steps: []Step{{ID: "first", Prompt: "fail"}},
	}

	result, err := engine.Run(context.Background(), workflow)
	require.NoError(t, err)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, `step "first"`)
	assert.Contains(t, result.Error, "agent exploded")
	require.Len(t, state.completions, 1)
	assert.Equal(t, "failed", state.completions[0].Status)
	require.NotEmpty(t, state.stepUpdates)
	last := state.stepUpdates[len(state.stepUpdates)-1]
	assert.Equal(t, "first", last.stepID)
	assert.Equal(t, "failed", last.status)
	assert.Contains(t, last.errMsg, "agent exploded")
}

func TestEngine_RunAsync_PrecreatesStepsAndStatusCanBeQueried(t *testing.T) {
	t.Parallel()

	state := newMemoryRunStore()
	runner := newBlockingAgentRunner("ok")
	engine := NewEngine(runner, state, nil, 1, time.Minute, zap.NewNop().Sugar())
	t.Cleanup(func() {
		runner.releaseAll()
		_ = engine.Shutdown(context.Background())
	})
	workflow := &Workflow{
		Name: "async",
		Steps: []Step{
			{ID: "a", Prompt: "a"},
			{ID: "b", Prompt: "b"},
		},
	}

	runID, err := engine.RunAsync(context.Background(), workflow)
	require.NoError(t, err)
	assert.Equal(t, "run-1", runID)

	assert.ElementsMatch(t, []string{"a", "b"}, state.stepCreateSnapshot())
	immediateStatus, err := engine.Status(context.Background(), runID)
	require.NoError(t, err)
	assert.NotEqual(t, "completed", immediateStatus.Status)
	require.Len(t, immediateStatus.StepStatuses, 2)
	assert.ElementsMatch(t, []string{"a", "b"}, []string{
		immediateStatus.StepStatuses[0].StepID,
		immediateStatus.StepStatuses[1].StepID,
	})

	runner.releaseAll()
	state.waitForCompletion(t)
	assert.ElementsMatch(t, []string{"a", "b"}, state.stepCreateSnapshot())
	status, err := engine.Status(context.Background(), runID)
	require.NoError(t, err)
	assert.Equal(t, "completed", status.Status)

	runs, err := engine.ListRuns(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	assert.Equal(t, runID, runs[0].RunID)

	require.NoError(t, engine.Shutdown(context.Background()))
}

func TestEngine_ResumeReturnsExistingResultsAndRejectsCompletedRun(t *testing.T) {
	t.Parallel()

	state := newMemoryRunStore()
	state.runs["running-run"] = &RunStatus{
		RunID:        "running-run",
		WorkflowName: "resume-me",
		Status:       "failed",
		TotalSteps:   2,
	}
	state.stepResults["running-run"] = map[string]string{"done": "cached"}
	state.runs["completed-run"] = &RunStatus{
		RunID:        "completed-run",
		WorkflowName: "done",
		Status:       "completed",
	}
	engine := NewEngine(nil, state, nil, 1, time.Minute, zap.NewNop().Sugar())

	result, err := engine.Resume(context.Background(), "running-run")
	require.NoError(t, err)
	assert.Equal(t, "resume-me", result.WorkflowName)
	assert.Equal(t, "failed", result.Status)
	assert.Equal(t, map[string]string{"done": "cached"}, result.StepResults)

	_, err = engine.Resume(context.Background(), "completed-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already completed")
}

func TestEngine_CancelReturnsBoundaryErrors(t *testing.T) {
	t.Parallel()

	state := newMemoryRunStore()
	engine := NewEngine(nil, state, nil, 1, time.Minute, zap.NewNop().Sugar())

	err := engine.Cancel("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found or not running")

	state.completeRunErr = fmt.Errorf("store unavailable")
	engine.mu.Lock()
	engine.cancels["run-1"] = func() {}
	engine.mu.Unlock()

	err = engine.Cancel("run-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update cancelled run")
}
