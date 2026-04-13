## 1. Phase B Rollback Mechanism

- [x] 1.1 Add `cleanupEntry` and `cleanupStack` types with push/rollback/clear methods to `internal/app/app.go`
- [x] 1.2 Wire cleanup pushes for OutputStore (B4b) and Gateway (B5) in Phase B
- [x] 1.3 Add `cleanups.rollback()` call on B6 agent creation failure
- [x] 1.4 Add `cleanups.clear()` at end of successful Phase B
- [x] 1.5 Add tests: reverse order, clear-without-execute, empty rollback, partial rollback, integration test

## 2. TextGenerator Consolidation

- [x] 2.1 Create `internal/llm/text_generator.go` with unified `TextGenerator` interface
- [x] 2.2 Update `internal/learning/conversation_analyzer.go` to use `llm.TextGenerator`
- [x] 2.3 Update `internal/learning/session_learner.go` to use `llm.TextGenerator`
- [x] 2.4 Update `internal/memory/observer.go` and `reflector.go` to use `llm.TextGenerator`
- [x] 2.5 Update `internal/graph/extractor.go` to use `llm.TextGenerator`
- [x] 2.6 Update `internal/librarian/types.go`, `inquiry_processor.go`, `observation_analyzer.go` to use `llm.TextGenerator`
- [x] 2.7 Update affected test files (`buffer_test.go`, `observer_test.go`, `reflector_test.go`)

## 3. CLI Test Harness

- [x] 3.1 Create `internal/testutil/cli_harness.go` with FakeCfgLoader, FakeBootLoader, ExecCmd
- [x] 3.2 Add tests for `internal/cli/memory/`
- [x] 3.3 Add tests for `internal/cli/graph/`
- [x] 3.4 Add tests for `internal/cli/learning/`
- [x] 3.5 Add tests for `internal/cli/librarian/`
- [x] 3.6 Add tests for `internal/cli/approval/`
- [x] 3.7 Add tests for `internal/cli/cron/`

## 4. EventBus Callback Migration

- [x] 4.1 Add `NeedsGraph` field to `ContentSavedEvent` and `ReputationChangedEvent` to `eventbus/events.go`
- [x] 4.2 Add `SetEventBus`/`publishContentSaved` to `knowledge/store.go` with correct NeedsGraph per path
- [x] 4.3 Add `SetEventBus`/`publishContentSaved` to `memory/store.go` with NeedsGraph=true
- [x] 4.4 Migrate `learning/graph_engine.go` to EventBus (SetEventBus/publishTriples)
- [x] 4.5 Migrate `learning/conversation_analyzer.go` and `session_learner.go` to EventBus
- [x] 4.6 Migrate `librarian/proactive_buffer.go` to EventBus
- [x] 4.7 Migrate `p2p/reputation/store.go` to EventBus (ReputationChangedEvent)
- [x] 4.8 Update `wiring_embedding.go` to use SubscribeTyped[ContentSavedEvent]
- [x] 4.9 Update `wiring_graph.go` to subscribe with NeedsGraph filter
- [x] 4.10 Update `wiring_knowledge.go` to call SetEventBus on stores and engines
- [x] 4.11 Update `wiring_librarian.go` and `wiring_memory.go` to call SetEventBus
- [x] 4.12 Update `wiring_p2p.go` to subscribe to ReputationChangedEvent
- [x] 4.13 Add `bus` field to `intelligenceModule` struct and pass to init functions
- [x] 4.14 Update learning tests (`graph_engine_test.go`, `conversation_analyzer_test.go`)
- [x] 4.15 Add regression test `TestContentSavedEvent_NeedsGraph` in `knowledge/store_test.go`

## 5. Economy Tool Builder Extraction

- [x] 5.1 Create `internal/economy/tools.go` with `BuildTools()` function
- [x] 5.2 Move 5 sub-builders (budget, risk, negotiation, escrow, pricing) from app to economy
- [x] 5.3 Update `app/modules.go` to call `economy.BuildTools()` in networkModule.Init()
- [x] 5.4 Delete `app/tools_economy.go`
- [x] 5.5 Add `internal/economy/tools_test.go` with nil-guard and no-app-import tests

## 6. P2P Config Reader Interface

- [x] 6.1 Create `internal/p2p/config.go` with `ConfigReader` interface (6 methods)
- [x] 6.2 Update `internal/p2p/node.go` to use `ConfigReader` instead of `config.P2PConfig`
- [x] 6.3 Add getter methods to `config.P2PConfig` in `internal/config/types_p2p.go`
- [x] 6.4 Verify `internal/p2p/` no longer imports `internal/config/`
