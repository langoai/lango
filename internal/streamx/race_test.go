package streamx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRace(t *testing.T) {
	t.Parallel()

	t.Run("first stream wins", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		// "fast" yields immediately, "slow" is delayed.
		fast := testStream([]int{42}, nil)
		slow := Stream[int](func(yield func(int, error) bool) {
			time.Sleep(100 * time.Millisecond)
			yield(99, nil)
		})

		streams := map[string]Stream[int]{
			"fast": fast,
			"slow": slow,
		}

		raced := Race[int](ctx, streams)
		var tags []Tag[int]
		for v, err := range raced {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tags = append(tags, v)
		}

		if len(tags) == 0 {
			t.Fatal("want at least 1 event, got 0")
		}
		if tags[0].Event != 42 {
			t.Errorf("want first event 42, got %d", tags[0].Event)
		}
		if tags[0].Source != "fast" {
			t.Errorf("want source 'fast', got %q", tags[0].Source)
		}
	})

	t.Run("single stream", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{
			"only": testStream([]int{7}, nil),
		}

		raced := Race[int](ctx, streams)
		count := 0
		for _, err := range raced {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			count++
		}
		if count != 1 {
			t.Fatalf("want 1 event, got %d", count)
		}
	})

	t.Run("empty streams map", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{}

		raced := Race[int](ctx, streams)
		count := 0
		for range raced {
			count++
		}
		if count != 0 {
			t.Fatalf("want 0 events, got %d", count)
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		errBoom := errors.New("boom")

		streams := map[string]Stream[int]{
			"bad": testStream(nil, errBoom),
		}

		raced := Race[int](ctx, streams)
		var gotErr error
		for _, err := range raced {
			if err != nil {
				gotErr = err
			}
		}
		if gotErr == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately.

		streams := map[string]Stream[int]{
			"a": testStream([]int{1}, nil),
		}

		raced := Race[int](ctx, streams)
		count := 0
		for range raced {
			count++
		}
		// With pre-cancelled context, we may get 0 or 1 events depending on
		// scheduling. The key is that it doesn't hang.
		if count > 1 {
			t.Fatalf("want at most 1 event, got %d", count)
		}
	})

	t.Run("two streams same speed", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{
			"a": testStream([]int{1}, nil),
			"b": testStream([]int{2}, nil),
		}

		raced := Race[int](ctx, streams)
		var tags []Tag[int]
		for v, err := range raced {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tags = append(tags, v)
		}

		if len(tags) != 1 {
			t.Fatalf("want 1 event, got %d", len(tags))
		}
		// Either "a" or "b" could win; just check the event is valid.
		if tags[0].Event != 1 && tags[0].Event != 2 {
			t.Errorf("want event 1 or 2, got %d", tags[0].Event)
		}
	})

	t.Run("multi-event winner drains fully", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// "fast" yields 3 events immediately; "slow" blocks until cancelled.
		fast := testStream([]int{10, 20, 30}, nil)
		slow := Stream[int](func(yield func(int, error) bool) {
			<-ctx.Done()
		})

		streams := map[string]Stream[int]{
			"fast": fast,
			"slow": slow,
		}

		raced := Race(ctx, streams)
		var tags []Tag[int]
		for v, err := range raced {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tags = append(tags, v)
		}

		if len(tags) != 3 {
			t.Fatalf("want 3 events from winner, got %d", len(tags))
		}
		for i, want := range []int{10, 20, 30} {
			if tags[i].Event != want {
				t.Errorf("event[%d]: want %d, got %d", i, want, tags[i].Event)
			}
			if tags[i].Source != "fast" {
				t.Errorf("event[%d]: want source 'fast', got %q", i, tags[i].Source)
			}
		}
	})
}
