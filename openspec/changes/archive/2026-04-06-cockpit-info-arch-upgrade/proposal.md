## Why

As features accumulated in the cockpit TUI through Phase 1-4, state became excessively concentrated in hub files. ChatModel has 23 struct fields and 21 msg cases, cockpit.go Update() has 8+ intercept blocks inline, and 7 utility functions are duplicated across 6 files. Changes to chat.go and cockpit.go are overly concentrated when adding new features, creating a structural problem that needs to be resolved.

## What Changes

- Extract CPR filter, pending indicator, approval state from ChatModel as independent types (23 → 18 fields)
- Remove package-level globals (`dialogScrollOffset`, `dialogSplitMode`), move to approvalState struct
- Consolidate 7 duplicate utility functions (truncate, formatDuration, formatNumber, relativeTime, etc.) into `internal/cli/tui/format.go` shared package
- Move sidebar menu item definitions from hardcoded `sidebar.New()` to centralized `AllPageMetas()` meta table in `router.go`
- Extract 8 inline intercept blocks in cockpit.go Update() into named methods
- Pure refactoring: no behavior changes

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
