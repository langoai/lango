package cron

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// --- local mocks for delivery ---

type mockChannelSender struct {
	mu       sync.Mutex
	messages []struct{ channel, msg string }
	err      error
}

func (m *mockChannelSender) SendMessage(_ context.Context, channel string, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, struct{ channel, msg string }{channel, message})
	return m.err
}

type mockTypingIndicator struct {
	mu       sync.Mutex
	channels []string
	err      error
	stopped  int
}

func (m *mockTypingIndicator) StartTyping(_ context.Context, channel string) (func(), error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	m.channels = append(m.channels, channel)
	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.stopped++
	}, nil
}

// --- tests ---

func TestNewDelivery(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	sender := &mockChannelSender{}
	typing := &mockTypingIndicator{}

	d := NewDelivery(sender, typing, logger)

	require.NotNil(t, d)
	assert.Equal(t, sender, d.sender)
	assert.Equal(t, typing, d.typing)
}

func TestDelivery_Deliver_NilSender(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	d := NewDelivery(nil, nil, logger)

	result := &JobResult{
		JobID:   "j1",
		JobName: "test",
	}

	err := d.Deliver(context.Background(), result, []string{"ch-1"})
	assert.NoError(t, err)
}

func TestDelivery_Deliver_EmptyTargets(t *testing.T) {
	t.Parallel()

	sender := &mockChannelSender{}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(sender, nil, logger)

	result := &JobResult{
		JobID:   "j1",
		JobName: "test",
	}

	err := d.Deliver(context.Background(), result, nil)
	assert.NoError(t, err)

	sender.mu.Lock()
	assert.Empty(t, sender.messages)
	sender.mu.Unlock()
}

func TestDelivery_Deliver_SingleTarget(t *testing.T) {
	t.Parallel()

	sender := &mockChannelSender{}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(sender, nil, logger)

	result := &JobResult{
		JobID:     "j1",
		JobName:   "my-job",
		Response:  "all good",
		StartedAt: time.Now(),
		Duration:  time.Second,
	}

	err := d.Deliver(context.Background(), result, []string{"slack:general"})
	require.NoError(t, err)

	sender.mu.Lock()
	require.Len(t, sender.messages, 1)
	assert.Equal(t, "slack:general", sender.messages[0].channel)
	assert.Contains(t, sender.messages[0].msg, "my-job")
	assert.Contains(t, sender.messages[0].msg, "all good")
	sender.mu.Unlock()
}

func TestDelivery_Deliver_MultipleTargets(t *testing.T) {
	t.Parallel()

	sender := &mockChannelSender{}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(sender, nil, logger)

	result := &JobResult{
		JobID:    "j1",
		JobName:  "multi-job",
		Response: "done",
	}

	err := d.Deliver(context.Background(), result, []string{"ch-1", "ch-2", "ch-3"})
	require.NoError(t, err)

	sender.mu.Lock()
	assert.Len(t, sender.messages, 3)
	sender.mu.Unlock()
}

func TestDelivery_Deliver_WithError(t *testing.T) {
	t.Parallel()

	sender := &mockChannelSender{}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(sender, nil, logger)

	result := &JobResult{
		JobID:   "j1",
		JobName: "error-job",
		Error:   fmt.Errorf("something went wrong"),
	}

	err := d.Deliver(context.Background(), result, []string{"ch-1"})
	require.NoError(t, err)

	sender.mu.Lock()
	require.Len(t, sender.messages, 1)
	assert.Contains(t, sender.messages[0].msg, "something went wrong")
	sender.mu.Unlock()
}

func TestDelivery_Deliver_SendError(t *testing.T) {
	t.Parallel()

	sender := &mockChannelSender{err: fmt.Errorf("network error")}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(sender, nil, logger)

	result := &JobResult{
		JobID:    "j1",
		JobName:  "fail-delivery",
		Response: "result",
	}

	err := d.Deliver(context.Background(), result, []string{"ch-1", "ch-2"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ch-1")
	assert.Contains(t, err.Error(), "ch-2")
	assert.Contains(t, err.Error(), "network error")
}

func TestDelivery_DeliverStart(t *testing.T) {
	t.Parallel()

	sender := &mockChannelSender{}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(sender, nil, logger)

	d.DeliverStart(context.Background(), "my-cron-job", []string{"ch-1", "ch-2"})

	sender.mu.Lock()
	require.Len(t, sender.messages, 2)
	assert.Contains(t, sender.messages[0].msg, "my-cron-job")
	assert.Contains(t, sender.messages[0].msg, "Starting")
	sender.mu.Unlock()
}

func TestDelivery_DeliverStart_NilSender(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	d := NewDelivery(nil, nil, logger)

	// Should not panic.
	d.DeliverStart(context.Background(), "job", []string{"ch-1"})
}

func TestDelivery_DeliverStart_EmptyTargets(t *testing.T) {
	t.Parallel()

	sender := &mockChannelSender{}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(sender, nil, logger)

	d.DeliverStart(context.Background(), "job", nil)

	sender.mu.Lock()
	assert.Empty(t, sender.messages)
	sender.mu.Unlock()
}

func TestDelivery_DeliverStart_SendError(t *testing.T) {
	t.Parallel()

	sender := &mockChannelSender{err: fmt.Errorf("send failed")}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(sender, nil, logger)

	// Should not panic, just logs the error.
	d.DeliverStart(context.Background(), "job", []string{"ch-1"})
}

func TestDelivery_StartTyping(t *testing.T) {
	t.Parallel()

	typing := &mockTypingIndicator{}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(nil, typing, logger)

	stop := d.StartTyping(context.Background(), []string{"ch-1", "ch-2"})
	require.NotNil(t, stop)

	typing.mu.Lock()
	assert.Len(t, typing.channels, 2)
	typing.mu.Unlock()

	stop()

	typing.mu.Lock()
	assert.Equal(t, 2, typing.stopped)
	typing.mu.Unlock()
}

func TestDelivery_StartTyping_NilTyping(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	d := NewDelivery(nil, nil, logger)

	stop := d.StartTyping(context.Background(), []string{"ch-1"})
	require.NotNil(t, stop)

	// Should be a no-op, should not panic.
	stop()
}

func TestDelivery_StartTyping_EmptyTargets(t *testing.T) {
	t.Parallel()

	typing := &mockTypingIndicator{}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(nil, typing, logger)

	stop := d.StartTyping(context.Background(), nil)
	require.NotNil(t, stop)
	stop()
}

func TestDelivery_StartTyping_Error(t *testing.T) {
	t.Parallel()

	typing := &mockTypingIndicator{err: fmt.Errorf("typing failed")}
	logger := zap.NewNop().Sugar()
	d := NewDelivery(nil, typing, logger)

	stop := d.StartTyping(context.Background(), []string{"ch-1"})
	require.NotNil(t, stop)

	// Should not panic even if typing start failed.
	stop()
}

func TestFormatDeliveryMessage_Success(t *testing.T) {
	t.Parallel()

	result := &JobResult{
		JobName:  "test-job",
		Response: "everything is fine",
	}

	msg := formatDeliveryMessage(result)

	assert.Contains(t, msg, "[Cron] test-job")
	assert.Contains(t, msg, "everything is fine")
	assert.NotContains(t, msg, "Error")
}

func TestFormatDeliveryMessage_Error(t *testing.T) {
	t.Parallel()

	result := &JobResult{
		JobName: "fail-job",
		Error:   fmt.Errorf("bad things happened"),
	}

	msg := formatDeliveryMessage(result)

	assert.Contains(t, msg, "[Cron] fail-job")
	assert.Contains(t, msg, "Error")
	assert.Contains(t, msg, "bad things happened")
}
