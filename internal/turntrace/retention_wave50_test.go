package turntrace

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWave50RetentionLoggerReturnsNamedLogger(t *testing.T) {
	t.Parallel()

	log := retentionLogger()
	require.NotNil(t, log)
	assert.IsType(t, &zap.SugaredLogger{}, log)
}

func TestWave50RetentionCleanupPurgesSuccessFailedAndExcessTraces(t *testing.T) {
	t.Parallel()

	store := &wave50RetentionStore{
		oldResponses: [][]string{
			{"success-old"},
			{"failed-old"},
			{"excess-old"},
		},
		traceCount: 5,
	}
	cleaner := NewRetentionCleaner(store, RetentionConfig{
		MaxAge:                24 * time.Hour,
		MaxTraces:             3,
		FailedTraceMultiplier: 3,
	})

	cleaner.cleanup()

	require.Len(t, store.oldCalls, 3)
	assert.True(t, store.oldCalls[0].onlySuccess)
	assert.Equal(t, 500, store.oldCalls[0].limit)
	assert.False(t, store.oldCalls[1].onlySuccess)
	assert.Equal(t, 500, store.oldCalls[1].limit)
	assert.False(t, store.oldCalls[2].onlySuccess)
	assert.Equal(t, 2, store.oldCalls[2].limit)

	assert.Equal(t, [][]string{
		{"success-old"},
		{"failed-old"},
		{"excess-old"},
	}, store.purged)
	assert.Equal(t, 1, store.traceCountCalls)
}

func TestWave50RetentionCleanupContinuesAfterQueryAndPurgeErrors(t *testing.T) {
	t.Parallel()

	store := &wave50RetentionStore{
		oldResponses: [][]string{
			nil,
			{"failed-old"},
			{"excess-old"},
		},
		oldErrs: []error{
			errors.New("success query failed"),
			nil,
			nil,
		},
		purgeErrs: []error{
			errors.New("failed purge failed"),
			nil,
		},
		traceCount: 4,
	}
	cleaner := NewRetentionCleaner(store, RetentionConfig{
		MaxAge:                time.Hour,
		MaxTraces:             2,
		FailedTraceMultiplier: 2,
	})

	require.NotPanics(t, cleaner.cleanup)

	require.Len(t, store.oldCalls, 3)
	assert.Equal(t, [][]string{
		{"failed-old"},
		{"excess-old"},
	}, store.purged)
	assert.Equal(t, 1, store.traceCountCalls)
}

func TestWave50RetentionCleanupStopsExcessPassWhenTraceCountFails(t *testing.T) {
	t.Parallel()

	store := &wave50RetentionStore{
		oldResponses: [][]string{
			nil,
			nil,
		},
		traceCountErr: errors.New("count failed"),
	}
	cleaner := NewRetentionCleaner(store, RetentionConfig{MaxAge: time.Hour, MaxTraces: 1})

	cleaner.cleanup()

	require.Len(t, store.oldCalls, 2)
	assert.Equal(t, 1, store.traceCountCalls)
	assert.Empty(t, store.purged)
}

type wave50RetentionStore struct {
	oldResponses [][]string
	oldErrs      []error
	oldCalls     []wave50OldTraceCall

	purgeErrs []error
	purged    [][]string

	traceCount      int
	traceCountErr   error
	traceCountCalls int
}

type wave50OldTraceCall struct {
	cutoff      time.Time
	onlySuccess bool
	limit       int
}

func (s *wave50RetentionStore) CreateTrace(context.Context, Trace) error { return nil }
func (s *wave50RetentionStore) AppendEvent(context.Context, Event) error { return nil }
func (s *wave50RetentionStore) FinishTrace(context.Context, string, Outcome, string, string, string, string, time.Time) error {
	return nil
}
func (s *wave50RetentionStore) RecentFailures(context.Context, int) ([]Trace, error) {
	return nil, nil
}
func (s *wave50RetentionStore) IsolationLeakCount(context.Context, []string) (int, error) {
	return 0, nil
}
func (s *wave50RetentionStore) EventsForTrace(context.Context, string) ([]Event, error) {
	return nil, nil
}
func (s *wave50RetentionStore) TracesForSession(context.Context, string) ([]Trace, error) {
	return nil, nil
}
func (s *wave50RetentionStore) RecentByOutcome(context.Context, Outcome, time.Time, int) ([]Trace, error) {
	return nil, nil
}

func (s *wave50RetentionStore) PurgeTraces(_ context.Context, traceIDs []string) error {
	copied := append([]string(nil), traceIDs...)
	s.purged = append(s.purged, copied)
	idx := len(s.purged) - 1
	if idx < len(s.purgeErrs) {
		return s.purgeErrs[idx]
	}
	return nil
}

func (s *wave50RetentionStore) TraceCount(context.Context) (int, error) {
	s.traceCountCalls++
	if s.traceCountErr != nil {
		return 0, s.traceCountErr
	}
	return s.traceCount, nil
}

func (s *wave50RetentionStore) OldTraces(_ context.Context, cutoff time.Time, onlySuccess bool, limit int) ([]string, error) {
	s.oldCalls = append(s.oldCalls, wave50OldTraceCall{
		cutoff:      cutoff,
		onlySuccess: onlySuccess,
		limit:       limit,
	})
	idx := len(s.oldCalls) - 1
	if idx < len(s.oldErrs) && s.oldErrs[idx] != nil {
		return nil, s.oldErrs[idx]
	}
	if idx < len(s.oldResponses) {
		return append([]string(nil), s.oldResponses[idx]...), nil
	}
	return nil, nil
}
