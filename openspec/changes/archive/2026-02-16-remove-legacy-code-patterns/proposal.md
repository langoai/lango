## Why

This is a new project with no external users, so backward compatibility wrappers, deprecated fields, and migration logic add unnecessary complexity. Six legacy patterns were identified across the Go codebase that should be removed to simplify the code and prevent confusion.

## What Changes

- **BREAKING**: Remove `NewSQLiteStore` compatibility wrapper (dead code, no callers)
- **BREAKING**: Remove `InterceptorConfig.ApprovalRequired` deprecated field and `migrateApprovalPolicy()` migration function
- **BREAKING**: Remove `AgentConfig.SystemPromptPath` deprecated field and legacy single-file prompt fallback
- **BREAKING**: Remove `EmbeddingConfig.Provider` legacy type-search fallback in `ResolveEmbeddingProvider()`; keep `Provider` field only for local (Ollama) embeddings
- **BREAKING**: Remove `EventsAdapter` legacy 100-message hardcap fallback; use default token budget (32000) when no explicit budget is set
- Clean up misleading backward-compatibility comment in `ent_store.go` salt column handling

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ent-session-store`: Remove `NewSQLiteStore` wrapper function and clarify salt column comment
- `approval-policy`: Remove deprecated `ApprovalRequired` field and migration logic
- `embedding-rag`: Remove legacy `Provider` type-search fallback; `Provider` field now only supports "local"
- `cli-onboard`: Remove `SystemPromptPath` form field and legacy embedding provider fallback in state update
- `adk-architecture`: Replace 100-message hardcap with default token budget truncation

## Impact

- **Config files**: Existing configs using `approvalRequired`, `systemPromptPath`, or `embedding.provider` (non-local) will have those fields silently ignored
- **Code**: Changes span `internal/config`, `internal/session`, `internal/app`, `internal/adk`, `internal/cli/onboard`, `internal/cli/doctor/checks`
- **Tests**: Migration tests removed, truncation tests updated, form tests updated to reflect removed fields
