## 1. CPR Filter Extraction (Unit 1A)

- [x] 1.1 Create `internal/cli/chat/cpr_filter.go` with cprFilter struct, cprState type/constants, Filter/Flush/HandleTimeout methods
- [x] 1.2 Move cprState type and 7 constants from chat.go to cpr_filter.go
- [x] 1.3 Move cprTimeoutMsg type and cprTimeout constant from chat.go to cpr_filter.go
- [x] 1.4 Replace ChatModel fields `cprDetect cprState` + `cprBuf []tea.KeyMsg` with `cpr cprFilter`
- [x] 1.5 Rewrite cprTimeoutMsg handler in Update() to use m.cpr.HandleTimeout() + replayKeys()
- [x] 1.6 Replace filterCPR call with m.cpr.Filter(msg) in Update() tea.KeyMsg case
- [x] 1.7 Rewrite cprFlush() as thin wrapper: m.cpr.Flush() + replayKeys()
- [x] 1.8 Remove original filterCPR and cprFlush method bodies (~80 lines)
- [x] 1.9 Update chat_test.go references from m.cprDetect → m.cpr.state, m.cprBuf → m.cpr.buf
- [x] 1.10 Verify go build/test/vet pass for chat package

## 2. Page Format Utilities Consolidation (Unit 1B)

- [x] 2.1 Create `internal/cli/tui/format.go` with Truncate, WordWrap, FormatNumber, FormatTokens, FormatDuration, RelativeTime, RelativeTimeHuman
- [x] 2.2 Create `internal/cli/tui/format_test.go` with table-driven tests for all 7 functions
- [x] 2.3 Replace `truncate` in pages/tools.go with tui.Truncate
- [x] 2.4 Replace `formatTokens` and `wordWrap` in pages/tasks.go with tui.FormatTokens and tui.WordWrap
- [x] 2.5 Replace `formatDuration` and `formatNumber` in pages/status.go with tui.FormatDuration and tui.FormatNumber
- [x] 2.6 Replace `relativeTime` in pages/approvals.go with tui.RelativeTime
- [x] 2.7 Replace `sessionsRelativeTime` in pages/sessions.go with tui.RelativeTimeHuman (preserving "just now" behavior)
- [x] 2.8 Replace `formatCompact`, `formatUptime`, `truncateName` in cockpit/contextpanel.go with tui.FormatNumber, tui.FormatDuration, tui.Truncate
- [x] 2.9 Update test files (tools_test, tasks_test, status_test, approvals_test, sessions_test, contextpanel_test) for tui.* calls
- [x] 2.10 Verify go build/test/vet pass for all affected packages

## 3. Pending Indicator Extraction (Unit 2A)

- [x] 3.1 Create `internal/cli/chat/pending.go` with pendingIndicator struct: Activate, Dismiss, IsActive, Elapsed, TickCmd methods
- [x] 3.2 Replace ChatModel fields `pendingStart time.Time` + `pendingActive bool` with `pending pendingIndicator`
- [x] 3.3 Update dismissPending() to delegate to m.pending.Dismiss()
- [x] 3.4 Update handleIdleKey() Enter case to use m.pending.Activate() + m.pending.TickCmd()
- [x] 3.5 Update RenderParts() and recalcLayout() to use m.pending.IsActive() and m.pending.Elapsed()
- [x] 3.6 Update PendingIndicatorTickMsg handler to use m.pending.IsActive()
- [x] 3.7 Verify go build/test/vet pass for chat package

## 4. Sidebar Meta Centralization (Unit 2B)

- [x] 4.1 Add AllPageMetas() function to cockpit/router.go returning 7 sidebar.MenuItem entries in current hardcoded order
- [x] 4.2 Change sidebar.New() signature to accept `items []MenuItem` parameter
- [x] 4.3 Update cockpit.New() to call sidebar.New(AllPageMetas())
- [x] 4.4 Create cockpit/router_test.go with PageID↔AllPageMetas↔PageIDFromString round-trip tests
- [x] 4.5 Add TestSidebarClick_UnregisteredPage_NoOp to cockpit_test.go
- [x] 4.6 Update sidebar_test.go to pass items to New()
- [x] 4.7 Verify go build/test/vet pass for cockpit packages

## 5. Approval Sub-Model Extraction (Unit 3A)

- [x] 5.1 Create `internal/cli/chat/approval_state.go` with approvalState struct: Reset, Clear, HasPending, IsConfirmExpired, StartConfirm, CancelConfirm, ScrollDiff, ToggleSplit methods
- [x] 5.2 Replace 4 ChatModel approval fields with `approval approvalState`
- [x] 5.3 Delete package globals `dialogScrollOffset` and `dialogSplitMode` from approval_dialog.go
- [x] 5.4 Update renderApprovalDialog signature to accept scrollOffset and splitMode parameters
- [x] 5.5 Update handleApprovalDialogKey to accept *approvalState parameter
- [x] 5.6 Update renderApproval dispatcher in approval.go to pass scroll/split params
- [x] 5.7 Update handleApprovingKey in chat.go for approval sub-model access
- [x] 5.8 Update RenderParts and recalcLayout for approval sub-model access
- [x] 5.9 Update approval_dialog_test.go to use approvalState instances instead of globals
- [x] 5.10 Update approval_origin_test.go for new renderApprovalDialog signature
- [x] 5.11 Update chat_test.go references from m.pendingApproval → m.approval.pending, m.approvalConfirmPending → m.approval.confirmPending
- [x] 5.12 Verify go build/test/vet pass for chat package

## 6. Cockpit Router Method Extraction (Unit 3B)

- [x] 6.1 Extract handleContextTick method from Update() inline block
- [x] 6.2 Extract handleChannelMessage method from Update() inline block
- [x] 6.3 Extract handleApprovalRequest method from Update() inline block
- [x] 6.4 Extract markTurnStarted method from Update() inline block
- [x] 6.5 Extract handleDelegation method from Update() inline block
- [x] 6.6 Extract handleBudgetWarning method from Update() inline block
- [x] 6.7 Extract handleRecovery method from Update() inline block
- [x] 6.8 Extract handleDone method from Update() inline block
- [x] 6.9 Rewrite Update() as clean type-switch dispatch table (~30 lines)
- [x] 6.10 Verify go build/test/vet pass for cockpit package

## 7. Final Verification

- [x] 7.1 Run `go build ./...` — full project build passes
- [x] 7.2 Run `go test ./...` — full project test suite passes (0 failures)
- [x] 7.3 Run `go vet ./...` — no warnings
- [x] 7.4 Verify ChatModel struct has exactly 18 fields
- [x] 7.5 Verify no package-level globals dialogScrollOffset/dialogSplitMode exist
- [x] 7.6 Codex review passes with no regressions found
