# Cockpit Loading / Failure Recovery Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the landed cockpit retry UX so operators can clearly see when a retry is running and get better immediate failure feedback, while keeping the existing replay backend path unchanged.

**Architecture:** Keep the existing cockpit dead-letter page, filter bar, retry action, confirm flow, and replay bridge. Extend the page-local retry state machine with an explicit `running` state, show that state in the detail pane, guard replay-trigger input while running, and surface the backend error string through the existing status-message path. Keep success-refresh semantics unchanged and leave richer failure panels or action history out of scope.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add retry running-state rendering and input guarding.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover running-state rendering, duplicate-trigger guard, and failure reset behavior.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe loading/failure feedback as landed.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark loading/failure feedback as landed and narrow the remaining operator feedback work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-cockpit-loading-failure-recovery-feedback/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Add Running-State Feedback to the Cockpit Retry Action

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- detail-pane `Retry action` enters `running...` state while retry is in flight
- duplicate retry/confirm triggers are blocked while running
- short help stays consistent with running-state semantics

- [ ] **Step 2: Run the focused cockpit-page tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement running-state feedback**

Extend the page so that:

- add a retry action state machine with at least:
  - idle
  - confirm
  - running
- render `Retry action: running...` in the detail pane while in flight
- block replay-trigger input while running
- keep the existing confirm flow and success refresh semantics

- [ ] **Step 4: Re-run the focused cockpit-page tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the running-state slice**

Run:

```bash
git add internal/cli/cockpit/pages/deadletters.go internal/cli/cockpit/pages/deadletters_test.go
git -c commit.gpgsign=false commit -m "feat: add cockpit retry loading state"
```

### Task 2: Improve Failure Feedback While Preserving Current Data

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- retry failure returns to enabled idle state
- failure error string is shown through the status message path
- failure does not auto-refresh or clear current backlog/detail data

- [ ] **Step 2: Run the focused cockpit-page tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement failure feedback semantics**

Update the page so that:

- failure clears the running state
- failure returns the action to idle/retryable state
- failure keeps the current backlog/detail in place
- failure surfaces the backend error string through the existing status message

- [ ] **Step 4: Re-run the focused cockpit-page tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the failure-feedback slice**

Run:

```bash
git add internal/cli/cockpit/pages/deadletters.go internal/cli/cockpit/pages/deadletters_test.go
git -c commit.gpgsign=false commit -m "feat: improve cockpit retry failure feedback"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-cockpit-loading-failure-recovery-feedback/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- cockpit retry running-state feedback
- blocked duplicate retry input while running
- failure string surfacing
- post-failure return to idle state

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks loading/failure feedback as landed and narrows the remaining operator feedback work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed loading/failure-feedback slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-cockpit-loading-failure-recovery-feedback/proposal.md`
- `openspec/changes/archive/2026-04-25-cockpit-loading-failure-recovery-feedback/design.md`
- `openspec/changes/archive/2026-04-25-cockpit-loading-failure-recovery-feedback/tasks.md`
- `openspec/changes/archive/2026-04-25-cockpit-loading-failure-recovery-feedback/specs/docs-only/spec.md`

- [ ] **Step 5: Run full verification**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
```

Expected:

```text
ok
Build finished
```

- [ ] **Step 6: Commit the docs/OpenSpec slice**

Run:

```bash
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-cockpit-loading-failure-recovery-feedback
git -c commit.gpgsign=false commit -m "specs: archive cockpit loading failure recovery feedback"
```

## Self-Review

- Spec coverage:
  - loading model: Task 1
  - failure feedback model: Task 2
  - post-failure reset semantics: Task 2
  - interaction guarding: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no success banner work
  - no structured failure panel
  - no action history
