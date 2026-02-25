# Design: Codebase Consolidation

## Overview
Three independent consolidation efforts executed in parallel, plus one skipped due to constraints.

## Phase 1: Generic AsyncBuffer

### New Package: `internal/asyncbuf/`

Two generic types replacing 5 duplicate buffer implementations:

**BatchBuffer[T]** — Timer-based batch collection with configurable flush:
- `BatchConfig{QueueSize, BatchSize, BatchTimeout}`
- `ProcessBatchFunc[T] func(batch []T)`
- Non-blocking `Enqueue` with drop counting
- Drain-on-shutdown semantics

**TriggerBuffer[T]** — Per-item async processing:
- `TriggerConfig{QueueSize}`
- `ProcessFunc[T] func(item T)`
- Non-blocking `Enqueue`, drain-on-shutdown

### Migration Strategy
Each existing buffer becomes a thin wrapper:
- `EmbeddingBuffer` → `asyncbuf.BatchBuffer[EmbedRequest]`
- `GraphBuffer` → `asyncbuf.BatchBuffer[GraphRequest]`
- `Buffer` (memory) → `asyncbuf.TriggerBuffer[string]`
- `AnalysisBuffer` → `asyncbuf.TriggerBuffer[AnalysisRequest]`
- `ProactiveBuffer` → `asyncbuf.TriggerBuffer[string]`

All public APIs remain identical. Domain-specific logic stays in the wrapper's process callback.

## Phase 2: Package Merges

### 2a. ctxutil → types
`Detach()` (28 LOC) moved to `internal/types/context.go` as `DetachContext()`. Better naming in the broader `types` namespace.

### 2b. passphrase → security/passphrase
All files moved under `internal/security/passphrase/`. Package name unchanged. Logical grouping with security domain.

### 2c. zkp → p2p/zkp
All files (including `circuits/` subdirectory) moved under `internal/p2p/zkp/`. ZKP is exclusively used for P2P proof verification.

## Phase 3: CLI UX

### Command Groups
```
Core:            serve, version, health
Configuration:   config, settings, onboard, doctor
Data & AI:       memory, graph, agent
Infrastructure:  security, p2p, cron, workflow, payment
```

### Cross-References
Each config-related command's `Long` description includes "See Also" pointing to the other three.

## Phase 4: Type Consolidation (SKIPPED)
`MessageProvider func(sessionKey string) ([]session.Message, error)` is duplicated in memory, learning, and librarian packages. Cannot consolidate into `types` because `session` imports `types` (for `MessageRole`), creating a cycle.
