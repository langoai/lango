# Dispute Runtime Completion Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the dispute / settlement / escrow runtime enough to land explicit keep-hold / re-escalation semantics, richer settlement progression, and broader dispute-linked escrow lifecycle behavior.

**Architecture:** Keep receipts as the canonical dispute-core authority, then align dispute, settlement, and escrow services around that contract. The workstream is receipts-first: Worker A settles canonical state and progression invariants, Worker B aligns settlement / escrow services and tool contracts, and Worker C truth-aligns docs/OpenSpec.

**Tech Stack:** Go, `internal/receipts`, dispute / adjudication / settlement / escrow domain services, `internal/app/tools_meta*.go`, Zensical docs, OpenSpec

---

## File Map

### Worker A: Receipts / Canonical Dispute-Core State

- Modify: `internal/receipts/*`
  - Add or extend canonical dispute / settlement / escrow state and evidence invariants.
- Modify: focused tests in `internal/receipts/*`
  - Cover canonical transitions and invariant enforcement.

### Worker B: Settlement / Escrow / Tool Alignment

- Modify: `internal/disputehold/*`
- Modify: `internal/escrowadjudication/*`
- Modify: `internal/escrowrelease/*`
- Modify: `internal/escrowrefund/*`
- Modify: `internal/settlementprogression/*`
- Modify: `internal/settlementexecution/*`
- Modify: `internal/partialsettlementexecution/*`
- Modify: `internal/app/tools_meta*.go` only where the tool contracts must reflect the landed canonical runtime
- Modify: focused service and integration tests adjacent to the above packages

### Worker C: Docs / OpenSpec / README

- Modify: `docs/architecture/*`
  - Update dispute / settlement / escrow architecture pages to match landed runtime behavior.
- Modify: `docs/cli/*` and `README.md` only if user-visible runtime-facing behavior changes.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync docs-only requirements.
- Create: `openspec/changes/archive/2026-04-26-dispute-runtime-completion-workstream/**`
  - Proposal, design, tasks, and docs-only delta spec.

## Task Breakdown

### Task 1: Map and Test Current Canonical Gaps

**Owner:** Worker A

**Files:**
- Modify: focused tests under `internal/receipts/*`

- [ ] **Step 1: Add or extend failing receipts tests**

Cover the current missing canonical behavior:

- keep-hold or re-escalation after adjudication pressure or incomplete downstream resolution
- richer settlement progression under disagreement pressure
- dispute-linked escrow lifecycle transitions that are still implicit or incomplete

Focus on canonical state and evidence behavior first, not tool UX.

- [ ] **Step 2: Run focused receipts tests and verify they fail**

Run the narrowest package-level test commands that cover the new canonical-state assertions.

Expected:

```text
FAIL
```

### Task 2: Land Canonical Keep-Hold / Re-Escalation Semantics

**Owner:** Worker A

**Files:**
- Modify: `internal/receipts/*`
- Modify: receipts tests from Task 1

- [ ] **Step 1: Implement canonical keep-hold / re-escalation behavior**

Implementation rules:

- make continued hold or renewed escalation explicit in canonical receipts state
- preserve already-landed adjudication evidence where correct
- avoid encoding continued disagreement as ad hoc execution failure side effects

- [ ] **Step 2: Re-run focused receipts tests and verify they pass**

Run the same focused receipts test commands from Task 1 plus any direct coverage added for keep-hold / re-escalation behavior.

Expected:

```text
ok
```

### Task 3: Extend Settlement Progression for Richer Disagreement Outcomes

**Owner:** Worker A

**Files:**
- Modify: `internal/receipts/*`
- Modify: related receipts tests

- [ ] **Step 1: Add or extend failing settlement-progression tests**

Cover:

- partial settlement aftermath under renewed disagreement
- continued review or renewed dispute pressure after partial or attempted resolution
- transaction-level progression remaining canonical

- [ ] **Step 2: Run focused settlement-progression tests and verify they fail**

Run the narrowest package-level test commands for the affected receipts progression behavior.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement richer canonical settlement progression**

Implementation rules:

- keep transaction-level progression canonical
- avoid leaking service-specific incidental states into the canonical model
- make disagreement aftermath semantics explicit enough for downstream services to consume

- [ ] **Step 4: Re-run focused settlement-progression tests and verify they pass**

Run the same focused progression test commands.

Expected:

```text
ok
```

### Task 4: Complete Dispute-Linked Escrow Lifecycle Behavior

**Owner:** Worker B

**Files:**
- Modify: `internal/disputehold/*`
- Modify: `internal/escrowadjudication/*`
- Modify: `internal/escrowrelease/*`
- Modify: `internal/escrowrefund/*`
- Modify: `internal/settlementprogression/*`
- Modify: `internal/settlementexecution/*`
- Modify: `internal/partialsettlementexecution/*`
- Modify: focused service tests

- [ ] **Step 1: Add or extend failing service tests**

Cover:

- dispute-linked escrow release / refund safety
- service behavior under canonical keep-hold / re-escalation states
- alignment between richer settlement progression and escrow execution paths

- [ ] **Step 2: Run focused service tests and verify they fail**

Run the narrowest package-level test commands for the touched service packages.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement service alignment**

Implementation rules:

- Worker A’s receipts contract is the source of truth
- service validation and execution paths should consume the canonical dispute-core state
- avoid inventing a parallel dispute state machine inside services

- [ ] **Step 4: Re-run focused service tests and verify they pass**

Run the same focused service test commands.

Expected:

```text
ok
```

### Task 5: Align Meta-Tool Contracts and Downstream Wording

**Owner:** Worker B

**Files:**
- Modify: `internal/app/tools_meta*.go`
- Modify: focused tool/integration tests

- [ ] **Step 1: Add or extend failing tool tests**

Cover:

- tool contracts reflecting the landed canonical dispute runtime
- no regression in existing release/refund/dispute-hold/adjudication tool behavior
- downstream wording or output aligned with the new canonical states where needed

- [ ] **Step 2: Run focused tool tests and verify they fail**

Run focused package-level tests for the touched meta-tool paths.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement downstream contract alignment**

Implementation rules:

- receipts / canonical domain state remains the source of truth
- tools expose the landed runtime more clearly without inventing a second state model

- [ ] **Step 4: Re-run focused tool tests and verify they pass**

Run the same focused tool test commands.

Expected:

```text
ok
```

### Task 6: Truth-Align Docs / OpenSpec

**Owner:** Worker C

**Files:**
- Modify: verified dispute / settlement / escrow docs under `docs/architecture/*`
- Modify: `docs/cli/*` and `README.md` only if runtime-facing behavior changes
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-26-dispute-runtime-completion-workstream/**`

- [ ] **Step 1: Audit landed dispute-core behavior before writing docs**

Verify from code:

- keep-hold / re-escalation behavior is actually modeled
- richer settlement progression semantics are actually landed
- dispute-linked escrow lifecycle completion is actually landed
- downstream wording changes only describe real behavior

- [ ] **Step 2: Update public docs**

Document:

- the landed canonical dispute runtime
- the richer settlement progression semantics
- the completed dispute-linked escrow lifecycle behavior

- [ ] **Step 3: Sync main OpenSpec requirements**

Update `openspec/specs/docs-only/spec.md` so the docs-only requirements reflect the landed dispute-runtime work and narrow the remaining backlog accordingly.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-26-dispute-runtime-completion-workstream/proposal.md`
- `openspec/changes/archive/2026-04-26-dispute-runtime-completion-workstream/design.md`
- `openspec/changes/archive/2026-04-26-dispute-runtime-completion-workstream/tasks.md`
- `openspec/changes/archive/2026-04-26-dispute-runtime-completion-workstream/specs/docs-only/spec.md`

### Task 7: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review canonical dispute-core changes**

Check:

- keep-hold / re-escalation semantics are explicit
- settlement progression is richer but still coherent
- dispute-linked escrow lifecycle behavior is more complete
- the canonical state remains centered in receipts

- [ ] **Step 2: Review downstream service and tool alignment**

Check:

- services and tools consume the same canonical dispute-core contract
- no second state model was invented downstream

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
git -c commit.gpgsign=false commit -m "feat: complete dispute runtime flow"
```
