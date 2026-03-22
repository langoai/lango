package app

import (
	"context"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/eventbus"
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

// wireGraphCallbacks subscribes to content.saved and triples.extracted events to feed the graph buffer.
// It also creates the Entity Extractor pipeline and Memory GraphHooks.
func wireGraphCallbacks(gc *graphComponents, kc *knowledgeComponents, mc *memoryComponents, sv *supervisor.Supervisor, cfg *config.Config, bus *eventbus.Bus) {
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

	// Subscribe to content.saved events to create graph triples and extract entities.
	// Only events with NeedsGraph=true trigger graph operations, preserving the
	// original callback behavior: new knowledge creation and memory saves graph,
	// while knowledge updates and learning saves are embed-only.
	if bus != nil {
		eventbus.SubscribeTyped(bus, func(evt eventbus.ContentSavedEvent) {
			if !evt.NeedsGraph {
				return
			}
			// Basic containment triple.
			gc.buffer.Enqueue(graph.GraphRequest{
				Triples: []graph.Triple{
					{
						Subject:   evt.Collection + ":" + evt.ID,
						Predicate: graph.Contains,
						Object:    "collection:" + evt.Collection,
						Metadata:  evt.Metadata,
					},
				},
			})

			// Async entity extraction via LLM.
			if extractor != nil && evt.Content != "" {
				go func() {
					ctx := context.Background()
					triples, err := extractor.Extract(ctx, evt.Content, evt.ID)
					if err != nil {
						logger().Debugw("entity extraction error", "id", evt.ID, "error", err)
						return
					}
					if len(triples) > 0 {
						gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
					}
				}()
			}
		})

		// Subscribe to triples.extracted events to enqueue graph triples.
		eventbus.SubscribeTyped(bus, func(evt eventbus.TriplesExtractedEvent) {
			graphTriples := make([]graph.Triple, len(evt.Triples))
			for i, t := range evt.Triples {
				graphTriples[i] = graph.Triple{
					Subject:   t.Subject,
					Predicate: t.Predicate,
					Object:    t.Object,
					Metadata:  t.Metadata,
				}
			}
			gc.buffer.Enqueue(graph.GraphRequest{Triples: graphTriples})
		})
	}

	// Wire Memory GraphHooks for temporal/session triples.
	if mc != nil {
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
