## Why

Chat Response Speed Optimization features (streaming, embedding cache, context limits) have been implemented but are not documented in README.md. Users cannot discover these capabilities or configure new settings without documentation.

## What Changes

- Update Gateway feature description to mention real-time streaming
- Add `observationalMemory.maxReflectionsInContext` and `maxObservationsInContext` config fields to the reference table
- Add Embedding Cache subsection under Embedding & RAG
- Add Context Limits bullet to Observational Memory description
- Add WebSocket Events subsection documenting `agent.thinking`, `agent.chunk`, `agent.done` events

## Capabilities

### New Capabilities

(none - documentation only)

### Modified Capabilities

- `docs-only`: README.md update to document streaming, embedding cache, and context limit features

## Impact

- **Files**: `README.md` only
- **Code**: No code changes - documentation only
- **APIs**: No API changes (documents existing WebSocket events)
- **Dependencies**: None
