package streamx

import (
	"context"
	"errors"
	"testing"
)

func TestFanIn(t *testing.T) {
	t.Parallel()

	t.Run("multiple streams", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{
			"a": testStream([]int{1, 2}, nil),
			"b": testStream([]int{3, 4, 5}, nil),
		}

		result, err := FanIn[int](ctx, streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("want 2 sources, got %d", len(result))
		}
		if len(result["a"]) != 2 {
			t.Errorf("source a: want 2 events, got %d", len(result["a"]))
		}
		if len(result["b"]) != 3 {
			t.Errorf("source b: want 3 events, got %d", len(result["b"]))
		}

		// Verify order preserved within each source.
		wantA := []int{1, 2}
		for i, v := range result["a"] {
			if v != wantA[i] {
				t.Errorf("a[%d]: want %d, got %d", i, wantA[i], v)
			}
		}
		wantB := []int{3, 4, 5}
		for i, v := range result["b"] {
			if v != wantB[i] {
				t.Errorf("b[%d]: want %d, got %d", i, wantB[i], v)
			}
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

		result, err := FanIn[int](ctx, streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 3 {
			t.Fatalf("want 3 sources, got %d", len(result))
		}
		if len(result["x"]) != 1 {
			t.Errorf("source x: want 1 event, got %d", len(result["x"]))
		}
		if len(result["y"]) != 2 {
			t.Errorf("source y: want 2 events, got %d", len(result["y"]))
		}
		if len(result["z"]) != 3 {
			t.Errorf("source z: want 3 events, got %d", len(result["z"]))
		}
	})

	t.Run("empty streams map", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{}

		result, err := FanIn[int](ctx, streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("want empty map, got %d entries", len(result))
		}
	})

	t.Run("single stream", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{
			"only": testStream([]int{7, 8, 9}, nil),
		}

		result, err := FanIn[int](ctx, streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("want 1 source, got %d", len(result))
		}
		if len(result["only"]) != 3 {
			t.Errorf("want 3 events, got %d", len(result["only"]))
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		errBoom := errors.New("boom")
		streams := map[string]Stream[int]{
			"ok":  testStream([]int{1, 2}, nil),
			"bad": testStream([]int{3}, errBoom),
		}

		_, err := FanIn[int](ctx, streams)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("empty streams within map", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		streams := map[string]Stream[int]{
			"empty": testStream(nil, nil),
			"full":  testStream([]int{1}, nil),
		}

		result, err := FanIn[int](ctx, streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result["full"]) != 1 {
			t.Errorf("full: want 1 event, got %d", len(result["full"]))
		}
		// Empty source may or may not be in the map (nil slice is fine).
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately.

		// An infinite stream that respects context via the combinator.
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

		_, err := FanIn[int](ctx, streams)
		if err == nil {
			t.Fatal("expected error from cancelled context, got nil")
		}
	})
}
