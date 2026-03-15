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

func TestExecutor_Execute_InjectsHistoryContext(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "new response"}
	store := newMockStore()
	// Pre-populate history.
	store.history = []HistoryEntry{
		{JobID: "job-h1", Result: "previous output 1"},
		{JobID: "job-h1", Result: "previous output 2"},
	}
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	job := Job{
		ID:          "job-h1",
		Name:        "history-job",
		Prompt:      "give me a bible verse",
		SessionMode: "isolated",
	}

	result := executor.Execute(context.Background(), job)
	require.NotNil(t, result)
	assert.NoError(t, result.Error)

	// The runner should have received an enriched prompt containing history.
	runner.mu.Lock()
	require.Len(t, runner.calls, 1)
	prompt := runner.calls[0]
	runner.mu.Unlock()

	assert.Contains(t, prompt, "Previous outputs")
	assert.Contains(t, prompt, "previous output 1")
	assert.Contains(t, prompt, "previous output 2")
	assert.Contains(t, prompt, "give me a bible verse")

	// History should be saved with the original prompt, not the enriched one.
	store.mu.Lock()
	require.Len(t, store.history, 3) // 2 pre-existing + 1 new
	assert.Equal(t, "give me a bible verse", store.history[2].Prompt)
	store.mu.Unlock()
}

func TestExecutor_Execute_NoHistory_OriginalPrompt(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "ok"}
	store := newMockStore()
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	job := Job{
		ID:          "job-noh",
		Name:        "no-history-job",
		Prompt:      "original prompt only",
		SessionMode: "isolated",
	}

	executor.Execute(context.Background(), job)

	runner.mu.Lock()
	require.Len(t, runner.calls, 1)
	assert.Equal(t, "original prompt only", runner.calls[0])
	runner.mu.Unlock()
}

func TestExecutor_Execute_HistoryQueryError_Graceful(t *testing.T) {
	t.Parallel()

	runner := &mockAgentRunner{response: "ok"}
	store := newMockStore()
	store.listHistoryErr = fmt.Errorf("db read failed")
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	job := Job{
		ID:          "job-herr",
		Name:        "history-error-job",
		Prompt:      "fallback prompt",
		SessionMode: "isolated",
	}

	result := executor.Execute(context.Background(), job)
	require.NotNil(t, result)
	assert.NoError(t, result.Error)

	// Should fall back to original prompt.
	runner.mu.Lock()
	require.Len(t, runner.calls, 1)
	assert.Equal(t, "fallback prompt", runner.calls[0])
	runner.mu.Unlock()
}
