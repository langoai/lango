## Context

All three embedding providers (Google, OpenAI, Local/Ollama) store a `dimensions` field from config but never pass it to the underlying API calls. The SQLite vec table is created with the configured dimension (e.g., `float[128]`), while the API returns the model's native dimension (e.g., 3072), causing every embedding insert and RAG search to fail with a dimension mismatch error.

## Goals / Non-Goals

**Goals:**
- Ensure all embedding providers pass the configured `dimensions` to their respective API calls
- Fix the dimension mismatch that prevents all RAG functionality from working

**Non-Goals:**
- Changing default dimension values or adding dimension validation
- Migrating or recreating the SQLite vec table (it's empty since all inserts failed)
- Adding new embedding providers or modifying the provider interface

## Decisions

### Pass dimensions at the API call level

Each provider already stores `dimensions` but discards it when constructing API requests. The fix is to pass this value through to each provider's native API parameter:

- **Google**: `EmbedContentConfig.OutputDimensionality` (pointer to int32)
- **OpenAI**: `EmbeddingRequest.Dimensions` (int field)
- **Local/Ollama**: Same as OpenAI — uses the OpenAI-compatible client with `EmbeddingRequest.Dimensions`

**Rationale**: This is the minimal change that fixes the root cause. Each API supports dimension truncation natively, so the returned vectors will match the configured dimension exactly.

## Risks / Trade-offs

- [Ollama model support] Not all Ollama models support the `dimensions` parameter. → Models that don't support it may ignore the field or return an error. Users would need to set dimensions to match the model's native output.
- [Dimension value correctness] If a user configures dimensions larger than the model's native output, the API will return an error. → This is expected behavior and the error message from the API will be clear.
