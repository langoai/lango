package app

import (
	"context"
	"time"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	cronpkg "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/workflow"
)

// agentRunnerAdapter adapts app.runAgent to cron.AgentRunner / background.AgentRunner / workflow.AgentRunner.
type agentRunnerAdapter struct {
	app *App
}

func (r *agentRunnerAdapter) Run(ctx context.Context, sessionKey, promptText string) (string, error) {
	return r.app.runAgent(ctx, sessionKey, promptText)
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
func initBackground(cfg *config.Config, app *App) *background.Manager {
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

	logger().Infow("background task manager initialized",
		"maxConcurrentTasks", maxTasks,
		"yieldMs", cfg.Background.YieldMs,
	)

	return mgr
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
		)
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
