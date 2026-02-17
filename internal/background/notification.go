package background

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ChannelNotifier sends notifications to communication channels.
type ChannelNotifier interface {
	SendMessage(ctx context.Context, channel string, message string) error
}

// Notification handles sending completion or failure notifications for background tasks.
type Notification struct {
	notifier ChannelNotifier
	logger   *zap.SugaredLogger
}

// NewNotification creates a new Notification with the given notifier and logger.
func NewNotification(notifier ChannelNotifier, logger *zap.SugaredLogger) *Notification {
	return &Notification{
		notifier: notifier,
		logger:   logger,
	}
}

// Notify sends a notification about a completed or failed task to its origin channel.
func (n *Notification) Notify(ctx context.Context, task *Task) error {
	snap := task.Snapshot()

	if snap.OriginChannel == "" {
		n.logger.Debugw("skip notification: no origin channel", "taskID", snap.ID)
		return nil
	}

	msg := formatNotification(snap)

	if err := n.notifier.SendMessage(ctx, snap.OriginChannel, msg); err != nil {
		return fmt.Errorf("send notification for task %s: %w", snap.ID, err)
	}

	n.logger.Infow("notification sent", "taskID", snap.ID, "channel", snap.OriginChannel, "status", snap.StatusText)
	return nil
}

// NotifyStart sends a notification that a background task has started execution.
func (n *Notification) NotifyStart(ctx context.Context, task *Task) error {
	snap := task.Snapshot()

	if snap.OriginChannel == "" {
		n.logger.Debugw("skip start notification: no origin channel", "taskID", snap.ID)
		return nil
	}

	msg := fmt.Sprintf("Background task started: %s", truncate(snap.Prompt, 50))

	if err := n.notifier.SendMessage(ctx, snap.OriginChannel, msg); err != nil {
		return fmt.Errorf("send start notification for task %s: %w", snap.ID, err)
	}

	n.logger.Infow("start notification sent", "taskID", snap.ID, "channel", snap.OriginChannel)
	return nil
}

func formatNotification(snap TaskSnapshot) string {
	promptSummary := truncate(snap.Prompt, 50)

	switch snap.Status {
	case Done:
		resultSummary := truncate(snap.Result, 500)
		return fmt.Sprintf("Background task completed: %s\nResult: %s", promptSummary, resultSummary)

	case Failed:
		return fmt.Sprintf("Background task failed: %s\nError: %s", promptSummary, snap.Error)

	case Cancelled:
		return fmt.Sprintf("Background task cancelled: %s", promptSummary)

	default:
		return fmt.Sprintf("Background task update [%s]: %s", snap.StatusText, promptSummary)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
