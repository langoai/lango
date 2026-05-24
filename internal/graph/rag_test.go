package graph

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type ragFakeRetriever struct {
	results []ContentResult
	err     error
	query   string
	opts    ContentRetrieveOptions
}

func (r *ragFakeRetriever) Retrieve(_ context.Context, query string, opts ContentRetrieveOptions) ([]ContentResult, error) {
	r.query = query
	r.opts = opts
	if r.err != nil {
		return nil, r.err
	}
	return r.results, nil
}

type ragFakeStore struct {
	triplesByStart map[string][]Triple
	errByStart     map[string]error
	calls          []ragTraverseCall
}

type ragTraverseCall struct {
	start      string
	maxDepth   int
	predicates []string
}

func (s *ragFakeStore) AddTriple(context.Context, Triple) error                  { return nil }
func (s *ragFakeStore) AddTriples(context.Context, []Triple) error               { return nil }
func (s *ragFakeStore) RemoveTriple(context.Context, Triple) error               { return nil }
func (s *ragFakeStore) QueryBySubject(context.Context, string) ([]Triple, error) { return nil, nil }
func (s *ragFakeStore) QueryByObject(context.Context, string) ([]Triple, error)  { return nil, nil }
func (s *ragFakeStore) QueryBySubjectPredicate(context.Context, string, string) ([]Triple, error) {
	return nil, nil
}
func (s *ragFakeStore) Count(context.Context) (int, error)                     { return 0, nil }
func (s *ragFakeStore) PredicateStats(context.Context) (map[string]int, error) { return nil, nil }
func (s *ragFakeStore) AllTriples(context.Context) ([]Triple, error)           { return nil, nil }
func (s *ragFakeStore) ClearAll(context.Context) error                         { return nil }
func (s *ragFakeStore) Close() error                                           { return nil }

func (s *ragFakeStore) Traverse(_ context.Context, start string, maxDepth int, predicates []string) ([]Triple, error) {
	s.calls = append(s.calls, ragTraverseCall{
		start:      start,
		maxDepth:   maxDepth,
		predicates: append([]string(nil), predicates...),
	})
	if err := s.errByStart[start]; err != nil {
		return nil, err
	}
	return s.triplesByStart[start], nil
}

func TestGraphRAGServiceRetrieveExpandsUniqueGraphNodes(t *testing.T) {
	t.Parallel()

	retriever := &ragFakeRetriever{
		results: []ContentResult{
			{Collection: "learning", SourceID: "one", Content: "first", Score: 0.9},
			{Collection: "session", SourceID: "two", Content: "second", Score: 0.8},
		},
	}
	store := &ragFakeStore{
		triplesByStart: map[string][]Triple{
			"learning:one": {
				{Subject: "learning:one", Predicate: RelatedTo, Object: "issue:123", ObjectType: "Issue"},
				{Subject: "learning:one", Predicate: SimilarTo, Object: "session:two", ObjectType: "Session"},
			},
			"session:two": {
				{Subject: "fix:abc", Predicate: ResolvedBy, Object: "session:two", SubjectType: "Fix"},
				{Subject: "session:two", Predicate: CausedBy, Object: "issue:123", ObjectType: "Issue"},
			},
		},
		errByStart: map[string]error{},
	}
	opts := ContentRetrieveOptions{Collections: []string{"learning"}, Limit: 2, SessionKey: "session-key"}
	service := NewGraphRAGService(retriever, store, 0, 2, zap.NewNop().Sugar())

	got, err := service.Retrieve(context.Background(), "query text", opts)
	require.NoError(t, err)

	assert.Equal(t, "query text", retriever.query)
	assert.Equal(t, opts, retriever.opts)
	assert.Equal(t, retriever.results, got.ContentResults)
	require.Len(t, got.GraphResults, 2)
	assert.Equal(t, GraphNode{
		ID:        "issue:123",
		NodeType:  "Issue",
		Predicate: RelatedTo,
		FromNode:  "learning:one",
		Depth:     1,
	}, got.GraphResults[0])
	assert.Equal(t, GraphNode{
		ID:        "fix:abc",
		NodeType:  "Fix",
		Predicate: ResolvedBy,
		FromNode:  "session:two",
		Depth:     1,
	}, got.GraphResults[1])
	require.Len(t, store.calls, 2)
	assert.Equal(t, 2, store.calls[0].maxDepth)
	assert.Equal(t, []string{RelatedTo, ResolvedBy, CausedBy, SimilarTo}, store.calls[0].predicates)
}

func TestGraphRAGServiceRetrieveHandlesRetrieverAndTraverseErrors(t *testing.T) {
	t.Parallel()

	t.Run("retriever error returns empty result without graph traversal", func(t *testing.T) {
		retriever := &ragFakeRetriever{err: errors.New("fts unavailable")}
		store := &ragFakeStore{triplesByStart: map[string][]Triple{}, errByStart: map[string]error{}}
		service := NewGraphRAGService(retriever, store, 1, 1, zap.NewNop().Sugar())

		got, err := service.Retrieve(context.Background(), "query", ContentRetrieveOptions{})
		require.NoError(t, err)
		assert.Empty(t, got.ContentResults)
		assert.Empty(t, got.GraphResults)
		assert.Empty(t, store.calls)
	})

	t.Run("traverse error skips failed seed and keeps later expansions", func(t *testing.T) {
		retriever := &ragFakeRetriever{results: []ContentResult{
			{Collection: "learning", SourceID: "bad", Content: "bad"},
			{Collection: "learning", SourceID: "good", Content: "good"},
		}}
		store := &ragFakeStore{
			triplesByStart: map[string][]Triple{
				"learning:good": {
					{Subject: "learning:good", Predicate: RelatedTo, Object: "note:ok"},
				},
			},
			errByStart: map[string]error{"learning:bad": errors.New("graph offline")},
		}
		service := NewGraphRAGService(retriever, store, 1, 10, zap.NewNop().Sugar())

		got, err := service.Retrieve(context.Background(), "query", ContentRetrieveOptions{})
		require.NoError(t, err)
		require.Len(t, got.GraphResults, 1)
		assert.Equal(t, "note:ok", got.GraphResults[0].ID)
		require.Len(t, store.calls, 2)
		assert.Equal(t, "learning:bad", store.calls[0].start)
		assert.Equal(t, "learning:good", store.calls[1].start)
	})
}

func TestGraphRAGServiceAssembleSectionFormatsContentAndGraphResults(t *testing.T) {
	t.Parallel()

	service := NewGraphRAGService(nil, nil, 0, 0, zap.NewNop().Sugar())

	assert.Empty(t, service.AssembleSection(nil))

	section := service.AssembleSection(&GraphRAGResult{
		ContentResults: []ContentResult{
			{Collection: "learning", SourceID: "one", Content: "useful context"},
			{Collection: "learning", SourceID: "empty"},
		},
		GraphResults: []GraphNode{
			{ID: "issue:123", NodeType: "Issue", Predicate: RelatedTo, FromNode: "learning:one"},
			{ID: "note:456", Predicate: SimilarTo, FromNode: "learning:one"},
		},
	})

	assert.Contains(t, section, "## Retrieved Context")
	assert.Contains(t, section, "### [learning] one")
	assert.Contains(t, section, "useful context")
	assert.NotContains(t, section, "### [learning] empty")
	assert.Contains(t, section, "## Graph-Expanded Context")
	assert.Contains(t, section, "- **Issue:issue:123** (via related_to from learning:one)")
	assert.Contains(t, section, "- **note:456** (via similar_to from learning:one)")
}
