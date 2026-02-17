package cron

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// AgentRunner is the interface for executing agent turns.
// This avoids import cycles -- wiring.go will provide the concrete implementation.
type AgentRunner interface {
	Run(ctx context.Context, sessionKey string, prompt string) (string, error)
}

// Executor runs cron jobs by delegating to an AgentRunner and persisting results.
type Executor struct {
	runner   AgentRunner
	delivery *Delivery
	store    Store
	logger   *zap.SugaredLogger
}

// NewExecutor creates a new Executor.
func NewExecutor(runner AgentRunner, delivery *Delivery, store Store, logger *zap.SugaredLogger) *Executor {
	return &Executor{
		runner:   runner,
		delivery: delivery,
		store:    store,
		logger:   logger,
	}
}

// Execute runs a single cron job and returns the result.
// It persists the execution history and delivers results to configured channels.
func (e *Executor) Execute(ctx context.Context, job Job) *JobResult {
	startedAt := time.Now()

	sessionKey := buildSessionKey(job)

	e.logger.Infow("executing cron job",
		"job", job.Name,
		"session_key", sessionKey,
		"session_mode", job.SessionMode,
	)

	// Notify start (best-effort).
	if len(job.DeliverTo) > 0 && e.delivery != nil {
		e.delivery.DeliverStart(ctx, job.Name, job.DeliverTo)
	}

	response, err := e.runner.Run(ctx, sessionKey, job.Prompt)
	duration := time.Since(startedAt)

	result := &JobResult{
		JobID:     job.ID,
		JobName:   job.Name,
		Response:  response,
		Error:     err,
		StartedAt: startedAt,
		Duration:  duration,
	}

	// Persist history entry.
	e.saveHistory(ctx, job, result)

	// Update last_run_at on the job.
	if entStore, ok := e.store.(*EntStore); ok {
		if updateErr := entStore.updateLastRunAt(ctx, job.ID, startedAt); updateErr != nil {
			e.logger.Warnw("update last_run_at: %v", updateErr)
		}
	}

	// Deliver to channels if configured.
	if len(job.DeliverTo) > 0 && e.delivery != nil {
		if deliverErr := e.delivery.Deliver(ctx, result, job.DeliverTo); deliverErr != nil {
			e.logger.Warnw("deliver cron result",
				"job", job.Name,
				"error", deliverErr,
			)
		}
	}

	if err != nil {
		e.logger.Warnw("cron job completed with error",
			"job", job.Name,
			"duration", duration,
			"error", err,
		)
	} else {
		e.logger.Infow("cron job completed",
			"job", job.Name,
			"duration", duration,
		)
	}

	return result
}

// saveHistory persists the execution result to the history store.
func (e *Executor) saveHistory(ctx context.Context, job Job, result *JobResult) {
	completedAt := result.StartedAt.Add(result.Duration)

	entry := HistoryEntry{
		JobID:       job.ID,
		JobName:     job.Name,
		Prompt:      job.Prompt,
		StartedAt:   result.StartedAt,
		CompletedAt: &completedAt,
	}

	if result.Error != nil {
		entry.Status = "failed"
		entry.ErrorMessage = result.Error.Error()
	} else {
		entry.Status = "completed"
		entry.Result = result.Response
	}

	if err := e.store.SaveHistory(ctx, entry); err != nil {
		e.logger.Errorw("save cron job history",
			"job", job.Name,
			"error", err,
		)
	}
}

// buildSessionKey generates a session key based on the job's session mode.
func buildSessionKey(job Job) string {
	switch job.SessionMode {
	case "main":
		return fmt.Sprintf("cron:%s", job.Name)
	default: // "isolated"
		return fmt.Sprintf("cron:%s:%s", job.Name, strconv.FormatInt(time.Now().UnixMilli(), 10))
	}
}
