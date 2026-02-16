## Context

The codebase accumulated backward-compatibility patterns during early development: deprecated config fields with migration/fallback logic, a dead-code wrapper function, a hardcoded message cap, and misleading comments. As a new project with no external consumers, these patterns add complexity without benefit.

## Goals / Non-Goals

**Goals:**
- Remove all identified legacy patterns from Go code
- Simplify `ResolveEmbeddingProvider()` to only support ProviderID and local
- Replace hardcoded 100-message cap with token-budget-based truncation
- Update all tests to reflect the simplified behavior

**Non-Goals:**
- Changing any user-facing behavior for correctly configured systems
- Modifying auto-generated ent code
- Updating archived openspec documents

## Decisions

1. **EmbeddingConfig.Provider kept for "local" only** — The `Provider` field is retained because local (Ollama) embeddings don't need a providers map entry. The legacy type-search fallback (searching providers by type name) is removed. Users of non-local embeddings must use `ProviderID`.

2. **Default token budget of 32000** — When `EventsAdapter` has no explicit budget, it defaults to 32000 tokens instead of a 100-message hardcap. This aligns with the observational memory default and provides content-aware truncation.

3. **Config fields silently ignored** — Removed fields (`approvalRequired`, `systemPromptPath`, non-local `embedding.provider`) will be silently ignored by `mapstructure` unmarshaling. No migration path needed for a pre-release project.

## Risks / Trade-offs

- [Configs with `approvalRequired: true` lose that setting] → Acceptable for pre-release; `approvalPolicy: "dangerous"` is the default and provides equivalent protection.
- [Configs using `embedding.provider: "openai"` stop resolving] → Users must switch to `embedding.providerID` referencing their providers map entry.
- [Default 32000 token budget may include more messages than the old 100-message cap] → This is intentional; token-based truncation is more accurate than message count.
