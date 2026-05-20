package background

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type recordingNotifier struct {
	err       error
	channel   string
	message   string
	callCount int
}

func (n *recordingNotifier) SendMessage(_ context.Context, channel string, message string) error {
	n.callCount++
	n.channel = channel
	n.message = message
	return n.err
}

type recordingTyping struct {
	err       error
	stopped   bool
	callCount int
	channel   string
}

func (t *recordingTyping) StartTyping(_ context.Context, channel string) (func(), error) {
	t.callCount++
	t.channel = channel
	if t.err != nil {
		return nil, t.err
	}
	return func() { t.stopped = true }, nil
}

func TestMonitorSummarizesTaskStatuses(t *testing.T) {
	mgr := NewManager(&mockRunner{}, nil, 10, time.Minute, testLogger())
	addMonitorTask(mgr, "pending", Pending)
	addMonitorTask(mgr, "running", Running)
	addMonitorTask(mgr, "done", Done)
	addMonitorTask(mgr, "failed", Failed)
	addMonitorTask(mgr, "cancelled", Cancelled)

	monitor := NewMonitor(mgr, testLogger())

	if got := monitor.ActiveCount(); got != 2 {
		t.Fatalf("ActiveCount() = %d, want 2", got)
	}
	summary := monitor.Summary()
	if summary.Total != 5 || summary.Pending != 1 || summary.Running != 1 ||
		summary.Done != 1 || summary.Failed != 1 || summary.Cancelled != 1 {
		t.Fatalf("Summary() = %+v, want one task in each status", summary)
	}
}

func TestNotificationSkipsEmptyOriginChannel(t *testing.T) {
	notifier := &recordingNotifier{}
	notification := NewNotification(notifier, nil, testLogger())

	err := notification.Notify(context.Background(), &Task{
		ID:     "task-1",
		Status: Done,
		Prompt: "summarize logs",
		Result: "ok",
	})
	if err != nil {
		t.Fatalf("Notify() returned error: %v", err)
	}
	if notifier.callCount != 0 {
		t.Fatalf("SendMessage calls = %d, want 0", notifier.callCount)
	}
}

func TestNotificationSendsStatusMessages(t *testing.T) {
	tests := []struct {
		name   string
		task   *Task
		expect string
	}{
		{
			name:   "done",
			task:   &Task{ID: "done", Status: Done, Prompt: "write report", Result: "finished", OriginChannel: "telegram"},
			expect: "Background task completed: write report\nResult: finished",
		},
		{
			name:   "failed",
			task:   &Task{ID: "failed", Status: Failed, Prompt: "deploy", Error: "boom", OriginChannel: "slack"},
			expect: "Background task failed: deploy\nError: boom",
		},
		{
			name:   "cancelled",
			task:   &Task{ID: "cancelled", Status: Cancelled, Prompt: "index data", OriginChannel: "discord"},
			expect: "Background task cancelled: index data",
		},
		{
			name:   "pending",
			task:   &Task{ID: "pending", Status: Pending, Prompt: "wait", OriginChannel: "telegram"},
			expect: "Background task update [pending]: wait",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifier := &recordingNotifier{}
			notification := NewNotification(notifier, nil, testLogger())

			if err := notification.Notify(context.Background(), tt.task); err != nil {
				t.Fatalf("Notify() returned error: %v", err)
			}
			if notifier.message != tt.expect {
				t.Fatalf("message = %q, want %q", notifier.message, tt.expect)
			}
			if notifier.channel != tt.task.OriginChannel {
				t.Fatalf("channel = %q, want %q", notifier.channel, tt.task.OriginChannel)
			}
		})
	}
}

func TestNotificationPropagatesSendErrors(t *testing.T) {
	notifier := &recordingNotifier{err: errors.New("channel unavailable")}
	notification := NewNotification(notifier, nil, testLogger())

	err := notification.Notify(context.Background(), &Task{
		ID:            "task-1",
		Status:        Done,
		Prompt:        "summarize",
		Result:        "ok",
		OriginChannel: "telegram",
	})

	if err == nil {
		t.Fatal("Notify() returned nil error, want send error")
	}
	if !strings.Contains(err.Error(), "task-1") || !strings.Contains(err.Error(), "channel unavailable") {
		t.Fatalf("Notify() error = %q, want task id and send error", err.Error())
	}
}

func TestNotifyStartAndTypingIndicator(t *testing.T) {
	notifier := &recordingNotifier{}
	typing := &recordingTyping{}
	notification := NewNotification(notifier, typing, testLogger())

	err := notification.NotifyStart(context.Background(), &Task{
		ID:            "task-1",
		Prompt:        "prepare release",
		OriginChannel: "telegram",
	})
	if err != nil {
		t.Fatalf("NotifyStart() returned error: %v", err)
	}
	if notifier.message != "Background task started: prepare release" {
		t.Fatalf("start message = %q, want expected message", notifier.message)
	}

	stop := notification.StartTyping(context.Background(), "telegram")
	stop()
	if typing.callCount != 1 || typing.channel != "telegram" || !typing.stopped {
		t.Fatalf("typing = calls:%d channel:%q stopped:%v, want one stopped telegram indicator", typing.callCount, typing.channel, typing.stopped)
	}
}

func TestStartTypingReturnsNoopForUnavailableTyping(t *testing.T) {
	notification := NewNotification(&recordingNotifier{}, nil, testLogger())
	notification.StartTyping(context.Background(), "telegram")()

	typing := &recordingTyping{err: errors.New("typing unavailable")}
	notification = NewNotification(&recordingNotifier{}, typing, testLogger())
	notification.StartTyping(context.Background(), "telegram")()
	if typing.callCount != 1 {
		t.Fatalf("typing calls = %d, want 1", typing.callCount)
	}
}

func addMonitorTask(mgr *Manager, id string, status Status) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.tasks[id] = &Task{ID: id, Status: status}
}
