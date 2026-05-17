package cron

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	croncore "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
)

const testCronJobID = "11111111-1111-1111-1111-111111111111"

type recordingCronStore struct {
	created          []croncore.Job
	updated          []croncore.Job
	deleted          []string
	listHistoryJobID string
	jobsByName       map[string]croncore.Job
	history          []croncore.HistoryEntry
}

func (s *recordingCronStore) Create(_ context.Context, job croncore.Job) error {
	if job.ID == "" {
		job.ID = testCronJobID
	}
	s.created = append(s.created, job)
	if s.jobsByName == nil {
		s.jobsByName = make(map[string]croncore.Job)
	}
	s.jobsByName[job.Name] = job
	return nil
}

func (s *recordingCronStore) Get(_ context.Context, id string) (*croncore.Job, error) {
	for _, job := range s.jobsByName {
		if job.ID == id {
			copy := job
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("job %q not found", id)
}

func (s *recordingCronStore) GetByName(_ context.Context, name string) (*croncore.Job, error) {
	job, ok := s.jobsByName[name]
	if !ok {
		return nil, fmt.Errorf("job %q not found", name)
	}
	return &job, nil
}

func (s *recordingCronStore) List(context.Context) ([]croncore.Job, error) {
	jobs := make([]croncore.Job, 0, len(s.jobsByName))
	for _, job := range s.jobsByName {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s *recordingCronStore) ListEnabled(ctx context.Context) ([]croncore.Job, error) {
	jobs, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	enabled := make([]croncore.Job, 0, len(jobs))
	for _, job := range jobs {
		if job.Enabled {
			enabled = append(enabled, job)
		}
	}
	return enabled, nil
}

func (s *recordingCronStore) Update(_ context.Context, job croncore.Job) error {
	s.updated = append(s.updated, job)
	if s.jobsByName == nil {
		s.jobsByName = make(map[string]croncore.Job)
	}
	s.jobsByName[job.Name] = job
	return nil
}

func (s *recordingCronStore) Upsert(
	ctx context.Context,
	job croncore.Job,
) (*croncore.Job, bool, error) {
	if existing, ok := s.jobsByName[job.Name]; ok {
		job.ID = existing.ID
		return &job, true, s.Update(ctx, job)
	}
	if err := s.Create(ctx, job); err != nil {
		return nil, false, err
	}
	created := s.created[len(s.created)-1]
	return &created, false, nil
}

func (s *recordingCronStore) Delete(_ context.Context, id string) error {
	s.deleted = append(s.deleted, id)
	return nil
}

func (s *recordingCronStore) SaveHistory(context.Context, croncore.HistoryEntry) error {
	return nil
}

func (s *recordingCronStore) ListHistory(
	_ context.Context,
	jobID string,
	_ int,
) ([]croncore.HistoryEntry, error) {
	s.listHistoryJobID = jobID
	return s.history, nil
}

func (s *recordingCronStore) ListAllHistory(
	context.Context,
	int,
) ([]croncore.HistoryEntry, error) {
	return s.history, nil
}

func bootLoaderWithCronStore(store croncore.Store) func() (*bootstrap.Result, error) {
	return bootLoaderWithCronStoreConfig(store, config.DefaultConfig())
}

func bootLoaderWithCronStoreConfig(
	store croncore.Store,
	cfg *config.Config,
) func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config: cfg,
			Storage: storage.NewFacade(
				nil,
				nil,
				storage.WithCronFactory(func() croncore.Store {
					return store
				}),
			),
		}, nil
	}
}

func executeCronCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestNewCronCmd_Structure(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	require.NotNil(t, cmd)
	assert.Equal(t, "cron", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewCronCmd_Subcommands(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	expected := []string{"add", "list", "delete", "pause", "resume", "history"}
	subCmds := make(map[string]bool, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestListCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	out, err := executeCronCommand(t, cmd, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "No cron jobs found.")
}

func TestListCmd_BootError(t *testing.T) {
	cmd := NewCronCmd(testutil.FailBootLoader(assert.AnError))

	_, err := executeCronCommand(t, cmd, "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap")
}

func TestHistoryCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	out, err := executeCronCommand(t, cmd, "history")
	require.NoError(t, err)
	assert.Contains(t, out, "No execution history found.")
}

func TestHistoryCmd_BootError(t *testing.T) {
	cmd := NewCronCmd(testutil.FailBootLoader(assert.AnError))

	_, err := executeCronCommand(t, cmd, "history")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap")
}

func TestAddCmd_MissingPrompt(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "add", "--name", "test", "--schedule", "0 9 * * *")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--prompt is required")
}

func TestAddCmd_MissingName(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "add", "--prompt", "do something", "--schedule", "0 9 * * *")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

func TestAddCmd_MissingSchedule(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "add", "--name", "test", "--prompt", "do something")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "one of --schedule, --every, or --at is required")
}

func TestAddCmd_MultipleSchedules(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "add",
		"--name", "test",
		"--prompt", "do something",
		"--schedule", "0 9 * * *",
		"--every", "1h",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one of --schedule, --every, or --at")
}

func TestAddCmd_HappyPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	out, err := executeCronCommand(t, cmd, "add",
		"--name", "test-job",
		"--prompt", "hello world",
		"--schedule", "0 9 * * *",
	)
	require.NoError(t, err)
	assert.Contains(t, out, `Cron job "test-job" created`)
	assert.Contains(t, out, "cron 0 9 * * *")
}

func TestAddCmd_DocumentedDeliverToAndTimeoutFlags(t *testing.T) {
	store := &recordingCronStore{}
	cmd := NewCronCmd(bootLoaderWithCronStore(store))

	out, err := executeCronCommand(t, cmd, "add",
		"--name", "daily-summary",
		"--prompt", "summarize",
		"--schedule", "0 9 * * *",
		"--deliver-to", "telegram",
		"--timeout", "5m",
	)

	require.NoError(t, err)
	require.Len(t, store.created, 1)
	assert.Equal(t, []string{"telegram"}, store.created[0].DeliverTo)
	assert.Equal(t, 5*time.Minute, store.created[0].Timeout)
	assert.Contains(t, out, `Cron job "daily-summary" created`)
}

func TestAddCmd_DefaultsToIsolatedSessionMode(t *testing.T) {
	store := &recordingCronStore{}
	cmd := NewCronCmd(bootLoaderWithCronStore(store))

	_, err := executeCronCommand(t, cmd, "add",
		"--name", "isolated-by-default",
		"--prompt", "summarize",
		"--schedule", "0 9 * * *",
	)

	require.NoError(t, err)
	require.Len(t, store.created, 1)
	assert.Equal(t, "isolated", store.created[0].SessionMode)
}

func TestAddCmd_UsesConfiguredDefaultSessionMode(t *testing.T) {
	store := &recordingCronStore{}
	cfg := config.DefaultConfig()
	cfg.Cron.DefaultSessionMode = "main"
	cmd := NewCronCmd(bootLoaderWithCronStoreConfig(store, cfg))

	_, err := executeCronCommand(t, cmd, "add",
		"--name", "configured-main",
		"--prompt", "summarize",
		"--schedule", "0 9 * * *",
	)

	require.NoError(t, err)
	require.Len(t, store.created, 1)
	assert.Equal(t, "main", store.created[0].SessionMode)
}

func TestAddCmd_IsolatedFalseUsesMainSessionMode(t *testing.T) {
	store := &recordingCronStore{}
	cmd := NewCronCmd(bootLoaderWithCronStore(store))

	_, err := executeCronCommand(t, cmd, "add",
		"--name", "shared-session",
		"--prompt", "summarize",
		"--schedule", "0 9 * * *",
		"--isolated=false",
	)

	require.NoError(t, err)
	require.Len(t, store.created, 1)
	assert.Equal(t, "main", store.created[0].SessionMode)
}

func TestAddCmd_UpdatesExistingJobByName(t *testing.T) {
	existing := croncore.Job{
		ID:           testCronJobID,
		Name:         "daily-summary",
		ScheduleType: "cron",
		Schedule:     "0 8 * * *",
		Prompt:       "old prompt",
		SessionMode:  "isolated",
		Enabled:      true,
	}
	store := &recordingCronStore{
		jobsByName: map[string]croncore.Job{
			existing.Name: existing,
		},
	}
	cmd := NewCronCmd(bootLoaderWithCronStore(store))

	out, err := executeCronCommand(t, cmd, "add",
		"--name", "daily-summary",
		"--prompt", "new prompt",
		"--schedule", "0 9 * * *",
	)

	require.NoError(t, err)
	assert.Empty(t, store.created)
	require.Len(t, store.updated, 1)
	assert.Equal(t, testCronJobID, store.updated[0].ID)
	assert.Equal(t, "0 9 * * *", store.updated[0].Schedule)
	assert.Contains(t, out, `Cron job "daily-summary" updated`)
}

func TestAddCmd_InvalidTimeoutDoesNotCreateJob(t *testing.T) {
	store := &recordingCronStore{}
	cmd := NewCronCmd(bootLoaderWithCronStore(store))

	_, err := executeCronCommand(t, cmd, "add",
		"--name", "bad-timeout",
		"--prompt", "summarize",
		"--schedule", "0 9 * * *",
		"--timeout", "not-a-duration",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `parse --timeout "not-a-duration"`)
	assert.Empty(t, store.created)
}

func TestAddCmd_WithEvery(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	out, err := executeCronCommand(t, cmd, "add",
		"--name", "interval-job",
		"--prompt", "check status",
		"--every", "30m",
	)
	require.NoError(t, err)
	assert.Contains(t, out, `Cron job "interval-job" created`)
	assert.Contains(t, out, "every 30m")
}

func TestPauseResumeDeleteCmds_AcceptIDFlag(t *testing.T) {
	job := croncore.Job{
		ID:           testCronJobID,
		Name:         "daily-summary",
		ScheduleType: "cron",
		Schedule:     "0 9 * * *",
		Prompt:       "summarize",
		SessionMode:  "isolated",
		Enabled:      true,
	}

	pauseStore := &recordingCronStore{jobsByName: map[string]croncore.Job{job.Name: job}}
	_, err := executeCronCommand(
		t,
		NewCronCmd(bootLoaderWithCronStore(pauseStore)),
		"pause",
		"--id",
		"daily-summary",
	)
	require.NoError(t, err)
	require.Len(t, pauseStore.updated, 1)
	assert.False(t, pauseStore.updated[0].Enabled)

	resumeJob := job
	resumeJob.Enabled = false
	resumeStore := &recordingCronStore{jobsByName: map[string]croncore.Job{job.Name: resumeJob}}
	_, err = executeCronCommand(
		t,
		NewCronCmd(bootLoaderWithCronStore(resumeStore)),
		"resume",
		"--id",
		"daily-summary",
	)
	require.NoError(t, err)
	require.Len(t, resumeStore.updated, 1)
	assert.True(t, resumeStore.updated[0].Enabled)

	deleteStore := &recordingCronStore{jobsByName: map[string]croncore.Job{job.Name: job}}
	_, err = executeCronCommand(
		t,
		NewCronCmd(bootLoaderWithCronStore(deleteStore)),
		"delete",
		"--id",
		"daily-summary",
	)
	require.NoError(t, err)
	assert.Equal(t, []string{testCronJobID}, deleteStore.deleted)
}

func TestPauseResumeDeleteCmds_KeepPositionalSelectors(t *testing.T) {
	job := croncore.Job{
		ID:           testCronJobID,
		Name:         "daily-summary",
		ScheduleType: "cron",
		Schedule:     "0 9 * * *",
		Prompt:       "summarize",
		SessionMode:  "isolated",
		Enabled:      true,
	}

	pauseStore := &recordingCronStore{jobsByName: map[string]croncore.Job{job.Name: job}}
	_, err := executeCronCommand(
		t,
		NewCronCmd(bootLoaderWithCronStore(pauseStore)),
		"pause",
		"daily-summary",
	)
	require.NoError(t, err)
	require.Len(t, pauseStore.updated, 1)
	assert.False(t, pauseStore.updated[0].Enabled)

	resumeJob := job
	resumeJob.Enabled = false
	resumeStore := &recordingCronStore{jobsByName: map[string]croncore.Job{job.Name: resumeJob}}
	_, err = executeCronCommand(
		t,
		NewCronCmd(bootLoaderWithCronStore(resumeStore)),
		"resume",
		"daily-summary",
	)
	require.NoError(t, err)
	require.Len(t, resumeStore.updated, 1)
	assert.True(t, resumeStore.updated[0].Enabled)

	deleteStore := &recordingCronStore{jobsByName: map[string]croncore.Job{job.Name: job}}
	_, err = executeCronCommand(
		t,
		NewCronCmd(bootLoaderWithCronStore(deleteStore)),
		"delete",
		"daily-summary",
	)
	require.NoError(t, err)
	assert.Equal(t, []string{testCronJobID}, deleteStore.deleted)
}

func TestHistoryCmd_AcceptsIDFlag(t *testing.T) {
	job := croncore.Job{
		ID:           testCronJobID,
		Name:         "daily-summary",
		ScheduleType: "cron",
		Schedule:     "0 9 * * *",
		Prompt:       "summarize",
		SessionMode:  "isolated",
		Enabled:      true,
	}
	store := &recordingCronStore{jobsByName: map[string]croncore.Job{job.Name: job}}
	cmd := NewCronCmd(bootLoaderWithCronStore(store))

	_, err := executeCronCommand(t, cmd, "history", "--id", "daily-summary")

	require.NoError(t, err)
	assert.Equal(t, testCronJobID, store.listHistoryJobID)
}

func TestControlCmd_RejectsAmbiguousSelector(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "pause", "daily-summary", "--id", "other")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provide either <id-or-name> or --id, not both")
}

func TestDeleteCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "delete")
	require.Error(t, err)
}

func TestPauseCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "pause")
	require.Error(t, err)
}

func TestResumeCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "resume")
	require.Error(t, err)
}

func TestShortID(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "abcdefgh-1234-5678", want: "abcdefgh"},
		{give: "short", want: "short"},
		{give: "12345678", want: "12345678"},
		{give: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, shortID(tt.give))
		})
	}
}
