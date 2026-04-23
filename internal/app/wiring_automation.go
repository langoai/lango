package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	cronpkg "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turnrunner"
	"github.com/langoai/lango/internal/turntrace"
	"github.com/langoai/lango/internal/workflow"
)

// agentRunnerAdapter adapts app.runAgent to cron.AgentRunner / background.AgentRunner / workflow.AgentRunner.
type agentRunnerAdapter struct {
	app *App
}

func (r *agentRunnerAdapter) Run(ctx context.Context, sessionKey, promptText string) (string, error) {
	if r.app.TurnRunner == nil {
		return "", fmt.Errorf("turn runner is not initialized")
	}
	result, err := r.app.TurnRunner.Run(ctx, turnrunner.Request{
		SessionKey: sessionKey,
		Input:      promptText,
		Entrypoint: "automation",
	})
	if err != nil {
		return "", err
	}
	if result.Outcome != turntrace.OutcomeSuccess {
		return result.ResponseText, errors.New(result.UserMessage)
	}
	return result.ResponseText, nil
}

// initCron creates the cron scheduling system if enabled.
func initCron(cfg *config.Config, store session.Store, app *App) *cronpkg.Scheduler {
	if !cfg.Cron.Enabled {
		logger().Info("cron scheduling disabled")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("cron scheduling requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()
	cronStore := cronpkg.NewEntStore(client)
	sender := newChannelSender(app)
	delivery := cronpkg.NewDelivery(sender, sender, logger())
	runner := &agentRunnerAdapter{app: app}
	executor := cronpkg.NewExecutor(runner, delivery, cronStore, logger())

	maxJobs := cfg.Cron.MaxConcurrentJobs
	if maxJobs <= 0 {
		maxJobs = 5
	}

	tz := cfg.Cron.Timezone
	if tz == "" {
		tz = "UTC"
	}

	defaultJobTimeout := cfg.Cron.DefaultJobTimeout
	if defaultJobTimeout <= 0 {
		defaultJobTimeout = 30 * time.Minute
	}

	scheduler := cronpkg.New(cronStore, executor, cronpkg.SchedulerConfig{
		Timezone:       tz,
		MaxJobs:        maxJobs,
		DefaultTimeout: defaultJobTimeout,
		Logger:         logger(),
	})

	logger().Infow("cron scheduling initialized",
		"timezone", tz,
		"maxConcurrentJobs", maxJobs,
	)

	return scheduler
}

// initBackground creates the background task manager if enabled.
func initBackground(cfg *config.Config, app *App, receiptStore *receipts.Store) *background.Manager {
	if !cfg.Background.Enabled {
		logger().Info("background tasks disabled")
		return nil
	}

	runner := &agentRunnerAdapter{app: app}
	sender := newChannelSender(app)
	notify := background.NewNotification(sender, sender, logger())

	maxTasks := cfg.Background.MaxConcurrentTasks
	if maxTasks <= 0 {
		maxTasks = 3
	}

	taskTimeout := cfg.Background.TaskTimeout
	if taskTimeout <= 0 {
		taskTimeout = 30 * time.Minute
	}

	mgr := background.NewManager(runner, notify, maxTasks, taskTimeout, logger())
	mgr.WithRetryKeyDeriver(func(prompt string, _ background.Origin) string {
		return parsePostAdjudicationRetryKey(prompt)
	})
	if receiptStore != nil {
		mgr.WithRetryHook(func(ctx context.Context, snap background.TaskSnapshot, exhausted bool, resubmit func()) {
			transactionReceiptID, outcome, ok := parseRetryKeyParts(snap.RetryKey)
			if !ok {
				return
			}
			if exhausted {
				if err := receiptStore.RecordPostAdjudicationDeadLetter(ctx, receipts.PostAdjudicationDeadLetterRequest{
					TransactionReceiptID: transactionReceiptID,
					Outcome:              outcome,
					AttemptCount:         snap.AttemptCount,
					Reason:               snap.Error,
				}); err != nil {
					logger().Warnw("post-adjudication dead-letter evidence failed", "taskID", snap.ID, "error", err)
				}
				return
			}
			if err := receiptStore.RecordPostAdjudicationRetryScheduled(ctx, receipts.PostAdjudicationRetryScheduledRequest{
				TransactionReceiptID: transactionReceiptID,
				Outcome:              outcome,
				AttemptCount:         snap.AttemptCount,
				NextRetryAt:          snap.NextRetryAt,
			}); err != nil {
				logger().Warnw("post-adjudication retry evidence failed", "taskID", snap.ID, "error", err)
			}
			resubmit()
		})
	}
	if app.RunLedgerStore != nil && cfg.RunLedger.Enabled && cfg.RunLedger.WriteThrough {
		mgr.WithProjection(runledger.NewBackgroundWriteThrough(
			app.RunLedgerStore,
			runledger.RolloutConfig{Stage: runledger.StageWriteThrough},
		).WithMaxHistory(cfg.RunLedger.MaxRunHistory))
	}

	logger().Infow("background task manager initialized",
		"maxConcurrentTasks", maxTasks,
		"yieldMs", cfg.Background.YieldMs,
	)

	return mgr
}

func parsePostAdjudicationRetryKey(prompt string) string {
	transactionReceiptID := ""
	switch {
	case strings.Contains(prompt, "release_escrow_settlement"):
		if idx := strings.Index(prompt, "transaction_receipt_id="); idx >= 0 {
			fields := strings.Fields(strings.TrimSpace(prompt[idx+len("transaction_receipt_id="):]))
			if len(fields) > 0 {
				transactionReceiptID = fields[0]
			}
		}
		if transactionReceiptID != "" {
			return transactionReceiptID + ":" + string(receipts.EscrowAdjudicationRelease)
		}
	case strings.Contains(prompt, "refund_escrow_settlement"):
		if idx := strings.Index(prompt, "transaction_receipt_id="); idx >= 0 {
			fields := strings.Fields(strings.TrimSpace(prompt[idx+len("transaction_receipt_id="):]))
			if len(fields) > 0 {
				transactionReceiptID = fields[0]
			}
		}
		if transactionReceiptID != "" {
			return transactionReceiptID + ":" + string(receipts.EscrowAdjudicationRefund)
		}
	}
	return ""
}

func parseRetryKeyParts(retryKey string) (string, receipts.EscrowAdjudicationDecision, bool) {
	parts := strings.SplitN(strings.TrimSpace(retryKey), ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", false
	}
	outcome := receipts.EscrowAdjudicationDecision(parts[1])
	if outcome != receipts.EscrowAdjudicationRelease && outcome != receipts.EscrowAdjudicationRefund {
		return "", "", false
	}
	return parts[0], outcome, true
}

// initWorkflow creates the workflow engine if enabled.
func initWorkflow(cfg *config.Config, store session.Store, app *App, rlv *runLedgerValues) *workflow.Engine {
	if !cfg.Workflow.Enabled {
		logger().Info("workflow engine disabled")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("workflow engine requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()
	var state workflow.RunStore = workflow.NewStateStore(client, logger())
	if rlv != nil && rlv.store != nil && cfg.RunLedger.Enabled && cfg.RunLedger.WriteThrough {
		state = runledger.NewWorkflowWriteThrough(
			rlv.store,
			workflow.NewStateStore(client, logger()),
			runledger.RolloutConfig{Stage: runledger.StageWriteThrough},
		).WithMaxHistory(cfg.RunLedger.MaxRunHistory)
	}
	runner := &agentRunnerAdapter{app: app}
	sender := newChannelSender(app)

	maxConcurrent := cfg.Workflow.MaxConcurrentSteps
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}

	defaultTimeout := cfg.Workflow.DefaultTimeout
	if defaultTimeout <= 0 {
		defaultTimeout = 10 * time.Minute
	}

	engine := workflow.NewEngine(runner, state, sender, maxConcurrent, defaultTimeout, logger())

	logger().Infow("workflow engine initialized",
		"maxConcurrentSteps", maxConcurrent,
		"defaultTimeout", defaultTimeout,
	)

	return engine
}
