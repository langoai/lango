## 1. Token Counter Infrastructure

- [x] 1.1 Create `internal/memory/types.go` with Observation, Reflection, and MemoryConfig types
- [x] 1.2 Create `internal/memory/token.go` with character-based token estimation (ASCII 1:4, CJK 1:2 ratio)
- [x] 1.3 Create `internal/memory/token.go` message token counting (content + role overhead + tool calls)
- [x] 1.4 Write tests for token counter (`internal/memory/token_test.go`) covering ASCII, CJK, mixed, empty, and message batch scenarios

## 2. Ent Schema for Observations and Reflections

- [x] 2.1 Create `internal/ent/schema/observation.go` Ent schema (id, session_key, content, token_count, source_start_index, source_end_index, created_at)
- [x] 2.2 Create `internal/ent/schema/reflection.go` Ent schema (id, session_key, content, token_count, generation, created_at)
- [x] 2.3 Run `go generate ./internal/ent` to generate Ent code
- [x] 2.4 Verify auto-migration works with new schemas

## 3. OM Data Store

- [x] 3.1 Create `internal/memory/store.go` with OM data access methods (SaveObservation, ListObservations, DeleteObservations, SaveReflection, ListReflections)
- [x] 3.2 Write tests for OM store (`internal/memory/store_test.go`) covering CRUD operations

## 4. Observer Agent

- [x] 4.1 Create `internal/memory/observer.go` with Observer struct and observation generation logic
- [x] 4.2 Implement observer prompt template for compressing messages into observation notes
- [x] 4.3 Integrate with provider system for LLM calls (configurable provider/model)
- [x] 4.4 Write tests for Observer (`internal/memory/observer_test.go`) including trigger logic and graceful degradation

## 5. Reflector Agent

- [x] 5.1 Create `internal/memory/reflector.go` with Reflector struct and condensation logic
- [x] 5.2 Implement reflector prompt template for condensing observations into reflections
- [x] 5.3 Implement multi-generation reflection support
- [x] 5.4 Write tests for Reflector (`internal/memory/reflector_test.go`) including trigger and observation cleanup

## 6. Async Observation Buffer

- [x] 6.1 Create `internal/memory/buffer.go` with goroutine-based async buffer (triggerCh, doneCh, WaitGroup integration)
- [x] 6.2 Implement graceful shutdown (drain pending, complete in-progress)
- [x] 6.3 Write tests for buffer (`internal/memory/buffer_test.go`) covering async trigger, shutdown, and concurrent safety

## 7. Configuration

- [x] 7.1 Add `ObservationalMemoryConfig` to `internal/config/types.go` (enabled, provider, model, messageTokenThreshold, observationTokenThreshold, maxMessageTokenBudget)
- [x] 7.2 Wire config defaults (enabled: false, messageTokenThreshold: 1000, observationTokenThreshold: 2000, maxMessageTokenBudget: 8000)

## 8. Context Retriever Integration

- [x] 8.1 Add `LayerObservations` and `LayerReflections` constants to `internal/knowledge/types.go`
- [x] 8.2 Modify `ContextAwareModelAdapter` in `internal/adk/context_model.go` to accept and inject observations/reflections into system prompt
- [x] 8.3 Add "Conversation Memory" section to prompt assembly with reflections before observations

## 9. ADK EventsAdapter Token-Budget Truncation

- [x] 9.1 Modify `EventsAdapter` in `internal/adk/state.go` to accept token counter and budget config
- [x] 9.2 Replace hard 100-message cap with token-budget-based dynamic truncation (iterate most-recent-first until budget exhausted)
- [x] 9.3 Add fallback to 100-message cap when OM is disabled
- [x] 9.4 Write tests for token-budget truncation (`internal/adk/state_test.go`)

## 10. Application Wiring

- [x] 10.1 Initialize OM components in `internal/app/app.go` (store, observer, reflector, buffer)
- [x] 10.2 Start observer buffer goroutine with WaitGroup registration
- [x] 10.3 Hook observation trigger into message append flow
- [x] 10.4 Wire OM data into ContextAwareModelAdapter initialization

## 11. Integration Testing

- [x] 11.1 Write end-to-end test: message append → threshold trigger → observation generated → context assembly includes observation
- [x] 11.2 Write test: observation accumulation → reflection trigger → observations replaced by reflection
- [x] 11.3 Write test: OM disabled → behavior identical to current system (100-message cap, no observations)
