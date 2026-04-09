# Embedding & RAG Specification

## Purpose

Capability spec for embedding-rag. See requirements below for scope and behavior contracts.

## Requirements

### REQ-EMB-001: Embedding Provider Interface
The system SHALL provide an `EmbeddingProvider` interface supporting batch text-to-vector conversion with provider ID, embed, and dimensions methods. Each provider SHALL pass the configured `dimensions` value to its underlying API call so that returned vectors match the configured dimension.

**Scenarios:**
- Given OpenAI API key and model, when Embed is called with texts, then float32 vectors of configured dimensions are returned
- Given Google API key and model, when Embed is called, then vectors matching configured dimensions are returned
- Given local Ollama endpoint, when Embed is called, then vectors from the local model are returned
- Given an unknown provider type, when creating via registry, then an error is returned
- Given GoogleProvider with dimensions=N, when Embed is called, then EmbedContent API includes OutputDimensionality=N
- Given OpenAIProvider with dimensions=N, when Embed is called, then EmbeddingRequest includes Dimensions=N
- Given LocalProvider with dimensions=N, when Embed is called, then EmbeddingRequest includes Dimensions=N
- Given any provider returns vectors, then vector dimension matches SQLite vec table float[N] schema

### REQ-EMB-002: Vector Store
The system SHALL provide a `VectorStore` interface supporting Upsert, Search (by collection + cosine similarity), and Delete operations.

**Scenarios:**
- Given a VectorRecord with ID, collection, embedding, and metadata, when Upsert is called, then the record is stored and retrievable
- Given an existing record ID, when Upsert is called again, then the previous record is replaced
- Given stored vectors, when Search is called with a query vector, then results are returned sorted by distance ascending
- Given stored records, when Delete is called with IDs, then those records are removed

### REQ-EMB-003: sqlite-vec Integration
The system SHALL use sqlite-vec virtual tables within the shared SQLite database for vector storage. Dimensions are determined at init time from the provider.

### REQ-EMB-004: Async Embedding Buffer
The system SHALL process embedding requests asynchronously via a background goroutine with batched provider calls and graceful shutdown.

**Scenarios:**
- Given embedding buffer is started, when Enqueue is called, then the request is processed in the background
- Given multiple enqueued requests, when batch timeout or size threshold is reached, then they are sent to the provider in one batch
- Given buffer is stopped, when Stop is called, then remaining queued items are drained before exit

### REQ-EMB-005: Store Integration
Knowledge, Memory (Observation/Reflection), and Learning stores SHALL emit embed callbacks on save operations. Callbacks are optional; nil means no embedding (backward compatible).

### REQ-EMB-006: RAG Service
The system SHALL provide a RAGService that:
1. Embeds a query string
2. Searches across configurable collections in parallel using errgroup
3. Resolves original content from source stores
4. Returns results merged, sorted by distance, and limited after all collections complete

Individual collection search errors SHALL be logged and treated as non-fatal.

#### Scenario: Parallel collection search
- **WHEN** a query is submitted against multiple collections
- **THEN** all collections SHALL be searched concurrently and results merged after all complete

#### Scenario: Single collection failure
- **WHEN** one collection search fails during parallel execution
- **THEN** the error SHALL be logged as a warning and results from other collections SHALL still be returned

#### Scenario: Results sorted and limited
- **WHEN** parallel searches complete
- **THEN** results SHALL be sorted by ascending distance and limited to the configured maximum

### REQ-EMB-007: Agent Context Injection
When RAG is enabled, the ContextAwareModelAdapter SHALL inject a "Semantic Context (RAG)" section into the system prompt before each LLM call.

### REQ-EMB-008: Configuration
Embedding settings SHALL be configurable via the `embedding` section in config, including provider, model, dimensions, local endpoint, and RAG options.

### REQ-EMB-009: Doctor Check
The doctor command SHALL include an Embedding/RAG check that validates provider configuration and API key availability.

### REQ-EMB-010: ProviderID-based Embedding Provider Resolution
The `EmbeddingConfig` SHALL support a `ProviderID` field that references a key in the `Config.Providers` map. When `ProviderID` is set, the embedding backend type and API key SHALL be resolved from the referenced provider's `Type` and `APIKey` fields using the `ProviderTypeToEmbeddingType` mapping.

**Scenarios:**
- Given `embedding.providerID` is `"gemini-1"` and `providers["gemini-1"]` has type `"gemini"` and a valid API key, then the embedding backend type is `"google"` and the API key is the provider's API key
- Given `embedding.providerID` is `"my-openai"` and `providers["my-openai"]` has type `"openai"` and a valid API key, then the embedding backend type is `"openai"` and the API key is the provider's API key
- Given `embedding.providerID` is `"my-ollama"` and `providers["my-ollama"]` has type `"ollama"`, then the embedding backend type is `"local"` and no API key is required
- Given `embedding.providerID` references a provider with type `"anthropic"` (no embedding support), then the resolver returns empty backend type and empty API key
- Given `embedding.providerID` is set to an ID that does not exist in the providers map, then the resolver returns empty backend type and empty API key

### REQ-EMB-011: Embedding Provider Resolution
The system SHALL resolve the embedding backend via two paths:
1. `ProviderID` — looks up the provider in the providers map and resolves backend type and API key.
2. `Provider = "local"` — uses local (Ollama) embeddings with no API key.

If neither `ProviderID` nor `Provider = "local"` is set, the embedding system SHALL be disabled.

**Scenarios:**
- Given `embedding.providerID` is set to a valid key in the providers map, then the backend type and API key are resolved from that provider entry
- Given `embedding.provider` is set to `"local"`, then the backend type is `"local"` with no API key
- Given both `embedding.providerID` and `embedding.provider` are empty, then the embedding system is disabled

### REQ-EMB-012: MaxDistance Filtering
The system SHALL support a MaxDistance configuration (default 0.0 = disabled). When enabled, vector search results with distance exceeding MaxDistance SHALL be excluded from RAG context.

**Scenarios:**
- Given MaxDistance is set to 0.5 and a search result has distance 0.7, then that result SHALL be excluded from the returned results
- Given MaxDistance is 0.0 (default), then all results SHALL be returned regardless of distance (backward compatible)

### REQ-EMB-013: Session-Scoped Metadata Filtering
The system SHALL support filtering vector search results by metadata key-value pairs, enabling session-scoped retrieval.

**Scenarios:**
- Given a RAG query includes a session key, then results SHALL be filtered to include only entries matching that session's metadata
- Given a RAG query has no session key, then all results SHALL be returned without metadata filtering (backward compatible)

### REQ-EMB-014: VectorStore Search Options
The VectorStore.Search method SHALL accept an optional `*SearchOptions` parameter for metadata filtering. Nil means no filtering.

**Scenarios:**
- Given Search is called with nil SearchOptions, then behavior SHALL be identical to the current implementation
- Given Search is called with MetadataFilter containing key-value pairs, then results SHALL be post-filtered to match all specified metadata pairs, with 3x over-fetch to compensate
