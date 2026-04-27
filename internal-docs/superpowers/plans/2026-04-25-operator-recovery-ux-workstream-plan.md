# Operator Recovery UX Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Polish the landed dead-letter retry experience across CLI and cockpit so operators can clearly distinguish retry acceptance, precheck rejection, and invocation failure.

**Architecture:** Keep the current retry control plane unchanged and refine only the operator-facing recovery UX. The workstream updates `internal/cli/status` and `internal/cli/cockpit` to present clearer retry states and result payloads, then truth-aligns public docs and OpenSpec in one closeout. CLI and cockpit continue to share the same retryability precheck (`can_retry`) and the same mutation path (`retry_post_adjudication_execution`).

**Tech Stack:** Go, Cobra CLI, Bubble Tea cockpit UI, existing dead-letter detail/retry bridge, Zensical docs, OpenSpec

---

## File Map

### Worker A: CLI Code / Tests

- Modify: `internal/cli/status/status.go`
  - Refine retry success/failure output and `json` result shape.
- Modify: `internal/cli/status/status_test.go`
  - Cover precheck failure wording, invocation failure wording, and refined success output.
- Modify: `internal/cli/status/render.go` only if a small shared output helper becomes necessary.

### Worker B: Cockpit Code / Tests

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Refine retry state text, message priority, and success/failure copy.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover retry state transitions and refined messages.
- Modify: `internal/cli/cockpit/deps.go` only if the cockpit needs a cleaner retry result surface from the existing bridge.

### Worker C: Docs / OpenSpec / README

- Modify: `docs/cli/status.md`
  - Document refined retry output semantics.
- Modify: `docs/cli/index.md`
  - Keep quick references aligned if retry wording or examples change.
- Modify: `README.md`
  - Update user-facing dead-letter operator examples only if the surfaced retry wording appears there.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the refined operator recovery UX.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow the remaining operator-surface follow-on work after this workstream lands.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync public docs requirements for refined retry UX.
- Create: `openspec/changes/archive/2026-04-25-operator-recovery-ux-workstream/**`
  - Proposal, design, tasks, and delta specs.

## Task Breakdown

### Task 1: Refine CLI Retry Recovery UX

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI tests**

Add or extend coverage for:

- `can_retry=false` precheck failures using clearer operator-facing wording
- invocation failures that are distinct from precheck failures
- refined success output that reads as retry acceptance/request, not completion
- refined `json` output with explicit success/failure result semantics
- unchanged confirm / `--yes` gating

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the CLI retry UX refinements**

Refine:

- default retry success message
- default retry failure message
- precheck failure wording
- `json` result payload shape

Keep unchanged:

- command name
- confirm prompt behavior
- `--yes`
- detail precheck path
- retry invocation path

- [ ] **Step 4: Re-run the focused status CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
ok
```

### Task 2: Refine Cockpit Retry Recovery UX

**Owner:** Worker B

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write or extend failing cockpit tests**

Add or extend coverage for:

- clearer `confirm -> running -> success/failure` state text
- success messaging after retry acceptance and refresh
- failure messaging after invocation failure
- message priority when retry state and refresh behavior overlap
- unchanged duplicate retry guard and selection-preservation behavior

- [ ] **Step 2: Run the focused cockpit tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the cockpit retry UX refinements**

Refine:

- retry action state copy
- success message wording
- failure message wording
- state-transition presentation in the detail pane

Keep unchanged:

- inline confirm
- running guard
- success refresh
- failure return to idle
- selection-preservation semantics

- [ ] **Step 4: Re-run the focused cockpit tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/... -count=1
```

Expected:

```text
ok
```

### Task 3: Truth-Align Docs / README / OpenSpec

**Owner:** Worker C

**Files:**
- Modify: verified CLI/public docs
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-operator-recovery-ux-workstream/**`

- [ ] **Step 1: Audit the final retry UX in code before editing docs**

Verify:

- CLI retry success wording
- CLI retry failure wording
- CLI `json` result semantics
- cockpit retry success/failure wording
- unchanged retry control-plane semantics

- [ ] **Step 2: Update CLI/public docs**

Document the refined operator recovery UX:

- CLI retry precheck behavior
- CLI success/failure output semantics
- cockpit retry state wording
- success meaning as retry acceptance/request, not full completion

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed Operator Recovery UX workstream.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-25-operator-recovery-ux-workstream/proposal.md`
- `openspec/changes/archive/2026-04-25-operator-recovery-ux-workstream/design.md`
- `openspec/changes/archive/2026-04-25-operator-recovery-ux-workstream/tasks.md`
- `openspec/changes/archive/2026-04-25-operator-recovery-ux-workstream/specs/docs-only/spec.md`

### Task 4: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review Worker A changes**

Check:

- precheck failure semantics
- invocation failure semantics
- success wording
- `json` result consistency

- [ ] **Step 2: Review Worker B changes**

Check:

- cockpit state-transition wording
- success/failure message priority
- no regression in running/confirm guards

- [ ] **Step 3: Review Worker C changes**

Check:

- docs match actual CLI/cockpit behavior
- README additions only describe wired surfaces
- OpenSpec language matches the landed implementation

- [ ] **Step 4: Run full verification**

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
