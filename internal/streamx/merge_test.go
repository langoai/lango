package streamx

import (
	"context"
	"errors"
	"sort"
	"testing"
)

func TestMerge(t *testing.T) {
	t.Parallel()

	t.Run("multiple streams", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{
			"a": testStream([]int{1, 2}, nil),
			"b": testStream([]int{3, 4}, nil),
		}

		merged := Merge[int](ctx, streams)
		var tags []Tag[int]
		for v, err := range merged {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tags = append(tags, v)
		}

		if len(tags) != 4 {
			t.Fatalf("want 4 events, got %d", len(tags))
		}

		// Collect events per source.
		bySource := make(map[string][]int)
		for _, tag := range tags {
			bySource[tag.Source] = append(bySource[tag.Source], tag.Event)
		}

		// Within each source, order is preserved.
		wantA := []int{1, 2}
		wantB := []int{3, 4}
		if len(bySource["a"]) != 2 || bySource["a"][0] != wantA[0] || bySource["a"][1] != wantA[1] {
			t.Errorf("source a: want %v, got %v", wantA, bySource["a"])
		}
		if len(bySource["b"]) != 2 || bySource["b"][0] != wantB[0] || bySource["b"][1] != wantB[1] {
			t.Errorf("source b: want %v, got %v", wantB, bySource["b"])
		}
	})

	t.Run("three streams", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{
			"x": testStream([]int{10}, nil),
			"y": testStream([]int{20, 30}, nil),
			"z": testStream([]int{40, 50, 60}, nil),
		}

		merged := Merge[int](ctx, streams)
		var events []int
		for v, err := range merged {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			events = append(events, v.Event)
		}

		sort.Ints(events)
		want := []int{10, 20, 30, 40, 50, 60}
		if len(events) != len(want) {
			t.Fatalf("want %d events, got %d", len(want), len(events))
		}
		for i := range want {
			if events[i] != want[i] {
				t.Errorf("events[%d]: want %d, got %d", i, want[i], events[i])
			}
		}
	})

	t.Run("empty streams map", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{}

		merged := Merge[int](ctx, streams)
		count := 0
		for range merged {
			count++
		}
		if count != 0 {
			t.Fatalf("want 0 events, got %d", count)
		}
	})

	t.Run("single stream", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{
			"only": testStream([]int{7, 8, 9}, nil),
		}

		merged := Merge[int](ctx, streams)
		var events []int
		for v, err := range merged {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			events = append(events, v.Event)
		}

		if len(events) != 3 {
			t.Fatalf("want 3 events, got %d", len(events))
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		errBoom := errors.New("boom")
		streams := map[string]Stream[int]{
			"ok":  testStream([]int{1}, nil),
			"bad": testStream([]int{2}, errBoom),
		}

		merged := Merge[int](ctx, streams)
		var gotErr error
		for _, err := range merged {
			if err != nil {
				gotErr = err
				break
			}
		}
		if gotErr == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())

		// A stream that would block forever without cancellation.
		infinite := Stream[int](func(yield func(int, error) bool) {
			i := 0
			for {
				i++
				if !yield(i, nil) {
					return
				}
			}
		})

		streams := map[string]Stream[int]{
			"inf": infinite,
		}

		merged := Merge[int](ctx, streams)
		count := 0
		for _, err := range merged {
			if err != nil {
				break
			}
			count++
			if count >= 3 {
				cancel()
				break
			}
		}
		// Just verify we got some events and didn't hang.
		if count < 3 {
			t.Fatalf("want at least 3 events before cancel, got %d", count)
		}
		// Ensure cancel is called (deferred above or explicit).
		cancel()
	})
}
