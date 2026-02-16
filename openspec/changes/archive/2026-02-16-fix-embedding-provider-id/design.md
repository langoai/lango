## Context

The embedding system resolves API keys using hardcoded provider name lookups (`cfg.Providers["openai"]`, `cfg.Providers["gemini"]`). Users who register providers with custom IDs (e.g., `"gemini-1"`, `"my-openai"`) cannot use them for embedding because the lookup logic cannot find them. The fix introduces an explicit `ProviderID` reference that resolves type and key from the providers map.

## Goals / Non-Goals

**Goals:**
- Allow users to reference any registered provider ID for embedding configuration
- Auto-resolve embedding backend type and API key from the provider config
- Maintain full backward compatibility with existing `embedding.provider` (type-based) configs
- Update onboard TUI to show user's actual providers instead of hardcoded list
- Update doctor checks to validate using the unified resolver

**Non-Goals:**
- Changing the embedding provider interface or implementations
- Adding new embedding providers
- Migrating existing configs automatically (legacy configs continue to work as-is)

## Decisions

### 1. Add `ProviderID` field alongside existing `Provider` field

**Rationale**: Adding a new field rather than changing the semantics of the existing `Provider` field ensures zero breakage for existing users. `ProviderID` takes precedence when set; otherwise the legacy `Provider` type-based resolution is used.

**Alternative considered**: Overloading the `Provider` field to accept both types and IDs. Rejected because it creates ambiguity (is `"openai"` a type or an ID?).

### 2. Exported `ProviderTypeToEmbeddingType` map

**Rationale**: The mapping is needed by both the config resolver and the onboard TUI state handler. Exporting it avoids duplication and keeps the mapping in a single authoritative location.

### 3. `ResolveEmbeddingProvider()` method on `Config`

**Rationale**: Centralizes the resolution logic (ProviderID → providers map → type + key) in one place. All consumers (`initEmbedding`, doctor checks) call this single method instead of duplicating lookup logic.

### 4. Unsupported provider types return empty backend

**Rationale**: When a provider type (e.g., `"anthropic"`) has no embedding support, `ResolveEmbeddingProvider` returns `("", "")`. Callers handle this uniformly — wiring logs a warning and skips, doctor reports a failure.

## Risks / Trade-offs

- [Map iteration order in legacy fallback] → Legacy type-based lookup iterates the providers map, which has non-deterministic order. If multiple providers share the same type, the first match with a non-empty API key wins. This matches the previous behavior. → Acceptable for legacy path; users should migrate to `ProviderID` for deterministic resolution.
- [Onboard form key change] → The form key changed from `emb_provider` to `emb_provider_id`. The old key handler is preserved in `state_update.go` for safety, but the form no longer generates it. → No runtime impact since form keys are ephemeral in-memory values, not persisted.
