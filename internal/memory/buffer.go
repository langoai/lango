package memory

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/session"
)

// MessageProvider retrieves messages for a session key.
type MessageProvider func(sessionKey string) ([]session.Message, error)

// MessageCompactor replaces observed messages with a summary to reduce session size.
type MessageCompactor func(sessionKey string, upToIndex int, summary string) error

// Buffer manages background observation and reflection processing.
type Buffer struct {
	observer  *Observer
	reflector *Reflector
	store     *Store

	messageTokenThreshold     int
	observationTokenThreshold int
	getMessages               MessageProvider
	compactor                 MessageCompactor // optional: compact observed messages

	// lastObserved tracks the last observed message index per session.
	mu           sync.Mutex
	lastObserved map[string]int

	triggerCh chan string
	stopCh    chan struct{}
	done      chan struct{}
	logger    *zap.SugaredLogger
}

// NewBuffer creates a new asynchronous observation buffer.
func NewBuffer(
	observer *Observer,
	reflector *Reflector,
	store *Store,
	msgThreshold, obsThreshold int,
	getMessages MessageProvider,
	logger *zap.SugaredLogger,
) *Buffer {
	return &Buffer{
		observer:                  observer,
		reflector:                 reflector,
		store:                     store,
		messageTokenThreshold:     msgThreshold,
		observationTokenThreshold: obsThreshold,
		getMessages:               getMessages,
		lastObserved:              make(map[string]int),
		triggerCh:                 make(chan string, 16),
		stopCh:                    make(chan struct{}),
		done:                      make(chan struct{}),
		logger:                    logger,
	}
}

// Start launches the background goroutine. The WaitGroup is incremented so
// callers can wait for graceful shutdown.
func (b *Buffer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(b.done)
		b.run()
	}()
}

// Trigger sends a non-blocking signal to process the given session.
func (b *Buffer) Trigger(sessionKey string) {
	select {
	case b.triggerCh <- sessionKey:
	default:
		b.logger.Debugw("buffer trigger dropped (channel full)", "sessionKey", sessionKey)
	}
}

// SetCompactor enables message compaction after observation. When set,
// observed messages are replaced with a summary message in the session,
// effectively reducing the session's memory footprint.
func (b *Buffer) SetCompactor(c MessageCompactor) {
	b.compactor = c
}

// Stop signals the background goroutine to stop and waits for completion.
func (b *Buffer) Stop() {
	close(b.stopCh)
	<-b.done
}

func (b *Buffer) run() {
	for {
		select {
		case sessionKey := <-b.triggerCh:
			b.process(sessionKey)
		case <-b.stopCh:
			// Drain remaining triggers before exiting.
			for {
				select {
				case sessionKey := <-b.triggerCh:
					b.process(sessionKey)
				default:
					return
				}
			}
		}
	}
}

func (b *Buffer) process(sessionKey string) {
	ctx := context.Background()

	messages, err := b.getMessages(sessionKey)
	if err != nil {
		b.logger.Errorw("get messages for observation", "sessionKey", sessionKey, "error", err)
		return
	}

	lastIdx := b.getLastObserved(sessionKey)

	// Check if un-observed messages exceed the token threshold.
	if lastIdx+1 < len(messages) {
		unobserved := messages[lastIdx+1:]
		tokens := CountMessagesTokens(unobserved)
		if tokens >= b.messageTokenThreshold {
			obs, err := b.observer.Observe(ctx, sessionKey, messages, lastIdx)
			if err != nil {
				b.logger.Errorw("observer failed", "sessionKey", sessionKey, "error", err)
				return
			}
			if obs != nil {
				b.setLastObserved(sessionKey, obs.SourceEndIndex)

				// Compact observed messages if compactor is configured.
				if b.compactor != nil && obs.SourceEndIndex > 0 {
					if err := b.compactor(sessionKey, obs.SourceEndIndex, obs.Content); err != nil {
						b.logger.Warnw("compaction failed", "sessionKey", sessionKey, "error", err)
					} else {
						// Reset lastObserved since message indices shifted after compaction.
						b.setLastObserved(sessionKey, 0)
						b.logger.Debugw("messages compacted",
							"sessionKey", sessionKey,
							"upToIndex", obs.SourceEndIndex,
						)
					}
				}
			}
		}
	}

	// Check if accumulated observation tokens exceed the reflection threshold.
	observations, err := b.store.ListObservations(ctx, sessionKey)
	if err != nil {
		b.logger.Errorw("list observations for reflection check", "sessionKey", sessionKey, "error", err)
		return
	}

	var totalObsTokens int
	for _, obs := range observations {
		totalObsTokens += obs.TokenCount
	}

	if totalObsTokens >= b.observationTokenThreshold {
		_, err := b.reflector.Reflect(ctx, sessionKey)
		if err != nil {
			b.logger.Errorw("reflector failed", "sessionKey", sessionKey, "error", err)
		}
	}
}

func (b *Buffer) getLastObserved(sessionKey string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if idx, ok := b.lastObserved[sessionKey]; ok {
		return idx
	}

	// Rebuild from stored observations.
	ctx := context.Background()
	observations, err := b.store.ListObservations(ctx, sessionKey)
	if err != nil {
		b.logger.Errorw("rebuild last observed index", "sessionKey", sessionKey, "error", err)
		b.lastObserved[sessionKey] = -1
		return -1
	}

	maxIdx := -1
	for _, obs := range observations {
		if obs.SourceEndIndex > maxIdx {
			maxIdx = obs.SourceEndIndex
		}
	}

	// Also check reflections (they were built from observations with known ranges).
	b.lastObserved[sessionKey] = maxIdx
	return maxIdx
}

func (b *Buffer) setLastObserved(sessionKey string, idx int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastObserved[sessionKey] = idx
}
