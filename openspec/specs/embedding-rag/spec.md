# Embedding & RAG Specification

## Purpose

Capability spec for embedding-rag. See requirements below for scope and behavior contracts.

## Requirements

### REQ-EMB-001: Embedding Provider Interface
The system SHALL provide an `EmbeddingProvider` interface supporting batch text-to-vector conversion with provider ID, embed, and dimensions methods. Each provider SHALL pass the configured `dimensions` value to its underlying API call so that returned vectors match the configured dimension.

#### Scenario: OpenAI provider returns configured dimensions
- **WHEN** OpenAI credentials and model are configured and Embed is called with texts
- **THEN** float32 vectors of the configured dimensions SHALL be returned

#### Scenario: Google provider returns configured dimensions
- **WHEN** Google credentials and model are configured and Embed is called
- **THEN** vectors matching the configured dimensions SHALL be returned

#### Scenario: Local provider returns local model vectors
- **WHEN** a local Ollama-compatible endpoint is configured and Embed is called
- **THEN** vectors from the local model SHALL be returned

#### Scenario: Unknown provider type returns an error
- **WHEN** an embedding provider is created from an unknown provider type
- **THEN** provider creation SHALL return an error

#### Scenario: Google provider forwards output dimensionality
- **WHEN** `GoogleProvider` is configured with dimensions `N` and Embed is called
- **THEN** the EmbedContent API SHALL include `OutputDimensionality=N`

#### Scenario: OpenAI provider forwards dimensions
- **WHEN** `OpenAIProvider` is configured with dimensions `N` and Embed is called
- **THEN** the embedding request SHALL include `Dimensions=N`

#### Scenario: Local provider forwards dimensions
- **WHEN** `LocalProvider` is configured with dimensions `N` and Embed is called
- **THEN** the embedding request SHALL include `Dimensions=N`

#### Scenario: Returned vectors match SQLite schema dimension
- **WHEN** any embedding provider returns vectors
- **THEN** the vector dimension SHALL match the SQLite vec table `float[N]` schema

### REQ-EMB-002: Vector Store
The system SHALL provide a `VectorStore` interface supporting Upsert, Search (by collection + cosine similarity), and Delete operations.

#### Scenario: Upsert stores a new vector record
- **WHEN** Upsert is called with a VectorRecord containing ID, collection, embedding, and metadata
- **THEN** the record SHALL be stored and retrievable

#### Scenario: Upsert replaces an existing record
- **WHEN** Upsert is called again with an existing record ID
- **THEN** the previous record SHALL be replaced

#### Scenario: Search returns ordered nearest results
- **WHEN** Search is called with a query vector against stored vectors
- **THEN** results SHALL be returned sorted by ascending distance

#### Scenario: Delete removes stored records
- **WHEN** Delete is called with stored record IDs
- **THEN** those records SHALL be removed

### REQ-EMB-003: sqlite-vec Integration
The system SHALL use sqlite-vec virtual tables within the shared SQLite database for vector storage. Dimensions are determined at init time from the provider.

#### Scenario: sqlite-vec tables use provider-derived dimensions
- **WHEN** the vector store initializes against SQLite
- **THEN** sqlite-vec tables SHALL be created using dimensions derived from the configured provider

### REQ-EMB-004: Async Embedding Buffer
The system SHALL process embedding requests asynchronously via a background goroutine with batched provider calls and graceful shutdown.

#### Scenario: Enqueue processes in the background
- **WHEN** the embedding buffer is started and Enqueue is called
- **THEN** the request SHALL be processed in the background

#### Scenario: Batch threshold sends grouped provider call
- **WHEN** multiple requests are enqueued and the batch timeout or size threshold is reached
- **THEN** they SHALL be sent to the provider in one batch

#### Scenario: Stop drains remaining requests
- **WHEN** the embedding buffer is stopped
- **THEN** remaining queued items SHALL be drained before exit

### REQ-EMB-005: Store Integration
Knowledge, Memory (Observation/Reflection), and Learning stores SHALL emit embed callbacks on save operations. Callbacks are optional; nil means no embedding (backward compatible).

#### Scenario: Save operations emit embed callbacks when configured
- **WHEN** knowledge, memory, or learning stores save content and an embed callback is configured
- **THEN** the corresponding embed callback SHALL be invoked
- **AND** nil callbacks SHALL continue to behave as a no-op

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

#### Scenario: RAG section is injected into the prompt
- **WHEN** RAG is enabled and relevant semantic results are available
- **THEN** the ContextAwareModelAdapter SHALL inject a "Semantic Context (RAG)" section before the LLM call

### REQ-EMB-008: Configuration
Embedding settings SHALL be configurable via the `embedding` section in config, including provider, model, dimensions, local endpoint, and RAG options.

#### Scenario: Embedding config exposes provider and RAG settings
- **WHEN** embedding configuration is loaded from config
- **THEN** provider selection, model, dimensions, local endpoint, and RAG options SHALL be available under the `embedding` section

### REQ-EMB-009: Doctor Check
The doctor command SHALL include an Embedding/RAG check that validates provider configuration and API key availability.

#### Scenario: Doctor validates embedding configuration
- **WHEN** `lango doctor` evaluates embedding and RAG status
- **THEN** it SHALL validate provider configuration and API key availability for the embedding subsystem

### REQ-EMB-010: ProviderID-based Embedding Provider Resolution
The `EmbeddingConfig` SHALL support a `ProviderID` field that references a key in the `Config.Providers` map. When `ProviderID` is set, the embedding backend type and API key SHALL be resolved from the referenced provider's `Type` and `APIKey` fields using the `ProviderTypeToEmbeddingType` mapping.

#### Scenario: Gemini provider ID resolves to google backend
- **WHEN** `embedding.providerID` is `"gemini-1"` and `providers["gemini-1"]` has type `"gemini"` with a valid API key
- **THEN** the embedding backend type SHALL resolve to `"google"`
- **AND** the API key SHALL resolve from that provider entry

#### Scenario: OpenAI provider ID resolves to openai backend
- **WHEN** `embedding.providerID` is `"my-openai"` and `providers["my-openai"]` has type `"openai"` with a valid API key
- **THEN** the embedding backend type SHALL resolve to `"openai"`
- **AND** the API key SHALL resolve from that provider entry

#### Scenario: Ollama provider ID resolves to local backend
- **WHEN** `embedding.providerID` is `"my-ollama"` and `providers["my-ollama"]` has type `"ollama"`
- **THEN** the embedding backend type SHALL resolve to `"local"`
- **AND** no API key SHALL be required

#### Scenario: Unsupported provider type resolves empty backend
- **WHEN** `embedding.providerID` references a provider with type `"anthropic"`
- **THEN** the resolver SHALL return an empty backend type and empty API key

#### Scenario: Missing provider ID resolves empty backend
- **WHEN** `embedding.providerID` is set to an ID that does not exist in the providers map
- **THEN** the resolver SHALL return an empty backend type and empty API key

### REQ-EMB-011: Embedding Provider Resolution
The system SHALL resolve the embedding backend via two paths:
1. `ProviderID` — looks up the provider in the providers map and resolves backend type and API key.
2. `Provider = "local"` — uses local (Ollama) embeddings with no API key.

If neither `ProviderID` nor `Provider = "local"` is set, the embedding system SHALL be disabled.

#### Scenario: ProviderID path resolves provider credentials
- **WHEN** `embedding.providerID` is set to a valid key in the providers map
- **THEN** the backend type and API key SHALL resolve from that provider entry

#### Scenario: Local provider path resolves without API key
- **WHEN** `embedding.provider` is set to `"local"`
- **THEN** the backend type SHALL be `"local"`
- **AND** no API key SHALL be required

#### Scenario: Missing provider settings disable embedding
- **WHEN** both `embedding.providerID` and `embedding.provider` are empty
- **THEN** the embedding system SHALL be disabled

### REQ-EMB-012: MaxDistance Filtering
The system SHALL support a MaxDistance configuration (default 0.0 = disabled). When enabled, vector search results with distance exceeding MaxDistance SHALL be excluded from RAG context.

#### Scenario: Results above max distance are excluded
- **WHEN** MaxDistance is set to `0.5` and a search result has distance `0.7`
- **THEN** that result SHALL be excluded from the returned results

#### Scenario: Zero max distance disables filtering
- **WHEN** MaxDistance is `0.0`
- **THEN** all results SHALL be returned regardless of distance

### REQ-EMB-013: Session-Scoped Metadata Filtering
The system SHALL support filtering vector search results by metadata key-value pairs, enabling session-scoped retrieval.

#### Scenario: Session key filters results by metadata
- **WHEN** a RAG query includes a session key
- **THEN** results SHALL be filtered to include only entries matching that session metadata

#### Scenario: Missing session key leaves results unfiltered
- **WHEN** a RAG query has no session key
- **THEN** results SHALL be returned without metadata filtering

### REQ-EMB-014: VectorStore Search Options
The VectorStore.Search method SHALL accept an optional `*SearchOptions` parameter for metadata filtering. Nil means no filtering.

#### Scenario: Nil search options preserve current behavior
- **WHEN** Search is called with nil `SearchOptions`
- **THEN** behavior SHALL remain identical to the current implementation

#### Scenario: Metadata filter applies post-filtering
- **WHEN** Search is called with `MetadataFilter` containing key-value pairs
- **THEN** results SHALL be post-filtered to match all specified metadata pairs
- **AND** the search SHALL over-fetch by 3x to compensate
