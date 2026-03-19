## Context

RunLedger Phase 1은 scaffold으로 구현됐으나, 코드 리뷰에서 5개 결함이 발견됐다:

1. PEV 자동 실행이 `run_propose_step_result`에 연결되지 않아 step이 `verify_pending`에서 영구 정지
2. `run_approve_step`이 validator type 검증 없이 아무 step이든 통과 가능
3. validator가 main tree에서 실행 — WorkDir 지원 미비로 workspace isolation이 무의미
4. `checkRole`이 execution tool에 orchestrator도 허용 — access control 무력화
5. CLI/TUI/README 등 downstream artifact 미반영

이미 코드 구현은 완료된 상태이고, 이 설계 문서는 사후 기록용이다.

## Goals / Non-Goals

**Goals:**
- PEV auto-verification: propose → journal → verify → completion check를 하나의 호출로 연결
- Strict access control: orchestrator ↔ execution agent 역할 분리 엄격 적용
- Validator WorkDir: Phase 1에서 필드/지원 준비, Phase 3에서 한 줄 활성화
- Run completion: step 검증 후 acceptance criteria 자동 확인 → run 상태 전이
- Downstream: CLI, status dashboard, README, docs, openspec 반영

**Non-Goals:**
- 실제 worktree 활성화 (Phase 3)
- DB 트랜잭션 래핑 (Phase 2 — Ent store 전환 시)
- TUI 전용 RunLedger surface (config-driven이 아닌 runtime feature이므로 불필요)

## Decisions

### 1. PEV 자동 실행 위치: `buildRunProposeStepResult` 내부

`run_propose_step_result` handler에서 journal 기록 직후 `pev.Verify()`를 호출한다. 별도 이벤트 핸들러나 비동기 처리 대신 동기 호출을 선택한 이유:
- Phase 1은 MemoryStore(순차 호출) — 동기가 가장 단순
- PEV 결과가 tool 응답에 즉시 포함되어야 agent가 다음 행동을 결정 가능
- Phase 2에서 Ent store로 전환 시 `tx := client.Tx()` 래핑만 추가하면 됨

대안: 이벤트 버스로 비동기 트리거 → 복잡도 증가, Phase 1에 불필요

### 2. 에러 처리 이원화: 인프라 vs 비즈니스

- 인프라 실패 (validator 미등록, exec 실패): `return nil, fmt.Errorf(...)` — non-nil Go error
- 비즈니스 실패 (validation not passed): structured map payload, nil error

이 구분이 필요한 이유: agent가 Go error를 받으면 tool 호출 자체가 실패한 것으로 간주하고, structured payload를 받으면 정상 흐름 안에서 정책 결정을 할 수 있다.

### 3. orchestrator_approval 흐름: PEV 자동 실행 → 항상 failed → approve

orchestrator_approval validator는 항상 failed를 반환한다. PEV가 자동 실행하면 step은 `failed`로 전이된다. 따라서 `run_approve_step`은 `verify_pending`뿐 아니라 `failed` 상태도 허용해야 한다.

대안: PEV가 orchestrator_approval을 감지하면 skip → approve 전까지 verify_pending 유지 → 특수 케이스 로직이 PEV에 침투하므로 거부

### 4. checkRunCompletion 공통 함수 추출

`run_propose_step_result`과 `run_approve_step` 모두 step 완료 후 run completion을 확인해야 한다. 동일한 로직:
- `AllStepsSuccessful()` → acceptance criteria 검증 → `EventCriterionMet` journaling → completed/failed
- `AllStepsTerminal()` but not successful → run failed
- 진행 중 → running

### 5. WorkDir: 필드 추가 + validator 지원, 활성화는 Phase 3

`ValidatorSpec.WorkDir` 필드를 추가하고 모든 command-running validator에서 사용한다. Phase 1에서 WorkDir는 항상 빈 문자열(기존 동작 유지). Phase 3에서 `pev.WithWorkspace(NewWorkspaceManager())`로 한 줄 활성화.

## Risks / Trade-offs

- [PEV 동기 호출 성능] → Phase 1은 in-memory store라 무시 가능. Phase 2에서 validator timeout 설정으로 대응.
- [checkRunCompletion 중복 criterion_met journaling] → 이미 Met인 criteria도 매번 journal에 기록됨. Phase 2에서 "이미 Met이면 skip" 조건 추가.
- [orchestrator_approval 2단계 전이 (verify_pending → failed → approved)] → journal에 validation_failed 이벤트가 기록된 뒤 validation_passed가 따라옴. 재생 시 최종 상태는 정확하지만 이벤트 로그가 다소 장황함. 수용 가능한 trade-off.
