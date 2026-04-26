package background

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockRunner struct {
	result string
	err    error
	delay  time.Duration
}

func (m *mockRunner) Run(_ context.Context, _ string, _ string) (string, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.result, m.err
}

type sequenceRunner struct {
	mu        sync.Mutex
	responses []runnerResponse
	calls     int
}

type runnerResponse struct {
	result string
	err    error
}

func (r *sequenceRunner) Run(_ context.Context, _ string, _ string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := r.calls
	r.calls++

	if len(r.responses) == 0 {
		return "", nil
	}
	if idx >= len(r.responses) {
		resp := r.responses[len(r.responses)-1]
		return resp.result, resp.err
	}

	resp := r.responses[idx]
	return resp.result, resp.err
}

func (r *sequenceRunner) Calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

type mockProjection struct {
	mu         sync.Mutex
	id         string
	prepared   int
	synced     []TaskSnapshot
	prepareErr error
	syncErr    error
}

func (m *mockProjection) PrepareTask(_ context.Context, _ string, _ Origin) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.prepareErr != nil {
		return "", m.prepareErr
	}
	m.prepared++
	if m.id != "" {
		return m.id, nil
	}
	return "projected-id", nil
}

func (m *mockProjection) SyncTask(_ context.Context, snap TaskSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.syncErr != nil {
		return m.syncErr
	}
	m.synced = append(m.synced, snap)
	return nil
}

func (m *mockProjection) getPrepared() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.prepared
}

func (m *mockProjection) getSynced() []TaskSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]TaskSnapshot, len(m.synced))
	copy(cp, m.synced)
	return cp
}

func testLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func TestNewManager_Defaults(t *testing.T) {
	mgr := NewManager(&mockRunner{}, nil, 0, 0, testLogger())
	require.NotNil(t, mgr)
	assert.Equal(t, 10, mgr.maxTasks, "default maxTasks should be 10")
	assert.Equal(t, 30*time.Minute, mgr.taskTimeout, "default timeout should be 30m")
}

func TestNewManager_CustomValues(t *testing.T) {
	mgr := NewManager(&mockRunner{}, nil, 5, 10*time.Minute, testLogger())
	assert.Equal(t, 5, mgr.maxTasks)
	assert.Equal(t, 10*time.Minute, mgr.taskTimeout)
}

func TestManager_Submit_And_List(t *testing.T) {
	runner := &mockRunner{result: "done", delay: 50 * time.Millisecond}
	mgr := NewManager(runner, nil, 5, time.Minute, testLogger())

	id, err := mgr.Submit(context.Background(), "test prompt", Origin{Channel: "test"})
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	time.Sleep(10 * time.Millisecond)

	tasks := mgr.List()
	assert.Len(t, tasks, 1)
}

func TestManager_Submit_MaxTasksReached(t *testing.T) {
	runner := &mockRunner{delay: time.Second}
	mgr := NewManager(runner, nil, 1, time.Minute, testLogger())

	id1, err := mgr.Submit(context.Background(), "task1", Origin{})
	require.NoError(t, err)
	assert.NotEmpty(t, id1)

	time.Sleep(20 * time.Millisecond)

	_, err = mgr.Submit(context.Background(), "task2", Origin{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max concurrent tasks")
}

func TestManager_Cancel_NotFound(t *testing.T) {
	mgr := NewManager(&mockRunner{}, nil, 5, time.Minute, testLogger())
	err := mgr.Cancel("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_Status_NotFound(t *testing.T) {
	mgr := NewManager(&mockRunner{}, nil, 5, time.Minute, testLogger())
	snap, err := mgr.Status("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, snap)
}

func TestManager_Result_NotFound(t *testing.T) {
	mgr := NewManager(&mockRunner{}, nil, 5, time.Minute, testLogger())
	result, err := mgr.Result("nonexistent")
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestManager_Submit_And_Result(t *testing.T) {
	runner := &mockRunner{result: "hello world"}
	mgr := NewManager(runner, nil, 5, time.Minute, testLogger())

	id, err := mgr.Submit(context.Background(), "test", Origin{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	result, err := mgr.Result(id)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestManager_WithProjection_UsesPreparedIDAndSyncsLifecycle(t *testing.T) {
	runner := &mockRunner{result: "done"}
	projection := &mockProjection{id: "run-ledger-id"}
	mgr := NewManager(runner, nil, 5, time.Minute, testLogger()).
		WithProjection(projection)

	id, err := mgr.Submit(context.Background(), "test", Origin{})
	require.NoError(t, err)
	assert.Equal(t, "run-ledger-id", id)

	time.Sleep(150 * time.Millisecond)

	require.GreaterOrEqual(t, projection.getPrepared(), 1)
	synced := projection.getSynced()
	require.NotEmpty(t, synced)
	assert.Equal(t, "run-ledger-id", synced[len(synced)-1].ID)
	assert.Equal(t, Done, synced[len(synced)-1].Status)
}

func TestManager_Submit_RunnerError(t *testing.T) {
	runner := &mockRunner{err: fmt.Errorf("runner failed")}
	mgr := NewManager(runner, nil, 5, time.Minute, testLogger())

	id, err := mgr.Submit(context.Background(), "test", Origin{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	snap, err := mgr.Status(id)
	require.NoError(t, err)
	assert.Equal(t, Failed, snap.Status)
}

func TestManager_Status_PreservesRetryMetadata(t *testing.T) {
	mgr := NewManager(&mockRunner{}, nil, 5, time.Minute, testLogger())

	nextRetryAt := time.Now().Add(2 * time.Minute).Truncate(time.Millisecond)
	task := &Task{
		ID:           "task-retry-meta",
		Status:       Failed,
		RetryKey:     "receipt-123:release",
		AttemptCount: 2,
		NextRetryAt:  nextRetryAt,
	}
	mgr.tasks[task.ID] = task

	snap, err := mgr.Status(task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.RetryKey, snap.RetryKey)
	assert.Equal(t, task.AttemptCount, snap.AttemptCount)
	assert.True(t, snap.NextRetryAt.Equal(nextRetryAt))
}

func TestRetryPolicy_DefaultsPreserveBoundedExponentialBackoff(t *testing.T) {
	policy := DefaultRetryPolicy()

	assert.Equal(t, 3, policy.MaxRetryAttempts)
	assert.Equal(t, 25*time.Millisecond, policy.BaseDelay)
	assert.False(t, policy.ShouldScheduleRetry(0))
	assert.True(t, policy.ShouldScheduleRetry(1))
	assert.True(t, policy.ShouldScheduleRetry(3))
	assert.False(t, policy.ShouldScheduleRetry(4))
	assert.Equal(t, 25*time.Millisecond, policy.DelayForAttempt(0))
	assert.Equal(t, 25*time.Millisecond, policy.DelayForAttempt(1))
	assert.Equal(t, 50*time.Millisecond, policy.DelayForAttempt(2))
	assert.Equal(t, 100*time.Millisecond, policy.DelayForAttempt(3))
}

func TestManager_RetryHook_UsesConfiguredRetryPolicy(t *testing.T) {
	runner := &sequenceRunner{
		responses: []runnerResponse{
			{err: fmt.Errorf("attempt 1 failed")},
			{err: fmt.Errorf("attempt 2 failed")},
		},
	}

	var (
		mu    sync.Mutex
		snaps []TaskSnapshot
	)
	hook := func(_ context.Context, snap TaskSnapshot, exhausted bool, retry func()) {
		mu.Lock()
		snaps = append(snaps, snap)
		mu.Unlock()
		if !exhausted {
			retry()
		}
	}

	mgr := NewManager(runner, nil, 5, time.Minute, testLogger()).
		WithRetryPolicy(RetryPolicy{
			MaxRetryAttempts: 1,
			BaseDelay:        time.Millisecond,
		}).
		WithRetryHook(hook)

	id, err := mgr.Submit(context.Background(), "retry once", Origin{Channel: "test"})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return runner.Calls() == 2
	}, time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		snap, err := mgr.Status(id)
		if err != nil {
			return false
		}
		return snap.Status == Failed && snap.AttemptCount == 2
	}, time.Second, 10*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, snaps, 2)
	assert.Equal(t, 1, snaps[0].AttemptCount)
	assert.False(t, snaps[0].NextRetryAt.IsZero())
	assert.Equal(t, 2, snaps[1].AttemptCount)
	assert.True(t, snaps[1].NextRetryAt.IsZero())
}

func TestManager_RetryHook_ResubmitsWithExponentialBackoff(t *testing.T) {
	runner := &sequenceRunner{
		responses: []runnerResponse{
			{err: fmt.Errorf("attempt 1 failed")},
			{err: fmt.Errorf("attempt 2 failed")},
			{result: "ok"},
		},
	}

	var (
		mu    sync.Mutex
		snaps []TaskSnapshot
	)
	hook := func(_ context.Context, snap TaskSnapshot, exhausted bool, retry func()) {
		assert.False(t, exhausted)
		mu.Lock()
		snaps = append(snaps, snap)
		mu.Unlock()
		retry()
	}

	mgr := NewManager(runner, nil, 5, time.Minute, testLogger()).
		WithRetryHook(hook)

	id, err := mgr.Submit(context.Background(), "retry me", Origin{Channel: "test"})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return runner.Calls() == 3
	}, 2*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		snap, err := mgr.Status(id)
		if err != nil {
			return false
		}
		return snap.Status == Done
	}, 2*time.Second, 10*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, snaps, 2)
	assert.Equal(t, 1, snaps[0].AttemptCount)
	assert.Equal(t, 2, snaps[1].AttemptCount)
	assert.False(t, snaps[0].NextRetryAt.IsZero())
	assert.False(t, snaps[1].NextRetryAt.IsZero())
	assert.True(t, snaps[1].NextRetryAt.After(snaps[0].NextRetryAt))
	assert.Greater(t, snaps[0].NextRetryAt.Sub(snaps[0].StartedAt), 0*time.Millisecond)
	assert.Greater(t, snaps[1].NextRetryAt.Sub(snaps[1].StartedAt), snaps[0].NextRetryAt.Sub(snaps[0].StartedAt))
}

func TestManager_RetryHook_StopsAfterExhaustion(t *testing.T) {
	runner := &sequenceRunner{
		responses: []runnerResponse{
			{err: fmt.Errorf("attempt 1 failed")},
			{err: fmt.Errorf("attempt 2 failed")},
			{err: fmt.Errorf("attempt 3 failed")},
			{err: fmt.Errorf("attempt 4 failed")},
		},
	}

	var (
		mu    sync.Mutex
		snaps []TaskSnapshot
	)
	hook := func(_ context.Context, snap TaskSnapshot, exhausted bool, retry func()) {
		mu.Lock()
		snaps = append(snaps, snap)
		mu.Unlock()
		if !exhausted {
			retry()
		}
	}

	mgr := NewManager(runner, nil, 5, time.Minute, testLogger()).
		WithRetryHook(hook)

	id, err := mgr.Submit(context.Background(), "retry exhaustion", Origin{Channel: "test"})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return runner.Calls() == 4
	}, 2*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		snap, err := mgr.Status(id)
		if err != nil {
			return false
		}
		return snap.Status == Failed && snap.AttemptCount == 4
	}, 2*time.Second, 10*time.Millisecond)

	snap, err := mgr.Status(id)
	require.NoError(t, err)
	assert.Equal(t, Failed, snap.Status)
	assert.Equal(t, 4, snap.AttemptCount)
	assert.True(t, snap.NextRetryAt.IsZero())
	assert.NotEmpty(t, snap.Error)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, snaps, 4)
	assert.Equal(t, 1, snaps[0].AttemptCount)
	assert.Equal(t, 2, snaps[1].AttemptCount)
	assert.Equal(t, 3, snaps[2].AttemptCount)
	assert.Equal(t, 4, snaps[3].AttemptCount)
}

func TestManager_Submit_UsesRetryKeyDeriver(t *testing.T) {
	mgr := NewManager(&mockRunner{}, nil, 5, time.Minute, testLogger()).
		WithRetryKeyDeriver(func(prompt string, _ Origin) string {
			return "derived:" + prompt
		})

	id, err := mgr.Submit(context.Background(), "retry me", Origin{Channel: "test"})
	require.NoError(t, err)

	snap, err := mgr.Status(id)
	require.NoError(t, err)
	assert.Equal(t, "derived:retry me", snap.RetryKey)
}

func TestStatus_Valid(t *testing.T) {
	assert.True(t, Pending.Valid())
	assert.True(t, Running.Valid())
	assert.True(t, Done.Valid())
	assert.True(t, Failed.Valid())
	assert.True(t, Cancelled.Valid())
	assert.False(t, Status(0).Valid())
	assert.False(t, Status(99).Valid())
}

func TestStatus_String(t *testing.T) {
	assert.Equal(t, "pending", Pending.String())
	assert.Equal(t, "running", Running.String())
	assert.Equal(t, "done", Done.String())
	assert.Equal(t, "failed", Failed.String())
	assert.Equal(t, "cancelled", Cancelled.String())
	assert.Equal(t, "unknown", Status(0).String())
}

func TestTask_Fail_PreservesCancelledStatus(t *testing.T) {
	task := &Task{
		ID:     "t1",
		Status: Pending,
	}

	task.Cancel()
	assert.Equal(t, Cancelled, task.Status)

	task.Fail("some error")
	assert.Equal(t, Cancelled, task.Status)
	assert.Empty(t, task.Error)
}

func TestTask_Complete_PreservesCancelledStatus(t *testing.T) {
	task := &Task{
		ID:     "t2",
		Status: Pending,
	}

	task.Cancel()
	assert.Equal(t, Cancelled, task.Status)

	task.Complete("result")
	assert.Equal(t, Cancelled, task.Status)
	assert.Empty(t, task.Result)
}

func TestManager_Cancel_PreservesStatus(t *testing.T) {
	runner := &mockRunner{result: "done", delay: 2 * time.Second}
	mgr := NewManager(runner, nil, 5, time.Minute, testLogger())

	id, err := mgr.Submit(context.Background(), "slow task", Origin{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = mgr.Cancel(id)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	snap, err := mgr.Status(id)
	require.NoError(t, err)
	assert.Equal(t, Cancelled, snap.Status)
}

func TestTerminalTaskEviction(t *testing.T) {
	t.Parallel()

	mgr := NewManager(nil, nil, 1000, time.Minute, testLogger())

	totalTasks := maxTerminalTasks + 100 // 600 total

	// Directly populate terminal tasks to avoid goroutine overhead.
	baseTime := time.Now()
	for i := 0; i < totalTasks; i++ {
		id := fmt.Sprintf("task-%04d", i)
		task := &Task{
			ID:          id,
			Status:      Done,
			Prompt:      "prompt",
			Result:      "result",
			StartedAt:   baseTime.Add(time.Duration(i) * time.Second),
			CompletedAt: baseTime.Add(time.Duration(i)*time.Second + time.Millisecond),
		}
		mgr.tasks[id] = task
	}

	// Trigger eviction.
	mgr.mu.Lock()
	mgr.evictTerminalTasksLocked()
	mgr.mu.Unlock()

	// Should have exactly maxTerminalTasks remaining.
	assert.Equal(t, maxTerminalTasks, len(mgr.tasks))

	// The oldest 100 tasks (task-0000 through task-0099) should be evicted.
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("task-%04d", i)
		_, ok := mgr.tasks[id]
		assert.False(t, ok, "oldest task %s should have been evicted", id)
	}

	// The newest 500 tasks (task-0100 through task-0599) should remain.
	for i := 100; i < totalTasks; i++ {
		id := fmt.Sprintf("task-%04d", i)
		_, ok := mgr.tasks[id]
		assert.True(t, ok, "recent task %s should still be present", id)
	}
}

func TestTerminalTaskEviction_PreservesActiveTasks(t *testing.T) {
	t.Parallel()

	mgr := NewManager(nil, nil, 1000, time.Minute, testLogger())

	baseTime := time.Now()

	// Add maxTerminalTasks + 50 terminal tasks.
	for i := 0; i < maxTerminalTasks+50; i++ {
		id := fmt.Sprintf("done-%04d", i)
		mgr.tasks[id] = &Task{
			ID:          id,
			Status:      Done,
			CompletedAt: baseTime.Add(time.Duration(i) * time.Second),
		}
	}

	// Add some active (non-terminal) tasks.
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("active-%d", i)
		mgr.tasks[id] = &Task{
			ID:        id,
			Status:    Running,
			StartedAt: baseTime,
		}
	}

	mgr.mu.Lock()
	mgr.evictTerminalTasksLocked()
	mgr.mu.Unlock()

	// All 5 active tasks should remain.
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("active-%d", i)
		_, ok := mgr.tasks[id]
		assert.True(t, ok, "active task %s should not be evicted", id)
	}

	// Terminal tasks should be capped at maxTerminalTasks.
	terminalCount := 0
	for _, task := range mgr.tasks {
		snap := task.Snapshot()
		if snap.Status == Done || snap.Status == Failed || snap.Status == Cancelled {
			terminalCount++
		}
	}
	assert.Equal(t, maxTerminalTasks, terminalCount)
}

func TestManagerShutdownHonorsContextDeadline(t *testing.T) {
	t.Parallel()

	mgr := NewManager(nil, nil, 1, time.Minute, zap.NewNop().Sugar())

	release := make(chan struct{})
	mgr.wg.Add(1)
	go func() {
		defer mgr.wg.Done()
		<-release
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := mgr.Shutdown(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	close(release)
	require.NoError(t, mgr.Shutdown(context.Background()))
}
