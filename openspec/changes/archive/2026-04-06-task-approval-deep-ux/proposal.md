## Why

v1의 task strip과 approval dialog는 시작점이다. 실제 운영에는 task detail 조회, cancel/retry 액션, approval 이력 추적, grant 관리, destructive action guardrail이 필요하다. 현재는 task 목록이 read-only이고, approval 결과가 로그에만 남으며, grant를 TUI에서 관리할 수 없다.

## What Changes

- Task detail inline expansion: Enter 키로 full prompt/result/error 조회
- Task cancel/retry 액션: `c`=cancel, `r`=retry (원 prompt/origin 재사용)
- TaskInfo 확장: Result, Error, OriginChannel, TokensUsed 필드 추가
- TaskActioner 인터페이스: TUI에서 background manager 액션 호출
- ApprovalHistoryStore: in-memory ring buffer (500) for approval decision tracking
- Approval middleware history recording: 모든 결정 지점(bypass, granted, denied, timeout, spending limiter)에서 기록
- Approvals cockpit page: history + grants 통합 (섹션 전환, grant revoke)
- GrantStore.List(): 활성 grant 목록 조회 (expired 제외, sorted)
- Approval rule explanation: "Why: ..." 설명 (safety level + category 기반 고정 템플릿)
- Double-press guardrail: critical-risk에서 `a`/`s` 모두 이중 확인 (액션 추적)
- ApprovalRequestMsg cockpit intercept: 비-chat 페이지에서도 chat으로 자동 전환

## Capabilities

### New Capabilities
- `task-detail-actions`: Task detail inline expansion + cancel/retry 액션 in cockpit Tasks page
- `approval-history-view`: In-memory approval history store + Approvals cockpit page
- `approval-guardrails`: Destructive action double-press confirmation + rule explanation

### Modified Capabilities
- `tui-task-surface`: TaskInfo 확장 (Result, Error, OriginChannel, TokensUsed) + TaskActioner 인터페이스
- `tui-approval-tiers`: Rule explanation 필드 추가, double-press guardrail, confirm action 추적
- `persistent-approval-grant`: GrantStore.List() 메서드 + grant revoke UI

## Impact

- `internal/approval/`: history.go (NEW), grant.go (List), viewmodel.go (RuleExplanation)
- `internal/toolchain/mw_approval.go`: WithApproval 시그니처 변경 (history 파라미터 추가), 7곳 history.Append
- `internal/app/`: types.go (ApprovalHistory 필드), app.go (HistoryStore 생성/주입)
- `internal/cli/chat/`: approval_dialog.go (explanation + confirm), approval_strip.go (destructive hint), chat.go (double-press state)
- `internal/cli/cockpit/`: pages/approvals.go (NEW page), pages/tasks.go (detail + actions), router.go, keymap.go, cockpit.go (approval intercept)
- `cmd/lango/main.go`: TaskActioner, Approvals page 등록, deps 확장
- No DB schema changes. All stores in-memory.
