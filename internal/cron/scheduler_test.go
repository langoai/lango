package cron

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// --- local mocks ---

type mockStore struct {
	mu      sync.Mutex
	jobs    map[string]Job
	history []HistoryEntry
	// control fields
	listEnabledErr error
	createErr      error
	getByNameErr   error
	deleteErr      error
	getErr         error
	updateErr      error
	upsertErr      error
	saveHistoryErr error
}

func newMockStore() *mockStore {
	return &mockStore{
		jobs: make(map[string]Job),
	}
}

func (m *mockStore) Create(_ context.Context, job Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return m.createErr
	}
	if job.ID == "" {
		job.ID = fmt.Sprintf("mock-%d", len(m.jobs)+1)
	}
	m.jobs[job.Name] = job
	return nil
}

func (m *mockStore) Get(_ context.Context, id string) (*Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, j := range m.jobs {
		if j.ID == id {
			return &j, nil
		}
	}
	return nil, fmt.Errorf("job %q not found", id)
}

func (m *mockStore) GetByName(_ context.Context, name string) (*Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	j, ok := m.jobs[name]
	if !ok {
		return nil, fmt.Errorf("job %q not found", name)
	}
	return &j, nil
}

func (m *mockStore) List(_ context.Context) ([]Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []Job
	for _, j := range m.jobs {
		result = append(result, j)
	}
	return result, nil
}

func (m *mockStore) ListEnabled(_ context.Context) ([]Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listEnabledErr != nil {
		return nil, m.listEnabledErr
	}
	var result []Job
	for _, j := range m.jobs {
		if j.Enabled {
			result = append(result, j)
		}
	}
	return result, nil
}

func (m *mockStore) Update(_ context.Context, job Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.updateErr != nil {
		return m.updateErr
	}
	m.jobs[job.Name] = job
	return nil
}

func (m *mockStore) Upsert(_ context.Context, job Job) (*Job, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.upsertErr != nil {
		return nil, false, m.upsertErr
	}
	// Check if job with same name exists.
	if existing, ok := m.jobs[job.Name]; ok {
		job.ID = existing.ID
		m.jobs[job.Name] = job
		return &job, true, nil
	}
	if job.ID == "" {
		job.ID = fmt.Sprintf("mock-%d", len(m.jobs)+1)
	}
	m.jobs[job.Name] = job
	return &job, false, nil
}

func (m *mockStore) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for name, j := range m.jobs {
		if j.ID == id {
			delete(m.jobs, name)
			return nil
		}
	}
	return nil
}

func (m *mockStore) SaveHistory(_ context.Context, entry HistoryEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveHistoryErr != nil {
		return m.saveHistoryErr
	}
	m.history = append(m.history, entry)
	return nil
}

func (m *mockStore) ListHistory(_ context.Context, jobID string, limit int) ([]HistoryEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []HistoryEntry
	for _, h := range m.history {
		if h.JobID == jobID {
			result = append(result, h)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockStore) ListAllHistory(_ context.Context, limit int) ([]HistoryEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	end := limit
	if end > len(m.history) {
		end = len(m.history)
	}
	return m.history[:end], nil
}

type mockAgentRunner struct {
	mu       sync.Mutex
	response string
	err      error
	calls    []string
}

func (m *mockAgentRunner) Run(_ context.Context, sessionKey string, prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, prompt)
	return m.response, m.err
}

func (m *mockAgentRunner) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// --- helper ---

func newTestScheduler(store *mockStore, runner *mockAgentRunner) *Scheduler {
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)
	return New(store, executor, "UTC", 5, 30*time.Minute, logger)
}

// --- scheduler tests ---

func TestNew_DefaultMaxJobs(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	s := New(store, executor, "", 0, 0, logger)

	assert.Equal(t, 5, s.maxJobs)
	assert.Equal(t, "UTC", s.timezone)
	assert.Equal(t, 30*time.Minute, s.defaultTimeout)
}

func TestNew_CustomValues(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)

	s := New(store, executor, "America/New_York", 10, 15*time.Minute, logger)

	assert.Equal(t, 10, s.maxJobs)
	assert.Equal(t, "America/New_York", s.timezone)
	assert.Equal(t, 15*time.Minute, s.defaultTimeout)
}

func TestNew_NegativeMaxJobs(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(&mockAgentRunner{}, nil, store, logger)

	s := New(store, executor, "UTC", -3, 0, logger)

	assert.Equal(t, 5, s.maxJobs)
}

func TestScheduler_StartStop(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	s := newTestScheduler(store, runner)

	err := s.Start(context.Background())
	require.NoError(t, err)

	s.Stop()
}

func TestScheduler_StartWithJobs(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.jobs["test-job"] = Job{
		ID:           "job-1",
		Name:         "test-job",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "do something",
		Enabled:      true,
	}
	runner := &mockAgentRunner{response: "ok"}
	s := newTestScheduler(store, runner)

	err := s.Start(context.Background())
	require.NoError(t, err)

	s.mu.RLock()
	assert.Len(t, s.entries, 1)
	s.mu.RUnlock()

	s.Stop()
}

func TestScheduler_StartWithInvalidTimezone(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(&mockAgentRunner{}, nil, store, logger)

	s := New(store, executor, "Invalid/Timezone", 5, 0, logger)

	err := s.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load timezone")
}

func TestScheduler_StartWithListEnabledError(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.listEnabledErr = fmt.Errorf("db connection failed")
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(&mockAgentRunner{}, nil, store, logger)

	s := New(store, executor, "UTC", 5, 0, logger)

	err := s.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load enabled jobs")
}

func TestScheduler_StartSkipsInvalidSchedule(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.jobs["bad-job"] = Job{
		ID:           "job-bad",
		Name:         "bad-job",
		ScheduleType: "unknown_type",
		Schedule:     "???",
		Enabled:      true,
	}
	runner := &mockAgentRunner{}
	s := newTestScheduler(store, runner)

	err := s.Start(context.Background())
	require.NoError(t, err)

	s.mu.RLock()
	assert.Empty(t, s.entries)
	s.mu.RUnlock()

	s.Stop()
}

func TestScheduler_StopWithoutStart(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{}
	s := newTestScheduler(store, runner)

	// Should not panic.
	s.Stop()
}

// --- Unit 1: Idempotent AddJob ---

func TestScheduler_AddJob_CreatesNew(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	s := newTestScheduler(store, runner)

	require.NoError(t, s.Start(context.Background()))
	defer s.Stop()

	updated, err := s.AddJob(context.Background(), Job{
		Name:         "new-job",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "do stuff",
		Enabled:      true,
	})
	require.NoError(t, err)
	assert.False(t, updated)

	s.mu.RLock()
	assert.Len(t, s.entries, 1)
	s.mu.RUnlock()
}

func TestScheduler_AddJob_UpdatesExisting(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	s := newTestScheduler(store, runner)

	require.NoError(t, s.Start(context.Background()))
	defer s.Stop()

	// Create first.
	updated, err := s.AddJob(context.Background(), Job{
		Name:         "my-job",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "original prompt",
		Enabled:      true,
	})
	require.NoError(t, err)
	assert.False(t, updated)

	// Upsert with same name, different prompt.
	updated, err = s.AddJob(context.Background(), Job{
		Name:         "my-job",
		ScheduleType: "every",
		Schedule:     "30m",
		Prompt:       "updated prompt",
		Enabled:      true,
	})
	require.NoError(t, err)
	assert.True(t, updated)

	// Should still have only 1 entry.
	s.mu.RLock()
	assert.Len(t, s.entries, 1)
	s.mu.RUnlock()

	// Verify the stored job was updated.
	store.mu.Lock()
	j := store.jobs["my-job"]
	store.mu.Unlock()
	assert.Equal(t, "updated prompt", j.Prompt)
	assert.Equal(t, "30m", j.Schedule)
}

// --- Unit 2: "at" schedule fires only once ---

func TestScheduler_AtJob_FiresOnlyOnce(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "done"}
	s := newTestScheduler(store, runner)

	require.NoError(t, s.Start(context.Background()))
	defer s.Stop()

	pastTime := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	updated, err := s.AddJob(context.Background(), Job{
		Name:         "one-time",
		ScheduleType: "at",
		Schedule:     pastTime,
		Prompt:       "run once",
		Enabled:      true,
	})
	require.NoError(t, err)
	assert.False(t, updated)

	// Wait for the job to fire (past-time schedule triggers at ~1s).
	time.Sleep(3 * time.Second)

	// sync.Once ensures exactly one execution despite @every 1s trigger.
	assert.Equal(t, 1, runner.callCount())
}

// --- Unit 3: Timeout ---

func TestScheduler_ExecuteWithSemaphore_UsesDefaultTimeout(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	s := newTestScheduler(store, runner)
	// Override with short timeout for testing.
	s.defaultTimeout = 5 * time.Second

	require.NoError(t, s.Start(context.Background()))
	defer s.Stop()

	s.executeWithSemaphore(Job{
		ID:           "test-id",
		Name:         "test-job",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "test",
	})

	assert.Equal(t, 1, runner.callCount())
}

func TestScheduler_ExecuteWithSemaphore_UsesJobTimeout(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	s := newTestScheduler(store, runner)
	s.defaultTimeout = 1 * time.Hour

	require.NoError(t, s.Start(context.Background()))
	defer s.Stop()

	s.executeWithSemaphore(Job{
		ID:           "test-id",
		Name:         "test-job",
		ScheduleType: "every",
		Schedule:     "1h",
		Prompt:       "test",
		Timeout:      10 * time.Second,
	})

	assert.Equal(t, 1, runner.callCount())
}

func TestScheduler_ExecuteWithSemaphore_ShutdownAbortsAcquisition(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	runner := &mockAgentRunner{response: "ok"}
	logger := zap.NewNop().Sugar()
	executor := NewExecutor(runner, nil, store, logger)
	// Create scheduler with semaphore size 1.
	s := New(store, executor, "UTC", 1, 5*time.Second, logger)

	require.NoError(t, s.Start(context.Background()))

	// Fill the semaphore.
	s.semaphore <- struct{}{}

	var executed atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.executeWithSemaphore(Job{
			ID:     "blocked",
			Name:   "blocked-job",
			Prompt: "test",
		})
		executed.Store(true)
	}()

	// Give the goroutine time to reach the select.
	time.Sleep(50 * time.Millisecond)

	// Shutdown should unblock it via shutdownCh.
	s.Stop()
	<-done

	// The job should not have executed (runner should have 0 calls).
	assert.Equal(t, 0, runner.callCount())
}

// --- buildCronSpec tests ---

func TestBuildCronSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     Job
		wantSpec string
		wantErr  bool
	}{
		{
			give:     Job{ScheduleType: "cron", Schedule: "*/5 * * * *"},
			wantSpec: "*/5 * * * *",
		},
		{
			give:     Job{ScheduleType: "every", Schedule: "30m"},
			wantSpec: "@every 30m",
		},
		{
			give:     Job{ScheduleType: "every", Schedule: "2h"},
			wantSpec: "@every 2h",
		},
		{
			give:    Job{ScheduleType: "every", Schedule: "not-a-duration"},
			wantErr: true,
		},
		{
			give:    Job{ScheduleType: "at", Schedule: "not-a-datetime"},
			wantErr: true,
		},
		{
			give:    Job{ScheduleType: "unknown", Schedule: "anything"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s/%s", tt.give.ScheduleType, tt.give.Schedule)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			spec, err := buildCronSpec(tt.give)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSpec, spec)
		})
	}
}

func TestBuildCronSpec_AtFutureTime(t *testing.T) {
	t.Parallel()

	futureTime := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	job := Job{ScheduleType: "at", Schedule: futureTime}

	spec, err := buildCronSpec(job)
	require.NoError(t, err)
	assert.Contains(t, spec, "@every ")
}

func TestBuildCronSpec_AtPastTime(t *testing.T) {
	t.Parallel()

	pastTime := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	job := Job{ScheduleType: "at", Schedule: pastTime}

	spec, err := buildCronSpec(job)
	require.NoError(t, err)
	// Past times get scheduled for 1 second.
	assert.Equal(t, "@every 1s", spec)
}

func TestZapPrintfAdapter(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	adapter := &zapPrintfAdapter{logger: logger}

	// Should not panic.
	adapter.Printf("test message: %s %d", "hello", 42)
}

func TestBuildSessionKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        Job
		wantPrefix  string
		wantContain string
	}{
		{
			give:       Job{Name: "test-job", SessionMode: "main"},
			wantPrefix: "cron:test-job",
		},
		{
			give:        Job{Name: "test-job", SessionMode: "isolated"},
			wantContain: "cron:test-job:",
		},
		{
			give:        Job{Name: "test-job", SessionMode: ""},
			wantContain: "cron:test-job:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give.SessionMode, func(t *testing.T) {
			t.Parallel()

			key := buildSessionKey(tt.give)

			if tt.wantPrefix != "" {
				assert.Equal(t, tt.wantPrefix, key)
			}
			if tt.wantContain != "" {
				assert.Contains(t, key, tt.wantContain)
			}
		})
	}
}
