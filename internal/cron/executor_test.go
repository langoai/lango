package cron

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewExecutor(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "ok"}
	store := newMockStore()
	logger := zap.NewNop().Sugar()

	e := NewExecutor(runner, nil, store, logger)

	require.NotNil(t, e)
	assert.Equal(t, runner, e.runner)
	assert.Equal(t, store, e.store)
}

func TestExecutor_Execute_HappyPath(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "task completed"}
	store := newMockStore()
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	job := Job{
		ID:           "job-1",
		Name:         "test-job",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "do the thing",
		SessionMode:  "isolated",
	}

	result := executor.Execute(context.Background(), job)

	require.NotNil(t, result)
	assert.Equal(t, "job-1", result.JobID)
	assert.Equal(t, "test-job", result.JobName)
	assert.Equal(t, "task completed", result.Response)
	assert.NoError(t, result.Error)
	assert.True(t, result.Duration > 0)

	// History should be saved.
	assert.Len(t, store.history, 1)
	assert.Equal(t, "completed", store.history[0].Status)
	assert.Equal(t, "task completed", store.history[0].Result)
}

func TestExecutor_Execute_RunnerError(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{err: fmt.Errorf("agent crashed")}
	store := newMockStore()
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	job := Job{
		ID:           "job-2",
		Name:         "failing-job",
		ScheduleType: "cron",
		Schedule:     "* * * * *",
		Prompt:       "do something risky",
		SessionMode:  "main",
	}

	result := executor.Execute(context.Background(), job)

	require.NotNil(t, result)
	assert.Error(t, result.Error)
	assert.Equal(t, "agent crashed", result.Error.Error())

	// History should record the failure.
	require.Len(t, store.history, 1)
	assert.Equal(t, "failed", store.history[0].Status)
	assert.Equal(t, "agent crashed", store.history[0].ErrorMessage)
}

func TestExecutor_Execute_WithDelivery(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "done"}
	store := newMockStore()
	sender := &mockChannelSender{}
	logger := zap.NewNop().Sugar()
	delivery := NewDelivery(sender, nil, logger)
	executor := NewExecutor(runner, delivery, store, logger)

	job := Job{
		ID:           "job-3",
		Name:         "delivery-job",
		ScheduleType: "every",
		Schedule:     "5m",
		Prompt:       "check status",
		SessionMode:  "isolated",
		DeliverTo:    []string{"channel-1"},
	}

	result := executor.Execute(context.Background(), job)

	require.NotNil(t, result)
	assert.NoError(t, result.Error)

	// Verify delivery happened (start notification + result delivery).
	sender.mu.Lock()
	assert.GreaterOrEqual(t, len(sender.messages), 1)
	sender.mu.Unlock()
}

func TestExecutor_Execute_NoDeliverTo_LogsWarning(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "ok"}
	store := newMockStore()
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	job := Job{
		ID:           "job-4",
		Name:         "no-delivery-job",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "check",
		SessionMode:  "isolated",
		DeliverTo:    nil,
	}

	result := executor.Execute(context.Background(), job)

	require.NotNil(t, result)
	assert.NoError(t, result.Error)
	assert.Equal(t, "ok", result.Response)
}

func TestExecutor_Execute_SaveHistoryError(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "ok"}
	store := newMockStore()
	store.saveHistoryErr = fmt.Errorf("db write failed")
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	job := Job{
		ID:           "job-5",
		Name:         "history-fail-job",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "run",
		SessionMode:  "isolated",
	}

	// Should not panic even if history save fails.
	result := executor.Execute(context.Background(), job)
	require.NotNil(t, result)
	assert.NoError(t, result.Error)
}

func TestExecutor_Execute_MainSessionMode(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "ok"}
	store := newMockStore()
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	job := Job{
		ID:          "job-6",
		Name:        "main-session",
		Prompt:      "test",
		SessionMode: "main",
	}

	result := executor.Execute(context.Background(), job)
	require.NotNil(t, result)
	assert.Equal(t, "ok", result.Response)
}
