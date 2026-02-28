package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeline_ExecutesInOrder(t *testing.T) {
	var order []string

	phases := []Phase{
		{
			Name: "phase-a",
			Run: func(_ context.Context, _ *State) error {
				order = append(order, "a")
				return nil
			},
		},
		{
			Name: "phase-b",
			Run: func(_ context.Context, _ *State) error {
				order = append(order, "b")
				return nil
			},
		},
		{
			Name: "phase-c",
			Run: func(_ context.Context, _ *State) error {
				order = append(order, "c")
				return nil
			},
		},
	}

	p := NewPipeline(phases...)
	result, err := p.Execute(context.Background(), Options{})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []string{"a", "b", "c"}, order)
}

func TestPipeline_CleanupRunsInReverseOnFailure(t *testing.T) {
	var cleanupOrder []string

	phases := []Phase{
		{
			Name: "phase-a",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			Cleanup: func(_ *State) {
				cleanupOrder = append(cleanupOrder, "a")
			},
		},
		{
			Name: "phase-b",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			Cleanup: func(_ *State) {
				cleanupOrder = append(cleanupOrder, "b")
			},
		},
		{
			Name: "phase-c",
			Run: func(_ context.Context, _ *State) error {
				return errors.New("phase-c failed")
			},
			Cleanup: func(_ *State) {
				cleanupOrder = append(cleanupOrder, "c")
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "phase-c")

	// Cleanup should run for a and b (completed) in reverse, NOT for c (failed).
	assert.Equal(t, []string{"b", "a"}, cleanupOrder)
}

func TestPipeline_CleanupNotCalledForFailedPhase(t *testing.T) {
	var cleaned []string

	phases := []Phase{
		{
			Name: "phase-a",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			Cleanup: func(_ *State) {
				cleaned = append(cleaned, "a")
			},
		},
		{
			Name: "phase-b",
			Run: func(_ context.Context, _ *State) error {
				return errors.New("boom")
			},
			Cleanup: func(_ *State) {
				cleaned = append(cleaned, "b")
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.Error(t, err)

	// Only phase-a cleanup should run, not phase-b.
	assert.Equal(t, []string{"a"}, cleaned)
}

func TestPipeline_StatePassedBetweenPhases(t *testing.T) {
	phases := []Phase{
		{
			Name: "set-home",
			Run: func(_ context.Context, s *State) error {
				s.Home = "/test/home"
				return nil
			},
		},
		{
			Name: "read-home",
			Run: func(_ context.Context, s *State) error {
				if s.Home != "/test/home" {
					return errors.New("expected Home to be /test/home")
				}
				s.LangoDir = s.Home + "/.lango"
				return nil
			},
		},
		{
			Name: "verify",
			Run: func(_ context.Context, s *State) error {
				if s.LangoDir != "/test/home/.lango" {
					return errors.New("expected LangoDir to be /test/home/.lango")
				}
				return nil
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.NoError(t, err)
}

func TestPipeline_NilCleanupSkipped(t *testing.T) {
	phases := []Phase{
		{
			Name: "no-cleanup",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			// Cleanup is nil — should not panic.
		},
		{
			Name: "fail",
			Run: func(_ context.Context, _ *State) error {
				return errors.New("fail")
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.Error(t, err)
	// No panic means nil cleanup was properly skipped.
}

func TestPipeline_ErrorWrapsPhaseNameAndCause(t *testing.T) {
	sentinel := errors.New("root cause")
	phases := []Phase{
		{
			Name: "important-phase",
			Run: func(_ context.Context, _ *State) error {
				return sentinel
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "important-phase")
	assert.True(t, errors.Is(err, sentinel))
}

func TestDefaultPhases_Returns7Phases(t *testing.T) {
	phases := DefaultPhases()
	require.Len(t, phases, 7)

	wantNames := []string{
		"ensure data directory",
		"detect encryption",
		"acquire passphrase",
		"open database",
		"load security state",
		"initialize crypto",
		"load profile",
	}

	for i, want := range wantNames {
		assert.Equal(t, want, phases[i].Name, "phase %d name", i)
	}
}

func TestDefaultPhases_OpenDatabaseHasCleanup(t *testing.T) {
	phases := DefaultPhases()
	// Only "open database" (index 3) should have a Cleanup function.
	for i, p := range phases {
		if p.Name == "open database" {
			assert.NotNil(t, p.Cleanup, "phase %d (%s) should have Cleanup", i, p.Name)
		}
	}
}

func TestPipeline_NoCleanupOnSuccess(t *testing.T) {
	var cleaned bool

	phases := []Phase{
		{
			Name: "only-phase",
			Run: func(_ context.Context, _ *State) error {
				return nil
			},
			Cleanup: func(_ *State) {
				cleaned = true
			},
		},
	}

	p := NewPipeline(phases...)
	_, err := p.Execute(context.Background(), Options{})
	require.NoError(t, err)

	// Cleanup should NOT run on success.
	assert.False(t, cleaned)
}
