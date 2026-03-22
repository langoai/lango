## Why

`internal/app/` has grown to 76 files / 17.7K LOC with no rollback on Phase B wiring failures, duplicate interfaces across packages, zero test coverage on 6 CLI packages, dual coupling patterns (EventBus + SetXxxCallback), tool builders tightly coupled to app-private glue structs, and config fan-in from 196 import sites. These structural inefficiencies slow down development velocity, make testing harder, and increase the risk of resource leaks during initialization failures.

## What Changes

- Add a `cleanupStack` rollback mechanism to `app.New()` Phase B — gateway and output-store are cleaned up on agent creation failure
- Consolidate 4 duplicate `TextGenerator` interface definitions into `internal/llm/` package
- Build shared CLI test harness (`internal/testutil/cli_harness.go`) and add tests for 6 zero-coverage CLI packages (memory, graph, learning, librarian, approval, cron)
- Migrate 9 app-level cross-domain `SetXxxCallback` calls to EventBus publish/subscribe pattern with `NeedsGraph` field to preserve graph routing semantics
- Extract `buildEconomyTools()` from `app/tools_economy.go` to `economy.BuildTools()` as a pilot for domain-owned tool builders
- Introduce `p2p.ConfigReader` interface so `internal/p2p/` no longer imports `internal/config/` directly

## Capabilities

### New Capabilities
- `phase-b-rollback`: Cleanup stack for Phase B post-build wiring with reverse-order rollback on failure
- `llm-text-generator`: Shared `TextGenerator` interface in `internal/llm/` replacing 4 duplicate definitions
- `cli-test-harness`: Shared CLI test infrastructure with fake config loader, in-memory stores, and stdout capture
- `domain-tool-builders`: Pattern for domain packages to own their tool builder functions (economy pilot)
- `narrow-config-reader`: Consumer-side config reader interfaces to decouple domain packages from monolithic config

### Modified Capabilities
- `eventbus`: Add `NeedsGraph` field to `ContentSavedEvent`, add `ReputationChangedEvent`, migrate 9 cross-domain callbacks
- `callback-wiring`: Remove 9 `SetXxxCallback` setters from knowledge, memory, learning, librarian, and reputation stores in favor of EventBus
- `tool-catalog`: Economy tools now registered via `economy.BuildTools()` instead of app-level `buildEconomyTools()`

## Impact

- `internal/app/app.go` — cleanupStack type, Phase B rollback, intelligenceModule bus wiring
- `internal/app/modules.go` — bus field on intelligenceModule, economy.BuildTools() call
- `internal/app/wiring_*.go` — 6 wiring files migrated from callback setters to EventBus subscriptions
- `internal/app/tools_economy.go` — deleted (moved to `internal/economy/tools.go`)
- `internal/llm/` — new package with shared TextGenerator interface
- `internal/knowledge/store.go`, `internal/memory/store.go` — SetEventBus replaces SetEmbedCallback/SetGraphCallback
- `internal/learning/` — graph_engine, conversation_analyzer, session_learner migrated to EventBus
- `internal/librarian/proactive_buffer.go` — migrated to EventBus
- `internal/p2p/reputation/store.go` — migrated to EventBus
- `internal/p2p/node.go` — accepts ConfigReader interface instead of config.P2PConfig
- `internal/p2p/config.go` — new ConfigReader interface
- `internal/config/types_p2p.go` — getter methods added for ConfigReader compliance
- `internal/eventbus/events.go` — NeedsGraph field, ReputationChangedEvent
- `internal/testutil/cli_harness.go` — new shared test infrastructure
- `internal/cli/{memory,graph,learning,librarian,approval,cron}/` — new test files
