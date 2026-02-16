## Why

The embedding system only accepts hardcoded provider types (`"openai"`, `"google"`, `"local"`) and resolves API keys by searching for hardcoded provider names (`cfg.Providers["openai"]`, `cfg.Providers["gemini"]`). When users register providers with custom IDs like `"gemini-1"` or `"my-openai"`, the system cannot find their API keys.

## What Changes

- Add `ProviderID` field to `EmbeddingConfig` allowing users to reference their own provider IDs directly.
- Add `ResolveEmbeddingProvider()` method on `Config` that resolves embedding backend type and API key from the provider ID or falls back to legacy type-based lookup.
- Add `ProviderTypeToEmbeddingType` mapping (provider type → embedding backend type).
- Update `initEmbedding` wiring to use the new resolver instead of hardcoded switch statements.
- Update onboard TUI embedding form to show the user's actual registered providers instead of hardcoded type list.
- Update `state_update.go` to handle the new `emb_provider_id` form key with auto-resolution of provider type.
- Update doctor embedding check to use `ResolveEmbeddingProvider()` for unified validation.

## Capabilities

### New Capabilities

### Modified Capabilities
- `embedding-rag`: Add `providerID` field support for referencing user-registered provider IDs. Backend type and API key are auto-resolved from the providers map.
- `cli-onboard`: Embedding form now shows user's registered provider IDs instead of hardcoded type list.
- `cli-doctor`: Embedding check uses unified provider resolver for validation.

## Impact

- `internal/config/types.go` — new field, mapping, and resolver method
- `internal/app/wiring.go` — `initEmbedding` uses resolver
- `internal/cli/onboard/forms_impl.go` — embedding form provider options
- `internal/cli/onboard/state_update.go` — new `emb_provider_id` handler
- `internal/cli/doctor/checks/embedding.go` — unified validation
- `README.md` — `embedding.providerID` configuration reference
- Backward compatible: existing configs with `embedding.provider` continue to work
