## Why

RunLedger Phase 1 코드 리뷰에서 5개 결함이 발견됐다. PEV 자동 실행 미연결(step이 verify_pending에서 영구 정지), validator type 검증 없는 approve, main tree에서 validator 실행(workspace isolation 무의미), orchestrator가 execution tool 호출 가능, CLI/TUI/README 등 downstream 미반영. 핵심 불변 원칙을 깨는 문제들이라 즉시 수정이 필요하다.

## What Changes

- `run_propose_step_result`에서 journal 기록 후 PEV 자동 실행 연결 (propose → verify → completion check 한 번에 전이)
- `run_approve_step`에 validator type 검증 추가 (orchestrator_approval 타입만 허용, verify_pending/failed 상태만 허용)
- `ValidatorSpec`에 `WorkDir` 필드 추가, 모든 command-running validator에 `cmd.Dir` 설정
- `PEVEngine`에 `WorkspaceManager` 필드 + `WithWorkspace()` 메서드 + `PrepareStepWorkspace()` 호출
- `checkRole`에서 execution-only tool에 대한 orchestrator 허용 로직 제거
- `checkRunCompletion` 공통 함수 추출 (AllStepsSuccessful → acceptance criteria → criterion_met journaling → completed/failed)
- `EventCriterionMet` 새 이벤트 타입 추가
- CLI: `lango run list|status|journal` 서브커맨드 추가
- `lango status` 대시보드에 RunLedger feature 추가
- README, docs, openspec spec 업데이트

## Capabilities

### New Capabilities

(none — all changes modify the existing run-ledger capability)

### Modified Capabilities

- `run-ledger`: PEV auto-verification, WorkDir injection, strict access control, run completion logic, EventCriterionMet, CLI downstream

## Impact

- `internal/runledger/tools.go` — PEV 연결, checkRunCompletion, checkRole 강화
- `internal/runledger/pev.go` — workspace 필드 + Verify에서 PrepareStepWorkspace 호출
- `internal/runledger/validators.go` — cmd.Dir = spec.WorkDir
- `internal/runledger/types.go` — ValidatorSpec.WorkDir
- `internal/runledger/workspace.go` — PrepareStepWorkspace 함수
- `internal/runledger/journal.go` — EventCriterionMet + CriterionMetPayload
- `internal/runledger/snapshot.go` — AllStepsSuccessful + applyEvent criterion_met
- `internal/ent/schema/run_journal.go` — criterion_met enum
- `internal/cli/run/run.go` — 새 CLI 패키지
- `cmd/lango/main.go` — run 명령 등록
- `internal/cli/status/status.go` — RunLedger feature
- README.md, docs/features/run-ledger.md, docs/features/index.md
