## Why

RunLedger는 아직 `MemoryStore` 중심의 Phase 1 scaffold/hardening 단계다. Task OS가 실제
정본이 되려면 journal/snapshot/step projection이 프로세스 생명주기를 넘어 살아남아야 하고,
workflow/background write path도 ledger를 먼저 거쳐야 한다. 이 단계 없이는 RunLedger가
"library"를 넘어 "runtime authority"가 될 수 없다.

## What Changes

- RunLedger에 Ent-backed persistent store를 도입한다.
- `run_id` 정본을 RunLedger가 단일 생성하고, workflow/background projection이 같은 ID를
  사용하도록 write-through adapter를 추가한다.
- projection drift 감지와 replay/rebuild 경로를 추가한다.
- `lango run journal`이 persistent store를 읽을 수 있도록 CLI를 확장한다.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `run-ledger`: persistent store, canonical run ID ownership, write-through projections,
  and projection rebuild/drift handling

## Impact

- `internal/runledger/store.go`
- `internal/runledger/writethrough.go`
- `internal/runledger/journal.go`
- `internal/runledger/snapshot.go`
- `internal/background/manager.go`
- `internal/workflow/state.go`
- `internal/cli/run/run.go`
- `openspec/specs/run-ledger/spec.md`
