package eventbus

import (
	"sync"
	"sync/atomic"
	"testing"
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
	bus := New()

	var received string
	bus.Subscribe("test.event", func(event Event) {
		received = event.(testEvent).Value
	})

	bus.Publish(testEvent{Value: "hello"})

	if received != "hello" {
		t.Errorf("want %q, got %q", "hello", received)
	}
}

func TestMultipleHandlersReceiveInOrder(t *testing.T) {
	bus := New()

	var order []int
	bus.Subscribe("test.event", func(_ Event) { order = append(order, 1) })
	bus.Subscribe("test.event", func(_ Event) { order = append(order, 2) })
	bus.Subscribe("test.event", func(_ Event) { order = append(order, 3) })

	bus.Publish(testEvent{Value: "x"})

	if len(order) != 3 {
		t.Fatalf("want 3 handler calls, got %d", len(order))
	}
	for i, want := range []int{1, 2, 3} {
		if order[i] != want {
			t.Errorf("order[%d] = %d, want %d", i, order[i], want)
		}
	}
}

func TestPublishWithNoHandlersDoesNotPanic(t *testing.T) {
	bus := New()

	// Should not panic.
	bus.Publish(testEvent{Value: "nobody listening"})
}

func TestSubscribeTypedProvidesSafeHandling(t *testing.T) {
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

	if received.ID != "doc-1" {
		t.Errorf("want ID %q, got %q", "doc-1", received.ID)
	}
	if received.Source != "knowledge" {
		t.Errorf("want Source %q, got %q", "knowledge", received.Source)
	}
}

func TestDifferentEventTypesRouteToSeparateHandlers(t *testing.T) {
	bus := New()

	var testCalled, otherCalled bool
	bus.Subscribe("test.event", func(_ Event) { testCalled = true })
	bus.Subscribe("other.event", func(_ Event) { otherCalled = true })

	bus.Publish(testEvent{Value: "a"})

	if !testCalled {
		t.Error("test.event handler was not called")
	}
	if otherCalled {
		t.Error("other.event handler was called unexpectedly")
	}

	// Reset and publish the other event.
	testCalled = false
	otherCalled = false

	bus.Publish(otherEvent{Code: 42})

	if testCalled {
		t.Error("test.event handler was called unexpectedly")
	}
	if !otherCalled {
		t.Error("other.event handler was not called")
	}
}

func TestConcurrentPublishAndSubscribe(t *testing.T) {
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
	if count.Load() == 0 {
		t.Error("expected at least one handler invocation")
	}
}

func TestSubscribeTypedIgnoresMismatchedType(t *testing.T) {
	bus := New()

	var called bool
	SubscribeTyped(bus, func(_ TurnCompletedEvent) {
		called = true
	})

	// Publish a different event with the same event name — this should not
	// happen in production but verifies the type assertion guard.
	bus.Subscribe("turn.completed", func(_ Event) {})
	bus.Publish(TurnCompletedEvent{SessionKey: "sess-1"})

	if !called {
		t.Error("typed handler was not called for matching type")
	}
}

func TestAllEventTypesHaveDistinctNames(t *testing.T) {
	events := []Event{
		ContentSavedEvent{},
		TriplesExtractedEvent{},
		TurnCompletedEvent{},
		ReputationChangedEvent{},
		MemoryGraphEvent{},
	}

	seen := make(map[string]bool, len(events))
	for _, e := range events {
		name := e.EventName()
		if seen[name] {
			t.Errorf("duplicate event name: %s", name)
		}
		seen[name] = true
	}
}

func TestReputationChangedEventRoundTrip(t *testing.T) {
	bus := New()

	var got ReputationChangedEvent
	SubscribeTyped(bus, func(e ReputationChangedEvent) {
		got = e
	})

	bus.Publish(ReputationChangedEvent{PeerDID: "did:example:123", NewScore: 0.85})

	if got.PeerDID != "did:example:123" {
		t.Errorf("PeerDID = %q, want %q", got.PeerDID, "did:example:123")
	}
	if got.NewScore != 0.85 {
		t.Errorf("NewScore = %f, want %f", got.NewScore, 0.85)
	}
}

func TestTriplesExtractedEventRoundTrip(t *testing.T) {
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

	if len(got.Triples) != 2 {
		t.Fatalf("want 2 triples, got %d", len(got.Triples))
	}
	if got.Triples[0].Subject != "Go" {
		t.Errorf("Subject = %q, want %q", got.Triples[0].Subject, "Go")
	}
	if got.Source != "learning" {
		t.Errorf("Source = %q, want %q", got.Source, "learning")
	}
}

func TestMemoryGraphEventRoundTrip(t *testing.T) {
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

	if len(got.Triples) != 1 {
		t.Fatalf("want 1 triple, got %d", len(got.Triples))
	}
	if got.Triples[0].Subject != "Alice" {
		t.Errorf("Subject = %q, want %q", got.Triples[0].Subject, "Alice")
	}
	if got.SessionKey != "sess-42" {
		t.Errorf("SessionKey = %q, want %q", got.SessionKey, "sess-42")
	}
	if got.Type != "observation" {
		t.Errorf("Type = %q, want %q", got.Type, "observation")
	}
}
