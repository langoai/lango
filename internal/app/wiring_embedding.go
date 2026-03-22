package app

import (
	"database/sql"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
)

// embeddingComponents holds optional embedding/RAG components.
type embeddingComponents struct {
	buffer     *embedding.EmbeddingBuffer
	ragService *embedding.RAGService
}

// initEmbedding creates the embedding pipeline and RAG service if configured.
func initEmbedding(cfg *config.Config, rawDB *sql.DB, kc *knowledgeComponents, mc *memoryComponents, bus *eventbus.Bus) *embeddingComponents {
	emb := cfg.Embedding
	if emb.Provider == "" {
		logger().Info("embedding system disabled (no provider configured)")
		return nil
	}

	backendType, apiKey := cfg.ResolveEmbeddingProvider()
	if backendType == "" {
		logger().Warnw("embedding provider type could not be resolved",
			"provider", emb.Provider)
		return nil
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
		return nil
	}

	provider := registry.Provider()
	dimensions := provider.Dimensions()

	// Create vector store using the shared database.
	if rawDB == nil {
		logger().Warn("embedding requires raw DB handle, skipping")
		return nil
	}
	vecStore, err := embedding.NewSQLiteVecStore(rawDB, dimensions)
	if err != nil {
		logger().Warnw("sqlite-vec store init failed, skipping", "error", err)
		return nil
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
	}
}
