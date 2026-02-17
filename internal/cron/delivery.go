package cron

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ChannelSender sends a message to a specific channel.
// This avoids import cycles -- wiring.go will provide the concrete implementation.
type ChannelSender interface {
	SendMessage(ctx context.Context, channel string, message string) error
}

// Delivery handles dispatching job results to configured channels.
type Delivery struct {
	sender ChannelSender
	logger *zap.SugaredLogger
}

// NewDelivery creates a new Delivery instance.
// If sender is nil, delivery will be a no-op (results are only stored in history).
func NewDelivery(sender ChannelSender, logger *zap.SugaredLogger) *Delivery {
	return &Delivery{
		sender: sender,
		logger: logger,
	}
}

// Deliver sends a job result to the specified target channels.
func (d *Delivery) Deliver(ctx context.Context, result *JobResult, targets []string) error {
	if d.sender == nil {
		d.logger.Warnw("no channel sender configured, skipping delivery",
			"job", result.JobName,
			"targets", targets,
		)
		return nil
	}

	if len(targets) == 0 {
		return nil
	}

	msg := formatDeliveryMessage(result)

	var errs []string
	for _, target := range targets {
		if err := d.sender.SendMessage(ctx, target, msg); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", target, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("deliver to channels: %s", strings.Join(errs, "; "))
	}
	return nil
}

// DeliverStart sends a notification that a cron job has started execution.
func (d *Delivery) DeliverStart(ctx context.Context, jobName string, targets []string) {
	if d.sender == nil {
		d.logger.Warn("no channel sender configured, skipping start notification",
			"job", jobName,
		)
		return
	}

	if len(targets) == 0 {
		return
	}

	msg := fmt.Sprintf("[Cron] Starting: %s", jobName)

	for _, target := range targets {
		if err := d.sender.SendMessage(ctx, target, msg); err != nil {
			d.logger.Warnw("deliver start notification",
				"job", jobName,
				"target", target,
				"error", err,
			)
		}
	}
}

// formatDeliveryMessage formats a JobResult into a human-readable message.
func formatDeliveryMessage(result *JobResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[Cron] %s\n", result.JobName))

	if result.Error != nil {
		sb.WriteString(fmt.Sprintf("Error: %v", result.Error))
	} else {
		sb.WriteString(result.Response)
	}

	return sb.String()
}
