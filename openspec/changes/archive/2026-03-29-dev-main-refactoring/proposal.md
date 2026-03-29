## Why

dev→main merge (1,073 files, +110K lines) 이후 코드 분석에서 TUI hot-path 성능 문제, RunLedger O(N) 탐색, store 패턴 중복, 타입 단언 불일치, 이벤트명 stringly-typed 등 12건의 개선점을 발견. UX 즉각 개선(P0)부터 코드 품질(P2), 최적화(P3)까지 4-phase 리팩토링으로 해결.

## What Changes

- Cache glamour markdown renderer to eliminate per-tick recreation (400ms interval)
- Skip redundant transcript re-render on cursor blink with width-change guard
- Add lazy step index to RunSnapshot.FindStep() for O(1) lookup (was O(N²))
- Persist SourceKind/SourceDescriptor in RunLedger journal for future reconciliation
- Introduce `internal/storeutil/` package with shared MarshalField/UnmarshalField/CopySlice/CopyMap helpers
- Add `RequireFloat64`/`OptionalFloat64` to toolparam, migrate 7 tools files from manual type assertions
- Extract 48 event name constants across 6 eventbus files
- Clarify Finding.SearchSource vs Source doc comments
- Add `internal/cli/tuicore/field_builder.go` with form field factory functions, migrate forms_knowledge.go and forms_p2p.go
- Add fast-path early return in ContextBudgetManager.ReallocateBudgets()
- Replace mutex-based coordinator merge with lock-free index-based approach
- Add tool stats dirty-flag sort guard in cockpit contextpanel with content-change detection
- Fix mockStore/uniqueMockStore missing ListSessions in adk tests

## Capabilities

### New Capabilities
- `store-util-helpers`: Shared JSON marshal/unmarshal and generic copy helpers for store implementations
- `tui-form-field-builder`: Factory functions for common TUI form field patterns (BoolInput, IntInput, SelectInput, etc.)

### Modified Capabilities
- `run-ledger`: SourceKind/SourceDescriptor persistence in journal, lazy step index in FindStep, storeutil integration
- `agentic-retrieval`: Lock-free coordinator merge, toolparam migration
- `tui-chat-rendering`: Cached markdown renderer, skip redundant re-render on cursor tick

## Impact

- **Core**: `internal/runledger/` (snapshot, journal, ent_store, writethrough, tools), `internal/adk/` (budget, state_test)
- **Retrieval**: `internal/retrieval/` (coordinator, finding), `internal/toolparam/`
- **TUI**: `internal/cli/chat/` (markdown, chatview), `internal/cli/cockpit/` (contextpanel), `internal/cli/settings/` (forms_knowledge, forms_p2p), `internal/cli/tuicore/`
- **Events**: `internal/eventbus/` (6 event files)
- **Stores**: `internal/provenance/ent_store.go`, `internal/runledger/ent_store.go`, NEW `internal/storeutil/`
- **Tools**: `internal/agentmemory/tools.go`, `internal/cron/tools.go`, `internal/tools/exec/tools.go`, `internal/tools/payment/payment.go`, `internal/p2p/team/tools_escrow.go`
