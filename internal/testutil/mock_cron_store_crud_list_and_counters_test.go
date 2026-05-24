package testutil_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/testutil"
)

func TestMockCronStoreCRUDListAndCounters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := testutil.NewMockCronStore()
	createdAt := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	first := cron.Job{
		ID:           "job-1",
		Name:         "daily-review",
		ScheduleType: "cron",
		Schedule:     "0 9 * * *",
		Prompt:       "review",
		Enabled:      true,
		CreatedAt:    createdAt,
	}
	second := cron.Job{
		ID:           "job-2",
		Name:         "paused-report",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "report",
		Enabled:      false,
		CreatedAt:    createdAt.Add(time.Hour),
	}

	require.NoError(t, store.Create(ctx, first))
	require.NoError(t, store.Create(ctx, second))

	gotFirst, err := store.Get(ctx, "job-1")
	require.NoError(t, err)
	assert.Equal(t, first, *gotFirst)

	gotByName, err := store.GetByName(ctx, "daily-review")
	require.NoError(t, err)
	assert.Equal(t, first, *gotByName)

	allJobs, err := store.List(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []cron.Job{first, second}, allJobs)

	enabledJobs, err := store.ListEnabled(ctx)
	require.NoError(t, err)
	assert.Equal(t, []cron.Job{first}, enabledJobs)

	second.Enabled = true
	second.Prompt = "updated report"
	require.NoError(t, store.Update(ctx, second))
	gotSecond, err := store.Get(ctx, "job-2")
	require.NoError(t, err)
	assert.Equal(t, second, *gotSecond)

	require.NoError(t, store.Delete(ctx, "job-1"))
	_, err = store.Get(ctx, "job-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `job "job-1" not found`)

	assert.Equal(t, 2, store.CreateCalls())
	assert.Equal(t, 1, store.JobCount())
}

func TestMockCronStoreUpsertCreatesAndUpdatesByName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := testutil.NewMockCronStore()

	created, updatedExisting, err := store.Upsert(ctx, cron.Job{
		ID:           "generated-id",
		Name:         "nightly",
		ScheduleType: "every",
		Schedule:     "24h",
		Enabled:      true,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.False(t, updatedExisting)
	assert.Equal(t, "generated-id", created.ID)
	assert.Equal(t, 1, store.CreateCalls())

	updated, updatedExisting, err := store.Upsert(ctx, cron.Job{
		ID:           "ignored-new-id",
		Name:         "nightly",
		ScheduleType: "every",
		Schedule:     "12h",
		Enabled:      false,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.True(t, updatedExisting)
	assert.Equal(t, "generated-id", updated.ID)
	assert.Equal(t, "12h", updated.Schedule)
	assert.Equal(t, 1, store.CreateCalls(), "updating by name must not count as a create")

	got, err := store.Get(ctx, "generated-id")
	require.NoError(t, err)
	assert.Equal(t, *updated, *got)
}

func TestMockCronStoreHistoryFiltersAndLimits(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := testutil.NewMockCronStore()
	startedAt := time.Date(2026, 5, 19, 13, 0, 0, 0, time.UTC)
	entries := []cron.HistoryEntry{
		{ID: "h1", JobID: "job-1", JobName: "one", Status: "completed", StartedAt: startedAt},
		{ID: "h2", JobID: "job-2", JobName: "two", Status: "failed", StartedAt: startedAt.Add(time.Minute)},
		{ID: "h3", JobID: "job-1", JobName: "one", Status: "running", StartedAt: startedAt.Add(2 * time.Minute)},
	}
	for _, entry := range entries {
		require.NoError(t, store.SaveHistory(ctx, entry))
	}

	jobHistory, err := store.ListHistory(ctx, "job-1", 1)
	require.NoError(t, err)
	assert.Equal(t, []cron.HistoryEntry{entries[0]}, jobHistory)

	allJobHistory, err := store.ListHistory(ctx, "job-1", 0)
	require.NoError(t, err)
	assert.Equal(t, []cron.HistoryEntry{entries[0], entries[2]}, allJobHistory)

	allHistory, err := store.ListAllHistory(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, entries[:2], allHistory)

	allHistory[0].Status = "mutated"
	allHistoryAgain, err := store.ListAllHistory(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, "completed", allHistoryAgain[0].Status)
	assert.Equal(t, 3, store.HistoryCount())
}

func TestMockCronStoreErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := testutil.NewMockCronStore()
	require.NoError(t, store.Create(ctx, cron.Job{ID: "job-1", Name: "existing"}))

	createErr := errors.New("create")
	getErr := errors.New("get")
	listErr := errors.New("list")
	updateErr := errors.New("update")
	deleteErr := errors.New("delete")
	historyErr := errors.New("history")

	store.CreateErr = createErr
	store.GetErr = getErr
	store.ListErr = listErr
	store.UpdateErr = updateErr
	store.DeleteErr = deleteErr
	store.SaveHistoryErr = historyErr

	require.ErrorIs(t, store.Create(ctx, cron.Job{ID: "job-2"}), createErr)
	_, err := store.Get(ctx, "job-1")
	require.ErrorIs(t, err, getErr)
	_, err = store.GetByName(ctx, "existing")
	require.ErrorIs(t, err, getErr)
	_, err = store.List(ctx)
	require.ErrorIs(t, err, listErr)
	_, err = store.ListEnabled(ctx)
	require.ErrorIs(t, err, listErr)
	require.ErrorIs(t, store.Update(ctx, cron.Job{ID: "job-1"}), updateErr)
	require.ErrorIs(t, store.Delete(ctx, "job-1"), deleteErr)
	require.ErrorIs(t, store.SaveHistory(ctx, cron.HistoryEntry{ID: "h1"}), historyErr)

	_, updatedExisting, err := store.Upsert(ctx, cron.Job{ID: "job-2", Name: "new"})
	require.ErrorIs(t, err, createErr)
	assert.False(t, updatedExisting)

	store.CreateErr = nil
	_, updatedExisting, err = store.Upsert(ctx, cron.Job{ID: "job-1", Name: "existing"})
	require.ErrorIs(t, err, updateErr)
	assert.False(t, updatedExisting)
}
