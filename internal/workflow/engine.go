package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/langoai/lango/internal/types"
	"go.uber.org/zap"
)

// AgentRunner executes agent prompts (avoids import cycles with orchestration).
type AgentRunner interface {
	Run(ctx context.Context, sessionKey string, prompt string) (string, error)
}

// ChannelSender sends results to communication channels.
type ChannelSender interface {
	SendMessage(ctx context.Context, channel string, message string) error
}

// Engine orchestrates DAG-based workflow execution.
type Engine struct {
	runner         AgentRunner
	state          *StateStore
	sender         ChannelSender
	maxConcurrent  int
	defaultTimeout time.Duration
	logger         *zap.SugaredLogger

	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

// NewEngine creates a new workflow execution engine.
func NewEngine(
	runner AgentRunner,
	state *StateStore,
	sender ChannelSender,
	maxConcurrent int,
	defaultTimeout time.Duration,
	logger *zap.SugaredLogger,
) *Engine {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}
	if defaultTimeout <= 0 {
		defaultTimeout = 5 * time.Minute
	}
	return &Engine{
		runner:         runner,
		state:          state,
		sender:         sender,
		maxConcurrent:  maxConcurrent,
		defaultTimeout: defaultTimeout,
		logger:         logger,
		cancels:        make(map[string]context.CancelFunc),
	}
}

// Run executes a workflow from start to finish synchronously.
// The context is detached from the parent to prevent cancellation
// when the originating request completes.
func (e *Engine) Run(ctx context.Context, w *Workflow) (*RunResult, error) {
	if err := Validate(w); err != nil {
		return nil, fmt.Errorf("validate workflow: %w", err)
	}

	dag, err := NewDAG(w.Steps)
	if err != nil {
		return nil, fmt.Errorf("build DAG: %w", err)
	}

	// Detach from parent context to prevent cascading cancellation.
	detached := types.DetachContext(ctx)

	runID, err := e.state.CreateRun(detached, w)
	if err != nil {
		return nil, fmt.Errorf("create run: %w", err)
	}

	return e.runDAG(detached, runID, w, dag)
}

// RunAsync validates, creates the run record and step records, then
// executes the DAG in a background goroutine. It returns the runID
// immediately so the caller can poll via Status().
func (e *Engine) RunAsync(ctx context.Context, w *Workflow) (string, error) {
	if err := Validate(w); err != nil {
		return "", fmt.Errorf("validate workflow: %w", err)
	}

	dag, err := NewDAG(w.Steps)
	if err != nil {
		return "", fmt.Errorf("build DAG: %w", err)
	}

	// Detach from parent context to prevent cascading cancellation.
	detached := types.DetachContext(ctx)

	runID, err := e.state.CreateRun(detached, w)
	if err != nil {
		return "", fmt.Errorf("create run: %w", err)
	}

	// Create step records before launching goroutine so the caller
	// can immediately query step status.
	for _, step := range w.Steps {
		if createErr := e.state.CreateStepRun(detached, runID, step, step.Prompt); createErr != nil {
			return "", fmt.Errorf("create step run %q: %w", step.ID, createErr)
		}
	}

	go func() {
		if _, runErr := e.runDAG(detached, runID, w, dag); runErr != nil {
			e.logger.Warnw("async workflow failed", "runID", runID, "error", runErr)
		}
	}()

	return runID, nil
}

// runDAG executes a workflow DAG to completion.
func (e *Engine) runDAG(ctx context.Context, runID string, w *Workflow, dag *DAG) (*RunResult, error) {
	// Register cancel function.
	ctx, cancel := context.WithCancel(ctx)
	e.mu.Lock()
	e.cancels[runID] = cancel
	e.mu.Unlock()

	defer func() {
		cancel()
		e.mu.Lock()
		delete(e.cancels, runID)
		e.mu.Unlock()
	}()

	// Create step records (idempotent — RunAsync pre-creates them,
	// Run calls runDAG directly so we create them here).
	for _, step := range w.Steps {
		// CreateStepRun is expected to be idempotent or tolerate duplicates.
		if createErr := e.state.CreateStepRun(ctx, runID, step, step.Prompt); createErr != nil {
			// Log but don't fail — RunAsync already created these.
			e.logger.Debugw("create step run (may already exist)", "step", step.ID, "error", createErr)
		}
	}

	// Update run status to running.
	if updateErr := e.state.UpdateRunStatus(ctx, runID, "running"); updateErr != nil {
		return nil, fmt.Errorf("update run status: %w", updateErr)
	}

	startedAt := time.Now()
	results := make(map[string]string, len(w.Steps))
	completed := make(map[string]bool, len(w.Steps))
	var runErr error

	// Build step lookup.
	stepMap := make(map[string]*Step, len(w.Steps))
	for i := range w.Steps {
		stepMap[w.Steps[i].ID] = &w.Steps[i]
	}

	// Execute DAG layer by layer.
	for len(completed) < len(w.Steps) {
		ready := dag.Ready(completed)
		if len(ready) == 0 {
			runErr = fmt.Errorf("no ready steps but %d/%d completed", len(completed), len(w.Steps))
			break
		}

		// Filter out already completed or in-flight steps.
		var toRun []string
		for _, id := range ready {
			if !completed[id] {
				toRun = append(toRun, id)
			}
		}
		if len(toRun) == 0 {
			break
		}

		// Execute ready steps in parallel with concurrency limit.
		sem := make(chan struct{}, e.maxConcurrent)
		var wg sync.WaitGroup
		var mu sync.Mutex
		var stepErrs []string

		for _, stepID := range toRun {
			wg.Add(1)
			go func(sid string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				step := stepMap[sid]
				stepResult, execErr := e.executeStep(ctx, runID, w.Name, step, results)

				mu.Lock()
				defer mu.Unlock()

				if execErr != nil {
					stepErrs = append(stepErrs, fmt.Sprintf("step %q: %s", sid, execErr))
					completed[sid] = true
				} else {
					results[sid] = stepResult
					completed[sid] = true
				}
			}(stepID)
		}

		wg.Wait()

		if len(stepErrs) > 0 {
			runErr = fmt.Errorf("step failures: %s", strings.Join(stepErrs, "; "))
			break
		}

		// Check context cancellation.
		if ctx.Err() != nil {
			runErr = ctx.Err()
			break
		}
	}

	// Deliver final results if configured.
	if runErr == nil && len(w.DeliverTo) > 0 && e.sender != nil {
		summary := e.buildSummary(w.Name, results)
		for _, target := range w.DeliverTo {
			if sendErr := e.sender.SendMessage(ctx, target, summary); sendErr != nil {
				e.logger.Warnw("deliver workflow result", "target", target, "error", sendErr)
			}
		}
	} else if runErr == nil && len(w.DeliverTo) == 0 {
		e.logger.Warnw("workflow completed but no delivery channel configured",
			"workflow", w.Name,
			"hint", "set deliver_to in YAML or configure workflow.defaultDeliverTo in settings",
		)
	}

	completedAt := time.Now()
	finalStatus := "completed"
	var errMsg string
	if runErr != nil {
		finalStatus = "failed"
		errMsg = runErr.Error()
	}

	if completeErr := e.state.CompleteRun(ctx, runID, finalStatus, errMsg); completeErr != nil {
		e.logger.Warnw("complete run record", "runID", runID, "error", completeErr)
	}

	return &RunResult{
		RunID:        runID,
		WorkflowName: w.Name,
		Status:       finalStatus,
		StepResults:  results,
		Error:        errMsg,
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
	}, nil
}

// executeStep runs a single workflow step with timeout and state tracking.
func (e *Engine) executeStep(
	ctx context.Context,
	runID string,
	workflowName string,
	step *Step,
	currentResults map[string]string,
) (string, error) {
	// Render prompt template.
	rendered, err := RenderPrompt(step.Prompt, currentResults)
	if err != nil {
		if updateErr := e.state.UpdateStepStatus(ctx, runID, step.ID, "failed", "", err.Error()); updateErr != nil {
			e.logger.Warnw("update step status after render failure", "step", step.ID, "error", updateErr)
		}
		return "", fmt.Errorf("render prompt for step %q: %w", step.ID, err)
	}

	// Update step status to running.
	if updateErr := e.state.UpdateStepStatus(ctx, runID, step.ID, "running", "", ""); updateErr != nil {
		e.logger.Warnw("update step status to running", "step", step.ID, "error", updateErr)
	}

	// Determine timeout.
	timeout := e.defaultTimeout
	if step.Timeout > 0 {
		timeout = step.Timeout
	}

	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Generate session key.
	sessionKey := fmt.Sprintf("workflow:%s:%s", workflowName, step.ID)

	// Execute via agent runner.
	result, err := e.runner.Run(stepCtx, sessionKey, rendered)
	if err != nil {
		if updateErr := e.state.UpdateStepStatus(ctx, runID, step.ID, "failed", "", err.Error()); updateErr != nil {
			e.logger.Warnw("update step status after execution failure", "step", step.ID, "error", updateErr)
		}
		return "", fmt.Errorf("execute step %q: %w", step.ID, err)
	}

	// Update step status to completed.
	if updateErr := e.state.UpdateStepStatus(ctx, runID, step.ID, "completed", result, ""); updateErr != nil {
		e.logger.Warnw("update step status to completed", "step", step.ID, "error", updateErr)
	}

	// Per-step delivery.
	if len(step.DeliverTo) > 0 && e.sender != nil {
		msg := fmt.Sprintf("[%s/%s] %s", workflowName, step.ID, result)
		for _, target := range step.DeliverTo {
			if sendErr := e.sender.SendMessage(ctx, target, msg); sendErr != nil {
				e.logger.Warnw("deliver step result", "step", step.ID, "target", target, "error", sendErr)
			}
		}
	}

	return result, nil
}

// Resume re-executes a workflow from where it left off.
func (e *Engine) Resume(ctx context.Context, runID string) (*RunResult, error) {
	status, err := e.state.GetRunStatus(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("get run status for resume: %w", err)
	}

	if status.Status == "completed" {
		return nil, fmt.Errorf("run %q already completed", runID)
	}

	// Get existing results.
	existingResults, err := e.state.GetStepResults(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("get step results for resume: %w", err)
	}

	e.logger.Infow("resuming workflow",
		"runID", runID,
		"workflow", status.WorkflowName,
		"completedSteps", len(existingResults),
		"totalSteps", status.TotalSteps,
	)

	// We cannot fully reconstruct the workflow from DB alone since the original
	// YAML definition (with step dependencies) is needed. Resume returns the
	// current status with existing results for the caller to re-invoke Run with
	// the same workflow definition.
	return &RunResult{
		RunID:        runID,
		WorkflowName: status.WorkflowName,
		Status:       status.Status,
		StepResults:  existingResults,
	}, nil
}

// Cancel requests cancellation of a running workflow.
func (e *Engine) Cancel(runID string) error {
	e.mu.Lock()
	cancel, ok := e.cancels[runID]
	e.mu.Unlock()

	if !ok {
		return fmt.Errorf("run %q not found or not running", runID)
	}

	cancel()

	ctx := context.Background()
	if updateErr := e.state.CompleteRun(ctx, runID, "cancelled", "cancelled by user"); updateErr != nil {
		return fmt.Errorf("update cancelled run: %w", updateErr)
	}

	return nil
}

// Status returns the current status of a workflow execution.
func (e *Engine) Status(ctx context.Context, runID string) (*RunStatus, error) {
	return e.state.GetRunStatus(ctx, runID)
}

// ListRuns returns the most recent workflow runs.
func (e *Engine) ListRuns(ctx context.Context, limit int) ([]RunStatus, error) {
	return e.state.ListRuns(ctx, limit)
}

// Shutdown cancels all running workflows.
func (e *Engine) Shutdown() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for runID, cancel := range e.cancels {
		cancel()
		e.logger.Infow("workflow cancelled during shutdown", "runID", runID)
	}
	e.logger.Info("workflow engine shut down")
}

// buildSummary formats a human-readable summary of workflow results.
func (e *Engine) buildSummary(workflowName string, results map[string]string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Workflow '%s' completed.\n\n", workflowName))
	for stepID, result := range results {
		b.WriteString(fmt.Sprintf("--- %s ---\n%s\n\n", stepID, result))
	}
	return b.String()
}
