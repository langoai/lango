package graph

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/asyncbuf"
)

// GraphRequest represents a request to add triples to the graph.
type GraphRequest struct {
	Triples []Triple
}

// GraphBuffer collects graph update requests and processes them in batches
// on a background goroutine. It follows the same lifecycle pattern as
// embedding.EmbeddingBuffer: Start -> Enqueue -> Stop.
//
// Note: GraphRequest items are expanded into individual Triples for batch
// processing, so the BatchBuffer operates on Triple slices internally.
type GraphBuffer struct {
	store  Store
	inner  *asyncbuf.BatchBuffer[GraphRequest]
	logger *zap.SugaredLogger
}

// NewGraphBuffer creates a new asynchronous graph update buffer.
func NewGraphBuffer(store Store, logger *zap.SugaredLogger) *GraphBuffer {
	b := &GraphBuffer{
		store:  store,
		logger: logger,
	}
	b.inner = asyncbuf.NewBatchBuffer[GraphRequest](asyncbuf.BatchConfig{
		QueueSize:    256,
		BatchSize:    64,
		BatchTimeout: 2 * time.Second,
	}, b.processBatchRequests, logger)
	return b
}

// Start launches the background goroutine. The WaitGroup is incremented
// so callers can wait for graceful shutdown.
func (b *GraphBuffer) Start(wg *sync.WaitGroup) {
	b.inner.Start(wg)
}

// Enqueue submits a graph update request. Non-blocking; drops if the queue is full.
func (b *GraphBuffer) Enqueue(req GraphRequest) {
	b.inner.Enqueue(req)
}

// DroppedCount returns the total number of dropped graph requests.
func (b *GraphBuffer) DroppedCount() int64 {
	return b.inner.DroppedCount()
}

// Stop signals the background goroutine to drain and exit.
func (b *GraphBuffer) Stop() {
	b.inner.Stop()
}

// processBatchRequests expands GraphRequests into triples and stores them.
func (b *GraphBuffer) processBatchRequests(batch []GraphRequest) {
	var triples []Triple
	for _, req := range batch {
		triples = append(triples, req.Triples...)
	}
	if len(triples) == 0 {
		return
	}
	b.processBatch(triples)
}

func (b *GraphBuffer) processBatch(batch []Triple) {
	ctx := context.Background()

	if err := b.store.AddTriples(ctx, batch); err != nil {
		b.logger.Errorw("batch graph update error", "count", len(batch), "error", err)
	}
}
