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

	// Give time for task to start.
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

	// Wait for the first task to become active.
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

	// Wait for completion.
	time.Sleep(100 * time.Millisecond)

	result, err := mgr.Result(id)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestManager_Submit_RunnerError(t *testing.T) {
	runner := &mockRunner{err: fmt.Errorf("runner failed")}
	mgr := NewManager(runner, nil, 5, time.Minute, testLogger())

	id, err := mgr.Submit(context.Background(), "test", Origin{})
	require.NoError(t, err)

	// Wait for completion.
	time.Sleep(100 * time.Millisecond)

	snap, err := mgr.Status(id)
	require.NoError(t, err)
	assert.Equal(t, Failed, snap.Status)
}

// Test Status enum.
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
