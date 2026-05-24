package cron

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronToolsAddUsesDefaultsTimeoutAndReportsUpdateStatus(t *testing.T) {
	store := newMockStore()
	scheduler := newTestScheduler(store, &mockAgentRunner{response: "ok"})
	tool := findCronTool(t, BuildTools(scheduler, []string{"telegram:default"}), "cron_add")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name":          "daily",
		"schedule_type": "every",
		"schedule":      "24h",
		"prompt":        "Summarize yesterday",
		"timeout":       "15m",
	})
	require.NoError(t, err)
	result := got.(map[string]interface{})
	assert.Equal(t, "created", result["status"])
	assert.Contains(t, result["message"], "every=24h")

	stored := store.jobs["daily"]
	assert.Equal(t, []string{"telegram:default"}, stored.DeliverTo)
	assert.Equal(t, 15*time.Minute, stored.Timeout)
	stored.DeliverTo[0] = "mutated"

	got, err = tool.Handler(context.Background(), map[string]interface{}{
		"name":          "daily",
		"schedule_type": "cron",
		"schedule":      "0 9 * * *",
		"prompt":        "Summarize today",
	})
	require.NoError(t, err)
	result = got.(map[string]interface{})
	assert.Equal(t, "updated", result["status"])
	assert.Equal(t, []string{"telegram:default"}, store.jobs["daily"].DeliverTo)
}

func TestCronToolsAddRejectsInvalidTimeoutBeforePersisting(t *testing.T) {
	store := newMockStore()
	scheduler := newTestScheduler(store, &mockAgentRunner{response: "ok"})
	tool := findCronTool(t, BuildTools(scheduler, nil), "cron_add")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name":          "daily",
		"schedule_type": "every",
		"schedule":      "24h",
		"prompt":        "Summarize yesterday",
		"timeout":       "soon",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Empty(t, store.jobs)
	assert.ErrorContains(t, err, `parse timeout "soon"`)
}

func TestCronToolsHistorySelectsScopedOrGlobalStoreMethod(t *testing.T) {
	store := newMockStore()
	store.history = []HistoryEntry{
		{JobID: "job-1", JobName: "daily", Status: "success"},
		{JobID: "job-2", JobName: "weekly", Status: "failed"},
	}
	scheduler := newTestScheduler(store, &mockAgentRunner{response: "ok"})
	tool := findCronTool(t, BuildTools(scheduler, nil), "cron_history")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"job_id": "job-1",
		"limit":  float64(5),
	})
	require.NoError(t, err)
	result := got.(map[string]interface{})
	assert.Equal(t, 1, result["count"])
	assert.Equal(t, "daily", result["entries"].([]HistoryEntry)[0].JobName)

	got, err = tool.Handler(context.Background(), map[string]interface{}{
		"limit": float64(1),
	})
	require.NoError(t, err)
	result = got.(map[string]interface{})
	assert.Equal(t, 1, result["count"])
	assert.Equal(t, "daily", result["entries"].([]HistoryEntry)[0].JobName)
}
