## 1. Data Layer — Task Extensions

- [x] 1.1 Extend TaskInfo with Result, Error, OriginChannel, TokensUsed fields
- [x] 1.2 Add TaskActioner interface (CancelTask, RetryTask)
- [x] 1.3 Implement bgTaskActioner adapter in main.go (Cancel delegates, Retry resubmits with original origin)
- [x] 1.4 Update bgTaskLister.ListTasks() to populate new TaskInfo fields from TaskSnapshot

## 2. Data Layer — Approval Extensions

- [x] 2.1 Create ApprovalHistoryStore (history.go) with ring buffer, Append, List, Count, CountByOutcome
- [x] 2.2 Add history_test.go (7 tests: append, list ordering, overflow, count, concurrent, default maxSize)
- [x] 2.3 Add GrantStore.List() returning []GrantInfo (expired excluded, sorted)
- [x] 2.4 Add grant_test.go tests for List (5 tests: empty, active, expired, revoked, sort order)

## 3. Wiring — History + Actioner

- [x] 3.1 Add ApprovalHistory field to App struct (types.go)
- [x] 3.2 Create HistoryStore in app.go and pass to WithApproval
- [x] 3.3 Extend WithApproval signature with history parameter
- [x] 3.4 Add history.Append at all 7 logApprovalEvent call sites (bypass, granted, denied, timeout, replay)
- [x] 3.5 Add spending limiter bypass recording (logApprovalEvent + history.Append)
- [x] 3.6 Fix middleware_test.go and policy_integration_test.go WithApproval calls
- [x] 3.7 Add ApprovalHistory and GrantStore to cockpit Deps
- [x] 3.8 Wire TaskActioner into NewTasksPage constructor
- [x] 3.9 Create bgTaskActioner instance in main.go and pass to NewTasksPage

## 4. UI — Task Detail + Actions

- [x] 4.1 Add detailMode, detailScroll, statusMsg state to TasksPage
- [x] 4.2 Implement Enter toggle for detail view with height clamp (tableMinHeight=6, detailMinHeight=8)
- [x] 4.3 Implement Esc to close detail, up/down scroll in detail mode
- [x] 4.4 Implement `c` key cancel action (Pending/Running only, via tea.Cmd)
- [x] 4.5 Implement `r` key retry action (Failed/Cancelled only, via tea.Cmd)
- [x] 4.6 Add taskActionResultMsg handling for status feedback
- [x] 4.7 Implement detail panel rendering (status, prompt, result, error, origin, tokens)
- [x] 4.8 Add ShortHelp context-aware bindings (enter, esc, c, r)
- [x] 4.9 Add tasks_test.go tests (17 tests: detail toggle, scroll, cancel, retry, nil actioner, feedback)

## 5. UI — Approvals Page

- [x] 5.1 Create ApprovalsPage struct with history/grants stores, dual-section layout
- [x] 5.2 Implement `/` key section toggle with independent cursors
- [x] 5.3 Implement history section rendering (Time, Tool, Summary, Outcome, Provider)
- [x] 5.4 Implement grants section rendering (Session, Tool, Granted)
- [x] 5.5 Implement `r` revoke grant and `R` revoke session actions
- [x] 5.6 Add PageApprovals to router.go, sidebar, keymap (ctrl+6), cockpit.go
- [x] 5.7 Register Approvals page in main.go with deps
- [x] 5.8 Add approvals_test.go tests (16 tests: display, tab, revoke, empty state, cursor)

## 6. UI — Approval Dialog Enhancement

- [x] 6.1 Add RuleExplanation field to ApprovalViewModel
- [x] 6.2 Implement buildRuleExplanation with 4 fixed templates
- [x] 6.3 Render "Why: ..." in approval_dialog.go between summary and params
- [x] 6.4 Add "(destructive)" label in approval_strip.go for critical risk
- [x] 6.5 Add approvalConfirmAction tracking to ChatModel (a/s action memory)
- [x] 6.6 Implement double-press guardrail for both `a` and `s` with action matching
- [x] 6.7 Add ApprovalRequestMsg intercept in cockpit.go with switchPage(PageChat)
- [x] 6.8 Add viewmodel_test.go tests (7 tests: rule explanation, ViewModel population)
- [x] 6.9 Add approval_dialog_test.go tests (3 tests: explanation, confirm prompt)
- [x] 6.10 Add approval_strip_test.go tests (3 tests: destructive label, confirm)
- [x] 6.11 Add chat_test.go tests (5 tests: critical double-press, session double-press, action tracking)
