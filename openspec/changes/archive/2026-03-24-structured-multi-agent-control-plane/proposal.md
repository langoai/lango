## Why

Lango의 멀티 에이전트 제어가 3곳에 분산되어 있어 디버깅, 정책 변경, 관측이 어려움:
- 라우팅/판단 프로토콜이 orchestrator prompt 산문에 묻혀 있음 (`orchestration/tools.go:606`)
- 예산 확장이 agent.go 내부의 ad-hoc 로직으로 하드코딩됨 (`adk/agent.go:270-328`)
- 복구 전략이 RunAndCollect 안에 inline으로 산재함 (`adk/agent.go:473-593`)

TUI가 기본 진입점이 된 이후(`847dcea`), TurnRunner 변화는 곧 TUI/Gateway/Channel 전체에 영향을 주므로 정책을 코드로 분리해야 할 시급성이 높아졌다.

## What Changes

- `internal/agentrt/` 신규 패키지 도입: `CoordinatingExecutor` (turnrunner.Executor 래퍼), `DelegationGuard` (circuit breaker), `BudgetPolicy` (observational 예산 미러링), `RecoveryPolicy` (재시도/재라우팅/직접응답/에스컬레이션 결정)
- `internal/turntrace/` 확장: typed event constants, Store interface 확장 (EventsForTrace, TracesForSession, RecentByOutcome, PurgeTraces, TraceCount, OldTraces), delegation graph 연산, agent metrics 연산, retention cleaner
- `internal/turnrunner/` 확장: TurnRunner Request에 OnDelegation/OnBudgetWarning callback 추가
- `internal/config/` 확장: OrchestrationConfig, BudgetCfg, RecoveryCfg, CircuitBreakerCfg, TraceStoreConfig
- `internal/cli/agent/` 확장: `lango agent trace list/trace <id>/graph/trace metrics` CLI 명령어
- `internal/cli/doctor/checks/multi_agent.go` 확장: loop 빈도, timeout 빈도, trace 성장률, 평균 턴 시간 체크
- `internal/gateway/server.go` 확장: agent.delegation, agent.budget_warning WebSocket events
- `internal/app/wiring_agentrt.go` 신규: structured mode 배선, RetentionCleaner lifecycle 등록

v1에서 **하지 않는 것**: agent.go 리팩토링 (budget/recovery authoritative 승격은 v2), per-agent direct execution, TaskQueue, Mailbox, Swarm, EventBus async

## Capabilities

### New Capabilities
- `agent-control-plane`: CoordinatingExecutor (turnrunner.Executor 래퍼) + DelegationGuard + BudgetPolicy + RecoveryPolicy — 멀티 에이전트 정책/관측 제어면
- `turntrace-diagnostics`: typed event constants, delegation graph, agent metrics, retention cleaner, Store 확장 — 턴 진단 인프라
- `agent-cli-diagnostics`: `lango agent trace/graph/trace metrics` CLI 명령어 — 운영자 진단 표면

### Modified Capabilities
- `agent-turn-tracing`: Store interface에 EventsForTrace, TracesForSession, RecentByOutcome, PurgeTraces, TraceCount, OldTraces 추가
- `agent-error-handling`: RecoveryPolicy가 기존 inline recovery 패턴을 코드 정책으로 포착 (agent.go 수정 없이 래퍼에서 적용)
- `multi-agent-orchestration`: DelegationGuard가 orchestrator의 delegation을 사후 관측, circuit breaker 상태 관리
- `agent-runtime`: CoordinatingExecutor가 turnrunner.Executor로 주입되어 기존 실행 경로를 래핑

## Impact

- **Code**: `internal/agentrt/` (신규), `internal/turntrace/`, `internal/turnrunner/`, `internal/config/`, `internal/cli/agent/`, `internal/cli/doctor/checks/`, `internal/gateway/`, `internal/app/`
- **APIs**: TurnRunner.Request에 OnDelegation/OnBudgetWarning callback 추가 (하위호환 — optional fields)
- **Config**: `agent.orchestration.mode` ("classic"|"structured"), `agent.orchestration.budget.*`, `agent.orchestration.recovery.*`, `agent.orchestration.circuitBreaker.*`, `observability.traceStore.*`
- **Dependencies**: 새 외부 의존성 없음 (stdlib + 기존 internal 패키지만 사용)
- **TUI**: `847dcea` 이후 TUI도 TurnRunner를 사용하므로 structured mode에서 정책 자동 적용. TUI statusbar 표시는 후속 작업
