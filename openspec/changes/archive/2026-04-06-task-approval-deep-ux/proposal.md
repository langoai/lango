## Why

The v1 task strip and approval dialog are just the starting point. Actual operations require task detail viewing, cancel/retry actions, approval history tracking, grant management, and destructive action guardrails. Currently, the task list is read-only, approval results only appear in logs, and grants cannot be managed from the TUI.

## What Changes

- Task detail inline expansion: Enter key to view full prompt/result/error
- Task cancel/retry actions: `c`=cancel, `r`=retry (reusing original prompt/origin)
- TaskInfo extension: Result, Error, OriginChannel, TokensUsed fields added
- TaskActioner interface: Call background manager actions from TUI
- ApprovalHistoryStore: in-memory ring buffer (500) for approval decision tracking
- Approval middleware history recording: Record at every decision point (bypass, granted, denied, timeout, spending limiter)
- Approvals cockpit page: history + grants unified (section switching, grant revoke)
- GrantStore.List(): Active grant list query (excluding expired, sorted)
- Approval rule explanation: "Why: ..." explanation (fixed template based on safety level + category)
- Double-press guardrail: Double confirmation for both `a`/`s` at critical-risk (action tracking)
- ApprovalRequestMsg cockpit intercept: Auto-switch to chat even from non-chat pages

## Capabilities

### New Capabilities
- `task-detail-actions`: Task detail inline expansion + cancel/retry actions in cockpit Tasks page
- `approval-history-view`: In-memory approval history store + Approvals cockpit page
- `approval-guardrails`: Destructive action double-press confirmation + rule explanation

### Modified Capabilities
- `tui-task-surface`: TaskInfo extension (Result, Error, OriginChannel, TokensUsed) + TaskActioner interface
- `tui-approval-tiers`: Rule explanation field added, double-press guardrail, confirm action tracking
- `persistent-approval-grant`: GrantStore.List() method + grant revoke UI

## Impact

- `internal/approval/`: history.go (NEW), grant.go (List), viewmodel.go (RuleExplanation)
- `internal/toolchain/mw_approval.go`: WithApproval signature change (history parameter added), 7 history.Append call sites
- `internal/app/`: types.go (ApprovalHistory field), app.go (HistoryStore creation/injection)
- `internal/cli/chat/`: approval_dialog.go (explanation + confirm), approval_strip.go (destructive hint), chat.go (double-press state)
- `internal/cli/cockpit/`: pages/approvals.go (NEW page), pages/tasks.go (detail + actions), router.go, keymap.go, cockpit.go (approval intercept)
- `cmd/lango/main.go`: TaskActioner, Approvals page registration, deps extension
- No DB schema changes. All stores in-memory.
