## Context

lango는 `google.golang.org/adk v0.6.0`을 사용하여 ADK 기반 에이전트 런타임을 구동한다. 34개 Go 파일이 7개 ADK 서브패키지를 임포트하며, 핵심 통합 지점은 `internal/adk/` 패키지에 집중되어 있다 (ModelAdapter, SessionServiceAdapter, Agent wrapper, tool adapters).

v1.0.0이 GA로 릴리스되었으며, 모듈 캐시에서 실제 Go 소스를 diff한 결과, lango가 사용하는 모든 인터페이스가 소스 레벨에서 동일하거나 additive 변경(variadic 파라미터, 새 optional struct 필드)만 존재한다.

## Goals / Non-Goals

**Goals:**
- ADK 의존성을 v0.6.0에서 v1.0.0 GA로 업그레이드
- 전체 빌드, vet, 테스트 스위트 통과 유지
- MCP spike test의 타입 참조 수정

**Non-Goals:**
- v1.0.0 신규 기능 채택 (AutoCreateSession, HITL, workflow agents, RunOption 등) — 별도 change로 진행
- ADK adapter 리팩토링 또는 아키텍처 변경
- 프로덕션 코드 변경

## Decisions

### D1: go.mod 단일 범프 전략

go.mod에서 ADK 버전을 직접 변경하고 `go mod tidy`로 전이 의존성을 해결한다.

**근거**: 모든 공개 인터페이스가 소스 호환이므로, 점진적 마이그레이션이 불필요하다. 단일 변경으로 충분하며, 이는 가장 작은 diff를 생성하고 리뷰 부담을 최소화한다.

**대안**: 중간 버전(v0.7.0 등)을 거쳐 점진적 업그레이드 — 중간 버전이 존재하지 않으므로 불가.

### D2: MCP spike test 타입 참조 수정

`mcptoolset.ConfirmationProvider` → `tool.ConfirmationProvider`로 변경한다.

**근거**: v1.0.0에서 `ConfirmationProvider` 타입이 `tool/mcptoolset` 패키지에서 `tool` 패키지로 이동되었다. spike test에서만 사용하며 프로덕션 영향 없음.

### D3: 프로덕션 코드 무변경

프로덕션 코드(`internal/adk/*.go`, `internal/orchestration/`, `internal/a2a/` 등)는 변경하지 않는다.

**근거**: 실제 diff 검증 결과:
- `session.Service` — 동일
- `model.LLM` — 동일
- `runner.Runner.Run()` — variadic `opts ...RunOption` 추가 (기존 호출 그대로 동작)
- `runner.Config` — `AutoCreateSession` 필드 추가 (zero value = false)
- `agent.Agent` — 동일
- `tool.Tool` / `functiontool.Config` — 동일
- `plugin.Config` — 동일

## Risks / Trade-offs

- **전이 의존성 충돌** → `go mod tidy`가 자동 해결. `grpc v1.78.0→v1.79.3`, `a2a-go v0.3.3→v0.3.10` 등 마이너 범프만 발생.
- **ADK 내부 동작 변경** → golden test, session test, model adapter test가 커버. 전체 테스트 통과로 검증 완료.
- **MCP spike test 의미 변경** → spike test는 프로덕션과 무관. 타입 참조만 변경, 로직 동일.
