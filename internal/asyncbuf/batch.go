package asyncbuf

import (
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// ProcessBatchFunc is called with a batch of items to process.
type ProcessBatchFunc[T any] func(batch []T)

// BatchConfig holds configuration for a BatchBuffer.
type BatchConfig struct {
	QueueSize    int
	BatchSize    int
	BatchTimeout time.Duration
}

// BatchBuffer collects items and processes them in batches on a background
// goroutine. It follows the Start -> Enqueue -> Stop lifecycle.
type BatchBuffer[T any] struct {
	processBatch ProcessBatchFunc[T]
	queue        chan T
	stopCh       chan struct{}
	done         chan struct{}
	batchSize    int
	batchTimeout time.Duration
	dropCount    atomic.Int64
	logger       *zap.SugaredLogger
}

// NewBatchBuffer creates a new batch-oriented async buffer.
func NewBatchBuffer[T any](cfg BatchConfig, fn ProcessBatchFunc[T], logger *zap.SugaredLogger) *BatchBuffer[T] {
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 256
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 32
	}
	if cfg.BatchTimeout <= 0 {
		cfg.BatchTimeout = 2 * time.Second
	}

	return &BatchBuffer[T]{
		processBatch: fn,
		queue:        make(chan T, cfg.QueueSize),
		stopCh:       make(chan struct{}),
		done:         make(chan struct{}),
		batchSize:    cfg.BatchSize,
		batchTimeout: cfg.BatchTimeout,
		logger:       logger,
	}
}

// Start launches the background goroutine. The WaitGroup is incremented
// so callers can wait for graceful shutdown.
func (b *BatchBuffer[T]) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(b.done)
		b.run()
	}()
}

// Enqueue submits an item. Non-blocking; drops if the queue is full.
func (b *BatchBuffer[T]) Enqueue(item T) {
	select {
	case b.queue <- item:
	default:
		b.dropCount.Add(1)
		b.logger.Warnw("batch buffer queue full, dropping item",
			"totalDropped", b.dropCount.Load())
	}
}

// DroppedCount returns the total number of dropped items.
func (b *BatchBuffer[T]) DroppedCount() int64 {
	return b.dropCount.Load()
}

// Stop signals the background goroutine to drain and exit.
func (b *BatchBuffer[T]) Stop() {
	close(b.stopCh)
	<-b.done
}

func (b *BatchBuffer[T]) run() {
	timer := time.NewTimer(b.batchTimeout)
	defer timer.Stop()

	var batch []T

	flush := func() {
		if len(batch) == 0 {
			return
		}
		b.processBatch(batch)
		batch = batch[:0]
	}

	for {
		select {
		case item := <-b.queue:
			batch = append(batch, item)
			if len(batch) >= b.batchSize {
				flush()
				timer.Reset(b.batchTimeout)
			}

		case <-timer.C:
			flush()
			timer.Reset(b.batchTimeout)

		case <-b.stopCh:
			// Drain remaining items.
			for {
				select {
				case item := <-b.queue:
					batch = append(batch, item)
				default:
					flush()
					return
				}
			}
		}
	}
}
