package streamx

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressBus_EmitAndSubscribe(t *testing.T) {
	bus := NewProgressBus()
	ch, cancel := bus.Subscribe("tool:")
	defer cancel()

	bus.Emit(ProgressEvent{Source: "tool:web_search", Type: ProgressStarted, Message: "searching"})
	bus.Emit(ProgressEvent{Source: "agent:operator", Type: ProgressStarted, Message: "running"})

	select {
	case ev := <-ch:
		assert.Equal(t, "tool:web_search", ev.Source)
		assert.Equal(t, ProgressStarted, ev.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}

	// agent: event should NOT arrive on tool: subscriber.
	select {
	case ev := <-ch:
		t.Fatalf("unexpected event: %v", ev)
	case <-time.After(50 * time.Millisecond):
		// expected
	}
}

func TestProgressBus_SubscribeAll(t *testing.T) {
	bus := NewProgressBus()
	ch, cancel := bus.SubscribeAll()
	defer cancel()

	bus.Emit(ProgressEvent{Source: "tool:fs_read", Type: ProgressCompleted})
	bus.Emit(ProgressEvent{Source: "agent:vault", Type: ProgressFailed})

	var received []string
	for i := 0; i < 2; i++ {
		select {
		case ev := <-ch:
			received = append(received, ev.Source)
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}
	assert.ElementsMatch(t, []string{"tool:fs_read", "agent:vault"}, received)
}

func TestProgressBus_Cancel(t *testing.T) {
	bus := NewProgressBus()
	ch, cancel := bus.Subscribe("tool:")

	cancel()

	// Channel should be closed.
	_, ok := <-ch
	assert.False(t, ok)

	// Double cancel should not panic.
	cancel()

	// Emit after cancel should not panic.
	bus.Emit(ProgressEvent{Source: "tool:test", Type: ProgressUpdate})
}

func TestProgressBus_BufferFullDropsEvent(t *testing.T) {
	bus := NewProgressBus()
	ch, cancel := bus.SubscribeAll()
	defer cancel()

	// Fill the buffer (capacity 64).
	for i := 0; i < 100; i++ {
		bus.Emit(ProgressEvent{Source: "tool:spam", Type: ProgressUpdate})
	}

	// Should have received up to 64 events.
	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			goto done
		}
	}
done:
	assert.Equal(t, 64, count)
}

func TestProgressBus_ConcurrentEmit(t *testing.T) {
	bus := NewProgressBus()
	ch, cancel := bus.SubscribeAll()
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Emit(ProgressEvent{Source: "tool:concurrent", Type: ProgressUpdate})
		}()
	}
	wg.Wait()

	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			goto done2
		}
	}
done2:
	assert.Equal(t, 10, count)
}

func TestProgressBus_FilterPrefixMatching(t *testing.T) {
	bus := NewProgressBus()
	ch, cancel := bus.Subscribe("bg:")
	defer cancel()

	bus.Emit(ProgressEvent{Source: "bg:task-123", Type: ProgressStarted})
	bus.Emit(ProgressEvent{Source: "bg:task-456", Type: ProgressCompleted})
	bus.Emit(ProgressEvent{Source: "tool:bg_submit", Type: ProgressUpdate}) // should NOT match

	var received []string
	for i := 0; i < 2; i++ {
		select {
		case ev := <-ch:
			received = append(received, ev.Source)
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}
	require.Len(t, received, 2)
	assert.Contains(t, received, "bg:task-123")
	assert.Contains(t, received, "bg:task-456")
}

func TestProgressBus_MultipleSubscribers(t *testing.T) {
	bus := NewProgressBus()
	ch1, cancel1 := bus.Subscribe("tool:")
	defer cancel1()
	ch2, cancel2 := bus.SubscribeAll()
	defer cancel2()

	bus.Emit(ProgressEvent{Source: "tool:test", Type: ProgressStarted})

	// Both should receive.
	select {
	case <-ch1:
	case <-time.After(time.Second):
		t.Fatal("ch1 timeout")
	}
	select {
	case <-ch2:
	case <-time.After(time.Second):
		t.Fatal("ch2 timeout")
	}
}
