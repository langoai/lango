package background

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"go.uber.org/zap"
)

// automationPrefix is prepended to prompts sent to the agent runner so that
// the orchestrator recognises them as automated tasks requiring tool execution.
const automationPrefix = "[Automated Task — Execute the following task using tools. Do NOT answer from general knowledge alone.]\n\n"

// AgentRunner executes agent prompts.
type AgentRunner interface {
	Run(ctx context.Context, sessionKey string, prompt string) (string, error)
}

// Projection mirrors background task lifecycle into another authority layer.
// RunLedger uses this to create canonical task IDs and persist transitions.
type Projection interface {
	PrepareTask(ctx context.Context, prompt string, origin Origin) (string, error)
	SyncTask(ctx context.Context, snap TaskSnapshot) error
}

// Origin identifies where a background task was initiated from.
type Origin struct {
	Channel string `json:"channel"`
	Session string `json:"session"`
}

// Manager handles lifecycle management of background tasks.
type Manager struct {
	tasks       map[string]*Task
	mu          sync.RWMutex
	wg          sync.WaitGroup
	maxTasks    int
	taskTimeout time.Duration
	runner      AgentRunner
	notify      *Notification
	projection  Projection
	sem         chan struct{} // concurrency limiter
	logger      *zap.SugaredLogger
}

// NewManager creates a new background task Manager.
// maxTasks limits the total number of non-terminal tasks.
// taskTimeout is the maximum duration for a single task (default: 30m).
// The semaphore size controls how many tasks can run concurrently (defaults to maxTasks if <= 0).
func NewManager(runner AgentRunner, notify *Notification, maxTasks int, taskTimeout time.Duration, logger *zap.SugaredLogger) *Manager {
	if maxTasks <= 0 {
		maxTasks = 10
	}
	if taskTimeout <= 0 {
		taskTimeout = 30 * time.Minute
	}
	return &Manager{
		tasks:       make(map[string]*Task, maxTasks),
		maxTasks:    maxTasks,
		taskTimeout: taskTimeout,
		runner:      runner,
		notify:      notify,
		sem:         make(chan struct{}, maxTasks),
		logger:      logger,
	}
}

// WithProjection configures an optional projection hook for task lifecycle mirroring.
func (m *Manager) WithProjection(projection Projection) *Manager {
	m.projection = projection
	return m
}

// Submit creates and enqueues a new background task. It returns the task ID on success.
func (m *Manager) Submit(ctx context.Context, prompt string, origin Origin) (string, error) {
	m.mu.Lock()

	if m.activeCountLocked() >= m.maxTasks {
		m.mu.Unlock()
		return "", fmt.Errorf("submit task: max concurrent tasks reached (%d)", m.maxTasks)
	}

	detached := types.DetachContext(ctx)
	taskCtx, cancelFn := context.WithTimeout(detached, m.taskTimeout)
	id := uuid.New().String()
	if m.projection != nil {
		preparedID, err := m.projection.PrepareTask(detached, prompt, origin)
		if err != nil {
			m.mu.Unlock()
			return "", fmt.Errorf("submit task: prepare projection: %w", err)
		}
		id = preparedID
	}

	task := &Task{
		ID:            id,
		Status:        Pending,
		Prompt:        prompt,
		OriginChannel: origin.Channel,
		OriginSession: origin.Session,
		cancelFn:      cancelFn,
	}
	m.tasks[id] = task
	m.mu.Unlock()

	m.syncProjection(detached, task)

	m.logger.Infow("task submitted", "taskID", id, "channel", origin.Channel)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.execute(taskCtx, task)
	}()

	return id, nil
}

// Cancel cancels a running or pending task by ID.
func (m *Manager) Cancel(id string) error {
	m.mu.RLock()
	task, ok := m.tasks[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cancel task: task %q not found", id)
	}

	snap := task.Snapshot()
	if snap.Status != Pending && snap.Status != Running {
		return fmt.Errorf("cancel task: task %q is already %s", id, snap.StatusText)
	}

	task.Cancel()
	m.syncProjection(context.Background(), task)
	m.logger.Infow("task cancelled", "taskID", id)
	return nil
}

// Status returns a snapshot of the task with the given ID.
func (m *Manager) Status(id string) (*TaskSnapshot, error) {
	m.mu.RLock()
	task, ok := m.tasks[id]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("task status: task %q not found", id)
	}

	snap := task.Snapshot()
	return &snap, nil
}

// List returns snapshots of all tasks.
func (m *Manager) List() []TaskSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := make([]TaskSnapshot, 0, len(m.tasks))
	for _, task := range m.tasks {
		snapshots = append(snapshots, task.Snapshot())
	}
	return snapshots
}

// Result returns the result of a completed task.
func (m *Manager) Result(id string) (string, error) {
	m.mu.RLock()
	task, ok := m.tasks[id]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("task result: task %q not found", id)
	}

	snap := task.Snapshot()
	if snap.Status != Done {
		return "", fmt.Errorf("task result: task %q is %s, not done", id, snap.StatusText)
	}

	return snap.Result, nil
}

func (m *Manager) execute(ctx context.Context, task *Task) {
	// Context-aware semaphore acquisition: abort if context cancelled.
	select {
	case m.sem <- struct{}{}:
	case <-ctx.Done():
		task.Fail("context cancelled waiting for semaphore")
		return
	}
	defer func() { <-m.sem }()

	task.SetRunning()
	m.syncProjection(ctx, task)
	m.logger.Infow("task running", "taskID", task.ID)

	// Send start notification (best-effort, use task context).
	if m.notify != nil {
		if notifyErr := m.notify.NotifyStart(ctx, task); notifyErr != nil {
			m.logger.Warnw("start notification send error", "taskID", task.ID, "error", notifyErr)
		}
	}

	// Show typing indicator while agent is processing.
	stopTyping := func() {}
	if m.notify != nil {
		stopTyping = m.notify.StartTyping(ctx, task.OriginChannel)
	}

	// Route tool approval requests to the originating channel.
	if task.OriginSession != "" {
		ctx = approval.WithApprovalTarget(ctx, task.OriginSession)
	} else if task.OriginChannel != "" && strings.Contains(task.OriginChannel, ":") {
		ctx = approval.WithApprovalTarget(ctx, task.OriginChannel)
	}

	sessionKey := "bg:" + task.ID
	ctx = session.WithRunContext(ctx, session.RunContext{
		SessionType: "background",
		RunID:       task.ID,
	})
	enrichedPrompt := automationPrefix + "Task: " + task.Prompt
	result, err := m.runner.Run(ctx, sessionKey, enrichedPrompt)
	stopTyping()

	// If the context was cancelled (user cancellation or timeout),
	// don't overwrite the Cancelled status set by Cancel().
	if ctx.Err() != nil {
		return
	}

	if err != nil {
		task.Fail(err.Error())
		m.syncProjection(types.DetachContext(ctx), task)
		m.logger.Warnw("task failed", "taskID", task.ID, "error", err)
	} else {
		task.Complete(result)
		m.syncProjection(types.DetachContext(ctx), task)
		m.logger.Infow("task completed", "taskID", task.ID)
	}

	// Send completion notification (best-effort, detach from task context).
	if m.notify != nil {
		if notifyErr := m.notify.Notify(types.DetachContext(ctx), task); notifyErr != nil {
			m.logger.Warnw("notification send error", "taskID", task.ID, "error", notifyErr)
		}
	}
}

// Shutdown cancels all Pending/Running tasks and waits for goroutines to finish.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	for _, task := range m.tasks {
		snap := task.Snapshot()
		if snap.Status == Pending || snap.Status == Running {
			task.Cancel()
		}
	}
	m.mu.Unlock()

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		m.logger.Warnw("background manager shutdown timed out", "error", ctx.Err())
		return ctx.Err()
	}

	m.logger.Info("background manager shut down")
	return nil
}

// activeCountLocked returns the number of non-terminal tasks. Caller must hold m.mu.
func (m *Manager) activeCountLocked() int {
	count := 0
	for _, task := range m.tasks {
		snap := task.Snapshot()
		if snap.Status == Pending || snap.Status == Running {
			count++
		}
	}
	return count
}

func (m *Manager) syncProjection(ctx context.Context, task *Task) {
	if m.projection == nil {
		return
	}
	if err := m.projection.SyncTask(ctx, task.Snapshot()); err != nil {
		m.logger.Warnw("background projection sync failed", "taskID", task.ID, "error", err)
	}
}
