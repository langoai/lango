---
title: Embedding & RAG
---

# Embedding & RAG

Lango supports vector embeddings for semantic search and Retrieval-Augmented Generation (RAG). When enabled, relevant knowledge entries are retrieved via vector similarity and injected into the agent's context.

## Embedding Providers

| Provider | Config Type | Default Model | Default Dimensions | Notes |
|----------|------------|---------------|-------------------|-------|
| **OpenAI** | `openai` | `text-embedding-3-small` | 1536 | Also supports `text-embedding-3-large` |
| **Google** | `google` | `text-embedding-004` | 768 | Via Google Generative AI API |
| **Local** | `local` | `nomic-embed-text` | 768 | Ollama-compatible, no API key required |

## Setup

### Interactive Setup

The easiest way to configure embedding is through the onboarding wizard:

```bash
lango onboard
```

Select **Embedding & RAG** from the setup menu.

### Config File

#### Using a Cloud Provider

Reference an existing entry from the `providers` map via `providerID`:

```yaml
providers:
  my-openai:
    type: openai
    apiKey: ${OPENAI_API_KEY}

embedding:
  providerID: my-openai
  model: text-embedding-3-small
  dimensions: 1536
  rag:
    enabled: true
    maxResults: 5
```

#### Using Local (Ollama) Embeddings

For local embeddings, set `provider` to `local` instead of using `providerID`:

```yaml
embedding:
  provider: local
  model: nomic-embed-text
  dimensions: 768
  local:
    baseUrl: http://localhost:11434/v1
  rag:
    enabled: true
    maxResults: 5
```

!!! tip "Local Setup"

    Make sure Ollama is running and the embedding model is pulled:

    ```bash
    ollama serve
    ollama pull nomic-embed-text
    ```

## RAG (Retrieval-Augmented Generation)

When `embedding.rag.enabled` is `true`, Lango performs semantic retrieval on every agent turn:

1. The user's message is embedded into a vector
2. The vector is compared against stored embeddings across collections
3. The most similar entries are retrieved and injected into the agent's context

### Collections

RAG searches across these embedding collections:

| Collection | Source |
|------------|--------|
| `knowledge` | Knowledge store entries |
| `observation` | Observational memory observations |
| `reflection` | Observational memory reflections |
| `learning` | Learning engine entries |

You can restrict which collections are searched:

```yaml
embedding:
  rag:
    enabled: true
    maxResults: 5
    collections:
      - knowledge
      - learning
```

Leave `collections` empty to search all collections.

### Distance Filtering

Set `maxDistance` to filter out low-relevance results:

```yaml
embedding:
  rag:
    enabled: true
    maxResults: 5
    maxDistance: 0.8  # Only include results with cosine distance <= 0.8
```

Set to `0.0` (default) to disable distance filtering.

## Configuration Reference

```yaml
embedding:
  providerID: ""          # Reference to providers map entry
  provider: ""            # "local" for Ollama (no providerID needed)
  model: ""               # Embedding model identifier
  dimensions: 0           # Vector dimensionality
  local:
    baseUrl: http://localhost:11434/v1
    model: ""             # Override model for local provider
  rag:
    enabled: false
    maxResults: 5
    collections: []       # Empty = search all
    maxDistance: 0.0       # 0.0 = no distance filter
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `providerID` | `string` | `""` | References a key in the `providers` map |
| `provider` | `string` | `""` | Set to `"local"` for Ollama embeddings |
| `model` | `string` | varies | Embedding model identifier |
| `dimensions` | `int` | varies | Vector dimensionality |
| `local.baseUrl` | `string` | `http://localhost:11434/v1` | Ollama API endpoint |
| `local.model` | `string` | `""` | Override model for local provider |
| `rag.enabled` | `bool` | `false` | Enable RAG context injection |
| `rag.maxResults` | `int` | `5` | Maximum results to inject per query |
| `rag.collections` | `[]string` | `[]` | Collections to search (empty = all) |
| `rag.maxDistance` | `float32` | `0.0` | Maximum cosine distance (0.0 = disabled) |

## Embedding Cache

Lango includes an automatic in-memory embedding cache to reduce redundant API calls:

| Parameter | Value |
|-----------|-------|
| TTL | 5 minutes |
| Max entries | 100 |
| Eviction | Expired entries first, then oldest |

The cache is transparent and requires no configuration. It applies to both query embeddings and content embeddings.

## Verification

Use `lango doctor` to verify your embedding configuration:

```bash
lango doctor
```

The doctor checks:

- Provider configuration is valid
- API key is set (for cloud providers)
- Embedding model is accessible
- Vector store is operational

## Related

- [Knowledge System](knowledge.md) -- Knowledge entries are embedded for RAG retrieval
- [Observational Memory](observational-memory.md) -- Observations and reflections are embedded
- [Knowledge Graph](knowledge-graph.md) -- Graph RAG combines vector + graph retrieval
