package alerting

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
)

func TestWebhookDelivery_Send(t *testing.T) {
	t.Parallel()

	var received map[string]interface{}
	var mu sync.Mutex

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := NewWebhookDelivery(srv.URL)

	evt := eventbus.AlertEvent{
		Type:     "policy_block_rate",
		Severity: "warning",
		Message:  "threshold exceeded",
		Details:  map[string]interface{}{"count": float64(15)},
	}

	err := wh.Send(t.Context(), evt)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "policy_block_rate", received["type"])
	assert.Equal(t, "warning", received["severity"])
}

func TestWebhookDelivery_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	wh := NewWebhookDelivery(srv.URL)
	err := wh.Send(t.Context(), eventbus.AlertEvent{Severity: "warning"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestDeliveryRouter_FanOut(t *testing.T) {
	t.Parallel()

	var calls int
	var mu sync.Mutex

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	bus := eventbus.New()
	NewDeliveryRouter(bus, []config.AlertDeliveryConfig{
		{Type: "webhook", WebhookURL: srv.URL, MinSeverity: "warning"},
	})

	bus.Publish(eventbus.AlertEvent{
		Type:      "test",
		Severity:  "warning",
		Timestamp: time.Now(),
	})

	// Give async delivery a moment.
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 1, calls)
	mu.Unlock()
}

func TestDeliveryRouter_MinSeverityFilter(t *testing.T) {
	t.Parallel()

	var calls int
	var mu sync.Mutex

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	bus := eventbus.New()
	NewDeliveryRouter(bus, []config.AlertDeliveryConfig{
		{Type: "webhook", WebhookURL: srv.URL, MinSeverity: "critical"},
	})

	// Warning should be filtered out.
	bus.Publish(eventbus.AlertEvent{
		Type:      "test",
		Severity:  "warning",
		Timestamp: time.Now(),
	})

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 0, calls, "warning should be filtered by minSeverity=critical")
	mu.Unlock()

	// Critical should pass.
	bus.Publish(eventbus.AlertEvent{
		Type:      "test",
		Severity:  "critical",
		Timestamp: time.Now(),
	})

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 1, calls, "critical should pass minSeverity=critical filter")
	mu.Unlock()
}

func TestDeliveryRouter_UnknownType_Skipped(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	r := NewDeliveryRouter(bus, []config.AlertDeliveryConfig{
		{Type: "unknown_channel"},
	})

	assert.Empty(t, r.channels, "unknown channel type should be skipped")
}
