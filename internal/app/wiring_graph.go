package app

import (
	"context"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/supervisor"
)

// graphComponents holds optional graph store components.
type graphComponents struct {
	store      graph.Store
	buffer     *graph.GraphBuffer
	ragService *graph.GraphRAGService
}

// initGraphStore creates the graph store if enabled.
func initGraphStore(cfg *config.Config) *graphComponents {
	if !cfg.Graph.Enabled {
		logger().Info("graph store disabled")
		return nil
	}

	dbPath := cfg.Graph.DatabasePath
	if dbPath == "" {
		// Default: graph.db next to session database.
		if cfg.Session.DatabasePath != "" {
			dbPath = filepath.Join(filepath.Dir(cfg.Session.DatabasePath), "graph.db")
		} else {
			dbPath = "graph.db"
		}
	}

	store, err := graph.NewBoltStore(dbPath)
	if err != nil {
		logger().Warnw("graph store init error, skipping", "error", err)
		return nil
	}

	buffer := graph.NewGraphBuffer(store, logger())

	logger().Infow("graph store initialized", "backend", "bolt", "path", dbPath)
	return &graphComponents{
		store:  store,
		buffer: buffer,
	}
}

// wireGraphCallbacks connects graph store callbacks to knowledge and memory stores.
// It also creates the Entity Extractor pipeline and Memory GraphHooks.
func wireGraphCallbacks(gc *graphComponents, kc *knowledgeComponents, mc *memoryComponents, sv *supervisor.Supervisor, cfg *config.Config) {
	if gc == nil || gc.buffer == nil {
		return
	}

	// Create Entity Extractor for async triple extraction from content.
	var extractor *graph.Extractor
	if sv != nil {
		provider := cfg.Agent.Provider
		mdl := cfg.Agent.Model
		proxy := supervisor.NewProviderProxy(sv, provider, mdl)
		generator := &providerTextGenerator{proxy: proxy}
		extractor = graph.NewExtractor(generator, logger())
		logger().Info("graph entity extractor initialized")
	}

	graphCB := func(id, collection, content string, metadata map[string]string) {
		// Basic containment triple.
		gc.buffer.Enqueue(graph.GraphRequest{
			Triples: []graph.Triple{
				{
					Subject:   collection + ":" + id,
					Predicate: graph.Contains,
					Object:    "collection:" + collection,
					Metadata:  metadata,
				},
			},
		})

		// Async entity extraction via LLM.
		if extractor != nil && content != "" {
			go func() {
				ctx := context.Background()
				triples, err := extractor.Extract(ctx, content, id)
				if err != nil {
					logger().Debugw("entity extraction error", "id", id, "error", err)
					return
				}
				if len(triples) > 0 {
					gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
				}
			}()
		}
	}

	if kc != nil {
		kc.store.SetGraphCallback(graphCB)
	}
	if mc != nil {
		mc.store.SetGraphCallback(graphCB)

		// Wire Memory GraphHooks for temporal/session triples.
		tripleCallback := func(triples []graph.Triple) {
			gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
		}
		hooks := memory.NewGraphHooks(tripleCallback, logger())
		mc.store.SetGraphHooks(hooks)
		logger().Info("memory graph hooks wired")
	}
}

// initGraphRAG creates the Graph RAG service if both graph store and vector RAG are available.
func initGraphRAG(cfg *config.Config, gc *graphComponents, ec *embeddingComponents) {
	if gc == nil || ec == nil || ec.ragService == nil {
		return
	}

	maxDepth := cfg.Graph.MaxTraversalDepth
	if maxDepth <= 0 {
		maxDepth = 2
	}
	maxExpand := cfg.Graph.MaxExpansionResults
	if maxExpand <= 0 {
		maxExpand = 10
	}

	// Create a VectorRetriever adapter from embedding.RAGService.
	adapter := &ragServiceAdapter{inner: ec.ragService}

	gc.ragService = graph.NewGraphRAGService(adapter, gc.store, maxDepth, maxExpand, logger())
	logger().Info("graph RAG hybrid retrieval initialized")
}

// ragServiceAdapter adapts embedding.RAGService to graph.VectorRetriever interface.
type ragServiceAdapter struct {
	inner *embedding.RAGService
}

func (a *ragServiceAdapter) Retrieve(ctx context.Context, query string, opts graph.VectorRetrieveOptions) ([]graph.VectorResult, error) {
	embOpts := embedding.RetrieveOptions{
		Collections: opts.Collections,
		Limit:       opts.Limit,
		SessionKey:  opts.SessionKey,
		MaxDistance: opts.MaxDistance,
	}

	results, err := a.inner.Retrieve(ctx, query, embOpts)
	if err != nil {
		return nil, err
	}

	graphResults := make([]graph.VectorResult, len(results))
	for i, r := range results {
		graphResults[i] = graph.VectorResult{
			Collection: r.Collection,
			SourceID:   r.SourceID,
			Content:    r.Content,
			Distance:   r.Distance,
		}
	}
	return graphResults, nil
}
