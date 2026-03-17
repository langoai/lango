package tooloutput

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputStore_StoreAndGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		giveContent string
	}{
		{
			give:        "simple text",
			giveContent: "hello world",
		},
		{
			give:        "multiline",
			giveContent: "line1\nline2\nline3",
		},
		{
			give:        "empty content",
			giveContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewOutputStore(time.Minute)

			ref := s.Store("test-tool", tt.giveContent)
			require.NotEmpty(t, ref)

			got, ok := s.Get(ref)
			require.True(t, ok)
			assert.Equal(t, tt.giveContent, got)
		})
	}
}

func TestOutputStore_GetMissing(t *testing.T) {
	t.Parallel()
	s := NewOutputStore(time.Minute)

	got, ok := s.Get("nonexistent-ref")
	assert.False(t, ok)
	assert.Empty(t, got)
}

func TestOutputStore_GetRange(t *testing.T) {
	t.Parallel()

	content := "line0\nline1\nline2\nline3\nline4"

	tests := []struct {
		give       string
		giveOffset int
		giveLimit  int
		wantLines  string
		wantTotal  int
		wantFound  bool
	}{
		{
			give:       "first 2 lines",
			giveOffset: 0,
			giveLimit:  2,
			wantLines:  "line0\nline1",
			wantTotal:  5,
			wantFound:  true,
		},
		{
			give:       "middle range",
			giveOffset: 1,
			giveLimit:  2,
			wantLines:  "line1\nline2",
			wantTotal:  5,
			wantFound:  true,
		},
		{
			give:       "offset beyond content",
			giveOffset: 10,
			giveLimit:  5,
			wantLines:  "",
			wantTotal:  5,
			wantFound:  true,
		},
		{
			give:       "limit zero returns all from offset",
			giveOffset: 2,
			giveLimit:  0,
			wantLines:  "line2\nline3\nline4",
			wantTotal:  5,
			wantFound:  true,
		},
		{
			give:       "limit exceeds remaining",
			giveOffset: 3,
			giveLimit:  100,
			wantLines:  "line3\nline4",
			wantTotal:  5,
			wantFound:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewOutputStore(time.Minute)
			ref := s.Store("test-tool", content)

			got, total, found := s.GetRange(ref, tt.giveOffset, tt.giveLimit)
			assert.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantTotal, total)
			assert.Equal(t, tt.wantLines, got)
		})
	}
}

func TestOutputStore_GetRange_NotFound(t *testing.T) {
	t.Parallel()
	s := NewOutputStore(time.Minute)

	got, total, found := s.GetRange("missing", 0, 10)
	assert.False(t, found)
	assert.Equal(t, 0, total)
	assert.Empty(t, got)
}

func TestOutputStore_Grep(t *testing.T) {
	t.Parallel()

	content := "ERROR: something failed\nINFO: all good\nERROR: another failure\nDEBUG: details"

	tests := []struct {
		give        string
		givePattern string
		wantMatches string
		wantFound   bool
	}{
		{
			give:        "matching lines",
			givePattern: "ERROR",
			wantMatches: "ERROR: something failed\nERROR: another failure",
			wantFound:   true,
		},
		{
			give:        "no matches",
			givePattern: "WARN",
			wantMatches: "",
			wantFound:   true,
		},
		{
			give:        "regex pattern",
			givePattern: `^(ERROR|DEBUG):`,
			wantMatches: "ERROR: something failed\nERROR: another failure\nDEBUG: details",
			wantFound:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewOutputStore(time.Minute)
			ref := s.Store("test-tool", content)

			got, found := s.Grep(ref, tt.givePattern)
			assert.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantMatches, got)
		})
	}
}

func TestOutputStore_Grep_NotFound(t *testing.T) {
	t.Parallel()
	s := NewOutputStore(time.Minute)

	got, found := s.Grep("missing", "pattern")
	assert.False(t, found)
	assert.Empty(t, got)
}

func TestOutputStore_Grep_InvalidPattern(t *testing.T) {
	t.Parallel()
	s := NewOutputStore(time.Minute)
	ref := s.Store("test-tool", "some content")

	got, found := s.Grep(ref, "[invalid")
	assert.True(t, found)
	assert.Empty(t, got)
}

func TestOutputStore_TTLExpiry(t *testing.T) {
	t.Parallel()
	ttl := 50 * time.Millisecond
	s := NewOutputStore(ttl)

	ref := s.Store("test-tool", "ephemeral data")

	// Entry exists immediately.
	_, ok := s.Get(ref)
	require.True(t, ok)

	// Wait for TTL to expire.
	time.Sleep(ttl + 10*time.Millisecond)

	// Manually trigger eviction (no cleanup goroutine running).
	s.evictExpired()

	_, ok = s.Get(ref)
	assert.False(t, ok)
}

func TestOutputStore_Name(t *testing.T) {
	t.Parallel()
	s := NewOutputStore(time.Minute)
	assert.Equal(t, "output-store", s.Name())
}

func TestOutputStore_StartStop(t *testing.T) {
	t.Parallel()
	s := NewOutputStore(100 * time.Millisecond)

	var wg sync.WaitGroup
	err := s.Start(context.Background(), &wg)
	require.NoError(t, err)

	// Store something while running.
	ref := s.Store("test-tool", "data")
	_, ok := s.Get(ref)
	require.True(t, ok)

	// Stop should not error.
	err = s.Stop(context.Background())
	require.NoError(t, err)

	// Wait for cleanup goroutine to exit.
	wg.Wait()
}
