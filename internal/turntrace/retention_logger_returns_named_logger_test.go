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

func TestRetentionLoggerReturnsNamedLogger(t *testing.T) {
	t.Parallel()

	log := retentionLogger()
	require.NotNil(t, log)
	assert.IsType(t, &zap.SugaredLogger{}, log)
}

func TestRetentionCleanupPurgesSuccessFailedAndExcessTraces(t *testing.T) {
	t.Parallel()

	store := &retentionLoggerReturnsNamedLoggerRetentionStore{
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

func TestRetentionCleanupContinuesAfterQueryAndPurgeErrors(t *testing.T) {
	t.Parallel()

	store := &retentionLoggerReturnsNamedLoggerRetentionStore{
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

func TestRetentionCleanupStopsExcessPassWhenTraceCountFails(t *testing.T) {
	t.Parallel()

	store := &retentionLoggerReturnsNamedLoggerRetentionStore{
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

type retentionLoggerReturnsNamedLoggerRetentionStore struct {
	oldResponses [][]string
	oldErrs      []error
	oldCalls     []retentionLoggerReturnsNamedLoggerOldTraceCall

	purgeErrs []error
	purged    [][]string

	traceCount      int
	traceCountErr   error
	traceCountCalls int
}

type retentionLoggerReturnsNamedLoggerOldTraceCall struct {
	cutoff      time.Time
	onlySuccess bool
	limit       int
}

func (s *retentionLoggerReturnsNamedLoggerRetentionStore) CreateTrace(context.Context, Trace) error {
	return nil
}
func (s *retentionLoggerReturnsNamedLoggerRetentionStore) AppendEvent(context.Context, Event) error {
	return nil
}
func (s *retentionLoggerReturnsNamedLoggerRetentionStore) FinishTrace(context.Context, string, Outcome, string, string, string, string, time.Time) error {
	return nil
}
func (s *retentionLoggerReturnsNamedLoggerRetentionStore) RecentFailures(context.Context, int) ([]Trace, error) {
	return nil, nil
}
func (s *retentionLoggerReturnsNamedLoggerRetentionStore) IsolationLeakCount(context.Context, []string) (int, error) {
	return 0, nil
}
func (s *retentionLoggerReturnsNamedLoggerRetentionStore) EventsForTrace(context.Context, string) ([]Event, error) {
	return nil, nil
}
func (s *retentionLoggerReturnsNamedLoggerRetentionStore) TracesForSession(context.Context, string) ([]Trace, error) {
	return nil, nil
}
func (s *retentionLoggerReturnsNamedLoggerRetentionStore) RecentByOutcome(context.Context, Outcome, time.Time, int) ([]Trace, error) {
	return nil, nil
}

func (s *retentionLoggerReturnsNamedLoggerRetentionStore) PurgeTraces(_ context.Context, traceIDs []string) error {
	copied := append([]string(nil), traceIDs...)
	s.purged = append(s.purged, copied)
	idx := len(s.purged) - 1
	if idx < len(s.purgeErrs) {
		return s.purgeErrs[idx]
	}
	return nil
}

func (s *retentionLoggerReturnsNamedLoggerRetentionStore) TraceCount(context.Context) (int, error) {
	s.traceCountCalls++
	if s.traceCountErr != nil {
		return 0, s.traceCountErr
	}
	return s.traceCount, nil
}

func (s *retentionLoggerReturnsNamedLoggerRetentionStore) OldTraces(_ context.Context, cutoff time.Time, onlySuccess bool, limit int) ([]string, error) {
	s.oldCalls = append(s.oldCalls, retentionLoggerReturnsNamedLoggerOldTraceCall{
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
