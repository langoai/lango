package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
)

func ptrString(v string) *string {
	return &v
}

type predicateCapturingStore struct {
	validator graph.PredicateValidatorFunc
}

func (s *predicateCapturingStore) SetPredicateValidator(v graph.PredicateValidatorFunc) {
	s.validator = v
}

func (s *predicateCapturingStore) AddTriple(context.Context, graph.Triple) error { return nil }

func (s *predicateCapturingStore) AddTriples(context.Context, []graph.Triple) error { return nil }

func (s *predicateCapturingStore) RemoveTriple(context.Context, graph.Triple) error { return nil }

func (s *predicateCapturingStore) QueryBySubject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *predicateCapturingStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *predicateCapturingStore) QueryBySubjectPredicate(context.Context, string, string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *predicateCapturingStore) Traverse(context.Context, string, int, []string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *predicateCapturingStore) Count(context.Context) (int, error) { return 0, nil }

func (s *predicateCapturingStore) PredicateStats(context.Context) (map[string]int, error) {
	return nil, nil
}

func (s *predicateCapturingStore) AllTriples(context.Context) ([]graph.Triple, error) {
	return nil, nil
}

func (s *predicateCapturingStore) ClearAll(context.Context) error { return nil }

func (s *predicateCapturingStore) Close() error { return nil }

type failingRuntimeGraphStore struct {
	err error
}

func (s *failingRuntimeGraphStore) AddTriple(context.Context, graph.Triple) error    { return s.err }
func (s *failingRuntimeGraphStore) AddTriples(context.Context, []graph.Triple) error { return s.err }
func (s *failingRuntimeGraphStore) RemoveTriple(context.Context, graph.Triple) error { return s.err }
func (s *failingRuntimeGraphStore) QueryBySubject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingRuntimeGraphStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingRuntimeGraphStore) QueryBySubjectPredicate(context.Context, string, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingRuntimeGraphStore) Traverse(context.Context, string, int, []string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingRuntimeGraphStore) Count(context.Context) (int, error) { return 0, nil }
func (s *failingRuntimeGraphStore) PredicateStats(context.Context) (map[string]int, error) {
	return nil, nil
}
func (s *failingRuntimeGraphStore) AllTriples(context.Context) ([]graph.Triple, error) {
	return nil, nil
}
func (s *failingRuntimeGraphStore) ClearAll(context.Context) error { return nil }
func (s *failingRuntimeGraphStore) Close() error                   { return nil }

type fakeContentSearchSource struct {
	calls   []fakeContentSearchCall
	results []knowledge.ScoredKnowledgeEntry
	err     error
}

type fakeContentSearchCall struct {
	query    string
	category string
	limit    int
}

func (s *fakeContentSearchSource) SearchKnowledgeScored(_ context.Context, query string, category string, limit int) ([]knowledge.ScoredKnowledgeEntry, error) {
	s.calls = append(s.calls, fakeContentSearchCall{query: query, category: category, limit: limit})
	if s.err != nil {
		return nil, s.err
	}
	return append([]knowledge.ScoredKnowledgeEntry(nil), s.results...), nil
}

func TestKnowledgeContentRetrieverRetrieveHandlesFiltersLimitsAndMapping(t *testing.T) {
	t.Parallel()

	t.Run("nil store and empty query skip search", func(t *testing.T) {
		t.Parallel()

		got, err := (&knowledgeContentRetriever{}).Retrieve(context.Background(), "query", graph.ContentRetrieveOptions{})
		require.NoError(t, err)
		assert.Nil(t, got)

		source := &fakeContentSearchSource{}
		got, err = (&knowledgeContentRetriever{store: source}).Retrieve(context.Background(), "", graph.ContentRetrieveOptions{})
		require.NoError(t, err)
		assert.Nil(t, got)
		assert.Empty(t, source.calls)
	})

	t.Run("non knowledge collection skips search", func(t *testing.T) {
		t.Parallel()

		source := &fakeContentSearchSource{}
		got, err := (&knowledgeContentRetriever{store: source}).Retrieve(context.Background(), "query", graph.ContentRetrieveOptions{
			Collections: []string{"learning"},
		})
		require.NoError(t, err)
		assert.Nil(t, got)
		assert.Empty(t, source.calls)
	})

	t.Run("default limit maps scored knowledge results", func(t *testing.T) {
		t.Parallel()

		source := &fakeContentSearchSource{results: []knowledge.ScoredKnowledgeEntry{
			{
				Entry: knowledge.KnowledgeEntry{
					Key:     "knowledge-1",
					Content: "Graph retrieval note",
				},
				Score: 0.75,
			},
		}}

		got, err := (&knowledgeContentRetriever{store: source}).Retrieve(context.Background(), "graph", graph.ContentRetrieveOptions{})
		require.NoError(t, err)
		assert.Equal(t, []fakeContentSearchCall{{query: "graph", limit: 5}}, source.calls)
		require.Len(t, got, 1)
		assert.Equal(t, graph.ContentResult{
			Collection: "knowledge",
			SourceID:   "knowledge-1",
			Content:    "Graph retrieval note",
			Score:      0.75,
		}, got[0])
	})

	t.Run("knowledge collection uses explicit limit and propagates errors", func(t *testing.T) {
		t.Parallel()

		source := &fakeContentSearchSource{err: errors.New("search failed")}
		got, err := (&knowledgeContentRetriever{store: source}).Retrieve(context.Background(), "graph", graph.ContentRetrieveOptions{
			Collections: []string{"learning", "knowledge"},
			Limit:       2,
		})
		require.ErrorContains(t, err, "search failed")
		assert.Nil(t, got)
		assert.Equal(t, []fakeContentSearchCall{{query: "graph", limit: 2}}, source.calls)
	})
}

func TestObserveExtractedTriples_PublishesAdmissionAndPreservesOriginalTriples(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var got eventbus.GraphAdmissionBatchEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionBatchEvent) {
		got = evt
	})

	policy := graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{
		Validator: func(name string) bool {
			return name == graph.CausedBy
		},
	}, zap.NewNop().Sugar())

	triples, emitWriteFailureBaseline := observeExtractedTriples(policy, bus, eventbus.TriplesExtractedEvent{
		Source: string(graph.AdmissionSourceConversationAnalysis),
		Triples: []eventbus.Triple{{
			Subject:   "a",
			Predicate: "invented_rel",
			Object:    "b",
		}},
	})

	require.Len(t, triples, 1)
	assert.True(t, emitWriteFailureBaseline)
	assert.Equal(t, "invented_rel", triples[0].Predicate)
	assert.Equal(t, eventbus.GraphAdmissionBatchEvent{
		Source:           string(graph.AdmissionSourceConversationAnalysis),
		ProducerGroup:    ptrString(string(graph.AdmissionProducerGroupLearning)),
		ValidatorSource:  string(graph.AdmissionValidatorSourceOntologyRegistry),
		BatchCount:       1,
		KnownCount:       0,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	}, got)
}

func TestObserveExtractedTriples_PublishesUnmappedAndPreservesOriginalTriples(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var got eventbus.GraphAdmissionUnmappedSourceEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionUnmappedSourceEvent) {
		got = evt
	})

	policy := graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{}, zap.NewNop().Sugar())

	triples, emitWriteFailureBaseline := observeExtractedTriples(policy, bus, eventbus.TriplesExtractedEvent{
		Source: "future_source",
		Triples: []eventbus.Triple{{
			Subject:   "a",
			Predicate: "invented_rel",
			Object:    "b",
		}},
	})

	require.Len(t, triples, 1)
	assert.True(t, emitWriteFailureBaseline)
	assert.Equal(t, "invented_rel", triples[0].Predicate)
	assert.Equal(t, eventbus.GraphAdmissionUnmappedSourceEvent{
		RawSource:  "future_source",
		BatchCount: 1,
	}, got)
}

func TestObserveExtractedTriples_ProducerMappingCoverage(t *testing.T) {
	t.Parallel()

	policy := graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{
		Validator: func(name string) bool {
			return name == graph.CausedBy
		},
	}, zap.NewNop().Sugar())

	testCases := []struct {
		name              string
		source            string
		expectedGroup     *string
		expectBatch       bool
		expectUnmappedRaw string
	}{
		{
			name:          "conversation_analysis maps to learning",
			source:        string(graph.AdmissionSourceConversationAnalysis),
			expectedGroup: ptrString(string(graph.AdmissionProducerGroupLearning)),
			expectBatch:   true,
		},
		{
			name:          "session_learning maps to learning",
			source:        string(graph.AdmissionSourceSessionLearning),
			expectedGroup: ptrString(string(graph.AdmissionProducerGroupLearning)),
			expectBatch:   true,
		},
		{
			name:          "learning maps to learning",
			source:        string(graph.AdmissionSourceLearning),
			expectedGroup: ptrString(string(graph.AdmissionProducerGroupLearning)),
			expectBatch:   true,
		},
		{
			name:          "proactive_librarian maps to librarian",
			source:        string(graph.AdmissionSourceProactiveLibrarian),
			expectedGroup: ptrString(string(graph.AdmissionProducerGroupLibrarian)),
			expectBatch:   true,
		},
		{
			name:              "fallback source emits unmapped",
			source:            "future_source",
			expectUnmappedRaw: "future_source",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bus := eventbus.New()
			var gotBatch *eventbus.GraphAdmissionBatchEvent
			var gotUnmapped *eventbus.GraphAdmissionUnmappedSourceEvent
			eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionBatchEvent) {
				event := evt
				gotBatch = &event
			})
			eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionUnmappedSourceEvent) {
				event := evt
				gotUnmapped = &event
			})

			triples, emitWriteFailureBaseline := observeExtractedTriples(policy, bus, eventbus.TriplesExtractedEvent{
				Source: tc.source,
				Triples: []eventbus.Triple{{
					Subject:   "a",
					Predicate: graph.CausedBy,
					Object:    "b",
				}},
			})

			require.Len(t, triples, 1)
			assert.Equal(t, graph.CausedBy, triples[0].Predicate)

			if tc.expectBatch {
				assert.True(t, emitWriteFailureBaseline)
				require.NotNil(t, gotBatch)
				assert.Equal(t, tc.source, gotBatch.Source)
				assert.Equal(t, tc.expectedGroup, gotBatch.ProducerGroup)
				assert.Equal(t, string(graph.AdmissionValidatorSourceOntologyRegistry), gotBatch.ValidatorSource)
				assert.Equal(t, 1, gotBatch.BatchCount)
				assert.Equal(t, 1, gotBatch.KnownCount)
				assert.Equal(t, 0, gotBatch.UnknownCount)
				assert.Equal(t, 0, gotBatch.UnvalidatedCount)
				assert.Nil(t, gotUnmapped)
				return
			}

			assert.Nil(t, gotBatch)
			require.NotNil(t, gotUnmapped)
			assert.True(t, emitWriteFailureBaseline)
			assert.Equal(t, tc.expectUnmappedRaw, gotUnmapped.RawSource)
			assert.Equal(t, 1, gotUnmapped.BatchCount)
		})
	}
}

func TestObserveContentSavedTriples_PublishesAdmissionAndPreservesOriginalTriples(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var got eventbus.GraphAdmissionBatchEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionBatchEvent) {
		got = evt
	})

	policy := graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{
		Validator: func(name string) bool {
			return name == graph.CausedBy
		},
	}, zap.NewNop().Sugar())

	triples, emitWriteFailureBaseline := observeContentSavedTriples(policy, bus, []graph.Triple{{
		Subject:   "a",
		Predicate: graph.CausedBy,
		Object:    "b",
	}})

	require.Len(t, triples, 1)
	assert.True(t, emitWriteFailureBaseline)
	assert.Equal(t, graph.CausedBy, triples[0].Predicate)
	assert.Equal(t, eventbus.GraphAdmissionBatchEvent{
		Source:           string(graph.AdmissionSourceContentSavedExtractor),
		ProducerGroup:    nil,
		ValidatorSource:  string(graph.AdmissionValidatorSourceOntologyRegistry),
		BatchCount:       1,
		KnownCount:       1,
		UnknownCount:     0,
		UnvalidatedCount: 0,
	}, got)
}

func TestGraphAdmissionWiring_UsesSharedValidatorClosure(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Ontology.Governance.AdmissionMode = config.OntologyAdmissionModeObserve

	var validatorCalls []string
	validator := func(name string) bool {
		validatorCalls = append(validatorCalls, name)
		return name == graph.CausedBy
	}

	policy := newGraphAdmissionPolicy(cfg, validator)
	require.NotNil(t, policy)

	store := &predicateCapturingStore{}
	injectGraphPredicateValidator(store, validator)
	require.NotNil(t, store.validator)

	bus := eventbus.New()
	var got eventbus.GraphAdmissionBatchEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionBatchEvent) {
		got = evt
	})

	assert.True(t, store.validator(graph.CausedBy))

	triples, emitWriteFailureBaseline := observeExtractedTriples(policy, bus, eventbus.TriplesExtractedEvent{
		Source: string(graph.AdmissionSourceConversationAnalysis),
		Triples: []eventbus.Triple{{
			Subject:   "a",
			Predicate: graph.CausedBy,
			Object:    "b",
		}},
	})

	require.Len(t, triples, 1)
	assert.True(t, emitWriteFailureBaseline)
	assert.Equal(t, []string{graph.CausedBy, graph.CausedBy}, validatorCalls)
	assert.Equal(t, string(graph.AdmissionValidatorSourceOntologyRegistry), got.ValidatorSource)
}

func TestGraphAdmissionWiring_UsesUnavailableValidatorSourceWhenOntologyUnavailable(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Ontology.Governance.AdmissionMode = config.OntologyAdmissionModeObserve

	policy := newGraphAdmissionPolicy(cfg, nil)
	require.NotNil(t, policy)

	bus := eventbus.New()
	var got eventbus.GraphAdmissionBatchEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionBatchEvent) {
		got = evt
	})

	triples, emitWriteFailureBaseline := observeExtractedTriples(policy, bus, eventbus.TriplesExtractedEvent{
		Source: string(graph.AdmissionSourceConversationAnalysis),
		Triples: []eventbus.Triple{{
			Subject:   "a",
			Predicate: "invented_rel",
			Object:    "b",
		}},
	})

	require.Len(t, triples, 1)
	assert.True(t, emitWriteFailureBaseline)
	assert.Equal(t, string(graph.AdmissionValidatorSourceUnavailable), got.ValidatorSource)
	assert.Equal(t, 0, got.KnownCount)
	assert.Equal(t, 0, got.UnknownCount)
	assert.Equal(t, 1, got.UnvalidatedCount)
}

func TestContentSavedDroppedUnknownObserver_PublishesOnlyWhenObserveModeOn(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		policy        *graph.AdmissionPolicy
		expectPublish bool
	}{
		{
			name: "observe mode on publishes",
			policy: graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{
				Validator: func(name string) bool { return name == graph.CausedBy },
			}, zap.NewNop().Sugar()),
			expectPublish: true,
		},
		{
			name:          "observe mode off suppresses",
			policy:        nil,
			expectPublish: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bus := eventbus.New()
			var published []eventbus.GraphExtractorDroppedUnknownEvent
			eventbus.SubscribeTyped(bus, func(evt eventbus.GraphExtractorDroppedUnknownEvent) {
				published = append(published, evt)
			})

			observer := contentSavedDroppedUnknownObserver(tc.policy, bus)
			observer(graph.DroppedUnknownPredicateEvent{
				Source:    string(graph.AdmissionSourceContentSavedExtractor),
				SourceID:  "doc-1",
				Subject:   "a",
				Predicate: "invented_rel",
				Object:    "b",
			})

			if tc.expectPublish {
				require.Len(t, published, 1)
				assert.Equal(t, eventbus.GraphExtractorDroppedUnknownEvent{
					Source:    string(graph.AdmissionSourceContentSavedExtractor),
					SourceID:  "doc-1",
					Subject:   "a",
					Predicate: "invented_rel",
					Object:    "b",
				}, published[0])
				return
			}

			assert.Empty(t, published)
		})
	}
}

func TestWireGraphCallbacks_ContentSavedContainmentPublishesWriteFailureBaselineInObserveMode(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	store := &failingRuntimeGraphStore{err: errors.New("boom")}
	buffer := graph.NewGraphBuffer(store, zap.NewNop().Sugar())
	buffer.SetEventBus(bus)

	var got []eventbus.GraphAdmissionWriteFailureEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionWriteFailureEvent) {
		got = append(got, evt)
	})

	var wg sync.WaitGroup
	buffer.Start(&wg)

	gc := &graphComponents{
		store:  store,
		buffer: buffer,
		admissionPolicy: graph.NewAdmissionPolicy(graph.AdmissionPolicyConfig{
			Validator: func(string) bool { return true },
		}, zap.NewNop().Sugar()),
	}
	cfg := config.DefaultConfig()
	wireGraphCallbacks(gc, nil, nil, nil, cfg, bus, nil)

	bus.Publish(eventbus.ContentSavedEvent{
		ID:         "doc-1",
		Collection: "knowledge",
		NeedsGraph: true,
	})

	buffer.Stop()
	wg.Wait()

	require.Len(t, got, 1)
	assert.Equal(t, eventbus.GraphAdmissionWriteFailureEvent{BatchCount: 1}, got[0])
}

func TestRuntimeGraphTripleCallback_PublishesWriteFailureBaselineInObserveMode(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	store := &failingRuntimeGraphStore{err: errors.New("boom")}
	buffer := graph.NewGraphBuffer(store, zap.NewNop().Sugar())
	buffer.SetEventBus(bus)

	var got []eventbus.GraphAdmissionWriteFailureEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.GraphAdmissionWriteFailureEvent) {
		got = append(got, evt)
	})

	var wg sync.WaitGroup
	buffer.Start(&wg)

	hooks := memory.NewGraphHooks(runtimeGraphTripleCallback(buffer, true), zap.NewNop().Sugar())
	hooks.OnObservation(memory.Observation{
		ID:         uuid.New(),
		SessionKey: "sess-1",
		CreatedAt:  time.Now(),
	}, "")

	buffer.Stop()
	wg.Wait()

	require.Len(t, got, 1)
	assert.Equal(t, eventbus.GraphAdmissionWriteFailureEvent{BatchCount: 1}, got[0])
}
