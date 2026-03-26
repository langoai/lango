package app

import (
	"database/sql"
	"fmt"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/types"
)

// embeddingComponents holds optional embedding/RAG components.
type embeddingComponents struct {
	buffer     *embedding.EmbeddingBuffer
	ragService *embedding.RAGService
}

// initEmbedding creates the embedding pipeline and RAG service if configured.
func initEmbedding(cfg *config.Config, rawDB *sql.DB, kc *knowledgeComponents, mc *memoryComponents, bus *eventbus.Bus) (*embeddingComponents, *types.FeatureStatus) {
	const featureName = "Embedding & RAG"
	emb := cfg.Embedding

	if emb.Provider == "" {
		logger().Info("embedding system disabled (no provider configured)")
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: true,
			Reason:     "no provider configured",
			Suggestion: "set embedding.provider or add an OpenAI/Gemini provider",
		}
	}

	backendType, apiKey := cfg.ResolveEmbeddingProvider()
	if backendType == "" {
		logger().Warnw("embedding provider type could not be resolved",
			"provider", emb.Provider)
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     fmt.Sprintf("provider %q could not be resolved", emb.Provider),
			Suggestion: "check providers map for a valid embedding-capable provider",
		}
	}

	providerCfg := embedding.ProviderConfig{
		Provider:   backendType,
		Model:      emb.Model,
		Dimensions: emb.Dimensions,
		APIKey:     apiKey,
		BaseURL:    emb.Local.BaseURL,
	}

	registry, err := embedding.NewRegistry(providerCfg, nil, logger())
	if err != nil {
		logger().Warnw("embedding provider init failed, skipping", "error", err)
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     fmt.Sprintf("provider init failed: %v", err),
			Suggestion: "check API key and provider configuration",
		}
	}

	provider := registry.Provider()
	dimensions := provider.Dimensions()

	// Create vector store using the shared database.
	if rawDB == nil {
		logger().Warn("embedding requires raw DB handle, skipping")
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     "raw DB handle not available",
			Suggestion: "ensure database is initialized before embedding",
		}
	}
	vecStore, err := embedding.NewSQLiteVecStore(rawDB, dimensions)
	if err != nil {
		logger().Warnw("sqlite-vec store init failed, skipping", "error", err)
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     fmt.Sprintf("sqlite-vec init failed: %v", err),
			Suggestion: "check sqlite-vec extension availability",
		}
	}

	embLogger := logger()

	// Create buffer.
	buffer := embedding.NewEmbeddingBuffer(provider, vecStore, embLogger)

	// Create resolver and RAG service.
	var ks *knowledge.Store
	var ms *memory.Store
	if kc != nil {
		ks = kc.store
	}
	if mc != nil {
		ms = mc.store
	}
	resolver := embedding.NewStoreResolver(ks, ms)
	ragService := embedding.NewRAGService(provider, vecStore, resolver, embLogger)

	// Subscribe to content.saved events to trigger async embedding.
	if bus != nil {
		eventbus.SubscribeTyped(bus, func(evt eventbus.ContentSavedEvent) {
			buffer.Enqueue(embedding.EmbedRequest{
				ID:         evt.ID,
				Collection: evt.Collection,
				Content:    evt.Content,
				Metadata:   evt.Metadata,
			})
		})
	}

	logger().Infow("embedding system initialized",
		"provider", emb.Provider,
		"backendType", backendType,
		"dimensions", dimensions,
		"ragEnabled", emb.RAG.Enabled,
	)

	return &embeddingComponents{
		buffer:     buffer,
		ragService: ragService,
	}, &types.FeatureStatus{Name: featureName, Enabled: true, Healthy: true}
}
