package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/cronjob"
	"github.com/langoai/lango/internal/ent/cronjobhistory"
)

// Store defines the persistence interface for cron jobs and their history.
type Store interface {
	Create(ctx context.Context, job Job) error
	Get(ctx context.Context, id string) (*Job, error)
	GetByName(ctx context.Context, name string) (*Job, error)
	List(ctx context.Context) ([]Job, error)
	ListEnabled(ctx context.Context) ([]Job, error)
	Update(ctx context.Context, job Job) error
	Delete(ctx context.Context, id string) error
	SaveHistory(ctx context.Context, entry HistoryEntry) error
	ListHistory(ctx context.Context, jobID string, limit int) ([]HistoryEntry, error)
	ListAllHistory(ctx context.Context, limit int) ([]HistoryEntry, error)
}

// EntStore implements Store using the Ent ORM client.
type EntStore struct {
	client *ent.Client
}

// NewEntStore creates a new EntStore backed by the given Ent client.
func NewEntStore(client *ent.Client) *EntStore {
	return &EntStore{client: client}
}

// Create persists a new cron job.
func (s *EntStore) Create(ctx context.Context, job Job) error {
	builder := s.client.CronJob.Create().
		SetName(job.Name).
		SetScheduleType(cronjob.ScheduleType(job.ScheduleType)).
		SetSchedule(job.Schedule).
		SetPrompt(job.Prompt).
		SetSessionMode(job.SessionMode).
		SetTimezone(job.Timezone).
		SetEnabled(job.Enabled)

	if len(job.DeliverTo) > 0 {
		builder.SetDeliverTo(job.DeliverTo)
	}

	if job.LastRunAt != nil {
		builder.SetLastRunAt(*job.LastRunAt)
	}
	if job.NextRunAt != nil {
		builder.SetNextRunAt(*job.NextRunAt)
	}

	if job.ID != "" {
		id, err := uuid.Parse(job.ID)
		if err != nil {
			return fmt.Errorf("parse job id %q: %w", job.ID, err)
		}
		builder.SetID(id)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create cron job %q: %w", job.Name, err)
	}
	return nil
}

// Get retrieves a cron job by ID.
func (s *EntStore) Get(ctx context.Context, id string) (*Job, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse job id %q: %w", id, err)
	}

	row, err := s.client.CronJob.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get cron job %q: %w", id, err)
	}

	j := entCronJobToDomain(row)
	return &j, nil
}

// GetByName retrieves a cron job by its unique name.
func (s *EntStore) GetByName(ctx context.Context, name string) (*Job, error) {
	row, err := s.client.CronJob.Query().
		Where(cronjob.Name(name)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cron job by name %q: %w", name, err)
	}

	j := entCronJobToDomain(row)
	return &j, nil
}

// List returns all cron jobs.
func (s *EntStore) List(ctx context.Context) ([]Job, error) {
	rows, err := s.client.CronJob.Query().
		Order(cronjob.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cron jobs: %w", err)
	}

	return entCronJobsToDomain(rows), nil
}

// ListEnabled returns only enabled cron jobs.
func (s *EntStore) ListEnabled(ctx context.Context) ([]Job, error) {
	rows, err := s.client.CronJob.Query().
		Where(cronjob.EnabledEQ(true)).
		Order(cronjob.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled cron jobs: %w", err)
	}

	return entCronJobsToDomain(rows), nil
}

// Update modifies an existing cron job.
func (s *EntStore) Update(ctx context.Context, job Job) error {
	uid, err := uuid.Parse(job.ID)
	if err != nil {
		return fmt.Errorf("parse job id %q: %w", job.ID, err)
	}

	builder := s.client.CronJob.UpdateOneID(uid).
		SetName(job.Name).
		SetScheduleType(cronjob.ScheduleType(job.ScheduleType)).
		SetSchedule(job.Schedule).
		SetPrompt(job.Prompt).
		SetSessionMode(job.SessionMode).
		SetTimezone(job.Timezone).
		SetEnabled(job.Enabled)

	if len(job.DeliverTo) > 0 {
		builder.SetDeliverTo(job.DeliverTo)
	} else {
		builder.ClearDeliverTo()
	}

	if job.LastRunAt != nil {
		builder.SetLastRunAt(*job.LastRunAt)
	} else {
		builder.ClearLastRunAt()
	}

	if job.NextRunAt != nil {
		builder.SetNextRunAt(*job.NextRunAt)
	} else {
		builder.ClearNextRunAt()
	}

	_, err = builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("update cron job %q: %w", job.ID, err)
	}
	return nil
}

// Delete removes a cron job by ID.
func (s *EntStore) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("parse job id %q: %w", id, err)
	}

	err = s.client.CronJob.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete cron job %q: %w", id, err)
	}
	return nil
}

// SaveHistory persists a job execution history entry.
func (s *EntStore) SaveHistory(ctx context.Context, entry HistoryEntry) error {
	jobID, err := uuid.Parse(entry.JobID)
	if err != nil {
		return fmt.Errorf("parse history job id %q: %w", entry.JobID, err)
	}

	builder := s.client.CronJobHistory.Create().
		SetJobID(jobID).
		SetJobName(entry.JobName).
		SetStatus(cronjobhistory.Status(entry.Status)).
		SetPrompt(entry.Prompt).
		SetTokensUsed(entry.TokensUsed).
		SetStartedAt(entry.StartedAt)

	if entry.Result != "" {
		builder.SetResult(entry.Result)
	}
	if entry.ErrorMessage != "" {
		builder.SetErrorMessage(entry.ErrorMessage)
	}
	if entry.CompletedAt != nil {
		builder.SetCompletedAt(*entry.CompletedAt)
	}
	if entry.ID != "" {
		id, err := uuid.Parse(entry.ID)
		if err != nil {
			return fmt.Errorf("parse history id %q: %w", entry.ID, err)
		}
		builder.SetID(id)
	}

	_, err = builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("save cron job history: %w", err)
	}
	return nil
}

// ListHistory returns execution history for a specific job, ordered by most recent first.
func (s *EntStore) ListHistory(ctx context.Context, jobID string, limit int) ([]HistoryEntry, error) {
	uid, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("parse history job id %q: %w", jobID, err)
	}

	rows, err := s.client.CronJobHistory.Query().
		Where(cronjobhistory.JobIDEQ(uid)).
		Order(ent.Desc(cronjobhistory.FieldStartedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list history for job %q: %w", jobID, err)
	}

	return entHistoriesToDomain(rows), nil
}

// ListAllHistory returns execution history across all jobs, ordered by most recent first.
func (s *EntStore) ListAllHistory(ctx context.Context, limit int) ([]HistoryEntry, error) {
	rows, err := s.client.CronJobHistory.Query().
		Order(ent.Desc(cronjobhistory.FieldStartedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all cron history: %w", err)
	}

	return entHistoriesToDomain(rows), nil
}

// entCronJobToDomain converts an Ent CronJob entity to the domain Job type.
func entCronJobToDomain(e *ent.CronJob) Job {
	j := Job{
		ID:           e.ID.String(),
		Name:         e.Name,
		ScheduleType: string(e.ScheduleType),
		Schedule:     e.Schedule,
		Prompt:       e.Prompt,
		SessionMode:  e.SessionMode,
		Timezone:     e.Timezone,
		Enabled:      e.Enabled,
		CreatedAt:    e.CreatedAt,
	}

	if len(e.DeliverTo) > 0 {
		dt := make([]string, len(e.DeliverTo))
		copy(dt, e.DeliverTo)
		j.DeliverTo = dt
	}

	if e.LastRunAt != nil {
		t := *e.LastRunAt
		j.LastRunAt = &t
	}
	if e.NextRunAt != nil {
		t := *e.NextRunAt
		j.NextRunAt = &t
	}

	return j
}

// entCronJobsToDomain converts a slice of Ent CronJob entities to domain types.
func entCronJobsToDomain(rows []*ent.CronJob) []Job {
	jobs := make([]Job, 0, len(rows))
	for _, r := range rows {
		jobs = append(jobs, entCronJobToDomain(r))
	}
	return jobs
}

// entHistoryToDomain converts an Ent CronJobHistory entity to the domain HistoryEntry type.
func entHistoryToDomain(e *ent.CronJobHistory) HistoryEntry {
	h := HistoryEntry{
		ID:           e.ID.String(),
		JobID:        e.JobID.String(),
		JobName:      e.JobName,
		Status:       string(e.Status),
		Prompt:       e.Prompt,
		Result:       e.Result,
		ErrorMessage: e.ErrorMessage,
		TokensUsed:   e.TokensUsed,
		StartedAt:    e.StartedAt,
	}

	if e.CompletedAt != nil {
		t := *e.CompletedAt
		h.CompletedAt = &t
	}

	return h
}

// entHistoriesToDomain converts a slice of Ent CronJobHistory entities to domain types.
func entHistoriesToDomain(rows []*ent.CronJobHistory) []HistoryEntry {
	entries := make([]HistoryEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, entHistoryToDomain(r))
	}
	return entries
}

// Compile-time interface check.
var _ Store = (*EntStore)(nil)

// updateLastRunAt updates the last_run_at timestamp for a job.
func (s *EntStore) updateLastRunAt(ctx context.Context, id string, t time.Time) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("parse job id %q: %w", id, err)
	}

	_, err = s.client.CronJob.UpdateOneID(uid).
		SetLastRunAt(t).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update last_run_at for job %q: %w", id, err)
	}
	return nil
}
