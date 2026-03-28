## Why

The retrieval pipeline (Steps 2-6) has no tracking of which context items were injected into which LLM turn. When debugging retrieval quality or analyzing context relevance, there is no way to answer "what knowledge was the model given for turn X?" This observability gap blocks future score-tuning work (Step 13).

## What Changes

- Add `ContextInjectedEvent` to the event bus, published after context assembly in `GenerateContent` with TurnID, session key, structured knowledge items (with scores/sources), and per-section token estimates
- Add `TurnID` propagation via `session.WithTurnID`/`TurnIDFromContext` (same pattern as SessionKey), set by TurnRunner before executor call
- Add `ContextLayer.String()` method for human-readable layer names in events/logs
- Add `FeedbackProcessor` that subscribes to the event and performs structured logging (layer distribution, source distribution, token totals) — read-only, no score mutations
- Add `RetrievalConfig.Feedback` toggle (independent of `Enabled`/`Shadow`)

## Capabilities

### New Capabilities
- `retrieval-feedback-observability`: Context injection event tracking, TurnID propagation, feedback processor for structured observability logging

### Modified Capabilities
- `agentic-retrieval`: Add `Feedback` field to `RetrievalConfig`, wire event bus into `ContextAwareModelAdapter`
- `knowledge-store`: Add `ContextLayer.String()` method

## Impact

- **MODIFY**: `internal/adk/context_model.go` — bus field, WithEventBus(), event publication, buildContextInjectedItems helper
- **MODIFY**: `internal/session/context.go` — TurnID context key functions
- **MODIFY**: `internal/turnrunner/runner.go` — TurnID propagation into context
- **MODIFY**: `internal/knowledge/types.go` — ContextLayer.String() method
- **MODIFY**: `internal/config/types.go`, `internal/config/loader.go` — Feedback config field
- **MODIFY**: `internal/app/wiring.go` — WithEventBus wiring in both ctxAdapter branches
- **MODIFY**: `internal/app/wiring_knowledge.go` — initFeedbackProcessor helper
- **NEW**: `internal/eventbus/retrieval_events.go` — ContextInjectedEvent, ContextInjectedItem
- **NEW**: `internal/retrieval/feedback_processor.go` — FeedbackProcessor
- **NEW**: `internal/retrieval/feedback_processor_test.go` — processor tests
