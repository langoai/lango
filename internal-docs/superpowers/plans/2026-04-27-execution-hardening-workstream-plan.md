# Execution Hardening Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the highest-risk dead-letter/runtime wiring and execution holes without reopening broader roadmap scope.

**Architecture:** Keep the recently landed product/runtime behavior intact and patch only the narrow, high-risk holes: shell adapter forwarding, replay principal injection, background panic handling, reputation persistence safety, and classifier consistency. The workstream stays surgical and avoids unrelated refactors.

**Tech Stack:** Go, `cmd/lango`, `internal/cli/status`, `internal/cli/cockpit`, `internal/background`, `internal/p2p/reputation`, Zensical docs, OpenSpec

---

## File Map

### Worker A: Background / Reputation Hardening

- Modify: `internal/background/*`
  - Panic recovery and task lifecycle safety.
- Modify: `internal/p2p/reputation/*`
  - Per-peer serialization and score-clamping safety.
- Modify: focused tests adjacent to those packages.

### Worker B: Cockpit / CLI Wiring Hardening

- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`
- Modify: `internal/cli/status/*`
- Modify: `internal/cli/cockpit/*`
- Modify: any tiny helper needed for local principal fallback
- Modify: focused tests adjacent to those packages

### Worker C: Docs / OpenSpec / README

- Modify: `docs/cli/status.md`
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-27-dead-letter-runtime-stabilization/**`

## Task Breakdown

### Task 1: Add or Extend Red Tests for the Hardening Gaps

**Owner:** Worker A + Worker B

**Files:**
- Modify: focused tests under `cmd/lango`, `internal/background`, `internal/p2p/reputation`, `internal/cli/status`, `internal/cli/cockpit`

- [ ] **Step 1: Add wiring and retry-context tests**

Cover:

- cockpit dead-letter filter forwarding through `cmd/lango`
- retry invocation with an empty principal context using a stable fallback

- [ ] **Step 2: Add execution and persistence safety tests**

Cover:

- background runner panic does not orphan task state
- concurrent reputation updates on one peer preserve counts
- `NaN` score does not propagate through trust-entry logic

- [ ] **Step 3: Run focused tests and verify they fail**

Run the narrowest package-level test commands that cover the new assertions.

Expected:

```text
FAIL
```

### Task 2: Land Background / Reputation Hardening

**Owner:** Worker A

**Files:**
- Modify: `internal/background/*`
- Modify: `internal/p2p/reputation/*`
- Modify: focused tests from Task 1

- [ ] **Step 1: Implement background panic hardening**

Implementation rules:

- recover from runner panic
- fail the task explicitly
- avoid orphaned running state
- preserve current retry/dead-letter semantics unless intentionally documented otherwise

- [ ] **Step 2: Implement reputation persistence hardening**

Implementation rules:

- serialize updates per peer
- clamp invalid score values such as `NaN`
- avoid broad store refactors

- [ ] **Step 3: Re-run focused background / reputation tests and verify they pass**

Run the same focused test commands.

Expected:

```text
ok
```

### Task 3: Land Cockpit / CLI Wiring Hardening

**Owner:** Worker B

**Files:**
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`
- Modify: `internal/cli/status/*`
- Modify: `internal/cli/cockpit/*`
- Modify: tiny principal helper only if needed
- Modify: focused tests from Task 1

- [ ] **Step 1: Implement shell adapter forwarding fix**

Implementation rules:

- forward all dead-letter list options through the cockpit shell adapter
- keep bridge/page interfaces unchanged unless a tiny helper is cleaner

- [ ] **Step 2: Implement retry principal hardening**

Implementation rules:

- inject a stable local default principal when the runtime context is otherwise empty
- preserve explicit principals when already present
- keep operator-visible behavior unchanged except that replay no longer fails immediately on missing actor context

- [ ] **Step 3: Implement classifier consistency fix**

Implementation rules:

- cockpit and CLI should use the same dispatch-family classifier
- do not invent a second local classifier

- [ ] **Step 4: Re-run focused shell / retry / classifier tests and verify they pass**

Run the same focused test commands.

Expected:

```text
ok
```

### Task 4: Truth-Align Docs / OpenSpec

**Owner:** Worker C

**Files:**
- Modify: `docs/cli/status.md`
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-27-dead-letter-runtime-stabilization/**`

- [ ] **Step 1: Audit landed stabilization behavior before writing docs**

Verify from code:

- cockpit filter forwarding is actually fixed
- retry default principal injection is actually landed
- dispatch-family classifier is shared across surfaces

- [ ] **Step 2: Update public docs**

Document:

- stabilized cockpit shell adapter forwarding
- retry principal fallback behavior
- shared dispatch-family classification

- [ ] **Step 3: Sync docs-only OpenSpec**

Update `openspec/specs/docs-only/spec.md` so the requirements match the stabilized behavior.

- [ ] **Step 4: Archive the completed stabilization work**

Create:

- `openspec/changes/archive/2026-04-27-dead-letter-runtime-stabilization/proposal.md`
- `openspec/changes/archive/2026-04-27-dead-letter-runtime-stabilization/design.md`
- `openspec/changes/archive/2026-04-27-dead-letter-runtime-stabilization/tasks.md`
- `openspec/changes/archive/2026-04-27-dead-letter-runtime-stabilization/specs/docs-only/spec.md`

### Task 5: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review hardening changes**

Check:

- no new feature scope slipped in
- stabilization behavior is explicit and narrow
- CLI / cockpit / background / reputation semantics stay coherent

- [ ] **Step 2: Run full verification**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
openspec validate docs-only --type spec --strict --no-interactive
```

Expected:

```text
go build ./... exits 0
go test ./... exits 0
.venv/bin/zensical build exits 0
openspec validate docs-only --type spec --strict --no-interactive exits 0
```

- [ ] **Step 3: Commit the implementation**

Commit message:

```bash
git -c commit.gpgsign=false commit -m "fix: stabilize dead letter runtime wiring"
```
