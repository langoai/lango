## Context

In Phase 1-3, the cockpit's channel integration and runtime visibility were completed. Task strip and approval dialog are at v1 level, supporting only read-only lists and single responses. This improves the task/approval operational experience using only in-memory stores without DB schema changes.

## Goals / Non-Goals

**Goals:**
- Tasks can be viewed in detail and managed with cancel/retry
- Approval history can be tracked and viewed within a session
- Grants can be managed (list/revoke) from the TUI
- Double confirmation guardrails are applied to destructive actions

**Non-Goals:**
- Task/approval history DB persistence (Phase 5 scope)
- Persistent grants across restart
- Approval rule CRUD
- Transcript deep links to tasks

## Decisions

### 1. TaskActioner interface + Manager.Submit re-invocation

**Decision**: No new `Retry` method added; resubmit using existing `Manager.Submit()` with the same prompt + origin.

**Why**: Submit already handles all submission logic (semaphore, origin routing, notifications). Retry reuses the original task's Prompt, OriginChannel, OriginSession as-is.

### 2. Inline detail expansion (not overlay)

**Decision**: On the Tasks page, pressing Enter shows a detail panel below the table. No overlay system introduced.

**Why**: Maintains Page interface contract. Height clamped: tableMinHeight=6, detailMinHeight=8. If total height <14, detail takes priority.

### 3. ApprovalHistoryStore at middleware level

**Decision**: HistoryStore is injected into `WithApproval()` middleware to record at every decision point.

**Why**: Covers not just the TUI provider but also session grant bypass, turn-local replay, and spending limiter auto-approve. Middleware is the sole chokepoint.

### 4. History + Grants unified Approvals page

**Decision**: Instead of two separate pages, a single Approvals page with two sections (history top, grants bottom).

**Why**: Already 7 pages. Prevent additional proliferation. `/` key for section switching (Tab is consumed by cockpit global sidebar toggle).

### 5. Double-press guardrail with action tracking

**Decision**: At critical-risk, both `a`/`s` require double confirmation. `approvalConfirmAction string` field tracks which action is pending. Only executes the action when the second press matches the same key.

**Why**: With a generic `confirmPending` bool only, pressing `s` first → `a` second would approve with the wrong scope (session grant instead of one-time).

### 6. Auto-switch to chat on ApprovalRequestMsg arrival

**Decision**: When ApprovalRequestMsg arrives on a non-chat page, call `switchPage(PageChat)`.

**Why**: If background task retry is initiated from the Tasks page, the approval is not visible, causing timeout. Auto-switching allows the user to respond immediately.

## Risks / Trade-offs

- **[Risk] In-memory history loss**: Lost on session termination → DB persistence added in Phase 5
- **[Risk] HistoryStore ring buffer overflow**: 500 cap → early history lost in long sessions. Adjust after operational observation
- **[Trade-off] `/` key section switching**: Tab would be ideal but conflicts with cockpit global key. `/` could also be used for search, so may change in the future
