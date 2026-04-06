## Why

Phase 1-4에서 cockpit TUI에 기능이 누적되면서 hub 파일에 상태가 과도하게 집중되었다. ChatModel은 23개 struct field와 21개 msg case를 가지고, cockpit.go Update()에는 8+ intercept 블록이 인라인으로 존재하며, 7개 유틸리티 함수가 6개 파일에 중복되어 있다. 새 기능 추가 시 chat.go와 cockpit.go에 변경이 과도하게 몰리는 구조적 문제를 해소해야 한다.

## What Changes

- ChatModel에서 CPR filter, pending indicator, approval state를 독립 타입으로 추출 (23 → 18 fields)
- Package-level globals (`dialogScrollOffset`, `dialogSplitMode`) 제거, approvalState struct로 이동
- 중복 유틸리티 함수 7개 (truncate, formatDuration, formatNumber, relativeTime 등)를 `internal/cli/tui/format.go` 공유 패키지로 통합
- Sidebar 메뉴 아이템 정의를 `sidebar.New()` 하드코딩에서 `router.go` 중앙 메타 테이블 `AllPageMetas()`로 이동
- cockpit.go Update() 인라인 intercept 블록 8개를 명명된 메서드로 추출
- 순수 리팩토링: 동작 변경 없음

## Capabilities

### New Capabilities
- `tui-format-utils`: Shared formatting utilities package (`tui/format.go`) — Truncate, WordWrap, FormatNumber, FormatTokens, FormatDuration, RelativeTime, RelativeTimeHuman

### Modified Capabilities
- `tui-cpr-filter`: CPR filter implementation restructured — standalone cprFilter struct with Filter/Flush/HandleTimeout methods
- `cockpit-sidebar`: Sidebar items sourced from centralized AllPageMetas() instead of hardcoded in New()
- `cockpit-shell`: Update() intercept blocks extracted to named handler methods
- `interactive-tui-chat`: ChatModel struct reduced from 23 to 18 fields via sub-model extraction (cprFilter, pendingIndicator, approvalState)
- `tui-approval-tiers`: Approval dialog state moved from package globals to approvalState struct; renderApprovalDialog/handleApprovalDialogKey signatures updated

## Impact

- `internal/cli/chat/`: chat.go (-116 lines), new cpr_filter.go, pending.go, approval_state.go
- `internal/cli/tui/`: new format.go, format_test.go
- `internal/cli/cockpit/`: cockpit.go (router method extraction), router.go (AllPageMetas), new router_test.go
- `internal/cli/cockpit/sidebar/`: sidebar.go (New takes items parameter)
- `internal/cli/cockpit/pages/`: 5 page files updated to use tui.* shared formatters
- `internal/cli/cockpit/contextpanel.go`: 3 local formatters replaced with tui.* shared versions
- No API changes, no dependency changes, no user-visible behavior changes
