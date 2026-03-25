## Context

멀티 에이전트 제어가 3곳에 분산 (orchestration prompt, agent.go budget, agent.go recovery). TUI가 기본 진입점이 된 이후(`847dcea`), TurnRunner 변경은 TUI/Gateway/Channel 전체에 영향. 현재 app에는 root `*adk.Agent` 하나만 존재하고, TurnRunner에 단일 executor로 주입됨.

기반 커밋: `847dcea` (dev), `app.New(boot, ...AppOption)` 패턴, `AppModeLocalChat`, `lifecycle.SetMaxPriority()` 도입됨.

## Goals / Non-Goals

**Goals:**
- 정책/관측 래퍼(`CoordinatingExecutor`)를 `turnrunner.Executor` 위에 올려 delegation 감시, 예산 미러링, 복구 결정을 코드로 분리
- turntrace 진단 인프라 확장 (typed events, delegation graph, metrics, retention)
- CLI 진단 표면 제공 (`lango agent trace/graph/trace metrics`)
- doctor health check 강화 (loop/timeout 빈도, trace 성장률)
- gateway WebSocket events (agent.delegation, agent.budget_warning)
- config로 정책 파라미터 외부화

**Non-Goals:**
- agent.go 리팩토링 (budget/recovery authoritative 승격은 v2)
- per-agent direct execution (root orchestrator만 경유)
- TaskQueue, Mailbox, Swarm, Pipeline 패턴
- EventBus async/priority 도입
- TUI statusbar에 delegation/budget 실시간 표시 (callback 인프라만 제공)
- prompt routing 축소 (v1에서는 root orchestrator LLM이 라우팅 소유 유지)

## Decisions

### D1: CoordinatingExecutor는 turnrunner.Executor를 구현

**선택:** `CoordinatingExecutor`가 `turnrunner.Executor` interface (`RunStreamingDetailed`)를 구현하는 래퍼.
**대안:** 자체 `Coordinate(sessionID, input) (string, error)` interface → TurnRunner가 chunk streaming, onEvent trace hook, idle timeout을 잃음.
**근거:** TurnRunner가 유일한 턴 경계. 기존 streaming/tracing 파이프라인을 깨지 않으려면 동일 interface를 구현해야 함.

### D2: DelegationGuard는 사후 관측기 (라우팅 소유 아님)

**선택:** `DelegationGuard`는 ADK event hook으로 delegation을 관측하고 circuit breaker 상태를 관리. 라우팅 결정 자체는 root orchestrator LLM이 소유.
**대안:** `StructuredRouter`가 사전에 agent를 선택 → per-agent direct execution 필요 (앱에 root agent 하나만 존재하므로 불가).
**근거:** v1에서 실현 가능한 범위. 정직한 이름(`DelegationGuard`)으로 역할 한정.

### D3: BudgetPolicy는 observational (authoritative 아님)

**선택:** inner executor(agent.go)의 hardcoded budget이 authoritative. BudgetPolicy는 event hook에서 turn/delegation을 미러링하고 threshold 알림만 발행.
**대안:** agent.go를 수정하여 budget 로직을 외부 정책에 위임 → v1 scope 초과, regression 위험.
**근거:** v1은 관측 계층. agent.go 수정은 v2.
**미러링 규칙:** inner budget과 동일 기준 — `hasFunctionCall(event)` && `!isDelegation(event)`만 turn으로 셈 (agent.go:350 참조). `RecordDelegation(target string)`으로 uniqueAgents 추적.

### D4: RecoveryPolicy는 실질적 제어권 보유

**선택:** inner executor 실패 시 `RecoveryPolicy.Decide()`가 재시도/힌트재시도/직접응답/에스컬레이션 결정. inner executor 재호출 여부를 외부에서 결정하므로 v1에서도 실질적 제어권.
**Actions:** `RecoveryRetry` (동일 입력 재시도), `RecoveryRetryWithHint` (root에 "다른 specialist 시도" 힌트 추가), `RecoveryDirectAnswer` (partial result 활용), `RecoveryEscalate` (에러 반환).
**`RecoveryRetryWithHint`는 reroute가 아님:** root orchestrator에 힌트를 추가한 입력으로 재시도하여 다른 선택을 유도.

### D5: turntrace Store 확장은 doctor 요구사항 포함

**선택:** `RecentByOutcome(ctx, outcome, since, limit)` 추가하여 doctor의 time-window + outcome-filter 조회 지원.
**대안:** doctor가 Ent client를 직접 사용 → Store interface 추상화 파괴.
**근거:** Store interface를 통한 일관된 접근.

### D6: Event hook 합성은 opts, onChunk 래핑 아님

**선택:** `adk.WithOnEvent()`로 policy hook을 opts에 합성. delegation event는 ADK event hook으로만 보임 (onChunk에서는 안 보임).
**근거:** TurnRunner의 traceRecorder도 `adk.WithOnEvent()`로 동작 (runner.go:227 참조). 동일 패턴 사용.

### D7: CoordinatingExecutor는 lifecycle component가 아님

**선택:** executor 래핑으로 주입 (`initAgentRuntime`이 반환한 executor를 TurnRunner에 전달). lifecycle.Registry에 등록하지 않음.
**근거:** lifecycle priority 제한(LocalChat의 `SetMaxPriority(PriorityBuffer)`)과 무관해야 함. RetentionCleaner만 lifecycle component로 등록.

## Risks / Trade-offs

- **[Risk] BudgetPolicy 미러링 오차** — event hook 타이밍과 inner budget 카운팅이 정확히 동기화되지 않을 수 있음 → **Mitigation:** inner budget과 동일 기준(hasFunctionCall && !isDelegation) 적용. 알림은 advisory, enforcement는 inner에 맡김.
- **[Risk] RecoveryRetryWithHint가 무한루프** — root orchestrator가 같은 specialist을 계속 선택 → **Mitigation:** maxRetries (default 2) 제한. 실패한 agent를 excludeAgents 힌트에 포함.
- **[Risk] runner.go를 Unit 1(event 상수)과 Unit 5(callback)가 동시 수정** → **Mitigation:** Phase 분리 (Unit 1 Phase 1, Unit 5 Phase 4). 수정 위치가 다름 (상수 교체 vs callback 추가).
- **[Risk] TUI에서 OnDelegation/OnBudgetWarning이 설정되지 않으면 누락** → **Mitigation:** callback은 optional (nil이면 no-op). TUI 표시는 후속 작업으로 명시.
- **[Trade-off] v1에서 budget enforcement 없음** — 관측만 제공, 실제 제어는 inner executor → v2에서 agent.go 리팩토링으로 해결.
