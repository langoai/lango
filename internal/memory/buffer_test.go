package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/session"
	_ "github.com/mattn/go-sqlite3"
)

func newTestBuffer(t *testing.T, gen TextGenerator, msgs []session.Message, msgThreshold, obsThreshold int) *Buffer {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	store := NewStore(client, logger)
	observer := NewObserver(gen, store, logger)
	reflector := NewReflector(gen, store, logger)

	getMessages := func(_ string) ([]session.Message, error) {
		return msgs, nil
	}

	return NewBuffer(observer, reflector, store, msgThreshold, obsThreshold, getMessages, logger)
}

func TestBufferStartStop(t *testing.T) {
	gen := &mockGenerator{response: "test observation"}
	buf := newTestBuffer(t, gen, nil, 100, 200)

	var wg sync.WaitGroup
	buf.Start(&wg)

	// Give the goroutine time to start.
	time.Sleep(10 * time.Millisecond)

	buf.Stop()
	wg.Wait()
}

func TestBufferTriggerProcessesObservation(t *testing.T) {
	messages := make([]session.Message, 20)
	for i := range messages {
		messages[i] = session.Message{
			Role:      "user",
			Content:   "This is a test message with enough content to generate tokens for observation threshold testing purposes",
			Timestamp: time.Now(),
		}
	}

	gen := &mockGenerator{response: "Compressed observation of the conversation."}
	buf := newTestBuffer(t, gen, messages, 10, 100000)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.Trigger("session-buf-1")

	// Allow processing time.
	time.Sleep(100 * time.Millisecond)

	buf.Stop()
	wg.Wait()

	// Verify observation was created.
	ctx := context.Background()
	obs, err := buf.store.ListObservations(ctx, "session-buf-1")
	require.NoError(t, err)
	assert.Len(t, obs, 1)
	assert.Equal(t, "Compressed observation of the conversation.", obs[0].Content)
}

func TestBufferConcurrentTriggers(t *testing.T) {
	messages := make([]session.Message, 10)
	for i := range messages {
		messages[i] = session.Message{
			Role:      "user",
			Content:   "Concurrent test message with some content for token counting",
			Timestamp: time.Now(),
		}
	}

	gen := &mockGenerator{response: "Concurrent observation."}
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	store := NewStore(client, logger)
	observer := NewObserver(gen, store, logger)
	reflector := NewReflector(gen, store, logger)

	getMessages := func(_ string) ([]session.Message, error) {
		return messages, nil
	}

	buf := NewBuffer(observer, reflector, store, 5, 100000, getMessages, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	// Fire many triggers concurrently.
	var triggerWg sync.WaitGroup
	for i := 0; i < 50; i++ {
		triggerWg.Add(1)
		go func() {
			defer triggerWg.Done()
			buf.Trigger("session-concurrent")
		}()
	}
	triggerWg.Wait()

	// Allow processing time.
	time.Sleep(200 * time.Millisecond)

	buf.Stop()
	wg.Wait()

	// No panics means success. Verify at least one observation was created.
	ctx := context.Background()
	obs, err := store.ListObservations(ctx, "session-concurrent")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(obs), 1)
}
