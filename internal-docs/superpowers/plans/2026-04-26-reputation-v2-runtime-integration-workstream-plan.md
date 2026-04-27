# Reputation V2 + Runtime Integration Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land a stronger reputation V2 contract, strengthen trust-entry semantics, and align key runtime consumers with that contract.

**Architecture:** Establish the canonical trust / reputation meaning first inside the reputation boundary, then align pricing, paygate, firewall, and team/runtime bridges to consume that meaning without inventing local interpretations. The workstream is reputation-contract first: Worker A settles the V2 model and canonical helpers, Worker B aligns runtime consumers, and Worker C truth-aligns docs/OpenSpec.

**Tech Stack:** Go, `internal/p2p/reputation`, `internal/app/wiring_p2p.go`, `internal/app/wiring_economy.go`, `internal/p2p/firewall`, `internal/p2p/team`, pricing/risk/paygate integration, Zensical docs, OpenSpec

---

## File Map

### Worker A: Reputation / Trust Contract Core

- Modify: `internal/p2p/reputation/*`
  - Strengthen canonical V2 reputation and trust-entry semantics.
- Modify: related trust / reputation helper readers or bridge helpers only where they belong to the canonical contract layer.
- Modify: focused trust-model tests adjacent to the above packages.

### Worker B: Runtime Consumer Integration

- Modify: `internal/app/wiring_p2p.go`
- Modify: `internal/app/wiring_economy.go`
- Modify: `internal/p2p/firewall/*`
- Modify: `internal/p2p/team/*`
- Modify: any focused pricing / paygate / selection consumer files that genuinely read the reputation contract
- Modify: focused integration tests adjacent to the above packages

### Worker C: Docs / OpenSpec / README

- Modify: `docs/architecture/*`
  - Update trust / reputation / runtime-integration pages to match landed behavior.
- Modify: `docs/features/*` where trust or reputation is surfaced.
- Modify: `docs/cli/*` and `README.md` only if user-visible runtime behavior changes.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync docs-only requirements.
- Create: `openspec/changes/archive/2026-04-26-reputation-v2-runtime-integration-workstream/**`
  - Proposal, design, tasks, and docs-only delta spec.

## Task Breakdown

### Task 1: Map and Test the Canonical Reputation V2 Gaps

**Owner:** Worker A

**Files:**
- Modify: focused tests under `internal/p2p/reputation/*`

- [ ] **Step 1: Add or extend failing trust-model tests**

Cover the current canonical gaps explicitly:

- owner-root trust versus earned agent/domain reputation
- bootstrap trust versus earned trust
- durable negative reputation versus temporary operational safety signals
- stronger treatment of adjudicated negative outcomes than unaudited operational failures

- [ ] **Step 2: Run focused trust-model tests and verify they fail**

Run the narrowest package-level test commands that cover the new trust / reputation assertions.

Expected:

```text
FAIL
```

### Task 2: Land the Canonical Reputation V2 Contract

**Owner:** Worker A

**Files:**
- Modify: `internal/p2p/reputation/*`
- Modify: related canonical helper readers if required
- Modify: trust-model tests from Task 1

- [ ] **Step 1: Implement the stronger canonical reputation contract**

Implementation rules:

- preserve owner-root trust as a continuity floor
- keep earned agent/domain reputation based on actual collaboration history
- prevent temporary operational safety events from automatically becoming permanent durable reputation damage
- make adjudicated negative outcomes eligible for stronger durable impact

- [ ] **Step 2: Re-run focused trust-model tests and verify they pass**

Run the same focused trust-model commands from Task 1 plus any direct coverage added for the canonical V2 contract.

Expected:

```text
ok
```

### Task 3: Strengthen the Trust-Entry Contract

**Owner:** Worker A

**Files:**
- Modify: `internal/p2p/reputation/*`
- Modify: related canonical trust-entry helpers
- Modify: related tests

- [ ] **Step 1: Add or extend failing trust-entry tests**

Cover:

- first-time peers versus returning peers
- low-trust but not-yet-banned peers
- temporarily unsafe peers
- bootstrap trust staying distinct from earned trust

- [ ] **Step 2: Run focused trust-entry tests and verify they fail**

Run the narrowest package-level test commands for the affected trust-entry behavior.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the stronger trust-entry behavior**

Implementation rules:

- do not collapse all trust-entry decisions into one scalar score
- keep the canonical contract readable and narrowly scoped
- avoid pushing runtime-consumer policy into the canonical reputation layer

- [ ] **Step 4: Re-run focused trust-entry tests and verify they pass**

Run the same focused trust-entry test commands.

Expected:

```text
ok
```

### Task 4: Align Runtime Consumers with the V2 Contract

**Owner:** Worker B

**Files:**
- Modify: `internal/app/wiring_p2p.go`
- Modify: `internal/app/wiring_economy.go`
- Modify: `internal/p2p/firewall/*`
- Modify: `internal/p2p/team/*`
- Modify: focused runtime consumer files only where needed
- Modify: focused integration tests

- [ ] **Step 1: Add or extend failing runtime-consumer tests**

Cover:

- pricing / risk reading the stronger reputation contract
- firewall / admission behavior staying coherent with the trust-entry contract
- team / coordination bridges reacting to reputation using the stronger semantics
- no regression in existing runtime consumers

- [ ] **Step 2: Run focused runtime-consumer tests and verify they fail**

Run focused package-level tests for the touched runtime-consumer packages.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement runtime consumer alignment**

Implementation rules:

- Worker A’s canonical contract is the source of truth
- keep consumer changes minimal and contract-aligned
- do not invent local trust or reputation meanings in consumer subsystems

- [ ] **Step 4: Re-run focused runtime-consumer tests and verify they pass**

Run the same focused runtime-consumer test commands.

Expected:

```text
ok
```

### Task 5: Integrate Adjudicated Outcomes and Operational Signals

**Owner:** Worker B

**Files:**
- Modify: the narrowest runtime-consumer files actually needed
- Modify: focused integration tests

- [ ] **Step 1: Add or extend failing integration tests**

Cover:

- adjudicated dispute outcomes affecting durable reputation more strongly than temporary operational signals
- operational safety invalidation not automatically causing permanent durable damage
- runtime consumers continuing to behave consistently after the stronger distinction is landed

- [ ] **Step 2: Run focused integration tests and verify they fail**

Run the narrowest package-level test commands for the touched integration points.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement adjudication/signal integration**

Implementation rules:

- preserve the distinction between temporary operational safety actions and durable negative reputation
- keep dispute-core truth as the source for stronger durable impact
- avoid broad operator-surface work in this task

- [ ] **Step 4: Re-run focused integration tests and verify they pass**

Run the same focused integration test commands.

Expected:

```text
ok
```

### Task 6: Truth-Align Docs / OpenSpec

**Owner:** Worker C

**Files:**
- Modify: verified trust / reputation / runtime docs under `docs/architecture/*`
- Modify: `docs/features/*`
- Modify: `docs/cli/*` and `README.md` only if surfaced behavior changes
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-26-reputation-v2-runtime-integration-workstream/**`

- [ ] **Step 1: Audit landed reputation/runtime behavior before writing docs**

Verify from code:

- the V2 contract is actually explicit
- trust-entry semantics are actually stronger and more separated
- runtime consumers are actually aligned with the stronger contract
- user-facing wording changes only describe real behavior

- [ ] **Step 2: Update public docs**

Document:

- the landed V2 trust / reputation model
- the stronger trust-entry semantics
- the runtime consumer alignment that is actually real

- [ ] **Step 3: Sync main OpenSpec requirements**

Update `openspec/specs/docs-only/spec.md` so the docs-only requirements reflect the landed V2 semantics and narrow the remaining backlog accordingly.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-26-reputation-v2-runtime-integration-workstream/proposal.md`
- `openspec/changes/archive/2026-04-26-reputation-v2-runtime-integration-workstream/design.md`
- `openspec/changes/archive/2026-04-26-reputation-v2-runtime-integration-workstream/tasks.md`
- `openspec/changes/archive/2026-04-26-reputation-v2-runtime-integration-workstream/specs/docs-only/spec.md`

### Task 7: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review canonical reputation changes**

Check:

- owner-root trust, bootstrap trust, earned trust, and durable negative impact are more clearly separated
- the canonical contract stayed centered in the reputation boundary

- [ ] **Step 2: Review runtime consumer alignment**

Check:

- consumers read the stronger trust / reputation contract consistently
- no local subsystem invented a competing trust model

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
git -c commit.gpgsign=false commit -m "feat: strengthen reputation runtime model"
```
