## Context

Steps 2-6 built FTS5 search, context budgeting, temporal versioning, structured taxonomy, and the agentic retrieval coordinator with shadow mode. The coordinator compares old vs new retrieval results, but there is no event-level tracking of what context was actually injected into each LLM turn. Step 7 adds this observability layer as the foundation for future score-tuning (Step 13).

## Goals / Non-Goals

**Goals:**
- Track which knowledge items were injected into which LLM turn via structured event
- Propagate TurnID through context so it's available at context assembly time
- Provide structured logging of injection metrics (layer/source distribution, token counts)
- Make feedback processor independent of knowledge/coordinator enabled state

**Non-Goals:**
- Score auto-adjustment (reserved for Step 13)
- Persistent storage of injection events (logging only for v1)
- RAG/memory/runSummary item-level decomposition (aggregate token counts only)
- PII-safe query persistence (processor must not log raw query text)

## Decisions

### DD1: TurnID context key in `session` package
TurnID propagated via `session.WithTurnID`/`TurnIDFromContext`, same pattern as SessionKey. Already imported by `adk/context_model.go`, no new dependency.
**Alternative**: `turntrace` package — rejected because TurnID is request-scoped metadata, not durable trace storage.

### DD2: Event type in `eventbus/retrieval_events.go`
New file following `observability_events.go` convention. `ContextInjectedItem.Layer` is `string` (not `ContextLayer`) to keep eventbus dependency-free, mirroring the `Triple` pattern.
**Alternative**: Define in `adk` package — rejected because `retrieval` imports would create cycle (`adk` imports `retrieval`).

### DD3: Items scope — knowledge structured results only
`ContextInjectedEvent.Items` contains only knowledge `ContextRetriever` structured results (`oldRetrieved`). RAG/memory/runSummary are opaque section strings — represented as aggregate token counts (`RAGTokens`, `MemoryTokens`, `RunSummaryTokens`).
**Alternative**: Item-level decomposition for all sections — rejected as over-engineering for v1.

### DD4: Publication point — after section assembly, before LLM call
Published after all sections are combined but before `inner.GenerateContent`. At this point all data is final (post-truncation). Guarded by `if m.bus != nil`. Empty retrieval still publishes (useful for "no context" debugging). TurnID may be empty (direct calls without TurnRunner) — still published, processor skips correlation.

### DD5: FeedbackProcessor — logging only, zero mutations
Subscribes to `ContextInjectedEvent`, logs structured metrics. MUST NOT modify scores, use counts, or any stored data. Pre-allocated fields slice avoids copy when prepending optional TurnID.

### DD6: Config — `retrieval.feedback` independent of coordinator
`RetrievalConfig.Feedback` operates independently of `Enabled`/`Shadow`. Feedback processor wired when `cfg.Retrieval.Feedback && bus != nil` regardless of knowledge/coordinator state.

### DD7: Bus wired unconditionally to ctxAdapter
`ctxAdapter.WithEventBus(deps.eventBus)` in both ctxAdapter branches (knowledge + OM-only). Feedback observability applies to all context injection, not just agentic retrieval.

## Risks / Trade-offs

- **[Synchronous bus on hot path]** → Event bus dispatch is synchronous; handler must be fast. FeedbackProcessor does logging only — negligible overhead. Document constraint for future subscribers.
- **[TurnID absent in direct calls]** → `GenerateContent(context.Background(), ...)` has no TurnID. Processor handles gracefully (empty string, skips correlation).
- **[PII in event Query field]** → Raw query stored in event struct (in-memory only, no persistence). Processor logs `query_length` only. Future persistence would need hash/truncation.
