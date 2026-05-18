package cron

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/cronjob"
	"github.com/langoai/lango/internal/ent/cronjobhistory"
	"github.com/langoai/lango/internal/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

func TestEntStoreWave11_CRUDListsAndUpdateLastRunAt(t *testing.T) {
	ctx := context.Background()
	store := newWave11EntStore(t)

	lastRunAt := time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)
	nextRunAt := time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)
	jobID := uuid.NewString()
	job := Job{
		ID:           jobID,
		Name:         "wave11-crud",
		ScheduleType: "cron",
		Schedule:     "0 9 * * *",
		Prompt:       "Summarize today's news",
		SessionMode:  "main",
		DeliverTo:    []string{"slack", "telegram"},
		Timezone:     "Asia/Seoul",
		Enabled:      true,
		Timeout:      5 * time.Minute,
		LastRunAt:    &lastRunAt,
		NextRunAt:    &nextRunAt,
	}

	require.NoError(t, store.Create(ctx, job))

	got, err := store.Get(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, got.ID)
	assert.Equal(t, "wave11-crud", got.Name)
	assert.Equal(t, "cron", got.ScheduleType)
	assert.Equal(t, "0 9 * * *", got.Schedule)
	assert.Equal(t, "Summarize today's news", got.Prompt)
	assert.Equal(t, "main", got.SessionMode)
	assert.Equal(t, []string{"slack", "telegram"}, got.DeliverTo)
	assert.Equal(t, "Asia/Seoul", got.Timezone)
	assert.True(t, got.Enabled)
	assert.Equal(t, 5*time.Minute, got.Timeout)
	require.NotNil(t, got.LastRunAt)
	require.NotNil(t, got.NextRunAt)
	assert.True(t, got.LastRunAt.Equal(lastRunAt))
	assert.True(t, got.NextRunAt.Equal(nextRunAt))
	assert.False(t, got.CreatedAt.IsZero())

	byName, err := store.GetByName(ctx, "wave11-crud")
	require.NoError(t, err)
	assert.Equal(t, got.ID, byName.ID)

	disabledJob := wave11Job("wave11-disabled")
	disabledJob.Enabled = false
	require.NoError(t, store.Create(ctx, disabledJob))

	allJobs, err := store.List(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"wave11-crud", "wave11-disabled"}, wave11JobNames(allJobs))

	enabledJobs, err := store.ListEnabled(ctx)
	require.NoError(t, err)
	require.Len(t, enabledJobs, 1)
	assert.Equal(t, "wave11-crud", enabledJobs[0].Name)

	updated := Job{
		ID:           jobID,
		Name:         "wave11-crud",
		ScheduleType: "every",
		Schedule:     "30m",
		Prompt:       "Updated prompt",
		SessionMode:  "isolated",
		Timezone:     "UTC",
		Enabled:      false,
	}
	require.NoError(t, store.Update(ctx, updated))

	got, err = store.Get(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, "every", got.ScheduleType)
	assert.Equal(t, "30m", got.Schedule)
	assert.Equal(t, "Updated prompt", got.Prompt)
	assert.Equal(t, "isolated", got.SessionMode)
	assert.Empty(t, got.DeliverTo)
	assert.Equal(t, time.Duration(0), got.Timeout)
	assert.Nil(t, got.LastRunAt)
	assert.Nil(t, got.NextRunAt)
	assert.False(t, got.Enabled)

	runAt := time.Date(2026, 5, 18, 11, 30, 0, 0, time.UTC)
	require.NoError(t, store.updateLastRunAt(ctx, jobID, runAt))

	got, err = store.Get(ctx, jobID)
	require.NoError(t, err)
	require.NotNil(t, got.LastRunAt)
	assert.True(t, got.LastRunAt.Equal(runAt))

	require.NoError(t, store.Delete(ctx, jobID))
	_, err = store.Get(ctx, jobID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get cron job")
}

func TestEntStoreWave11_UpsertCreatesAndUpdatesByName(t *testing.T) {
	ctx := context.Background()
	store := newWave11EntStore(t)

	created, updated, err := store.Upsert(ctx, wave11Job("wave11-upsert"))
	require.NoError(t, err)
	assert.False(t, updated)
	require.NotNil(t, created)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "wave11-upsert", created.Name)
	assert.False(t, created.CreatedAt.IsZero())

	createdAt := created.CreatedAt
	updatedJob := Job{
		Name:         "wave11-upsert",
		ScheduleType: "at",
		Schedule:     "2026-05-19T12:00:00Z",
		Prompt:       "Run once",
		SessionMode:  "main",
		DeliverTo:    []string{"discord"},
		Timezone:     "Asia/Seoul",
		Enabled:      false,
		Timeout:      90 * time.Second,
	}

	got, updated, err := store.Upsert(ctx, updatedJob)
	require.NoError(t, err)
	assert.True(t, updated)
	require.NotNil(t, got)
	assert.Equal(t, created.ID, got.ID)
	assert.True(t, got.CreatedAt.Equal(createdAt))
	assert.Equal(t, "at", got.ScheduleType)
	assert.Equal(t, "2026-05-19T12:00:00Z", got.Schedule)
	assert.Equal(t, "Run once", got.Prompt)
	assert.Equal(t, "main", got.SessionMode)
	assert.Equal(t, []string{"discord"}, got.DeliverTo)
	assert.Equal(t, "Asia/Seoul", got.Timezone)
	assert.False(t, got.Enabled)
	assert.Equal(t, 90*time.Second, got.Timeout)

	allJobs, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, allJobs, 1)
	assert.Equal(t, created.ID, allJobs[0].ID)
	assert.Equal(t, "Run once", allJobs[0].Prompt)
}

func TestEntStoreWave11_SaveAndListHistory(t *testing.T) {
	ctx := context.Background()
	store := newWave11EntStore(t)

	jobOne, updated, err := store.Upsert(ctx, wave11Job("wave11-history-one"))
	require.NoError(t, err)
	require.False(t, updated)
	jobTwo, updated, err := store.Upsert(ctx, wave11Job("wave11-history-two"))
	require.NoError(t, err)
	require.False(t, updated)

	oldStarted := time.Date(2026, 5, 18, 8, 0, 0, 0, time.UTC)
	newStarted := time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)
	otherStarted := time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 5, 18, 9, 1, 0, 0, time.UTC)

	oldEntry := HistoryEntry{
		ID:         uuid.NewString(),
		JobID:      jobOne.ID,
		JobName:    jobOne.Name,
		Status:     "completed",
		Prompt:     "Old prompt",
		Result:     "Old result",
		TokensUsed: 11,
		StartedAt:  oldStarted,
	}
	newEntry := HistoryEntry{
		ID:           uuid.NewString(),
		JobID:        jobOne.ID,
		JobName:      jobOne.Name,
		Status:       "failed",
		Prompt:       "New prompt",
		ErrorMessage: "boom",
		TokensUsed:   22,
		StartedAt:    newStarted,
		CompletedAt:  &completedAt,
	}
	otherEntry := HistoryEntry{
		ID:         uuid.NewString(),
		JobID:      jobTwo.ID,
		JobName:    jobTwo.Name,
		Status:     "running",
		Prompt:     "Other prompt",
		TokensUsed: 33,
		StartedAt:  otherStarted,
	}

	require.NoError(t, store.SaveHistory(ctx, oldEntry))
	require.NoError(t, store.SaveHistory(ctx, newEntry))
	require.NoError(t, store.SaveHistory(ctx, otherEntry))

	history, err := store.ListHistory(ctx, jobOne.ID, 1)
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Equal(t, newEntry.ID, history[0].ID)
	assert.Equal(t, "failed", history[0].Status)
	assert.Equal(t, "boom", history[0].ErrorMessage)
	require.NotNil(t, history[0].CompletedAt)
	assert.True(t, history[0].CompletedAt.Equal(completedAt))

	history, err = store.ListHistory(ctx, jobOne.ID, 10)
	require.NoError(t, err)
	require.Len(t, history, 2)
	assert.Equal(t, []string{newEntry.ID, oldEntry.ID}, []string{history[0].ID, history[1].ID})
	assert.Equal(t, "Old result", history[1].Result)
	assert.Equal(t, 11, history[1].TokensUsed)

	allHistory, err := store.ListAllHistory(ctx, 2)
	require.NoError(t, err)
	require.Len(t, allHistory, 2)
	assert.Equal(t, []string{otherEntry.ID, newEntry.ID}, []string{allHistory[0].ID, allHistory[1].ID})
}

func TestEntStoreWave11_ConversionHelpers(t *testing.T) {
	timeoutMs := int64(120000)
	lastRunAt := time.Date(2026, 5, 18, 6, 0, 0, 0, time.UTC)
	nextRunAt := time.Date(2026, 5, 18, 7, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 18, 5, 0, 0, 0, time.UTC)
	entJob := &ent.CronJob{
		ID:           uuid.New(),
		Name:         "wave11-convert",
		ScheduleType: cronjob.ScheduleTypeCron,
		Schedule:     "0 6 * * *",
		Prompt:       "Convert job",
		SessionMode:  "isolated",
		DeliverTo:    []string{"slack", "telegram"},
		Timezone:     "UTC",
		Enabled:      true,
		TimeoutMs:    &timeoutMs,
		LastRunAt:    &lastRunAt,
		NextRunAt:    &nextRunAt,
		CreatedAt:    createdAt,
	}

	job := entCronJobToDomain(entJob)
	entJob.DeliverTo[0] = "mutated"

	assert.Equal(t, entJob.ID.String(), job.ID)
	assert.Equal(t, "wave11-convert", job.Name)
	assert.Equal(t, "cron", job.ScheduleType)
	assert.Equal(t, "0 6 * * *", job.Schedule)
	assert.Equal(t, "Convert job", job.Prompt)
	assert.Equal(t, "isolated", job.SessionMode)
	assert.Equal(t, []string{"slack", "telegram"}, job.DeliverTo)
	assert.Equal(t, "UTC", job.Timezone)
	assert.True(t, job.Enabled)
	assert.Equal(t, 2*time.Minute, job.Timeout)
	require.NotNil(t, job.LastRunAt)
	require.NotNil(t, job.NextRunAt)
	assert.True(t, job.LastRunAt.Equal(lastRunAt))
	assert.True(t, job.NextRunAt.Equal(nextRunAt))
	assert.True(t, job.CreatedAt.Equal(createdAt))

	jobs := entCronJobsToDomain([]*ent.CronJob{entJob})
	require.Len(t, jobs, 1)
	assert.Equal(t, "wave11-convert", jobs[0].Name)

	completedAt := time.Date(2026, 5, 18, 6, 2, 0, 0, time.UTC)
	entHistory := &ent.CronJobHistory{
		ID:           uuid.New(),
		JobID:        entJob.ID,
		JobName:      "wave11-convert",
		Status:       cronjobhistory.StatusCompleted,
		Prompt:       "Convert history",
		Result:       "done",
		ErrorMessage: "",
		TokensUsed:   42,
		StartedAt:    lastRunAt,
		CompletedAt:  &completedAt,
	}

	history := entHistoryToDomain(entHistory)
	assert.Equal(t, entHistory.ID.String(), history.ID)
	assert.Equal(t, entJob.ID.String(), history.JobID)
	assert.Equal(t, "wave11-convert", history.JobName)
	assert.Equal(t, "completed", history.Status)
	assert.Equal(t, "Convert history", history.Prompt)
	assert.Equal(t, "done", history.Result)
	assert.Equal(t, 42, history.TokensUsed)
	assert.True(t, history.StartedAt.Equal(lastRunAt))
	require.NotNil(t, history.CompletedAt)
	assert.True(t, history.CompletedAt.Equal(completedAt))

	histories := entHistoriesToDomain([]*ent.CronJobHistory{entHistory})
	require.Len(t, histories, 1)
	assert.Equal(t, history.ID, histories[0].ID)
}

func TestEntStoreWave11_InvalidUUIDErrors(t *testing.T) {
	ctx := context.Background()
	store := newWave11EntStore(t)

	err := store.Create(ctx, Job{
		ID:           "not-a-uuid",
		Name:         "wave11-invalid-create",
		ScheduleType: "cron",
		Schedule:     "0 9 * * *",
		Prompt:       "Invalid create",
		SessionMode:  "isolated",
		Timezone:     "UTC",
		Enabled:      true,
	})
	require.ErrorContains(t, err, `parse job id "not-a-uuid"`)

	_, err = store.Get(ctx, "not-a-uuid")
	require.ErrorContains(t, err, `parse job id "not-a-uuid"`)

	err = store.Update(ctx, Job{ID: "not-a-uuid"})
	require.ErrorContains(t, err, `parse job id "not-a-uuid"`)

	err = store.Delete(ctx, "not-a-uuid")
	require.ErrorContains(t, err, `parse job id "not-a-uuid"`)

	err = store.updateLastRunAt(ctx, "not-a-uuid", time.Now())
	require.ErrorContains(t, err, `parse job id "not-a-uuid"`)

	err = store.SaveHistory(ctx, HistoryEntry{JobID: "not-a-uuid"})
	require.ErrorContains(t, err, `parse history job id "not-a-uuid"`)

	err = store.SaveHistory(ctx, HistoryEntry{
		ID:        "not-a-uuid",
		JobID:     uuid.NewString(),
		JobName:   "wave11-invalid-history",
		Status:    "running",
		Prompt:    "Invalid history",
		StartedAt: time.Now(),
	})
	require.ErrorContains(t, err, `parse history id "not-a-uuid"`)

	_, err = store.ListHistory(ctx, "not-a-uuid", 10)
	require.ErrorContains(t, err, `parse history job id "not-a-uuid"`)
}

func newWave11EntStore(t *testing.T) *EntStore {
	t.Helper()

	client := enttest.Open(
		t,
		"sqlite3",
		fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", uuid.NewString()),
	)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return NewEntStore(client)
}

func wave11Job(name string) Job {
	return Job{
		Name:         name,
		ScheduleType: "cron",
		Schedule:     "0 9 * * *",
		Prompt:       "Summarize today's news",
		SessionMode:  "isolated",
		Timezone:     "UTC",
		Enabled:      true,
	}
}

func wave11JobNames(jobs []Job) []string {
	names := make([]string, 0, len(jobs))
	for _, job := range jobs {
		names = append(names, job.Name)
	}
	return names
}
