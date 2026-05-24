package graph

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

type failingGraphStore struct {
	err error
}

func (s *failingGraphStore) AddTriple(context.Context, Triple) error    { return s.err }
func (s *failingGraphStore) AddTriples(context.Context, []Triple) error { return s.err }
func (s *failingGraphStore) RemoveTriple(context.Context, Triple) error { return s.err }
func (s *failingGraphStore) QueryBySubject(context.Context, string) ([]Triple, error) {
	return nil, nil
}
func (s *failingGraphStore) QueryByObject(context.Context, string) ([]Triple, error) {
	return nil, nil
}
func (s *failingGraphStore) QueryBySubjectPredicate(context.Context, string, string) ([]Triple, error) {
	return nil, nil
}
func (s *failingGraphStore) Traverse(context.Context, string, int, []string) ([]Triple, error) {
	return nil, nil
}
func (s *failingGraphStore) Count(context.Context) (int, error) { return 0, nil }
func (s *failingGraphStore) PredicateStats(context.Context) (map[string]int, error) {
	return nil, nil
}
func (s *failingGraphStore) AllTriples(context.Context) ([]Triple, error) {
	return nil, nil
}
func (s *failingGraphStore) ClearAll(context.Context) error { return nil }
func (s *failingGraphStore) Close() error                   { return nil }

func TestGraphBuffer_ProcessBatch_PublishesWriteFailureBaseline(t *testing.T) {
	t.Parallel()

	store := &failingGraphStore{err: errors.New("write failed")}
	buffer := NewGraphBuffer(store, zap.NewNop().Sugar())
	bus := eventbus.New()
	buffer.SetEventBus(bus)

	var got []eventbus.GraphAdmissionWriteFailureEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionWriteFailureEvent) {
		got = append(got, evt)
	})

	buffer.processBatchRequests([]GraphRequest{{
		Triples:                  []Triple{{Subject: "a", Predicate: "likes", Object: "b"}},
		EmitWriteFailureBaseline: true,
	}})

	require.Len(t, got, 1)
	assert.Equal(t, eventbus.GraphAdmissionWriteFailureEvent{BatchCount: 1}, got[0])
}

func TestGraphBuffer_ProcessBatch_SuppressesWriteFailureBaselineWithoutFlag(t *testing.T) {
	t.Parallel()

	store := &failingGraphStore{err: errors.New("write failed")}
	buffer := NewGraphBuffer(store, zap.NewNop().Sugar())
	bus := eventbus.New()
	buffer.SetEventBus(bus)

	var got []eventbus.GraphAdmissionWriteFailureEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionWriteFailureEvent) {
		got = append(got, evt)
	})

	buffer.processBatchRequests([]GraphRequest{{
		Triples: []Triple{{Subject: "a", Predicate: "likes", Object: "b"}},
	}})

	assert.Empty(t, got)
}
