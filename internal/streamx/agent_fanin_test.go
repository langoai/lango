package streamx

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testStringStream creates a Stream[string] that yields the given items, then
// optionally yields an error.
func testStringStream(items []string, err error) Stream[string] {
	return func(yield func(string, error) bool) {
		for _, v := range items {
			if !yield(v, nil) {
				return
			}
		}
		if err != nil {
			yield("", err)
		}
	}
}

func TestAgentStreamFanIn_TwoChildren(t *testing.T) {
	t.Parallel()

	bus := NewProgressBus()
	fanin := NewAgentStreamFanIn("session-1", bus)

	fanin.AddChild("child-a", testStringStream([]string{"hello", "world"}, nil))
	fanin.AddChild("child-b", testStringStream([]string{"foo", "bar"}, nil))

	ctx := context.Background()
	merged := fanin.MergedStream(ctx)

	var tags []Tag[string]
	for v, err := range merged {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tags = append(tags, v)
	}

	require.Len(t, tags, 4)

	// Collect events per source.
	bySource := make(map[string][]string)
	for _, tag := range tags {
		bySource[tag.Source] = append(bySource[tag.Source], tag.Event)
	}

	assert.Equal(t, []string{"hello", "world"}, bySource["child-a"])
	assert.Equal(t, []string{"foo", "bar"}, bySource["child-b"])
}

func TestAgentStreamFanIn_SingleChild(t *testing.T) {
	t.Parallel()

	fanin := NewAgentStreamFanIn("session-2", NewProgressBus())
	fanin.AddChild("only", testStringStream([]string{"one", "two", "three"}, nil))

	ctx := context.Background()
	merged := fanin.MergedStream(ctx)

	var events []string
	for v, err := range merged {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		events = append(events, v.Event)
	}

	assert.Equal(t, []string{"one", "two", "three"}, events)
}

func TestAgentStreamFanIn_EmptyChildren(t *testing.T) {
	t.Parallel()

	fanin := NewAgentStreamFanIn("session-3", NewProgressBus())
	// No children added.

	ctx := context.Background()
	merged := fanin.MergedStream(ctx)

	count := 0
	for range merged {
		count++
	}
	assert.Equal(t, 0, count)
}

func TestAgentStreamFanIn_ProgressBusLifecycle(t *testing.T) {
	t.Parallel()

	bus := NewProgressBus()
	ch, cancel := bus.Subscribe("agent:session-4:child:")
	defer cancel()

	fanin := NewAgentStreamFanIn("session-4", bus)
	fanin.AddChild("alpha", testStringStream([]string{"a1"}, nil))
	fanin.AddChild("beta", testStringStream([]string{"b1"}, nil))

	ctx := context.Background()
	merged := fanin.MergedStream(ctx)

	// Drain the merged stream to trigger child completions.
	for range merged {
	}

	// Collect all progress events with a short timeout.
	var events []ProgressEvent
	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-ch:
			events = append(events, ev)
			// We expect 2 started + 2 completed = 4 events total.
			if len(events) >= 4 {
				goto collected
			}
		case <-deadline:
			goto collected
		}
	}
collected:

	require.Len(t, events, 4)

	// Separate by type.
	var started, completed []string
	for _, ev := range events {
		switch ev.Type {
		case ProgressStarted:
			started = append(started, ev.Source)
		case ProgressCompleted:
			completed = append(completed, ev.Source)
		}
	}

	sort.Strings(started)
	sort.Strings(completed)

	assert.Equal(t, []string{
		"agent:session-4:child:alpha",
		"agent:session-4:child:beta",
	}, started)
	assert.Equal(t, []string{
		"agent:session-4:child:alpha",
		"agent:session-4:child:beta",
	}, completed)
}

func TestAgentStreamFanIn_OneChildError(t *testing.T) {
	t.Parallel()

	bus := NewProgressBus()
	ch, cancel := bus.Subscribe("agent:session-5:child:")
	defer cancel()

	errBoom := errors.New("child-b exploded")

	fanin := NewAgentStreamFanIn("session-5", bus)
	fanin.AddChild("child-a", testStringStream([]string{"a1", "a2"}, nil))
	fanin.AddChild("child-b", testStringStream([]string{"b1"}, errBoom))

	ctx := context.Background()
	merged := fanin.MergedStream(ctx)

	var events []Tag[string]
	var gotErr error
	for v, err := range merged {
		if err != nil {
			gotErr = err
			continue
		}
		events = append(events, v)
	}

	// We should have received some events before the error (at least child-a's).
	assert.NotEmpty(t, events)
	assert.Error(t, gotErr)

	// Collect progress events.
	var progressEvents []ProgressEvent
	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-ch:
			progressEvents = append(progressEvents, ev)
			if len(progressEvents) >= 4 {
				goto done
			}
		case <-deadline:
			goto done
		}
	}
done:

	// Should have started events for both children.
	var started []string
	var failed []string
	for _, ev := range progressEvents {
		switch ev.Type {
		case ProgressStarted:
			started = append(started, ev.Source)
		case ProgressFailed:
			failed = append(failed, ev.Source)
		}
	}

	sort.Strings(started)
	assert.Equal(t, []string{
		"agent:session-5:child:child-a",
		"agent:session-5:child:child-b",
	}, started)

	// child-b should have a failed event.
	assert.Contains(t, failed, "agent:session-5:child:child-b")
}

func TestAgentStreamFanIn_NilBus(t *testing.T) {
	t.Parallel()

	// nil bus should not panic.
	fanin := NewAgentStreamFanIn("session-6", nil)
	fanin.AddChild("child-a", testStringStream([]string{"x", "y"}, nil))

	ctx := context.Background()
	merged := fanin.MergedStream(ctx)

	var events []string
	for v, err := range merged {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		events = append(events, v.Event)
	}

	assert.Equal(t, []string{"x", "y"}, events)
}
