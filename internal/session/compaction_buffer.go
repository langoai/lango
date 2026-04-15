package session

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/asyncbuf"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/types"
)

// MessageCompactor is the subset of Store the buffer needs to perform
// compaction and measure its effect. Both EntStore and test stubs can
// satisfy it.
type MessageCompactor interface {
	Get(key string) (*Session, error)
	CompactMessages(key string, upToIndex int, summary string) error
}

// CompactionJob carries a single compaction request.
type CompactionJob struct {
	Key       string
	UpToIndex int
}

// CompactionBuffer processes compaction jobs asynchronously and exposes a
// per-session synchronization handle consumers can wait on before starting
// a new turn.
type CompactionBuffer struct {
	store  MessageCompactor
	bus    *eventbus.Bus
	logger *zap.SugaredLogger
	inner  *asyncbuf.TriggerBuffer[CompactionJob]

	mu       sync.Mutex
	inFlight map[string]chan struct{}
}

// NewCompactionBuffer creates a new compaction buffer. A nil bus is allowed
// for tests; completion events are silently dropped in that case.
func NewCompactionBuffer(store MessageCompactor, bus *eventbus.Bus, logger *zap.SugaredLogger) *CompactionBuffer {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	b := &CompactionBuffer{
		store:    store,
		bus:      bus,
		logger:   logger,
		inFlight: make(map[string]chan struct{}),
	}
	b.inner = asyncbuf.NewTriggerBuffer[CompactionJob](asyncbuf.TriggerConfig{
		QueueSize: 32,
	}, b.process, logger)
	return b
}

// Start launches the background worker.
func (b *CompactionBuffer) Start(wg *sync.WaitGroup) {
	b.inner.Start(wg)
}

// Stop signals graceful shutdown; pending jobs drain before exit.
func (b *CompactionBuffer) Stop() {
	b.inner.Stop()
}

// Drain waits up to timeout for pending jobs to finish. Returns nil if all
// pending and in-flight jobs complete, or an error on timeout.
func (b *CompactionBuffer) Drain(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		b.inner.Stop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("compaction buffer drain timeout after %s", timeout)
	}
}

// EnqueueCompaction schedules a compaction job. Non-blocking. The per-session
// in-flight channel is registered here so a concurrent turn can WaitForSession
// immediately after the caller returns from EnqueueCompaction, even if the
// worker has not yet pulled the job off the queue.
func (b *CompactionBuffer) EnqueueCompaction(key string, upToIndex int) {
	if key == "" {
		return
	}
	b.registerInFlight(key)
	b.inner.Enqueue(CompactionJob{Key: key, UpToIndex: upToIndex})
}

// WaitForSession blocks up to timeout waiting for any in-flight compaction
// for key to complete. Returns (true, 0) if the wait completed within the
// timeout or no compaction was in flight. Returns (false, elapsed) on
// timeout — caller should emit CompactionSlowEvent.
func (b *CompactionBuffer) WaitForSession(ctx context.Context, key string, timeout time.Duration) (bool, time.Duration) {
	b.mu.Lock()
	ch, ok := b.inFlight[key]
	b.mu.Unlock()
	if !ok {
		return true, 0
	}
	start := time.Now()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ch:
		return true, time.Since(start)
	case <-timer.C:
		return false, timeout
	case <-ctx.Done():
		return false, time.Since(start)
	}
}

func (b *CompactionBuffer) registerInFlight(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if existing, ok := b.inFlight[key]; ok {
		// An older job is still in flight. Replace the channel so the new
		// enqueue has its own completion handle; close the old one since a
		// newer job supersedes it for any subsequent WaitForSession callers.
		close(existing)
	}
	b.inFlight[key] = make(chan struct{})
}

func (b *CompactionBuffer) finalizeInFlight(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch, ok := b.inFlight[key]
	if !ok {
		return
	}
	close(ch)
	delete(b.inFlight, key)
}

func (b *CompactionBuffer) process(job CompactionJob) {
	defer b.finalizeInFlight(job.Key)

	before, err := b.store.Get(job.Key)
	if err != nil {
		b.logger.Warnw("compaction get session failed", "key", job.Key, "error", err)
		return
	}
	tokensBefore := estimateHistoryTokens(before.History)

	upTo := job.UpToIndex
	if upTo < 0 || upTo >= len(before.History) {
		upTo = len(before.History)/2 - 1
	}
	if upTo < 0 {
		b.logger.Debugw("compaction skipped — not enough history", "key", job.Key, "history", len(before.History))
		return
	}

	summary := fmt.Sprintf("Compacted %d earlier messages.", upTo+1)
	if err := b.store.CompactMessages(job.Key, upTo, summary); err != nil {
		b.logger.Warnw("compaction failed", "key", job.Key, "error", err)
		return
	}

	after, err := b.store.Get(job.Key)
	if err != nil {
		b.logger.Warnw("compaction re-read session failed", "key", job.Key, "error", err)
		return
	}
	tokensAfter := estimateHistoryTokens(after.History)

	reclaimed := tokensBefore - tokensAfter
	if reclaimed < 0 {
		reclaimed = 0
	}
	summaryTokens := types.EstimateTokens(summary)

	if b.bus != nil {
		b.bus.Publish(eventbus.CompactionCompletedEvent{
			SessionKey:      job.Key,
			UpToIndex:       upTo,
			SummaryTokens:   summaryTokens,
			ReclaimedTokens: reclaimed,
			Timestamp:       time.Now(),
		})
	}
	slog.Debug("compaction completed",
		"key", job.Key,
		"upTo", upTo,
		"reclaimed", reclaimed,
		"summaryTokens", summaryTokens,
	)
}

// estimateHistoryTokens approximates total tokens across a message slice
// using the same rough per-message overhead the learning buffer uses.
func estimateHistoryTokens(msgs []Message) int {
	total := 0
	for _, m := range msgs {
		total += 4 + types.EstimateTokens(m.Content)
	}
	return total
}
