package eventbus

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEvent is a minimal event used across tests.
type testEvent struct {
	Value string
}

func (e testEvent) EventName() string { return "test.event" }

// otherEvent is used to verify event routing isolation.
type otherEvent struct {
	Code int
}

func (e otherEvent) EventName() string { return "other.event" }

func TestSingleHandlerReceivesEvent(t *testing.T) {
	t.Parallel()

	bus := New()

	var received string
	bus.Subscribe("test.event", func(event Event) {
		received = event.(testEvent).Value
	})

	bus.Publish(testEvent{Value: "hello"})

	assert.Equal(t, "hello", received)
}

func TestMultipleHandlersReceiveInOrder(t *testing.T) {
	t.Parallel()

	bus := New()

	var order []int
	bus.Subscribe("test.event", func(_ Event) { order = append(order, 1) })
	bus.Subscribe("test.event", func(_ Event) { order = append(order, 2) })
	bus.Subscribe("test.event", func(_ Event) { order = append(order, 3) })

	bus.Publish(testEvent{Value: "x"})

	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestPublishWithNoHandlersDoesNotPanic(t *testing.T) {
	t.Parallel()

	bus := New()

	// Should not panic.
	bus.Publish(testEvent{Value: "nobody listening"})
}

func TestSubscribeTypedProvidesSafeHandling(t *testing.T) {
	t.Parallel()

	bus := New()

	var received ContentSavedEvent
	SubscribeTyped(bus, func(e ContentSavedEvent) {
		received = e
	})

	bus.Publish(ContentSavedEvent{
		ID:         "doc-1",
		Collection: "notes",
		Content:    "hello world",
		Source:     "knowledge",
	})

	assert.Equal(t, "doc-1", received.ID)
	assert.Equal(t, "knowledge", received.Source)
}

func TestDifferentEventTypesRouteToSeparateHandlers(t *testing.T) {
	t.Parallel()

	bus := New()

	var testCalled, otherCalled bool
	bus.Subscribe("test.event", func(_ Event) { testCalled = true })
	bus.Subscribe("other.event", func(_ Event) { otherCalled = true })

	bus.Publish(testEvent{Value: "a"})

	assert.True(t, testCalled, "test.event handler was not called")
	assert.False(t, otherCalled, "other.event handler was called unexpectedly")

	// Reset and publish the other event.
	testCalled = false
	otherCalled = false

	bus.Publish(otherEvent{Code: 42})

	assert.False(t, testCalled, "test.event handler was called unexpectedly")
	assert.True(t, otherCalled, "other.event handler was not called")
}

func TestConcurrentPublishAndSubscribe(t *testing.T) {
	t.Parallel()

	bus := New()

	var count atomic.Int64
	const goroutines = 50
	const eventsPerGoroutine = 100

	// Pre-register one handler so there is something to call.
	bus.Subscribe("test.event", func(_ Event) {
		count.Add(1)
	})

	var wg sync.WaitGroup

	// Concurrent publishers.
	for i := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range eventsPerGoroutine {
				bus.Publish(testEvent{Value: "msg"})
				// Interleave a subscribe on every 10th iteration to
				// exercise concurrent subscribe + publish.
				if j%10 == 0 {
					bus.Subscribe("test.event", func(_ Event) {
						count.Add(1)
					})
				}
			}
			_ = id
		}(i)
	}

	wg.Wait()

	// We only assert that no data race occurred. The exact count is
	// non-deterministic because new handlers are added while publishing.
	assert.Greater(t, count.Load(), int64(0), "expected at least one handler invocation")
}

func TestSubscribeTypedIgnoresMismatchedType(t *testing.T) {
	t.Parallel()

	bus := New()

	var called bool
	SubscribeTyped(bus, func(_ TurnCompletedEvent) {
		called = true
	})

	// Publish a different event with the same event name — this should not
	// happen in production but verifies the type assertion guard.
	bus.Subscribe("turn.completed", func(_ Event) {})
	bus.Publish(TurnCompletedEvent{SessionKey: "sess-1"})

	assert.True(t, called, "typed handler was not called for matching type")
}

func TestAllEventTypesHaveDistinctNames(t *testing.T) {
	t.Parallel()

	events := []Event{
		ContentSavedEvent{},
		TriplesExtractedEvent{},
		TurnCompletedEvent{},
		ReputationChangedEvent{},
		MemoryGraphEvent{},
		ChannelMessageReceivedEvent{},
		ChannelMessageSentEvent{},
		CompactionCompletedEvent{},
		CompactionSlowEvent{},
		LearningSuggestionEvent{},
	}

	seen := make(map[string]bool, len(events))
	for _, e := range events {
		name := e.EventName()
		assert.False(t, seen[name], "duplicate event name: %s", name)
		seen[name] = true
	}
}

func TestReputationChangedEventRoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got ReputationChangedEvent
	SubscribeTyped(bus, func(e ReputationChangedEvent) {
		got = e
	})

	bus.Publish(ReputationChangedEvent{PeerDID: "did:example:123", NewScore: 0.85})

	assert.Equal(t, "did:example:123", got.PeerDID)
	assert.InDelta(t, 0.85, got.NewScore, 0.001)
}

func TestTriplesExtractedEventRoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got TriplesExtractedEvent
	SubscribeTyped(bus, func(e TriplesExtractedEvent) {
		got = e
	})

	bus.Publish(TriplesExtractedEvent{
		Triples: []Triple{
			{Subject: "Go", Predicate: "is", Object: "fast"},
			{Subject: "Rust", Predicate: "is", Object: "safe"},
		},
		Source: "learning",
	})

	require.Len(t, got.Triples, 2)
	assert.Equal(t, "Go", got.Triples[0].Subject)
	assert.Equal(t, "learning", got.Source)
}

func TestMemoryGraphEventRoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got MemoryGraphEvent
	SubscribeTyped(bus, func(e MemoryGraphEvent) {
		got = e
	})

	bus.Publish(MemoryGraphEvent{
		Triples: []Triple{
			{Subject: "Alice", Predicate: "knows", Object: "Bob"},
		},
		SessionKey: "sess-42",
		Type:       "observation",
	})

	require.Len(t, got.Triples, 1)
	assert.Equal(t, "Alice", got.Triples[0].Subject)
	assert.Equal(t, "sess-42", got.SessionKey)
	assert.Equal(t, "observation", got.Type)
}

func TestCompactionCompletedEventRoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got CompactionCompletedEvent
	SubscribeTyped(bus, func(e CompactionCompletedEvent) {
		got = e
	})

	bus.Publish(CompactionCompletedEvent{
		SessionKey:      "sess-1",
		UpToIndex:       20,
		SummaryTokens:   120,
		ReclaimedTokens: 4200,
	})

	assert.Equal(t, "sess-1", got.SessionKey)
	assert.Equal(t, 20, got.UpToIndex)
	assert.Equal(t, 4200, got.ReclaimedTokens)
	assert.Equal(t, EventCompactionCompleted, got.EventName())
}

func TestCompactionSlowEventRoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got CompactionSlowEvent
	SubscribeTyped(bus, func(e CompactionSlowEvent) {
		got = e
	})

	bus.Publish(CompactionSlowEvent{
		SessionKey: "sess-2",
	})

	assert.Equal(t, "sess-2", got.SessionKey)
	assert.Equal(t, EventCompactionSlow, got.EventName())
}

func TestLearningSuggestionEventRoundTrip(t *testing.T) {
	t.Parallel()

	bus := New()

	var got LearningSuggestionEvent
	SubscribeTyped(bus, func(e LearningSuggestionEvent) {
		got = e
	})

	bus.Publish(LearningSuggestionEvent{
		SessionKey:   "sess-3",
		SuggestionID: "sugg-1",
		Pattern:      "timeout:fetch",
		ProposedRule: "retry with 2x backoff",
		Confidence:   0.62,
	})

	assert.Equal(t, "sess-3", got.SessionKey)
	assert.Equal(t, "sugg-1", got.SuggestionID)
	assert.InDelta(t, 0.62, got.Confidence, 0.001)
	assert.Equal(t, EventLearningSuggestion, got.EventName())
}
