## Context

Phase 1-3에서 cockpit의 채널 통합과 runtime visibility를 완성했다. Task strip과 approval dialog는 v1 수준으로, read-only 목록과 단일 응답만 지원한다. DB 스키마 변경 없이 in-memory store만으로 task/approval 운영 경험을 개선한다.

## Goals / Non-Goals

**Goals:**
- Task를 상세 조회하고 cancel/retry로 관리할 수 있다
- Approval 이력을 세션 내에서 추적하고 조회할 수 있다
- Grant를 TUI에서 관리(목록/revoke)할 수 있다
- Destructive action에 이중 확인 guardrail이 적용된다

**Non-Goals:**
- Task/approval history DB 지속성 (Phase 5 scope)
- Persistent grants across restart
- Approval rule CRUD
- Transcript deep links to tasks

## Decisions

### 1. TaskActioner 인터페이스 + Manager.Submit 재호출

**Decision**: 새 `Retry` 메서드 추가 없이 기존 `Manager.Submit()`으로 같은 prompt + origin 재제출.

**Why**: Submit이 이미 모든 제출 로직(세마포어, origin 라우팅, 알림)을 처리. Retry는 원 task의 Prompt, OriginChannel, OriginSession을 그대로 재사용.

### 2. Inline detail expansion (overlay 아님)

**Decision**: Tasks page에서 Enter로 테이블 아래에 detail panel 표시. Overlay 시스템 도입 안 함.

**Why**: Page interface 계약 유지. 높이 clamp: tableMinHeight=6, detailMinHeight=8. 총 높이 <14이면 detail 우선.

### 3. ApprovalHistoryStore at middleware level

**Decision**: HistoryStore를 `WithApproval()` middleware에 주입하여 모든 결정 지점에서 기록.

**Why**: TUI provider뿐 아니라 session grant bypass, turn-local replay, spending limiter auto-approve까지 포괄. Middleware가 유일한 chokepoint.

### 4. History + Grants 통합 Approvals page

**Decision**: 두 개 별도 page 대신 하나의 Approvals page에 두 섹션 (history 상단, grants 하단).

**Why**: 이미 7개 page. 추가 난립 방지. `/` 키로 섹션 전환 (Tab은 cockpit 글로벌 sidebar toggle에 소비됨).

### 5. Double-press guardrail with action tracking

**Decision**: Critical-risk에서 `a`/`s` 모두 이중 확인. `approvalConfirmAction string` 필드로 어떤 액션이 pending인지 추적. 두 번째 press가 같은 키일 때만 해당 액션 실행.

**Why**: Generic `confirmPending` bool만으로는 `s` 첫 press → `a` 두 번째 press 시 잘못된 scope(session grant 대신 일회성)로 승인됨.

### 6. ApprovalRequestMsg 도착 시 자동 chat 전환

**Decision**: 비-chat 페이지에서 ApprovalRequestMsg 도착 시 `switchPage(PageChat)` 호출.

**Why**: Background task retry가 Tasks page에서 시작되면 approval이 보이지 않아 timeout. 자동 전환으로 사용자가 즉시 응답 가능.

## Risks / Trade-offs

- **[Risk] In-memory history loss**: 세션 종료 시 소실 → Phase 5에서 DB 지속성 추가
- **[Risk] HistoryStore ring buffer overflow**: 500개 cap → 장시간 세션에서 초기 이력 소실. 운영 관찰 후 조정
- **[Trade-off] `/` 키 섹션 전환**: Tab이 이상적이지만 cockpit 글로벌 키와 충돌. `/`는 검색 의미로도 쓰일 수 있어 향후 변경 가능
