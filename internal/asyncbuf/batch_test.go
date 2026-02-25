package asyncbuf

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBatchBuffer_StartStop(t *testing.T) {
	logger := zap.NewNop().Sugar()
	buf := NewBatchBuffer[int](BatchConfig{}, func(_ []int) {}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.Stop()
	wg.Wait()
}

func TestBatchBuffer_BatchFlush(t *testing.T) {
	logger := zap.NewNop().Sugar()

	var mu sync.Mutex
	var batches [][]int

	buf := NewBatchBuffer[int](BatchConfig{
		QueueSize:    64,
		BatchSize:    5,
		BatchTimeout: 10 * time.Second, // large timeout so flush is triggered by batch size
	}, func(batch []int) {
		cp := make([]int, len(batch))
		copy(cp, batch)
		mu.Lock()
		batches = append(batches, cp)
		mu.Unlock()
	}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	for i := 0; i < 10; i++ {
		buf.Enqueue(i)
	}

	// Wait for batch-size flushes.
	time.Sleep(100 * time.Millisecond)
	buf.Stop()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	// Should have gotten at least 2 batches of 5.
	total := 0
	for _, b := range batches {
		total += len(b)
	}
	assert.Equal(t, 10, total)
}

func TestBatchBuffer_TimeoutFlush(t *testing.T) {
	logger := zap.NewNop().Sugar()

	var mu sync.Mutex
	var received []int

	buf := NewBatchBuffer[int](BatchConfig{
		QueueSize:    64,
		BatchSize:    100, // large batch size so flush is triggered by timeout
		BatchTimeout: 50 * time.Millisecond,
	}, func(batch []int) {
		mu.Lock()
		received = append(received, batch...)
		mu.Unlock()
	}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.Enqueue(1)
	buf.Enqueue(2)
	buf.Enqueue(3)

	// Wait for timeout flush.
	time.Sleep(200 * time.Millisecond)
	buf.Stop()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, received, 3)
	assert.Equal(t, []int{1, 2, 3}, received)
}

func TestBatchBuffer_DrainOnStop(t *testing.T) {
	logger := zap.NewNop().Sugar()

	var mu sync.Mutex
	var received []int

	buf := NewBatchBuffer[int](BatchConfig{
		QueueSize:    256,
		BatchSize:    1000, // large so nothing flushes before stop
		BatchTimeout: 10 * time.Second,
	}, func(batch []int) {
		mu.Lock()
		received = append(received, batch...)
		mu.Unlock()
	}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	for i := 0; i < 20; i++ {
		buf.Enqueue(i)
	}

	// Stop immediately — should drain.
	buf.Stop()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, received, 20)
}

func TestBatchBuffer_DropCounting(t *testing.T) {
	logger := zap.NewNop().Sugar()

	// Tiny queue + slow processor to force drops.
	var processCalls atomic.Int64
	buf := NewBatchBuffer[int](BatchConfig{
		QueueSize:    2,
		BatchSize:    1000,
		BatchTimeout: 10 * time.Second,
	}, func(_ []int) {
		processCalls.Add(1)
	}, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	// Enqueue more than queue capacity.
	for i := 0; i < 10; i++ {
		buf.Enqueue(i)
	}

	buf.Stop()
	wg.Wait()

	assert.Greater(t, buf.DroppedCount(), int64(0))
}

func TestBatchBuffer_DefaultConfig(t *testing.T) {
	logger := zap.NewNop().Sugar()
	buf := NewBatchBuffer[string](BatchConfig{}, func(_ []string) {}, logger)

	assert.Equal(t, 256, cap(buf.queue))
	assert.Equal(t, 32, buf.batchSize)
	assert.Equal(t, 2*time.Second, buf.batchTimeout)
}
