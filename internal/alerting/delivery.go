package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/logging"
)

var deliveryLogger = logging.SubsystemSugar("alerting.delivery")

// DeliveryChannel delivers alerts to an external system.
type DeliveryChannel interface {
	// Send delivers a single alert event. Implementations should not retry.
	Send(ctx context.Context, evt eventbus.AlertEvent) error
	// Type returns the channel type identifier (e.g. "webhook").
	Type() string
}

// WebhookDelivery delivers alerts as JSON HTTP POST requests.
type WebhookDelivery struct {
	url    string
	client *http.Client
}

// NewWebhookDelivery creates a webhook delivery channel targeting the given URL.
func NewWebhookDelivery(url string) *WebhookDelivery {
	return &WebhookDelivery{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *WebhookDelivery) Type() string { return "webhook" }

// Send posts the alert event as JSON to the webhook URL.
func (w *WebhookDelivery) Send(ctx context.Context, evt eventbus.AlertEvent) error {
	body, err := json.Marshal(map[string]interface{}{
		"type":       evt.Type,
		"severity":   evt.Severity,
		"message":    evt.Message,
		"details":    evt.Details,
		"sessionKey": evt.SessionKey,
		"timestamp":  evt.Timestamp.Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("marshal alert: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook POST: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// severityRank returns a numeric rank for severity comparison.
// Higher rank = more severe.
func severityRank(s string) int {
	switch s {
	case "critical":
		return 2
	case "warning":
		return 1
	default:
		return 0
	}
}

// DeliveryRouter subscribes to AlertEvent on the EventBus and fans out
// to configured delivery channels, filtering by minimum severity.
type DeliveryRouter struct {
	channels    []channelEntry
}

type channelEntry struct {
	channel     DeliveryChannel
	minSeverity int // numeric rank from severityRank
}

// NewDeliveryRouter creates a router with the given channels and subscribes to the bus.
func NewDeliveryRouter(bus *eventbus.Bus, channels []config.AlertDeliveryConfig) *DeliveryRouter {
	r := &DeliveryRouter{}

	for _, cfg := range channels {
		var ch DeliveryChannel
		switch cfg.Type {
		case "webhook":
			if cfg.WebhookURL == "" {
				deliveryLogger.Warnw("webhook channel missing URL, skipping")
				continue
			}
			ch = NewWebhookDelivery(cfg.WebhookURL)
		default:
			deliveryLogger.Warnw("unknown delivery channel type, skipping", "type", cfg.Type)
			continue
		}

		r.channels = append(r.channels, channelEntry{
			channel:     ch,
			minSeverity: severityRank(cfg.MinSeverity),
		})
	}

	if len(r.channels) > 0 {
		eventbus.SubscribeTyped[eventbus.AlertEvent](bus, r.handle)
		deliveryLogger.Infow("delivery router started", "channels", len(r.channels))
	}

	return r
}

func (r *DeliveryRouter) handle(evt eventbus.AlertEvent) {
	rank := severityRank(evt.Severity)

	for _, entry := range r.channels {
		if rank < entry.minSeverity {
			continue
		}

		// Dispatch asynchronously to avoid blocking the synchronous EventBus.
		go func(ch DeliveryChannel) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := ch.Send(ctx, evt); err != nil {
				deliveryLogger.Warnw("alert delivery failed",
					"channel", ch.Type(),
					"alertType", evt.Type,
					"error", err,
				)
			}
		}(entry.channel)
	}
}

