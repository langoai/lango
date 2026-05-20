package embedding

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider struct {
	mu        sync.Mutex
	embedding []float32
	err       error
	calls     int
	texts     [][]string
}

func (p *ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider) ID() string {
	return "ragServiceRetrieveMergesSortsLimitsAndResolvesContent6"
}

func (p *ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider) Dimensions() int {
	return len(p.embedding)
}

func (p *ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider) Embed(_ context.Context, texts []string) ([][]float32, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.calls++
	p.texts = append(p.texts, append([]string(nil), texts...))
	if p.err != nil {
		return nil, p.err
	}
	if p.embedding == nil {
		return nil, nil
	}
	return [][]float32{append([]float32(nil), p.embedding...)}, nil
}

func (p *ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

type ragServiceRetrieveMergesSortsLimitsAndResolvesContentSearchCall struct {
	collection string
	query      []float32
	limit      int
	opts       *SearchOptions
}

type ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore struct {
	mu      sync.Mutex
	hits    map[string][]SearchResult
	errs    map[string]error
	calls   []ragServiceRetrieveMergesSortsLimitsAndResolvesContentSearchCall
	deletes []string
}

func (s *ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore) Upsert(_ context.Context, _ []VectorRecord) error {
	return nil
}

func (s *ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore) Search(
	_ context.Context,
	collection string,
	query []float32,
	limit int,
	opts *SearchOptions,
) ([]SearchResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, ragServiceRetrieveMergesSortsLimitsAndResolvesContentSearchCall{
		collection: collection,
		query:      append([]float32(nil), query...),
		limit:      limit,
		opts:       cloneRagServiceRetrieveMergesSortsLimitsAndResolvesContentSearchOptions(opts),
	})
	if err := s.errs[collection]; err != nil {
		return nil, err
	}
	hits := append([]SearchResult(nil), s.hits[collection]...)
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}

func (s *ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore) Delete(_ context.Context, collection string, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		s.deletes = append(s.deletes, collection+"/"+id)
	}
	return nil
}

func (s *ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore) Close() error { return nil }

func (s *ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func (s *ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore) searchCallsByCollection() map[string]ragServiceRetrieveMergesSortsLimitsAndResolvesContentSearchCall {
	s.mu.Lock()
	defer s.mu.Unlock()

	calls := make(map[string]ragServiceRetrieveMergesSortsLimitsAndResolvesContentSearchCall, len(s.calls))
	for _, call := range s.calls {
		calls[call.collection] = ragServiceRetrieveMergesSortsLimitsAndResolvesContentSearchCall{
			collection: call.collection,
			query:      append([]float32(nil), call.query...),
			limit:      call.limit,
			opts:       cloneRagServiceRetrieveMergesSortsLimitsAndResolvesContentSearchOptions(call.opts),
		}
	}
	return calls
}

type ragServiceRetrieveMergesSortsLimitsAndResolvesContentResolver struct {
	mu      sync.Mutex
	content map[string]string
	errs    map[string]error
	calls   []string
}

func (r *ragServiceRetrieveMergesSortsLimitsAndResolvesContentResolver) ResolveContent(
	_ context.Context,
	collection string,
	id string,
) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := collection + "/" + id
	r.calls = append(r.calls, key)
	if err := r.errs[key]; err != nil {
		return "", err
	}
	return r.content[key], nil
}

func newRagServiceRetrieveMergesSortsLimitsAndResolvesContentRAGService(
	provider *ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider,
	store *ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore,
	resolver ContentResolver,
) *RAGService {
	return NewRAGService(provider, store, resolver, zap.NewNop().Sugar())
}

func cloneRagServiceRetrieveMergesSortsLimitsAndResolvesContentSearchOptions(opts *SearchOptions) *SearchOptions {
	if opts == nil {
		return nil
	}
	cloned := &SearchOptions{}
	if opts.MetadataFilter != nil {
		cloned.MetadataFilter = make(map[string]string, len(opts.MetadataFilter))
		for key, value := range opts.MetadataFilter {
			cloned.MetadataFilter[key] = value
		}
	}
	return cloned
}

func TestRAGService_RetrieveMergesSortsLimitsAndResolvesContent(t *testing.T) {
	provider := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider{embedding: []float32{0.25, 0.75}}
	store := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore{hits: map[string][]SearchResult{
		"knowledge": {
			{ID: "k-slower", Collection: "knowledge", Distance: 0.40},
			{ID: "k-best", Collection: "knowledge", Distance: 0.10},
		},
		"observation": {
			{ID: "o-mid", Collection: "observation", Distance: 0.20},
		},
		"reflection": {
			{ID: "r-late", Collection: "reflection", Distance: 0.30},
		},
	}}
	resolver := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentResolver{content: map[string]string{
		"knowledge/k-best":   "best knowledge",
		"knowledge/k-slower": "slower knowledge",
		"observation/o-mid":  "middle observation",
		"reflection/r-late":  "late reflection",
	}}
	svc := newRagServiceRetrieveMergesSortsLimitsAndResolvesContentRAGService(provider, store, resolver)

	results, err := svc.Retrieve(context.Background(), "ranked query", RetrieveOptions{Limit: 3})

	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, []RAGResult{
		{
			Collection: "knowledge",
			SourceID:   "k-best",
			Content:    "best knowledge",
			Distance:   0.10,
		},
		{
			Collection: "observation",
			SourceID:   "o-mid",
			Content:    "middle observation",
			Distance:   0.20,
		},
		{
			Collection: "reflection",
			SourceID:   "r-late",
			Content:    "late reflection",
			Distance:   0.30,
		},
	}, results)
	assert.Equal(t, 1, provider.callCount())

	calls := store.searchCallsByCollection()
	require.Len(t, calls, 4)
	for _, collection := range allCollections {
		call, ok := calls[collection]
		require.True(t, ok, "missing search for %s", collection)
		assert.Equal(t, []float32{0.25, 0.75}, call.query)
		assert.Equal(t, 6, call.limit)
		assert.Nil(t, call.opts)
	}
}

func TestRAGService_RetrieveAppliesSessionFilterAndMaxDistance(t *testing.T) {
	provider := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider{embedding: []float32{1, 0}}
	store := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore{hits: map[string][]SearchResult{
		"knowledge": {
			{ID: "near", Collection: "knowledge", Distance: 0.10},
			{ID: "edge", Collection: "knowledge", Distance: 0.25},
			{ID: "far", Collection: "knowledge", Distance: 0.26},
		},
	}}
	svc := newRagServiceRetrieveMergesSortsLimitsAndResolvesContentRAGService(provider, store, nil)

	results, err := svc.Retrieve(context.Background(), "filtered query", RetrieveOptions{
		Collections: []string{"knowledge"},
		Limit:       10,
		SessionKey:  "session-1",
		MaxDistance: 0.25,
	})

	require.NoError(t, err)
	assert.Equal(t, []RAGResult{
		{Collection: "knowledge", SourceID: "near", Distance: 0.10},
		{Collection: "knowledge", SourceID: "edge", Distance: 0.25},
	}, results)

	calls := store.searchCallsByCollection()
	require.Len(t, calls, 1)
	require.NotNil(t, calls["knowledge"].opts)
	assert.Equal(t, map[string]string{"session_key": "session-1"}, calls["knowledge"].opts.MetadataFilter)
	assert.Equal(t, 10, calls["knowledge"].limit)
}

func TestRAGService_RetrieveSkipsResultsWhenResolverFails(t *testing.T) {
	provider := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider{embedding: []float32{1, 0}}
	store := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore{hits: map[string][]SearchResult{
		"knowledge": {
			{ID: "good", Collection: "knowledge", Distance: 0.10},
			{ID: "bad", Collection: "knowledge", Distance: 0.20},
			{ID: "next", Collection: "knowledge", Distance: 0.30},
		},
	}}
	resolver := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentResolver{
		content: map[string]string{
			"knowledge/good": "resolved good",
			"knowledge/next": "resolved next",
		},
		errs: map[string]error{
			"knowledge/bad": errors.New("missing content"),
		},
	}
	svc := newRagServiceRetrieveMergesSortsLimitsAndResolvesContentRAGService(provider, store, resolver)

	results, err := svc.Retrieve(context.Background(), "resolver query", RetrieveOptions{
		Collections: []string{"knowledge"},
		Limit:       5,
	})

	require.NoError(t, err)
	assert.Equal(t, []RAGResult{
		{
			Collection: "knowledge",
			SourceID:   "good",
			Content:    "resolved good",
			Distance:   0.10,
		},
		{
			Collection: "knowledge",
			SourceID:   "next",
			Content:    "resolved next",
			Distance:   0.30,
		},
	}, results)
}

func TestRAGService_RetrieveEmptyQueryAvoidsProviderAndStore(t *testing.T) {
	provider := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider{embedding: []float32{1, 0}}
	store := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore{hits: map[string][]SearchResult{
		"knowledge": {{ID: "unused", Collection: "knowledge", Distance: 0.10}},
	}}
	svc := newRagServiceRetrieveMergesSortsLimitsAndResolvesContentRAGService(provider, store, nil)

	results, err := svc.Retrieve(context.Background(), "", RetrieveOptions{Limit: 5})

	require.NoError(t, err)
	assert.Nil(t, results)
	assert.Equal(t, 0, provider.callCount())
	assert.Equal(t, 0, store.callCount())
}

func TestRAGService_RetrieveEmptyEmbeddingAvoidsStoreSearch(t *testing.T) {
	provider := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider{}
	store := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore{hits: map[string][]SearchResult{
		"knowledge": {{ID: "unused", Collection: "knowledge", Distance: 0.10}},
	}}
	svc := newRagServiceRetrieveMergesSortsLimitsAndResolvesContentRAGService(provider, store, nil)

	results, err := svc.Retrieve(context.Background(), "empty embedding query", RetrieveOptions{Limit: 5})

	require.NoError(t, err)
	assert.Nil(t, results)
	assert.Equal(t, 1, provider.callCount())
	assert.Equal(t, 0, store.callCount())
}

func TestRAGService_RetrieveReturnsProviderError(t *testing.T) {
	providerErr := errors.New("provider offline")
	provider := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider{err: providerErr}
	store := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore{}
	svc := newRagServiceRetrieveMergesSortsLimitsAndResolvesContentRAGService(provider, store, nil)

	results, err := svc.Retrieve(context.Background(), "provider error query", RetrieveOptions{Limit: 5})

	require.Error(t, err)
	assert.ErrorIs(t, err, providerErr)
	assert.ErrorContains(t, err, "embed query")
	assert.Nil(t, results)
	assert.Equal(t, 0, store.callCount())
}

func TestRAGService_RetrieveTreatsStoreErrorsAsNonFatal(t *testing.T) {
	provider := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentEmbeddingProvider{embedding: []float32{1, 0}}
	store := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentVectorStore{
		hits: map[string][]SearchResult{
			"observation": {
				{ID: "survivor", Collection: "observation", Distance: 0.15},
			},
		},
		errs: map[string]error{
			"knowledge": errors.New("temporary search failure"),
		},
	}
	resolver := &ragServiceRetrieveMergesSortsLimitsAndResolvesContentResolver{content: map[string]string{
		"observation/survivor": "surviving content",
	}}
	svc := newRagServiceRetrieveMergesSortsLimitsAndResolvesContentRAGService(provider, store, resolver)

	results, err := svc.Retrieve(context.Background(), "partial failure query", RetrieveOptions{
		Collections: []string{"knowledge", "observation"},
		Limit:       5,
	})

	require.NoError(t, err)
	assert.Equal(t, []RAGResult{
		{
			Collection: "observation",
			SourceID:   "survivor",
			Content:    "surviving content",
			Distance:   0.15,
		},
	}, results)
	calls := store.searchCallsByCollection()
	require.Len(t, calls, 2)
	assert.Contains(t, calls, "knowledge")
	assert.Contains(t, calls, "observation")
}
