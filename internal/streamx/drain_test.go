package streamx

import (
	"errors"
	"testing"
)

// testStream creates a Stream[int] that yields the given items, then optionally
// yields an error. This helper is shared across test files via package-level
// visibility.
func testStream(items []int, err error) Stream[int] {
	return func(yield func(int, error) bool) {
		for _, v := range items {
			if !yield(v, nil) {
				return
			}
		}
		if err != nil {
			yield(0, err)
		}
	}
}

func TestDrain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		items    []int
		giveErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			give:    "multiple items",
			items:   []int{1, 2, 3},
			wantLen: 3,
		},
		{
			give:    "empty stream",
			items:   nil,
			wantLen: 0,
		},
		{
			give:    "single item",
			items:   []int{42},
			wantLen: 1,
		},
		{
			give:    "error after items",
			items:   []int{1, 2},
			giveErr: errors.New("boom"),
			wantLen: 2,
			wantErr: true,
		},
		{
			give:    "error with no items",
			items:   nil,
			giveErr: errors.New("boom"),
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := testStream(tt.items, tt.giveErr)
			got, err := Drain[int](s)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("want %d items, got %d", tt.wantLen, len(got))
			}
			for i, v := range got {
				if v != tt.items[i] {
					t.Errorf("item[%d]: want %d, got %d", i, tt.items[i], v)
				}
			}
		})
	}
}
