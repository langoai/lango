package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	robfigcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler manages cron job registration, lifecycle, and concurrent execution.
type Scheduler struct {
	cron           *robfigcron.Cron
	store          Store
	executor       *Executor
	mu             sync.RWMutex
	entries        map[string]robfigcron.EntryID // jobID -> cron entry
	semaphore      chan struct{}                 // limits concurrent job execution
	maxJobs        int
	defaultTimeout time.Duration
	timezone       string
	shutdownCh     chan struct{}
	stopOnce       sync.Once
	logger         *zap.SugaredLogger
}

// New creates a new Scheduler.
func New(store Store, executor *Executor, timezone string, maxJobs int, defaultTimeout time.Duration, logger *zap.SugaredLogger) *Scheduler {
	if maxJobs <= 0 {
		maxJobs = 5
	}
	if timezone == "" {
		timezone = "UTC"
	}
	if defaultTimeout <= 0 {
		defaultTimeout = 30 * time.Minute
	}

	return &Scheduler{
		store:          store,
		executor:       executor,
		entries:        make(map[string]robfigcron.EntryID),
		semaphore:      make(chan struct{}, maxJobs),
		maxJobs:        maxJobs,
		defaultTimeout: defaultTimeout,
		timezone:       timezone,
		shutdownCh:     make(chan struct{}),
		logger:         logger,
	}
}

// Start loads all enabled jobs from the database, registers them with the cron
// scheduler, and starts the scheduler. The provided context is used for the
// initial load; the scheduler itself runs until Stop is called.
func (s *Scheduler) Start(ctx context.Context) error {
	loc, err := time.LoadLocation(s.timezone)
	if err != nil {
		return fmt.Errorf("load timezone %q: %w", s.timezone, err)
	}

	s.cron = robfigcron.New(
		robfigcron.WithLocation(loc),
		robfigcron.WithLogger(robfigcron.PrintfLogger(&zapPrintfAdapter{s.logger})),
	)

	jobs, err := s.store.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("load enabled jobs: %w", err)
	}

	for _, job := range jobs {
		if err := s.registerJob(job); err != nil {
			s.logger.Warnw("skip job registration",
				"job", job.Name,
				"error", err,
			)
			continue
		}
	}

	s.cron.Start()
	s.logger.Infow("cron scheduler started",
		"timezone", s.timezone,
		"jobs_loaded", len(jobs),
		"max_concurrent", s.maxJobs,
	)

	return nil
}

// Stop gracefully shuts down the scheduler and waits for running jobs to drain.
// It is safe to call Stop multiple times.
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		if s.cron == nil {
			return
		}

		// Signal shutdown to unblock context-aware semaphore acquisition.
		close(s.shutdownCh)

		ctx := s.cron.Stop()
		<-ctx.Done()

		s.mu.Lock()
		s.entries = make(map[string]robfigcron.EntryID)
		s.mu.Unlock()

		s.logger.Info("cron scheduler stopped")
	})
}

// AddJob creates or updates a job in the store and registers it with the scheduler.
// Returns true if an existing job was updated, false if a new one was created.
func (s *Scheduler) AddJob(ctx context.Context, job Job) (bool, error) {
	stored, updated, err := s.store.Upsert(ctx, job)
	if err != nil {
		return false, fmt.Errorf("upsert cron job: %w", err)
	}

	// If updating, unregister the old cron entry before re-registering.
	if updated {
		s.unregisterJob(stored.ID)
	}

	if stored.Enabled && s.cron != nil {
		if err := s.registerJob(*stored); err != nil {
			return false, fmt.Errorf("register cron job %q: %w", job.Name, err)
		}
	}

	action := "added"
	if updated {
		action = "updated"
	}
	s.logger.Infow("cron job "+action,
		"job", stored.Name,
		"id", stored.ID,
		"schedule_type", stored.ScheduleType,
		"schedule", stored.Schedule,
	)

	return updated, nil
}

// RemoveJob removes a job from the scheduler and deletes it from the store.
func (s *Scheduler) RemoveJob(ctx context.Context, id string) error {
	s.unregisterJob(id)

	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete cron job %q: %w", id, err)
	}

	s.logger.Infow("cron job removed", "id", id)
	return nil
}

// PauseJob disables a job so it no longer fires.
func (s *Scheduler) PauseJob(ctx context.Context, id string) error {
	job, err := s.store.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get cron job %q: %w", id, err)
	}

	s.unregisterJob(id)

	job.Enabled = false
	if err := s.store.Update(ctx, *job); err != nil {
		return fmt.Errorf("disable cron job %q: %w", id, err)
	}

	s.logger.Infow("cron job paused", "job", job.Name, "id", id)
	return nil
}

// ResumeJob re-enables a paused job and registers it with the scheduler.
func (s *Scheduler) ResumeJob(ctx context.Context, id string) error {
	job, err := s.store.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get cron job %q: %w", id, err)
	}

	job.Enabled = true
	if err := s.store.Update(ctx, *job); err != nil {
		return fmt.Errorf("enable cron job %q: %w", id, err)
	}

	if s.cron != nil {
		if err := s.registerJob(*job); err != nil {
			return fmt.Errorf("register cron job %q: %w", id, err)
		}
	}

	s.logger.Infow("cron job resumed", "job", job.Name, "id", id)
	return nil
}

// ListJobs returns all cron jobs from the store.
func (s *Scheduler) ListJobs(ctx context.Context) ([]Job, error) {
	return s.store.List(ctx)
}

// History returns execution history for a specific job.
func (s *Scheduler) History(ctx context.Context, jobID string, limit int) ([]HistoryEntry, error) {
	return s.store.ListHistory(ctx, jobID, limit)
}

// AllHistory returns execution history across all jobs.
func (s *Scheduler) AllHistory(ctx context.Context, limit int) ([]HistoryEntry, error) {
	return s.store.ListAllHistory(ctx, limit)
}

// registerJob adds a job to the internal cron scheduler based on its schedule type.
func (s *Scheduler) registerJob(job Job) error {
	spec, err := buildCronSpec(job)
	if err != nil {
		return err
	}

	// Capture job by value for the closure.
	j := job

	// For "at" (one-time) jobs, wrap with sync.Once to guarantee single execution
	// regardless of how the cron library fires the trigger.
	var addFunc func()
	if j.ScheduleType == "at" {
		var once sync.Once
		addFunc = func() {
			once.Do(func() {
				s.unregisterJob(j.ID)
				s.executeWithSemaphore(j)
			})
		}
	} else {
		addFunc = func() {
			s.executeWithSemaphore(j)
		}
	}

	entryID, err := s.cron.AddFunc(spec, addFunc)
	if err != nil {
		return fmt.Errorf("add cron entry for job %q: %w", job.Name, err)
	}

	s.mu.Lock()
	s.entries[job.ID] = entryID
	s.mu.Unlock()

	return nil
}

// unregisterJob removes a job from the internal cron scheduler.
func (s *Scheduler) unregisterJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entries[id]; ok {
		if s.cron != nil {
			s.cron.Remove(entryID)
		}
		delete(s.entries, id)
	}
}

// executeWithSemaphore runs a job while respecting the concurrency limit.
func (s *Scheduler) executeWithSemaphore(job Job) {
	// Context-aware semaphore acquisition: abort if shutting down.
	select {
	case s.semaphore <- struct{}{}:
	case <-s.shutdownCh:
		return
	}
	defer func() { <-s.semaphore }()

	// Determine per-job or default timeout.
	timeout := s.defaultTimeout
	if job.Timeout > 0 {
		timeout = job.Timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s.executor.Execute(ctx, job)

	// For "at" (one-time) jobs, disable after execution.
	if job.ScheduleType == "at" {
		s.disableOneTimeJob(ctx, job)
	}
}

// disableOneTimeJob disables a one-time ("at") job after it has fired.
// Note: unregisterJob is already called by the sync.Once wrapper in registerJob.
func (s *Scheduler) disableOneTimeJob(ctx context.Context, job Job) {
	job.Enabled = false
	if err := s.store.Update(ctx, job); err != nil {
		s.logger.Warnw("disable one-time job after execution",
			"job", job.Name,
			"error", err,
		)
	}
}

// buildCronSpec converts a Job's schedule type and value into a robfig/cron spec string.
func buildCronSpec(job Job) (string, error) {
	switch job.ScheduleType {
	case "cron":
		return job.Schedule, nil

	case "every":
		// Parse as Go duration to validate, then use @every syntax.
		_, err := time.ParseDuration(job.Schedule)
		if err != nil {
			return "", fmt.Errorf("parse duration %q: %w", job.Schedule, err)
		}
		return "@every " + job.Schedule, nil

	case "at":
		// Parse ISO8601 datetime and compute delay from now.
		t, err := time.Parse(time.RFC3339, job.Schedule)
		if err != nil {
			return "", fmt.Errorf("parse datetime %q: %w", job.Schedule, err)
		}
		delay := time.Until(t)
		if delay <= 0 {
			// Already past -- schedule for immediate execution (1 second).
			delay = time.Second
		}
		return "@every " + delay.Round(time.Second).String(), nil

	default:
		return "", fmt.Errorf("unknown schedule type %q", job.ScheduleType)
	}
}

// zapPrintfAdapter adapts zap.SugaredLogger to the Printf interface expected
// by robfig/cron's PrintfLogger.
type zapPrintfAdapter struct {
	logger *zap.SugaredLogger
}

// Printf implements the interface required by robfig/cron.PrintfLogger.
func (a *zapPrintfAdapter) Printf(format string, args ...interface{}) {
	a.logger.Infof(format, args...)
}
