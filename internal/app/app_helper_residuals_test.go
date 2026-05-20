package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/eventbus"
)

func TestCompactionSyncHolderWaiterSwapAndNilReset(t *testing.T) {
	t.Parallel()

	holder := newCompactionSyncHolder()
	ok, waited := holder.WaitForSession(context.Background(), "sess-empty", time.Second)
	assert.True(t, ok)
	assert.Zero(t, waited)

	waiter := &appHelperResidualsCompactionWaiter{ok: false, waited: 42 * time.Millisecond}
	holder.SetWaiter(waiter)

	ok, waited = holder.WaitForSession(context.Background(), "sess-1", time.Second)
	assert.False(t, ok)
	assert.Equal(t, 42*time.Millisecond, waited)
	assert.Equal(t, "sess-1", waiter.sessionKey)
	assert.Equal(t, time.Second, waiter.timeout)

	holder.SetWaiter(nil)
	ok, waited = holder.WaitForSession(context.Background(), "sess-reset", time.Second)
	assert.True(t, ok)
	assert.Zero(t, waited)
}

func TestWireTeamMetricsBridgeAccumulatesDelegationAndCompletionCounters(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	metrics := &TeamMetrics{}
	wireTeamMetricsBridge(bus, metrics, logger())

	bus.Publish(eventbus.TeamTaskDelegatedEvent{
		TeamID:   "team-1",
		ToolName: "search",
		Workers:  3,
	})
	bus.Publish(eventbus.TeamTaskDelegatedEvent{
		TeamID:   "team-1",
		ToolName: "summarize",
		Workers:  2,
	})
	bus.Publish(eventbus.TeamTaskCompletedEvent{
		TeamID:     "team-1",
		ToolName:   "search",
		Successful: 2,
		Failed:     1,
		Duration:   5 * time.Second,
	})

	assert.Equal(t, int64(2), metrics.Delegations.Load())
	assert.Equal(t, int64(5), metrics.TotalWorkers.Load())
	assert.Equal(t, int64(2), metrics.TotalSuccesses.Load())
	assert.Equal(t, int64(1), metrics.TotalFailures.Load())
}

func TestTruncateID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "short", truncateID("short", 8))
	assert.Equal(t, "12345678", truncateID("123456789abcdef", 8))
	assert.Equal(t, "", truncateID("", 8))
}

type appHelperResidualsCompactionWaiter struct {
	ok         bool
	waited     time.Duration
	sessionKey string
	timeout    time.Duration
}

func (w *appHelperResidualsCompactionWaiter) WaitForSession(_ context.Context, key string, timeout time.Duration) (bool, time.Duration) {
	w.sessionKey = key
	w.timeout = timeout
	return w.ok, w.waited
}
