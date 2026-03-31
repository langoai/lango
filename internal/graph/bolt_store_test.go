package graph

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *BoltStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := NewBoltStore(path)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func TestBoltStore_AddAndQuery(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		give      string
		triples   []Triple
		wantBySub map[string]int // subject -> expected count
		wantByObj map[string]int // object -> expected count
		wantBySP  struct {
			subject   string
			predicate string
			count     int
		}
	}{
		{
			give: "basic relationships",
			triples: []Triple{
				{Subject: "session_1", Predicate: InSession, Object: "obs_1"},
				{Subject: "session_1", Predicate: InSession, Object: "obs_2"},
				{Subject: "session_1", Predicate: Contains, Object: "ref_1"},
				{Subject: "obs_1", Predicate: CausedBy, Object: "error_x"},
			},
			wantBySub: map[string]int{
				"session_1": 3,
				"obs_1":     1,
				"nonexist":  0,
			},
			wantByObj: map[string]int{
				"obs_1":   1,
				"obs_2":   1,
				"error_x": 1,
			},
			wantBySP: struct {
				subject   string
				predicate string
				count     int
			}{
				subject:   "session_1",
				predicate: InSession,
				count:     2,
			},
		},
		{
			give: "with metadata",
			triples: []Triple{
				{
					Subject:   "fix_1",
					Predicate: ResolvedBy,
					Object:    "error_1",
					Metadata:  map[string]string{"source": "obs_42"},
				},
			},
			wantBySub: map[string]int{
				"fix_1": 1,
			},
			wantByObj: map[string]int{
				"error_1": 1,
			},
			wantBySP: struct {
				subject   string
				predicate string
				count     int
			}{
				subject:   "fix_1",
				predicate: ResolvedBy,
				count:     1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			store := newTestStore(t)

			for _, triple := range tt.triples {
				require.NoError(t, store.AddTriple(ctx, triple))
			}

			// QueryBySubject
			for subject, wantCount := range tt.wantBySub {
				got, err := store.QueryBySubject(ctx, subject)
				require.NoError(t, err)
				assert.Len(t, got, wantCount, "QueryBySubject(%q)", subject)
			}

			// QueryByObject
			for object, wantCount := range tt.wantByObj {
				got, err := store.QueryByObject(ctx, object)
				require.NoError(t, err)
				assert.Len(t, got, wantCount, "QueryByObject(%q)", object)
			}

			// QueryBySubjectPredicate
			sp := tt.wantBySP
			got, err := store.QueryBySubjectPredicate(ctx, sp.subject, sp.predicate)
			require.NoError(t, err)
			assert.Len(t, got, sp.count, "QueryBySubjectPredicate(%q, %q)", sp.subject, sp.predicate)
		})
	}
}

func TestBoltStore_AddAndQuery_Metadata(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	triple := Triple{
		Subject:   "learning_1",
		Predicate: LearnedFrom,
		Object:    "session_5",
		Metadata:  map[string]string{"confidence": "high", "source": "reflector"},
	}
	require.NoError(t, store.AddTriple(ctx, triple))

	got, err := store.QueryBySubject(ctx, "learning_1")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "high", got[0].Metadata["confidence"])
	assert.Equal(t, "reflector", got[0].Metadata["source"])
}

func TestBoltStore_RemoveTriple(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	triple := Triple{Subject: "A", Predicate: RelatedTo, Object: "B"}
	require.NoError(t, store.AddTriple(ctx, triple))

	// Verify present before removal.
	got, err := store.QueryBySubject(ctx, "A")
	require.NoError(t, err)
	assert.Len(t, got, 1)

	// Remove and verify absent in all indexes.
	require.NoError(t, store.RemoveTriple(ctx, triple))

	bySub, err := store.QueryBySubject(ctx, "A")
	require.NoError(t, err)
	assert.Empty(t, bySub, "should be absent by subject")

	byObj, err := store.QueryByObject(ctx, "B")
	require.NoError(t, err)
	assert.Empty(t, byObj, "should be absent by object")

	bySP, err := store.QueryBySubjectPredicate(ctx, "A", RelatedTo)
	require.NoError(t, err)
	assert.Empty(t, bySP, "should be absent by subject+predicate")
}

func TestBoltStore_RemoveTriple_Nonexistent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Removing a non-existent triple should not error.
	err := store.RemoveTriple(ctx, Triple{Subject: "X", Predicate: RelatedTo, Object: "Y"})
	assert.NoError(t, err)
}

func TestBoltStore_Traverse(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Build graph: A → B → C → D (all via "follows").
	triples := []Triple{
		{Subject: "A", Predicate: Follows, Object: "B"},
		{Subject: "B", Predicate: Follows, Object: "C"},
		{Subject: "C", Predicate: Follows, Object: "D"},
	}
	for _, triple := range triples {
		require.NoError(t, store.AddTriple(ctx, triple))
	}

	tests := []struct {
		give       string
		startNode  string
		maxDepth   int
		predicates []string
		wantNodes  []string // expected objects in result (sorted)
	}{
		{
			give:      "depth 1 from A",
			startNode: "A",
			maxDepth:  1,
			wantNodes: []string{"B"},
		},
		{
			give:      "depth 2 from A finds B and C",
			startNode: "A",
			maxDepth:  2,
			wantNodes: []string{"B", "C"},
		},
		{
			give:      "depth 3 from A finds B, C, D",
			startNode: "A",
			maxDepth:  3,
			wantNodes: []string{"B", "C", "D"},
		},
		{
			give:      "depth 0 finds nothing",
			startNode: "A",
			maxDepth:  0,
			wantNodes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := store.Traverse(ctx, tt.startNode, tt.maxDepth, tt.predicates)
			require.NoError(t, err)

			// Collect discovered nodes (excluding start node).
			nodes := make(map[string]bool)
			for _, triple := range got {
				if triple.Object != tt.startNode {
					nodes[triple.Object] = true
				}
				if triple.Subject != tt.startNode {
					nodes[triple.Subject] = true
				}
			}

			var gotNodes []string
			for n := range nodes {
				gotNodes = append(gotNodes, n)
			}
			sort.Strings(gotNodes)
			sort.Strings(tt.wantNodes)

			assert.Equal(t, tt.wantNodes, gotNodes)
		})
	}
}

func TestBoltStore_TraverseWithPredicates(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Build a graph with mixed predicates:
	// A --follows--> B --follows--> C
	// A --related_to--> X
	triples := []Triple{
		{Subject: "A", Predicate: Follows, Object: "B"},
		{Subject: "B", Predicate: Follows, Object: "C"},
		{Subject: "A", Predicate: RelatedTo, Object: "X"},
	}
	for _, triple := range triples {
		require.NoError(t, store.AddTriple(ctx, triple))
	}

	tests := []struct {
		give       string
		predicates []string
		wantNodes  []string
	}{
		{
			give:       "filter by follows only",
			predicates: []string{Follows},
			wantNodes:  []string{"B", "C"},
		},
		{
			give:       "filter by related_to only",
			predicates: []string{RelatedTo},
			wantNodes:  []string{"X"},
		},
		{
			give:       "empty predicates returns all",
			predicates: nil,
			wantNodes:  []string{"B", "C", "X"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := store.Traverse(ctx, "A", 3, tt.predicates)
			require.NoError(t, err)

			nodes := make(map[string]bool)
			for _, triple := range got {
				if triple.Object != "A" {
					nodes[triple.Object] = true
				}
				if triple.Subject != "A" {
					nodes[triple.Subject] = true
				}
			}

			var gotNodes []string
			for n := range nodes {
				gotNodes = append(gotNodes, n)
			}
			sort.Strings(gotNodes)
			sort.Strings(tt.wantNodes)

			assert.Equal(t, tt.wantNodes, gotNodes)
		})
	}
}

func TestBoltStore_TraverseCyclePrevention(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Build a cycle: A → B → C → A.
	triples := []Triple{
		{Subject: "A", Predicate: Follows, Object: "B"},
		{Subject: "B", Predicate: Follows, Object: "C"},
		{Subject: "C", Predicate: Follows, Object: "A"},
	}
	for _, triple := range triples {
		require.NoError(t, store.AddTriple(ctx, triple))
	}

	// Even with a high depth, we should not loop forever.
	got, err := store.Traverse(ctx, "A", 10, nil)
	require.NoError(t, err)

	// Should find edges to B, C — and the cycle edge back to A, but not loop.
	assert.True(t, len(got) >= 2 && len(got) <= 6, "expected bounded results, got %d", len(got))
}

func TestBoltStore_AddTriples(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	triples := []Triple{
		{Subject: "S1", Predicate: RelatedTo, Object: "O1"},
		{Subject: "S1", Predicate: CausedBy, Object: "O2"},
		{Subject: "S2", Predicate: ResolvedBy, Object: "O3"},
	}
	require.NoError(t, store.AddTriples(ctx, triples))

	// Verify all three are present.
	got, err := store.QueryBySubject(ctx, "S1")
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = store.QueryBySubject(ctx, "S2")
	require.NoError(t, err)
	assert.Len(t, got, 1)

	// Verify reverse index.
	got, err = store.QueryByObject(ctx, "O2")
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "S1", got[0].Subject)
	assert.Equal(t, CausedBy, got[0].Predicate)
}

func TestBoltStore_AddTriples_Empty(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Adding empty slice should not error.
	err := store.AddTriples(ctx, nil)
	assert.NoError(t, err)
}

func TestBoltStore_QueryEmptyStore(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	got, err := store.QueryBySubject(ctx, "anything")
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = store.QueryByObject(ctx, "anything")
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = store.QueryBySubjectPredicate(ctx, "anything", RelatedTo)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestBoltStore_TypedTripleRoundTrip(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.AddTriple(ctx, Triple{
		Subject:     "error:timeout",
		SubjectType: "ErrorPattern",
		Predicate:   CausedBy,
		Object:      "tool:api_call",
		ObjectType:  "Tool",
	})
	require.NoError(t, err)

	triples, err := store.QueryBySubject(ctx, "error:timeout")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.Equal(t, "ErrorPattern", triples[0].SubjectType)
	assert.Equal(t, "Tool", triples[0].ObjectType)
	assert.Equal(t, "ErrorPattern", triples[0].Metadata["_subject_type"])
	assert.Equal(t, "Tool", triples[0].Metadata["_object_type"])
}

func TestBoltStore_UntypedTripleBackwardCompat(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Store without type fields (legacy behavior)
	err := store.AddTriple(ctx, Triple{
		Subject:   "a",
		Predicate: RelatedTo,
		Object:    "b",
	})
	require.NoError(t, err)

	triples, err := store.QueryBySubject(ctx, "a")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.Equal(t, "", triples[0].SubjectType)
	assert.Equal(t, "", triples[0].ObjectType)
}

func TestBoltStore_PredicateValidator(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	validator := func(name string) bool {
		return name == CausedBy || name == RelatedTo
	}
	store.SetPredicateValidator(validator)

	// Valid predicate accepted
	err := store.AddTriple(ctx, Triple{Subject: "a", Predicate: CausedBy, Object: "b"})
	require.NoError(t, err)

	// Invalid predicate rejected
	err = store.AddTriple(ctx, Triple{Subject: "a", Predicate: "fake_pred", Object: "b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown predicate")
}

func TestBoltStore_NoValidatorAcceptsAll(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// No validator set — any predicate accepted
	err := store.AddTriple(ctx, Triple{Subject: "a", Predicate: "any_pred", Object: "b"})
	require.NoError(t, err)

	triples, err := store.QueryBySubject(ctx, "a")
	require.NoError(t, err)
	assert.Len(t, triples, 1)
}

func TestBoltStore_TypedTripleOSPQuery(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.AddTriple(ctx, Triple{
		Subject:     "obs:1",
		SubjectType: "Observation",
		Predicate:   InSession,
		Object:      "session:abc",
		ObjectType:  "Session",
	})
	require.NoError(t, err)

	// Query by object — should restore type fields
	triples, err := store.QueryByObject(ctx, "session:abc")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.Equal(t, "Observation", triples[0].SubjectType)
	assert.Equal(t, "Session", triples[0].ObjectType)
}
