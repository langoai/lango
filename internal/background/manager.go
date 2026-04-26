package background

import (
	"context"
	"fmt"
	"sort"
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

// maxTerminalTasks is the maximum number of completed/failed/cancelled tasks
// retained in memory. When exceeded, the oldest terminal task is evicted.
const maxTerminalTasks = 500

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

// RetryHook is invoked after a failed task is either eligible for resubmission
// or has exhausted its retry budget. The manager computes the backoff and
// provides a callback for resubmitting the same task after the delay elapses.
type RetryHook func(ctx context.Context, snap TaskSnapshot, exhausted bool, resubmit func())

// RetryKeyDeriver derives an optional canonical retry identity for a task.
type RetryKeyDeriver func(prompt string, origin Origin) string

// Origin identifies where a background task was initiated from.
type Origin struct {
	Channel string `json:"channel"`
	Session string `json:"session"`
}

// Manager handles lifecycle management of background tasks.
type Manager struct {
	tasks           map[string]*Task
	mu              sync.RWMutex
	wg              sync.WaitGroup
	maxTasks        int
	taskTimeout     time.Duration
	runner          AgentRunner
	notify          *Notification
	projection      Projection
	retryHook       RetryHook
	retryKeyDeriver RetryKeyDeriver
	retryPolicy     RetryPolicy
	sem             chan struct{} // concurrency limiter
	logger          *zap.SugaredLogger
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
		retryPolicy: DefaultRetryPolicy(),
		sem:         make(chan struct{}, maxTasks),
		logger:      logger,
	}
}

// WithProjection configures an optional projection hook for task lifecycle mirroring.
func (m *Manager) WithProjection(projection Projection) *Manager {
	m.projection = projection
	return m
}

// WithRetryHook configures an optional retry callback for failed work.
func (m *Manager) WithRetryHook(hook RetryHook) *Manager {
	m.retryHook = hook
	return m
}

// WithRetryKeyDeriver configures an optional retry-key derivation hook.
func (m *Manager) WithRetryKeyDeriver(deriver RetryKeyDeriver) *Manager {
	m.retryKeyDeriver = deriver
	return m
}

// WithRetryPolicy configures the automatic retry policy for failed work.
func (m *Manager) WithRetryPolicy(policy RetryPolicy) *Manager {
	m.retryPolicy = policy.normalized()
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
			cancelFn()
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
	if m.retryKeyDeriver != nil {
		task.RetryKey = m.retryKeyDeriver(prompt, origin)
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
		m.mu.Lock()
		m.evictTerminalTasksLocked()
		m.mu.Unlock()
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
	result := ""
	var err error
	defer func() {
		if r := recover(); r != nil {
			stopTyping()
			err = fmt.Errorf("panic while running background task: %v", r)
			task.Fail(err.Error())
			m.syncProjection(types.DetachContext(ctx), task)
			m.logger.Errorw("task panicked", "taskID", task.ID, "panic", r)
			m.mu.Lock()
			m.evictTerminalTasksLocked()
			m.mu.Unlock()
			if m.notify != nil {
				if notifyErr := m.notify.Notify(types.DetachContext(ctx), task); notifyErr != nil {
					m.logger.Warnw("notification send error", "taskID", task.ID, "error", notifyErr)
				}
			}
		}
	}()
	result, err = m.runner.Run(ctx, sessionKey, enrichedPrompt)
	stopTyping()

	// If the context was cancelled (user cancellation or timeout),
	// don't overwrite the Cancelled status set by Cancel().
	if ctx.Err() != nil {
		m.mu.Lock()
		m.evictTerminalTasksLocked()
		m.mu.Unlock()
		return
	}

	if err != nil {
		task.Fail(err.Error())
		if m.scheduleRetry(types.DetachContext(ctx), task) {
			m.syncProjection(types.DetachContext(ctx), task)
			m.logger.Infow("task retry scheduled", "taskID", task.ID, "attempt", task.Snapshot().AttemptCount, "nextRetryAt", task.Snapshot().NextRetryAt)
		} else {
			if m.retryHook != nil {
				m.retryHook(types.DetachContext(ctx), task.Snapshot(), true, func() {})
			}
			m.syncProjection(types.DetachContext(ctx), task)
			m.logger.Warnw("task failed", "taskID", task.ID, "error", err)
		}
	} else {
		task.Complete(result)
		m.syncProjection(types.DetachContext(ctx), task)
		m.logger.Infow("task completed", "taskID", task.ID)
	}

	// Evict oldest terminal tasks if the cap is exceeded.
	m.mu.Lock()
	m.evictTerminalTasksLocked()
	m.mu.Unlock()

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

func (m *Manager) scheduleRetry(ctx context.Context, task *Task) bool {
	if m.retryHook == nil {
		return false
	}

	snap := task.Snapshot()
	policy := m.retryPolicy.normalized()
	if !policy.ShouldScheduleRetry(snap.AttemptCount) {
		return false
	}

	delay := policy.DelayForAttempt(snap.AttemptCount)
	nextRetryAt := time.Now().Add(delay)
	task.ScheduleRetry(nextRetryAt)

	taskID := task.ID
	expectedAttempt := snap.AttemptCount

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}

		m.mu.RLock()
		currentTask, ok := m.tasks[taskID]
		m.mu.RUnlock()
		if !ok {
			return
		}

		currentSnap := currentTask.Snapshot()
		if currentSnap.Status == Cancelled || currentSnap.AttemptCount != expectedAttempt {
			return
		}

		var once sync.Once
		m.retryHook(ctx, currentSnap, false, func() {
			once.Do(func() {
				m.resubmit(currentTask, ctx)
			})
		})
	}()

	return true
}

func (m *Manager) resubmit(task *Task, ctx context.Context) {
	if task == nil {
		return
	}

	taskCtx, cancelFn := context.WithTimeout(types.DetachContext(ctx), m.taskTimeout)

	task.mu.Lock()
	if task.Status == Cancelled {
		task.mu.Unlock()
		cancelFn()
		return
	}
	task.cancelFn = cancelFn
	task.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.execute(taskCtx, task)
	}()
}

// evictTerminalTasksLocked removes the oldest terminal tasks when the count
// exceeds maxTerminalTasks. Caller must hold m.mu (write lock).
func (m *Manager) evictTerminalTasksLocked() {
	type terminalEntry struct {
		id          string
		completedAt time.Time
	}

	var terminals []terminalEntry
	for id, task := range m.tasks {
		snap := task.Snapshot()
		switch snap.Status {
		case Done, Failed, Cancelled:
			ts := snap.CompletedAt
			if ts.IsZero() {
				ts = snap.StartedAt
			}
			terminals = append(terminals, terminalEntry{id: id, completedAt: ts})
		}
	}

	if len(terminals) <= maxTerminalTasks {
		return
	}

	sort.Slice(terminals, func(i, j int) bool {
		return terminals[i].completedAt.Before(terminals[j].completedAt)
	})

	evictCount := len(terminals) - maxTerminalTasks
	for i := 0; i < evictCount; i++ {
		delete(m.tasks, terminals[i].id)
	}
}
