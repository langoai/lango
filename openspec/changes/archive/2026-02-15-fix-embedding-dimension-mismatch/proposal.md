## Why

All three embedding providers (Google, OpenAI, Local) ignore the `dimensions` value from `config.json` when making API calls, causing a dimension mismatch error. The SQLite vec table is created with the configured dimension (e.g., `float[128]`), but the API returns the model's native dimension (e.g., 3072 for gemini-embedding-001), resulting in `Expected 128 dimensions but received 3072` errors on every RAG search and embedding insert.

## What Changes

- Pass `OutputDimensionality` config to Google's `EmbedContent` API call
- Pass `Dimensions` field in OpenAI's `EmbeddingRequest`
- Pass `Dimensions` field in Local (Ollama) provider's `EmbeddingRequest`

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `embedding-rag`: Embedding providers must pass the configured `dimensions` value to the underlying API so that returned vectors match the SQLite vec table schema.

## Impact

- `internal/embedding/google.go` — `EmbedContent` call updated with `EmbedContentConfig`
- `internal/embedding/openai.go` — `EmbeddingRequest` includes `Dimensions` field
- `internal/embedding/local.go` — `EmbeddingRequest` includes `Dimensions` field
- No schema migration needed: existing `vec_embeddings` table is empty (all prior inserts failed due to this bug)
