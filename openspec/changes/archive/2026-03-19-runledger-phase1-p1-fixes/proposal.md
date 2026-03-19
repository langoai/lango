## Why

RunLedger Phase 1은 기본 journal/snapshot/PEV 뼈대까지는 들어왔지만, 실제 실행 권한과
workspace lifecycle에서 P1급 구멍이 남아 있다. 특히 step ownership 검증 없는 proposal
journaling과 retry-safe 하지 않은 worktree branch naming은 Task OS의 핵심 불변 조건을
직접 깬다. 이 변경은 현재 Phase 1을 안전한 hardening 상태까지 끌어올리고, 이후
Phase 2~4 OpenSpec change들이 올라갈 수 있는 안정된 바닥을 만든다.

## What Changes

- `run_propose_step_result`가 journal append 전에 step 존재, owner agent, 허용 상태를
  검증하도록 수정한다.
- workspace preparation이 재시도와 반복 검증에서 깨지지 않도록 branch/path lifecycle을
  retry-safe 하게 바꾼다.
- RunLedger module이 Phase 1에서는 workspace isolation을 의도적으로 비활성 상태로
  둔다는 점을 코드와 문서에서 명시한다.
- RunLedger README/docs/OpenSpec를 현재 실제 동작과 일치하도록 정리한다.
- 후속 구현이 바로 가능하도록 Phase 2~4를 별도 OpenSpec change로 계획한다.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `run-ledger`: step proposal authorization, retry-safe workspace lifecycle, and
  explicit phase-gated workspace activation semantics

## Impact

- `internal/runledger/tools.go`
- `internal/runledger/workspace.go`
- `internal/app/modules_runledger.go`
- `internal/runledger/tools_test.go`
- `internal/runledger/workspace_test.go` (new)
- `README.md`
- `docs/features/run-ledger.md`
- `openspec/specs/run-ledger/spec.md`
