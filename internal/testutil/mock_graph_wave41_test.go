package testutil_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/testutil"
)

func TestMockGraphStoreCRUDQueriesCountersAndClear(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := testutil.NewMockGraphStore()
	knowsBob := graph.Triple{
		Subject:     "alice",
		Predicate:   "knows",
		Object:      "bob",
		SubjectType: "person",
		ObjectType:  "person",
		Metadata:    map[string]string{"source": "single"},
	}
	followsCarol := graph.Triple{
		Subject:   "alice",
		Predicate: "follows",
		Object:    "carol",
	}
	bobKnowsDana := graph.Triple{
		Subject:   "bob",
		Predicate: "knows",
		Object:    "dana",
	}

	require.NoError(t, store.AddTriple(ctx, knowsBob))
	require.NoError(t, store.AddTriples(ctx, []graph.Triple{followsCarol, bobKnowsDana}))

	assert.Equal(t, 2, store.AddCalls())
	assert.Equal(t, 3, store.TripleCount())

	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	bySubject, err := store.QueryBySubject(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, []graph.Triple{knowsBob, followsCarol}, bySubject)

	byObject, err := store.QueryByObject(ctx, "bob")
	require.NoError(t, err)
	assert.Equal(t, []graph.Triple{knowsBob}, byObject)

	bySubjectPredicate, err := store.QueryBySubjectPredicate(ctx, "alice", "follows")
	require.NoError(t, err)
	assert.Equal(t, []graph.Triple{followsCarol}, bySubjectPredicate)

	stats, err := store.PredicateStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, map[string]int{"follows": 1, "knows": 2}, stats)

	require.NoError(t, store.RemoveTriple(ctx, graph.Triple{
		Subject:   "missing",
		Predicate: "knows",
		Object:    "nobody",
	}))
	assert.Equal(t, 3, store.TripleCount())

	require.NoError(t, store.RemoveTriple(ctx, followsCarol))
	assert.Equal(t, 2, store.TripleCount())

	bySubjectPredicate, err = store.QueryBySubjectPredicate(ctx, "alice", "follows")
	require.NoError(t, err)
	assert.Empty(t, bySubjectPredicate)

	allTriples, err := store.AllTriples(ctx)
	require.NoError(t, err)
	require.Len(t, allTriples, 2)
	allTriples[0] = graph.Triple{
		Subject:   "mutated",
		Predicate: "mutated",
		Object:    "mutated",
	}
	allTriples = append(allTriples, graph.Triple{
		Subject:   "extra",
		Predicate: "extra",
		Object:    "extra",
	})

	allTriplesAgain, err := store.AllTriples(ctx)
	require.NoError(t, err)
	assert.Equal(t, []graph.Triple{knowsBob, bobKnowsDana}, allTriplesAgain)
	allTriplesAgain[0].Metadata["source"] = "mutated"

	allTriplesAfterMetadataMutation, err := store.AllTriples(ctx)
	require.NoError(t, err)
	assert.Equal(t, "single", allTriplesAfterMetadataMutation[0].Metadata["source"])

	require.NoError(t, store.ClearAll(ctx))
	assert.Equal(t, 0, store.TripleCount())

	count, err = store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	bySubject, err = store.QueryBySubject(ctx, "alice")
	require.NoError(t, err)
	assert.Empty(t, bySubject)

	require.NoError(t, store.Close())
}

func TestMockGraphStoreAddAndQueryErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := testutil.NewMockGraphStore()
	addErr := errors.New("add failed")
	queryErr := errors.New("query failed")
	triple := graph.Triple{
		Subject:   "alice",
		Predicate: "knows",
		Object:    "bob",
	}

	store.AddErr = addErr
	require.ErrorIs(t, store.AddTriple(ctx, triple), addErr)
	require.ErrorIs(t, store.AddTriples(ctx, []graph.Triple{triple}), addErr)
	assert.Equal(t, 2, store.AddCalls())
	assert.Equal(t, 0, store.TripleCount())

	store.AddErr = nil
	require.NoError(t, store.AddTriple(ctx, triple))
	assert.Equal(t, 3, store.AddCalls())
	assert.Equal(t, 1, store.TripleCount())

	store.QueryErr = queryErr
	_, err := store.QueryBySubject(ctx, "alice")
	require.ErrorIs(t, err, queryErr)

	_, err = store.QueryByObject(ctx, "bob")
	require.ErrorIs(t, err, queryErr)

	_, err = store.QueryBySubjectPredicate(ctx, "alice", "knows")
	require.ErrorIs(t, err, queryErr)

	_, err = store.Traverse(ctx, "alice", 2, []string{"knows"})
	require.ErrorIs(t, err, queryErr)

	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestMockGraphStoreTraverseReturnsIncidentTriples(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := testutil.NewMockGraphStore()
	fromStart := graph.Triple{
		Subject:   "start",
		Predicate: "knows",
		Object:    "neighbor",
	}
	toStart := graph.Triple{
		Subject:   "other",
		Predicate: "follows",
		Object:    "start",
	}
	unrelated := graph.Triple{
		Subject:   "other",
		Predicate: "knows",
		Object:    "neighbor",
	}

	require.NoError(t, store.AddTriples(ctx, []graph.Triple{fromStart, toStart, unrelated}))

	got, err := store.Traverse(ctx, "start", 0, nil)
	require.NoError(t, err)
	assert.Equal(t, []graph.Triple{fromStart, toStart}, got)

	got, err = store.Traverse(ctx, "start", 10, []string{"does-not-filter"})
	require.NoError(t, err)
	assert.Equal(t, []graph.Triple{fromStart, toStart}, got)

	got, err = store.Traverse(ctx, "missing", 10, []string{"knows"})
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestMockGraphStoreCopiesMetadataOnWriteAndRead(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := testutil.NewMockGraphStore()
	triple := graph.Triple{
		Subject:   "alice",
		Predicate: "knows",
		Object:    "bob",
		Metadata:  map[string]string{"source": "original"},
	}

	require.NoError(t, store.AddTriple(ctx, triple))
	triple.Metadata["source"] = "mutated-after-add"

	got, err := store.QueryBySubject(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "original", got[0].Metadata["source"])

	got[0].Metadata["source"] = "mutated-after-query"
	gotAgain, err := store.QueryBySubject(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, gotAgain, 1)
	assert.Equal(t, "original", gotAgain[0].Metadata["source"])
}
