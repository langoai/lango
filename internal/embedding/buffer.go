package embedding

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/asyncbuf"
)

// EmbedRequest represents a request to embed and store a text.
type EmbedRequest struct {
	ID         string
	Collection string
	Content    string
	Metadata   map[string]string
}

// EmbeddingBuffer collects embed requests and processes them in batches
// on a background goroutine. It follows the same lifecycle pattern as
// memory.Buffer: Start -> Enqueue -> Stop.
type EmbeddingBuffer struct {
	provider EmbeddingProvider
	store    VectorStore
	inner    *asyncbuf.BatchBuffer[EmbedRequest]
	logger   *zap.SugaredLogger
}

// NewEmbeddingBuffer creates a new asynchronous embedding buffer.
func NewEmbeddingBuffer(
	provider EmbeddingProvider,
	store VectorStore,
	logger *zap.SugaredLogger,
) *EmbeddingBuffer {
	b := &EmbeddingBuffer{
		provider: provider,
		store:    store,
		logger:   logger,
	}
	b.inner = asyncbuf.NewBatchBuffer[EmbedRequest](asyncbuf.BatchConfig{
		QueueSize:    256,
		BatchSize:    32,
		BatchTimeout: 2 * time.Second,
	}, b.processBatch, logger)
	return b
}

// Start launches the background goroutine. The WaitGroup is incremented
// so callers can wait for graceful shutdown.
func (b *EmbeddingBuffer) Start(wg *sync.WaitGroup) {
	b.inner.Start(wg)
}

// Enqueue submits an embed request. Non-blocking; drops if the queue is full.
func (b *EmbeddingBuffer) Enqueue(req EmbedRequest) {
	b.inner.Enqueue(req)
}

// DroppedCount returns the total number of dropped embed requests.
func (b *EmbeddingBuffer) DroppedCount() int64 {
	return b.inner.DroppedCount()
}

// Stop signals the background goroutine to drain and exit.
func (b *EmbeddingBuffer) Stop() {
	b.inner.Stop()
}

func (b *EmbeddingBuffer) processBatch(batch []EmbedRequest) {
	ctx := context.Background()

	texts := make([]string, len(batch))
	for i, r := range batch {
		texts[i] = r.Content
	}

	embeddings, err := b.provider.Embed(ctx, texts)
	if err != nil {
		b.logger.Errorw("batch embedding failed", "count", len(batch), "error", err)
		return
	}

	if len(embeddings) != len(batch) {
		b.logger.Errorw("embedding count mismatch", "expected", len(batch), "got", len(embeddings))
		return
	}

	records := make([]VectorRecord, len(batch))
	for i, r := range batch {
		records[i] = VectorRecord{
			ID:         r.ID,
			Collection: r.Collection,
			Embedding:  embeddings[i],
			Metadata:   r.Metadata,
		}
	}

	if err := b.store.Upsert(ctx, records); err != nil {
		b.logger.Errorw("batch upsert failed", "count", len(records), "error", err)
	}
}
