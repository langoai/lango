package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/types"
)

// predicateValidatable is an optional capability for graph stores that support
// runtime predicate validation. Used to inject ontology validators without
// coupling wiring code to concrete store types.
type predicateValidatable interface {
	SetPredicateValidator(graph.PredicateValidatorFunc)
}

// graphComponents holds optional graph store components.
type graphComponents struct {
	store      graph.Store
	buffer     *graph.GraphBuffer
	ragService *graph.GraphRAGService
}

// initGraphStore creates the graph store if enabled.
func initGraphStore(cfg *config.Config) (*graphComponents, *types.FeatureStatus) {
	const featureName = "Graph Store"

	if !cfg.Graph.Enabled {
		logger().Info("graph store disabled")
		return nil, &types.FeatureStatus{Name: featureName, Enabled: false, Healthy: true}
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
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     fmt.Sprintf("bolt store init failed: %v", err),
			Suggestion: "check graph.databasePath permissions and disk space",
		}
	}

	buffer := graph.NewGraphBuffer(store, logger())

	logger().Infow("graph store initialized", "backend", "bolt", "path", dbPath)
	return &graphComponents{
		store:  store,
		buffer: buffer,
	}, &types.FeatureStatus{Name: featureName, Enabled: true, Healthy: true}
}

// wireGraphCallbacks subscribes to content.saved and triples.extracted events to feed the graph buffer.
// It also creates the Entity Extractor pipeline and Memory GraphHooks.
func wireGraphCallbacks(gc *graphComponents, kc *knowledgeComponents, mc *memoryComponents, sv *supervisor.Supervisor, cfg *config.Config, bus *eventbus.Bus, ontologyValidator graph.PredicateValidatorFunc) {
	if gc == nil || gc.buffer == nil {
		return
	}

	// Inject predicate validator if the store implementation supports it.
	if ontologyValidator != nil {
		if pv, ok := gc.store.(predicateValidatable); ok {
			pv.SetPredicateValidator(ontologyValidator)
			logger().Info("ontology predicate validator injected into graph store")
		}
	}

	// Create Entity Extractor for async triple extraction from content.
	var extractor *graph.Extractor
	if sv != nil {
		provider := cfg.Agent.Provider
		mdl := cfg.Agent.Model
		proxy := supervisor.NewProviderProxy(sv, provider, mdl)
		generator := &providerTextGenerator{proxy: proxy}
		var opts []graph.ExtractorOption
		if ontologyValidator != nil {
			opts = append(opts, graph.WithPredicateValidator(ontologyValidator))
		}
		extractor = graph.NewExtractor(generator, logger(), opts...)
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
						Subject:     evt.Collection + ":" + evt.ID,
						SubjectType: evt.Collection,
						Predicate:   graph.Contains,
						Object:      "collection:" + evt.Collection,
						Metadata:    evt.Metadata,
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
					Subject:     t.Subject,
					Predicate:   t.Predicate,
					Object:      t.Object,
					SubjectType: t.SubjectType,
					ObjectType:  t.ObjectType,
					Metadata:    t.Metadata,
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

// initGraphRAG creates the Graph RAG service when both graph store and
// knowledge search are available.
func initGraphRAG(cfg *config.Config, gc *graphComponents, kc *knowledgeComponents) {
	if gc == nil || kc == nil || kc.store == nil {
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

	adapter := &knowledgeContentRetriever{store: kc.store}

	gc.ragService = graph.NewGraphRAGService(adapter, gc.store, maxDepth, maxExpand, logger())
	logger().Info("graph RAG hybrid retrieval initialized")
}

type contentSearchSource interface {
	SearchKnowledgeScored(ctx context.Context, query string, category string, limit int) ([]knowledge.ScoredKnowledgeEntry, error)
}

type knowledgeContentRetriever struct {
	store contentSearchSource
}

func (r *knowledgeContentRetriever) Retrieve(ctx context.Context, query string, opts graph.ContentRetrieveOptions) ([]graph.ContentResult, error) {
	if r.store == nil || query == "" {
		return nil, nil
	}
	if len(opts.Collections) > 0 {
		allowed := false
		for _, collection := range opts.Collections {
			if collection == "knowledge" {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, nil
		}
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	results, err := r.store.SearchKnowledgeScored(ctx, query, "", limit)
	if err != nil {
		return nil, err
	}

	graphResults := make([]graph.ContentResult, 0, len(results))
	for _, item := range results {
		graphResults = append(graphResults, graph.ContentResult{
			Collection: "knowledge",
			SourceID:   item.Entry.Key,
			Content:    item.Entry.Content,
			Score:      float32(item.Score),
		})
	}
	return graphResults, nil
}
