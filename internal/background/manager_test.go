package background

import (
	"context"
	"fmt"
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

type mockProjection struct {
	id         string
	prepared   int
	synced     []TaskSnapshot
	prepareErr error
	syncErr    error
}

func (m *mockProjection) PrepareTask(_ context.Context, _ string, _ Origin) (string, error) {
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
	if m.syncErr != nil {
		return m.syncErr
	}
	m.synced = append(m.synced, snap)
	return nil
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

	require.GreaterOrEqual(t, projection.prepared, 1)
	require.NotEmpty(t, projection.synced)
	assert.Equal(t, "run-ledger-id", projection.synced[len(projection.synced)-1].ID)
	assert.Equal(t, Done, projection.synced[len(projection.synced)-1].Status)
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
