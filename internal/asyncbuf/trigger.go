package asyncbuf

import (
	"sync"

	"go.uber.org/zap"
)

// ProcessFunc is called for each individual item.
type ProcessFunc[T any] func(item T)

// TriggerConfig holds configuration for a TriggerBuffer.
type TriggerConfig struct {
	QueueSize int
}

// TriggerBuffer processes items one at a time on a background goroutine.
// It follows the Start -> Enqueue -> Stop lifecycle.
type TriggerBuffer[T any] struct {
	process ProcessFunc[T]
	queue   chan T
	stopCh  chan struct{}
	done    chan struct{}
	logger  *zap.SugaredLogger
}

// NewTriggerBuffer creates a new per-item async buffer.
func NewTriggerBuffer[T any](cfg TriggerConfig, fn ProcessFunc[T], logger *zap.SugaredLogger) *TriggerBuffer[T] {
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 16
	}

	return &TriggerBuffer[T]{
		process: fn,
		queue:   make(chan T, cfg.QueueSize),
		stopCh:  make(chan struct{}),
		done:    make(chan struct{}),
		logger:  logger,
	}
}

// Start launches the background goroutine. The WaitGroup is incremented
// so callers can wait for graceful shutdown.
func (b *TriggerBuffer[T]) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(b.done)
		b.run()
	}()
}

// Enqueue submits an item. Non-blocking; drops if the queue is full.
func (b *TriggerBuffer[T]) Enqueue(item T) {
	select {
	case b.queue <- item:
	default:
		b.logger.Debugw("trigger buffer queue full, dropping item")
	}
}

// Stop signals the background goroutine to drain and exit.
func (b *TriggerBuffer[T]) Stop() {
	close(b.stopCh)
	<-b.done
}

func (b *TriggerBuffer[T]) run() {
	for {
		select {
		case item := <-b.queue:
			b.process(item)
		case <-b.stopCh:
			// Drain remaining items.
			for {
				select {
				case item := <-b.queue:
					b.process(item)
				default:
					return
				}
			}
		}
	}
}
