# Replay / Recovery Policy Runtime Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stabilize replay and recovery runtime policy by landing policy-driven defaults, normalizing async retry / dead-letter policy, and aligning replay with the recovery substrate.

**Architecture:** Keep adjudication and receipts canonical, then normalize the runtime policy layer that sits underneath `auto_execute`, `background_execute`, retry / dead-letter handling, and operator replay. The workstream is runtime-core first: Worker A settles the runtime contract, Worker B aligns downstream CLI/cockpit/meta-tool impact, and Worker C truth-aligns docs/OpenSpec.

**Tech Stack:** Go, `internal/postadjudicationreplay`, `internal/background`, `internal/app/tools_meta*.go`, `internal/receipts`, Cobra/Bubble Tea downstream alignment, Zensical docs, OpenSpec

---

## File Map

### Worker A: Runtime Policy / Retry Core

- Modify: `internal/postadjudicationreplay/*`
  - Normalize replay policy handling and align it with runtime recovery policy.
- Modify: `internal/background/*`
  - Clarify retry scheduling, retry limits, and dead-letter transition behavior as runtime policy.
- Modify: `internal/receipts/*` only where needed
  - Preserve canonical evidence / state semantics required by the runtime policy layer.
- Modify: focused runtime tests adjacent to the above packages.

### Worker B: Downstream Runtime Integration

- Modify: `internal/app/tools_meta*.go`
  - Align meta-tool runtime defaults and behavior with the normalized runtime policy contract.
- Modify: `internal/cli/status/*` and `internal/cli/cockpit/*` only where downstream wording or runtime-facing behavior must reflect the landed policy semantics.
- Modify: integration tests that exercise runtime-facing CLI/meta-tool behavior.

### Worker C: Docs / OpenSpec / README

- Modify: `docs/architecture/*`
  - Update runtime and replay/recovery architecture pages to match landed behavior.
- Modify: `docs/cli/*` and `README.md` only if surfaced runtime-facing behavior changes.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync public docs requirements.
- Create: `openspec/changes/archive/2026-04-26-replay-recovery-policy-runtime-workstream/**`
  - Proposal, design, tasks, and docs-only delta spec.

## Task Breakdown

### Task 1: Map and Test the Current Runtime Contracts

**Owner:** Worker A

**Files:**
- Modify: focused tests under `internal/postadjudicationreplay/*`, `internal/background/*`, and related runtime packages

- [ ] **Step 1: Add or extend failing runtime tests**

Cover the current gaps explicitly:

- how execution mode is chosen when `auto_execute` / `background_execute` flags are absent
- how retry scheduling and dead-letter transitions behave today
- how replay policy and retry / dead-letter policy differ today

Focus on behavior that must become explicit in the normalized runtime contract.

- [ ] **Step 2: Run focused runtime tests and verify they fail**

Run the narrowest package-level test commands that cover the new runtime-contract assertions.

Expected:

```text
FAIL
```

### Task 2: Land Policy-Driven Defaults

**Owner:** Worker A

**Files:**
- Modify: runtime-core files under `internal/background/*`, `internal/postadjudicationreplay/*`, and any minimal supporting files actually required
- Modify: runtime tests introduced in Task 1

- [ ] **Step 1: Implement policy-driven default selection**

Implementation rules:

- preserve current explicit surface flags and controls
- define coherent runtime defaults when explicit flags are absent
- keep adjudication as the canonical write layer
- avoid redesigning operator UX in this task

- [ ] **Step 2: Re-run focused runtime tests and verify they pass**

Run the same focused runtime test commands from Task 1 plus any direct coverage added for default selection.

Expected:

```text
ok
```

### Task 3: Normalize Async Retry / Dead-Letter Policy

**Owner:** Worker A

**Files:**
- Modify: `internal/background/*`
- Modify: related tests
- Modify: `internal/receipts/*` only if needed to preserve canonical evidence / state semantics

- [ ] **Step 1: Add or extend failing retry-policy tests**

Cover:

- retry scheduling behavior
- retry-cap handling
- dead-letter transition behavior
- policy behavior that is still too path-specific

- [ ] **Step 2: Run focused retry-policy tests and verify they fail**

Run the narrowest package-level test commands for the affected retry/dead-letter packages.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the normalized retry / dead-letter policy**

Implementation rules:

- preserve currently landed retry / dead-letter semantics at the user-visible level
- make the runtime policy units easier to reason about
- do not attempt a full generic background-manager-wide redesign in one step

- [ ] **Step 4: Re-run focused retry-policy tests and verify they pass**

Run the same focused retry-policy commands.

Expected:

```text
ok
```

### Task 4: Align Replay / Recovery Substrate Behavior

**Owner:** Worker A

**Files:**
- Modify: `internal/postadjudicationreplay/*`
- Modify: related runtime tests

- [ ] **Step 1: Add or extend failing replay-substrate tests**

Cover:

- canonical adjudication + dead-letter evidence gate preservation
- replay policy enforcement preservation
- alignment between replay-request semantics and the runtime recovery substrate
- removal of avoidable duplication between replay-specific and retry-specific policy handling

- [ ] **Step 2: Run focused replay-substrate tests and verify they fail**

Run the narrowest package-level test commands for replay runtime packages.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement replay / recovery substrate normalization**

Implementation rules:

- keep replay semantics operator-compatible
- reduce policy duplication where practical
- do not redesign the operator surfaces here

- [ ] **Step 4: Re-run focused replay-substrate tests and verify they pass**

Run the same focused replay test commands.

Expected:

```text
ok
```

### Task 5: Align Downstream Runtime Integration

**Owner:** Worker B

**Files:**
- Modify: `internal/app/tools_meta*.go`
- Modify: `internal/cli/status/*` and `internal/cli/cockpit/*` only if runtime-facing wording / behavior must reflect the normalized contract
- Modify: downstream integration tests

- [ ] **Step 1: Add or extend failing downstream integration tests**

Cover:

- meta-tool behavior reflecting runtime defaults correctly
- downstream CLI/cockpit/runtime-facing behavior staying coherent with the normalized contract
- no regression in replay / recovery invocation semantics

- [ ] **Step 2: Run focused downstream tests and verify they fail**

Run focused package-level tests for the touched downstream integration points.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement downstream alignment**

Implementation rules:

- Worker A’s runtime contract is the source of truth
- keep downstream changes minimal and contract-aligned
- avoid introducing a second runtime interpretation in the operator layer

- [ ] **Step 4: Re-run focused downstream tests and verify they pass**

Run the same focused downstream test commands.

Expected:

```text
ok
```

### Task 6: Truth-Align Docs / OpenSpec

**Owner:** Worker C

**Files:**
- Modify: verified runtime-facing docs under `docs/architecture/*`
- Modify: `docs/cli/*` and `README.md` only if needed
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-26-replay-recovery-policy-runtime-workstream/**`

- [ ] **Step 1: Audit landed runtime behavior before writing docs**

Verify from code:

- policy-driven defaults are actually defined
- retry / dead-letter policy normalization is actually landed
- replay / recovery substrate alignment is actually landed
- downstream wording changes only describe real behavior

- [ ] **Step 2: Update public docs**

Document:

- the landed runtime default behavior
- the normalized retry / dead-letter policy shape
- the replay / recovery substrate alignment that is now real

- [ ] **Step 3: Sync main OpenSpec requirements**

Update `openspec/specs/docs-only/spec.md` so the docs-only requirements reflect the landed runtime work and narrow the remaining backlog accordingly.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-26-replay-recovery-policy-runtime-workstream/proposal.md`
- `openspec/changes/archive/2026-04-26-replay-recovery-policy-runtime-workstream/design.md`
- `openspec/changes/archive/2026-04-26-replay-recovery-policy-runtime-workstream/tasks.md`
- `openspec/changes/archive/2026-04-26-replay-recovery-policy-runtime-workstream/specs/docs-only/spec.md`

### Task 7: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review runtime-core changes**

Check:

- policy-driven defaults are coherent
- retry / dead-letter policy is more explicit and less path-specific
- replay / recovery substrate semantics are more aligned
- dispute-core behavior was not accidentally pulled into this workstream

- [ ] **Step 2: Review downstream integration changes**

Check:

- meta-tool and operator-facing runtime behavior match Worker A’s contract
- no downstream layer invented its own runtime semantics

- [ ] **Step 3: Run full verification**

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

- [ ] **Step 4: Commit the implementation**

Commit message:

```bash
git -c commit.gpgsign=false commit -m "feat: normalize replay recovery runtime policy"
```
