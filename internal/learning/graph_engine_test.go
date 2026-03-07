package learning

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/graph"
)

// Verify GraphEngine satisfies ToolResultObserver.
var _ ToolResultObserver = (*GraphEngine)(nil)

// fakeGraphStore is a minimal in-memory graph store for testing.
type fakeGraphStore struct {
	triples []graph.Triple
}

func (s *fakeGraphStore) AddTriple(_ context.Context, t graph.Triple) error {
	s.triples = append(s.triples, t)
	return nil
}

func (s *fakeGraphStore) AddTriples(_ context.Context, ts []graph.Triple) error {
	s.triples = append(s.triples, ts...)
	return nil
}

func (s *fakeGraphStore) RemoveTriple(context.Context, graph.Triple) error { return nil }
func (s *fakeGraphStore) QueryBySubject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *fakeGraphStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *fakeGraphStore) QueryBySubjectPredicate(context.Context, string, string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *fakeGraphStore) Traverse(context.Context, string, int, []string) ([]graph.Triple, error) {
	return nil, nil
}
func (s *fakeGraphStore) Count(context.Context) (int, error)                     { return len(s.triples), nil }
func (s *fakeGraphStore) PredicateStats(context.Context) (map[string]int, error) { return nil, nil }
func (s *fakeGraphStore) ClearAll(context.Context) error                         { s.triples = nil; return nil }
func (s *fakeGraphStore) AllTriples(_ context.Context) ([]graph.Triple, error)    { return s.triples, nil }
func (s *fakeGraphStore) Close() error                                           { return nil }

func TestGraphEngine_RecordErrorGraph_WithCallback(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()

	var callbackTriples []graph.Triple
	ge := &GraphEngine{
		Engine:      &Engine{store: nil, logger: logger},
		graphStore:  nil, // no direct store — only callback
		propagation: 0.3,
		logger:      logger,
	}
	ge.SetGraphCallback(func(triples []graph.Triple) {
		callbackTriples = append(callbackTriples, triples...)
	})

	// Call recordErrorGraph directly (bypasses store.SearchLearningEntities since graphStore is nil).
	ge.recordErrorGraph(context.Background(), "test-session", "exec", fmt.Errorf("permission denied"))

	require.GreaterOrEqual(t, len(callbackTriples), 2, "want at least 2 triples")

	// Check that CausedBy and InSession triples are present.
	var hasCausedBy, hasInSession bool
	for _, triple := range callbackTriples {
		if triple.Predicate == graph.CausedBy {
			hasCausedBy = true
		}
		if triple.Predicate == graph.InSession {
			hasInSession = true
		}
	}

	assert.True(t, hasCausedBy, "want CausedBy triple")
	assert.True(t, hasInSession, "want InSession triple")
}

func TestGraphEngine_RecordErrorGraph_DirectStore(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()

	ge := &GraphEngine{
		Engine:      &Engine{store: nil, logger: logger},
		graphStore:  nil,
		propagation: 0.3,
		logger:      logger,
	}
	// No callback — triples go to store directly.
	ge.graphStore = nil // force callback path only
	ge.SetGraphCallback(nil)

	// With both nil, recordErrorGraph should just return (no panic).
	ge.recordErrorGraph(context.Background(), "s1", "tool1", fmt.Errorf("test error"))
	// No panic = success
}

func TestGraphEngine_RecordFix(t *testing.T) {
	t.Parallel()

	gs := &fakeGraphStore{}
	logger := zap.NewNop().Sugar()

	ge := &GraphEngine{
		Engine:      &Engine{store: nil, logger: logger},
		graphStore:  gs,
		propagation: 0.3,
		logger:      logger,
	}

	// Without callback — should use direct store.
	ge.RecordFix(context.Background(), "timeout error", "increase timeout", "session-1")

	require.Len(t, gs.triples, 2)

	var hasResolvedBy, hasLearnedFrom bool
	for _, triple := range gs.triples {
		if triple.Predicate == graph.ResolvedBy {
			hasResolvedBy = true
		}
		if triple.Predicate == graph.LearnedFrom {
			hasLearnedFrom = true
		}
	}

	assert.True(t, hasResolvedBy, "want ResolvedBy triple")
	assert.True(t, hasLearnedFrom, "want LearnedFrom triple")
}

func TestGraphEngine_RecordFixWithCallback(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()

	var callbackTriples []graph.Triple
	ge := &GraphEngine{
		Engine:      &Engine{store: nil, logger: logger},
		graphStore:  nil,
		propagation: 0.3,
		logger:      logger,
	}
	ge.SetGraphCallback(func(triples []graph.Triple) {
		callbackTriples = append(callbackTriples, triples...)
	})

	ge.RecordFix(context.Background(), "some error", "some fix", "session-2")

	require.Len(t, callbackTriples, 2, "want 2 triples via callback")
}

func TestSanitizeForNode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{give: "hello world", want: "hello_world"},
		{give: "foo@bar.com", want: "foo_bar_com"},
		{give: "a-b_c:d", want: "a-b_c:d"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := sanitizeForNode(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeForNode_Truncation(t *testing.T) {
	t.Parallel()

	long := ""
	for range 100 {
		long += "a"
	}

	result := sanitizeForNode(long)
	assert.Len(t, result, 64, "want max length 64")
}
