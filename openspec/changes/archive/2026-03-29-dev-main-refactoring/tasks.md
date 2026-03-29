## 1. TUI Chat Hot Path (P0)

- [x] 1.1 Cache glamour markdown renderer with width-keyed module-level cache (`internal/cli/chat/markdown.go`)
- [x] 1.2 Skip redundant markdown re-render on cursor tick, add prevWidth guard in setSize (`internal/cli/chat/chatview.go`)

## 2. RunLedger Performance (P1)

- [x] 2.1 Add lazy step index to RunSnapshot.FindStep with ensureStepIndex/invalidateStepIndex (`internal/runledger/snapshot.go`)
- [x] 2.2 Add 4 FindStep test cases + BenchmarkFindStep (`internal/runledger/snapshot_test.go`, `snapshot_bench_test.go`)
- [x] 2.3 Add SourceKind + SourceDescriptor to RunCreatedPayload and RunSnapshot (`internal/runledger/journal.go`, `snapshot.go`)
- [x] 2.4 Persist SourceKind/SourceDescriptor in workflow and background writethrough paths (`internal/runledger/writethrough.go`)
- [x] 2.5 Add SourceDescriptor deep copy and backing array independence test (`internal/runledger/snapshot.go`, `snapshot_test.go`)

## 3. Store Util Helpers (P1)

- [x] 3.1 Create `internal/storeutil/` package with MarshalField (fail-fast), UnmarshalField, CopySlice, CopyMap
- [x] 3.2 Add storeutil tests including marshal failure case (`internal/storeutil/json_test.go`, `copy_test.go`)
- [x] 3.3 Apply storeutil to `internal/provenance/ent_store.go` (metadata marshal/unmarshal)
- [x] 3.4 Apply storeutil to `internal/runledger/ent_store.go` (snapshot, evidence, validator marshal/unmarshal)

## 4. Toolparam Migration (P2)

- [x] 4.1 Add RequireFloat64 and OptionalFloat64 to `internal/toolparam/extract.go`
- [x] 4.2 Migrate `internal/agentmemory/tools.go` to toolparam helpers
- [x] 4.3 Migrate `internal/runledger/tools.go` to toolparam helpers
- [x] 4.4 Migrate `internal/cron/tools.go`, `internal/tools/exec/tools.go` to toolparam helpers
- [x] 4.5 Migrate `internal/tools/payment/payment.go`, `internal/p2p/team/tools_escrow.go` to toolparam helpers

## 5. Event Name Constants (P2)

- [x] 5.1 Extract 48 event name constants across 6 eventbus files (`events.go`, `team_events.go`, `retrieval_events.go`, `economy_events.go`, `observability_events.go`, `workspace_events.go`)

## 6. Code Quality (P2)

- [x] 6.1 Clarify Finding.SearchSource vs Source doc comments (`internal/retrieval/finding.go`)
- [x] 6.2 Create `internal/cli/tuicore/field_builder.go` with BoolInput, IntInput, SelectInput, TextInput, TextInputWithPlaceholder, PasswordInput, SearchSelectInput factories
- [x] 6.3 Migrate `internal/cli/settings/forms_knowledge.go` to use field builders
- [x] 6.4 Migrate `internal/cli/settings/forms_p2p.go` to use field builders

## 7. Optimizations (P3)

- [x] 7.1 Add fast-path early return in ReallocateBudgets when no empty sections (`internal/adk/budget.go`)
- [x] 7.2 Replace mutex merge with lock-free index-based merge in RetrievalCoordinator (`internal/retrieval/coordinator.go`)
- [x] 7.3 Add tool stats dirty-flag sort guard with content-change detection in contextpanel (`internal/cli/cockpit/contextpanel.go`)

## 8. Test Fix

- [x] 8.1 Add ListSessions to mockStore and uniqueMockStore in `internal/adk/state_test.go`
