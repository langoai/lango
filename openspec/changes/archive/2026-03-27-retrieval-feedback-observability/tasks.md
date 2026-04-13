## 1. Foundation Types

- [x] 1.1 Add `ContextLayer.String()` method to `internal/knowledge/types.go` with all 9 layers + unknown fallback
- [x] 1.2 Add `turnIDCtxKey`, `WithTurnID()`, `TurnIDFromContext()` to `internal/session/context.go`
- [x] 1.3 Add TurnID context tests to `internal/session/context_test.go`

## 2. Event Types

- [x] 2.1 Create `internal/eventbus/retrieval_events.go` with `ContextInjectedEvent` and `ContextInjectedItem`

## 3. TurnID Propagation

- [x] 3.1 Add `ctx = langosession.WithTurnID(ctx, traceID)` to `internal/turnrunner/runner.go`

## 4. Event Publication

- [x] 4.1 Add `bus *eventbus.Bus` field and `WithEventBus()` method to `ContextAwareModelAdapter`
- [x] 4.2 Add `buildContextInjectedItems()` helper to convert `RetrievalResult` to event items
- [x] 4.3 Publish `ContextInjectedEvent` after section assembly in `GenerateContent`

## 5. FeedbackProcessor

- [x] 5.1 Create `internal/retrieval/feedback_processor.go` with `NewFeedbackProcessor`, `Subscribe`, and `handleContextInjected`
- [x] 5.2 Create `internal/retrieval/feedback_processor_test.go` with table-driven tests

## 6. Config + Wiring

- [x] 6.1 Add `Feedback bool` to `RetrievalConfig` in `internal/config/types.go`
- [x] 6.2 Add `Feedback: false` default in `internal/config/loader.go`
- [x] 6.3 Add `initFeedbackProcessor(cfg, bus)` to `internal/app/wiring_knowledge.go`
- [x] 6.4 Wire `ctxAdapter.WithEventBus(deps.eventBus)` in both branches of `internal/app/wiring.go`
- [x] 6.5 Call `initFeedbackProcessor(cfg, deps.eventBus)` after ctxAdapter creation in `wiring.go`

## 7. Verification

- [x] 7.1 Run `CGO_ENABLED=1 go build -tags fts5 ./...` — full project builds
- [x] 7.2 Run tests for session, retrieval, knowledge packages — all pass
