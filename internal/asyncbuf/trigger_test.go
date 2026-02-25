package asyncbuf

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTriggerBuffer_StartStop(t *testing.T) {
	logger := zap.NewNop().Sugar()
	buf := NewTriggerBuffer[int](TriggerConfig{}, func(_ int) {}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.Stop()
	wg.Wait()
}

func TestTriggerBuffer_ProcessesItems(t *testing.T) {
	logger := zap.NewNop().Sugar()

	var mu sync.Mutex
	var received []string

	buf := NewTriggerBuffer[string](TriggerConfig{QueueSize: 32}, func(item string) {
		mu.Lock()
		received = append(received, item)
		mu.Unlock()
	}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.Enqueue("alpha")
	buf.Enqueue("beta")
	buf.Enqueue("gamma")

	// Wait for processing.
	time.Sleep(100 * time.Millisecond)
	buf.Stop()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, received)
}

func TestTriggerBuffer_DrainOnStop(t *testing.T) {
	logger := zap.NewNop().Sugar()

	var count atomic.Int64

	// Block processing until stop to ensure items are in the queue.
	gate := make(chan struct{})
	buf := NewTriggerBuffer[int](TriggerConfig{QueueSize: 64}, func(_ int) {
		<-gate
		count.Add(1)
	}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	// Enqueue items while processor is blocked.
	for i := 0; i < 5; i++ {
		buf.Enqueue(i)
	}

	// Unblock and stop — should drain all.
	close(gate)
	buf.Stop()
	wg.Wait()

	assert.Equal(t, int64(5), count.Load())
}

func TestTriggerBuffer_DefaultConfig(t *testing.T) {
	logger := zap.NewNop().Sugar()
	buf := NewTriggerBuffer[int](TriggerConfig{}, func(_ int) {}, logger)

	assert.Equal(t, 16, cap(buf.queue))
}

func TestTriggerBuffer_ConcurrentEnqueue(t *testing.T) {
	logger := zap.NewNop().Sugar()

	var count atomic.Int64

	buf := NewTriggerBuffer[int](TriggerConfig{QueueSize: 256}, func(_ int) {
		count.Add(1)
	}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	var enqueueWg sync.WaitGroup
	for i := 0; i < 50; i++ {
		enqueueWg.Add(1)
		go func(v int) {
			defer enqueueWg.Done()
			buf.Enqueue(v)
		}(i)
	}
	enqueueWg.Wait()

	time.Sleep(200 * time.Millisecond)
	buf.Stop()
	wg.Wait()

	assert.Equal(t, int64(50), count.Load())
}
