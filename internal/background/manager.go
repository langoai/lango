package background

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/langowarny/lango/internal/approval"
	"go.uber.org/zap"
)

// AgentRunner executes agent prompts.
type AgentRunner interface {
	Run(ctx context.Context, sessionKey string, prompt string) (string, error)
}

// Origin identifies where a background task was initiated from.
type Origin struct {
	Channel string `json:"channel"`
	Session string `json:"session"`
}

// Manager handles lifecycle management of background tasks.
type Manager struct {
	tasks    map[string]*Task
	mu       sync.RWMutex
	maxTasks int
	runner   AgentRunner
	notify   *Notification
	sem      chan struct{} // concurrency limiter
	logger   *zap.SugaredLogger
}

// NewManager creates a new background task Manager.
// maxTasks limits the total number of non-terminal tasks.
// The semaphore size controls how many tasks can run concurrently (defaults to maxTasks if <= 0).
func NewManager(runner AgentRunner, notify *Notification, maxTasks int, logger *zap.SugaredLogger) *Manager {
	if maxTasks <= 0 {
		maxTasks = 10
	}
	return &Manager{
		tasks:    make(map[string]*Task, maxTasks),
		maxTasks: maxTasks,
		runner:   runner,
		notify:   notify,
		sem:      make(chan struct{}, maxTasks),
		logger:   logger,
	}
}

// Submit creates and enqueues a new background task. It returns the task ID on success.
func (m *Manager) Submit(ctx context.Context, prompt string, origin Origin) (string, error) {
	m.mu.Lock()

	if m.activeCountLocked() >= m.maxTasks {
		m.mu.Unlock()
		return "", fmt.Errorf("submit task: max concurrent tasks reached (%d)", m.maxTasks)
	}

	taskCtx, cancelFn := context.WithCancel(ctx)
	id := uuid.New().String()

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

	m.logger.Infow("task submitted", "taskID", id, "channel", origin.Channel)

	go m.execute(taskCtx, task)

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
	// Acquire concurrency semaphore.
	m.sem <- struct{}{}
	defer func() { <-m.sem }()

	task.SetRunning()
	m.logger.Infow("task running", "taskID", task.ID)

	// Send start notification (best-effort).
	if m.notify != nil {
		if notifyErr := m.notify.NotifyStart(context.Background(), task); notifyErr != nil {
			m.logger.Warnw("start notification send error", "taskID", task.ID, "error", notifyErr)
		}
	}

	// Route tool approval requests to the originating channel.
	if task.OriginSession != "" {
		ctx = approval.WithApprovalTarget(ctx, task.OriginSession)
	} else if task.OriginChannel != "" && strings.Contains(task.OriginChannel, ":") {
		ctx = approval.WithApprovalTarget(ctx, task.OriginChannel)
	}

	sessionKey := "bg:" + task.ID
	result, err := m.runner.Run(ctx, sessionKey, task.Prompt)

	if err != nil {
		task.Fail(err.Error())
		m.logger.Warnw("task failed", "taskID", task.ID, "error", err)
	} else {
		task.Complete(result)
		m.logger.Infow("task completed", "taskID", task.ID)
	}

	// Send notification (best-effort).
	if m.notify != nil {
		if notifyErr := m.notify.Notify(context.Background(), task); notifyErr != nil {
			m.logger.Warnw("notification send error", "taskID", task.ID, "error", notifyErr)
		}
	}
}

// Shutdown cancels all Pending/Running tasks.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, task := range m.tasks {
		snap := task.Snapshot()
		if snap.Status == Pending || snap.Status == Running {
			task.Cancel()
		}
	}
	m.logger.Info("background manager shut down")
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
