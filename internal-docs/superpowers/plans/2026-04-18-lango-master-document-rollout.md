# Lango Master Document Rollout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish the top-level Lango master document, wire it into the docs site, and create the first audit and track entry-point documents that future execution plans can build on.

**Architecture:** This rollout intentionally stops at the documentation control plane. It publishes a canonical master document under `docs/architecture/`, adds a first audit ledger for `External Collaboration & Economic Exchange`, adds a first track charter for `P2P Knowledge Exchange`, and wires those documents into MkDocs navigation. Detailed audit judgments and implementation-track execution happen in follow-on plans so each later plan can stay narrow and testable.

**Tech Stack:** Markdown, MkDocs Material, existing `docs/architecture/` docs structure, repo-local verification with `rg` and `mkdocs build --strict`

---

## Scope Split

The approved design spans multiple independent workstreams. This plan only delivers the documentation control plane:

- published master document,
- published first audit ledger,
- published first track charter,
- docs navigation updates.

Do **not** expand this plan into full feature-audit judgments for all capability areas or into product implementation work. Those should be separate plans after this rollout lands.

## File Map

- Create: `docs/architecture/master-document.md`
  - Canonical top-level constitution, product path, capability areas, audit framework, and track portfolio.
- Create: `docs/architecture/external-collaboration-audit.md`
  - First audit ledger for `External Collaboration & Economic Exchange`.
- Create: `docs/architecture/p2p-knowledge-exchange-track.md`
  - First concrete track charter derived from the master document.
- Modify: `docs/architecture/index.md`
  - Surface the new pages on the architecture landing page.
- Modify: `mkdocs.yml`
  - Add the new pages to the site navigation.

## Task 1: Publish the Master Document

**Files:**
- Create: `docs/architecture/master-document.md`
- Test: `docs/architecture/master-document.md`

- [ ] **Step 1: Verify the master document does not already exist**

Run:

```bash
test -f docs/architecture/master-document.md && echo "exists" || echo "missing"
```

Expected:

```text
missing
```

- [ ] **Step 2: Create `docs/architecture/master-document.md` with the approved constitution and product path**

Write:

```md
# Lango Master Document

## Purpose

This document is the top-level product and strategy document for Lango.

It exists to:

- define what Lango fundamentally is,
- define how features are judged,
- define which execution tracks own which problems,
- keep future work aligned with one product story.

## Product Constitution

Lango is a sovereign peer-to-peer agent network, not just a local agent runtime.

### Constitutional Principles

1. External collaboration between sovereign agents is the central product identity.
2. Internal collaboration matters, but it is a reduced form of the external collaboration model.
3. External collaboration is economically native.
4. On-chain settlement is one trusted settlement mechanism, not the product itself.
5. The default external collaboration structure is leader-led.
6. The long-term product target is high autonomy, but the near-term product model is delegated autonomy.
7. User private conversations, confidential material, and sensitive internal information are never tradeable assets.
8. Tradeable knowledge and deliverables must stay inside an allowlist plus explicit exportability policy.
9. The first economic activity is knowledge exchange and result exchange, not broad remote execution.
10. Trust starts from cryptographic continuity of identity, but reputation should be centered on real collaboration and transaction history.
11. The owner carries root identity and final accountability, while individual agents accumulate separate role-specific and domain-specific reputation.

## Primary Product Path

### Phase 1: Knowledge Exchange

- Sovereign external agents participate through pseudonymous but cryptographically continuous identities.
- The first market activity is expertise access plus reviewable deliverables.
- The default experience is delivery of outputs such as summaries, research notes, design drafts, and code drafts.
- Tradeable artifacts remain bounded by allowlists and exportability policy.
- Default settlement is small upfront payment plus approval-based final settlement, anchored by on-chain stablecoin.

### Phase 2: Result Exchange with Controlled Execution

- Repeat interactions and stronger trust allow more structured result exchange.
- Limited execution may be opened only for higher-trust relationships with explicit approval and strong policy controls.
- Off-chain accrual, dynamic credit limits, clearer acceptance criteria, and formalized dispute handling mature here.

### Phase 3: Leader-Led Team Execution

- A leader agent forms an external team around a shared goal.
- Team composition uses standard role templates plus optional custom roles.
- Budgeting, contracting, and settlement remain leader-controlled by default, with narrow delegated authority when needed.
- Workspace sharing is selective and leader-owned by default.
- Artifact exchange is snapshot/package oriented by default, not unrestricted live co-editing.

### Phase 4: Long-Running Multi-Agent Projects

- Collaboration expands from one-off tasks to long-running, multi-session, multi-milestone projects.
- Provenance, ledgers, recurring settlement, durable work records, and persistent shared artifacts become central.

## Capability Areas

### Sovereign Runtime & Operator UX

Providers, agent runtime, channels, cockpit, settings, status, session handling, logging, and core tool surfaces.

### Knowledge & Intelligence Capital

Knowledge, observational memory, embedding/RAG, graph, ontology, librarian, agent memory, skills, and context systems.

### External Collaboration & Economic Exchange

P2P identity, trust, reputation, pricing, negotiation, settlement, team formation, and shared artifact exchange.

### Trust, Security & Policy

Auth, signer/KMS, privacy boundaries, exportability policy, approvals, gatekeeper, exec safety, sandboxing, and auditability.

### Execution, Continuity & Accountability

Workflow, background tasks, cron, RunLedger, provenance, and hooks.

### Extensibility & Developer Clarity

MCP, extension packs, architecture boundaries, docs truthfulness, CLI surface clarity, and duplicate cleanup.

## Feature Audit Framework

### Allowed Audit Judgments

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

### Audit Rules

- A feature is valuable if it directly strengthens the core product story.
- The strongest test is whether it materially supports external collaboration, knowledge exchange, result exchange, leader-led team execution, or long-running multi-agent collaboration.
- Features with overlapping responsibility should be merge candidates.
- Features aligned with the vision but weak in reliability, defaults, docs, trust boundaries, or UX should be stabilization candidates.
- Features that fit the vision but are too early for the current product path should be deferred.
- Features that add complexity without strengthening the core story should be removal candidates.

## Track Portfolio

### Stabilization Track

Truth alignment, durability, defaults, hardening, observability, and production correctness.

### Consolidation Track

Duplicate feature, config-surface, command-surface, and responsibility cleanup.

### UX Clarity Track

Settings, cockpit, CLI, docs, and defaults simplification.

### P2P Knowledge Exchange Track

The first concrete economic track for external sovereign agents.

### Leader-Led Team Execution Track

The next layer above knowledge exchange.

### Developer Clarity Track

Architecture boundary, naming, documentation, and codebase clarity improvements.

## Near-Term Priorities

1. Audit `External Collaboration & Economic Exchange`.
2. Audit `Trust, Security & Policy`.
3. Publish the first `P2P Knowledge Exchange` track document.
4. Move stabilization work under the master-document hierarchy.
5. Follow with consolidation and UX clarity work.
```

- [ ] **Step 3: Verify the required sections exist**

Run:

```bash
rg -n "^## Product Constitution$|^## Primary Product Path$|^## Capability Areas$|^## Feature Audit Framework$|^## Track Portfolio$|^## Near-Term Priorities$" docs/architecture/master-document.md
```

Expected:

```text
11:## Product Constitution
28:## Primary Product Path
56:## Capability Areas
78:## Feature Audit Framework
94:## Track Portfolio
111:## Near-Term Priorities
```

- [ ] **Step 4: Commit the published master document**

Run:

```bash
git add docs/architecture/master-document.md
git -c commit.gpgsign=false commit -m "docs: publish lango master document"
```

Expected:

```text
[<branch> <sha>] docs: publish lango master document
 1 file changed, <n> insertions(+)
 create mode 100644 docs/architecture/master-document.md
```

## Task 2: Publish the First External Collaboration Audit Ledger

**Files:**
- Create: `docs/architecture/external-collaboration-audit.md`
- Test: `docs/architecture/external-collaboration-audit.md`

- [ ] **Step 1: Verify the audit ledger file does not already exist**

Run:

```bash
test -f docs/architecture/external-collaboration-audit.md && echo "exists" || echo "missing"
```

Expected:

```text
missing
```

- [ ] **Step 2: Create `docs/architecture/external-collaboration-audit.md` with the audit order and ledger structure**

Write:

```md
# External Collaboration & Economic Exchange Audit

## Purpose

This document is the first detailed audit ledger under the Lango master document.

It exists to review the product area that most directly defines Lango:

- P2P identity,
- trust,
- reputation,
- pricing,
- negotiation,
- settlement,
- team formation,
- shared artifacts.

## Audit Order

1. P2P identity / trust / reputation
2. pricing / negotiation / settlement
3. team formation / role coordination
4. workspace / shared artifacts

## Audit Method

Each feature family should be judged by:

- capability area fit,
- product-path fit,
- current user-facing surface,
- duplication risk,
- trust or policy gaps,
- judgment,
- owning track.

Allowed judgments:

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

## Current Surface Map

| Feature family | Primary phase | Current surface clues | Audit status |
| --- | --- | --- | --- |
| P2P identity / trust / reputation | Phase 1 | `docs/features/p2p-network.md`, `docs/features/economy.md`, `internal/config/types_p2p.go`, `internal/cli/p2p/`, `internal/cli/settings/forms_p2p.go` | Ready for detailed audit |
| pricing / negotiation / settlement | Phase 1-2 | `docs/features/economy.md`, `docs/payments/usdc.md`, `docs/payments/x402.md`, `internal/config/types_economy.go`, `internal/cli/economy/`, `internal/cli/payment/` | Ready for detailed audit |
| team formation / role coordination | Phase 3 | `docs/features/p2p-network.md`, `docs/features/multi-agent.md`, `internal/config/types_p2p.go`, `internal/config/types_orchestration.go`, `internal/cli/p2p/`, `internal/cli/agent/` | Ready for detailed audit |
| workspace / shared artifacts | Phase 3-4 | `docs/features/p2p-network.md`, `docs/features/provenance.md`, `internal/config/types_p2p.go`, `internal/cli/p2p/`, `internal/cli/provenance/` | Ready for detailed audit |

## Baseline Decisions Already Locked

- External collaboration is economically native.
- Trust is mixed: cryptographic continuity first, transaction history at the center.
- Root accountability belongs to the owner.
- Agent-level reputation stays separate from owner-level root trust.
- Early trade is bounded by allowlists plus explicit exportability policy.
- The default early external exchange is deliverable-oriented, not broad execution.
- On-chain stablecoin is the trust anchor for settlement.
- Off-chain accrual opens only after trust is earned.
- Team formation is leader-led by default.
- Shared artifacts are leader-owned and selectively exposed by scope.

## Next Plan

The next implementation plan after this document lands should perform the detailed audit for the first row:

- P2P identity / trust / reputation
```

- [ ] **Step 3: Verify the audit file contains the expected ledger headings**

Run:

```bash
rg -n "^## Audit Order$|^## Audit Method$|^## Current Surface Map$|^## Baseline Decisions Already Locked$|^## Next Plan$" docs/architecture/external-collaboration-audit.md
```

Expected:

```text
12:## Audit Order
19:## Audit Method
33:## Current Surface Map
42:## Baseline Decisions Already Locked
55:## Next Plan
```

- [ ] **Step 4: Commit the audit ledger**

Run:

```bash
git add docs/architecture/external-collaboration-audit.md
git -c commit.gpgsign=false commit -m "docs: add external collaboration audit ledger"
```

Expected:

```text
[<branch> <sha>] docs: add external collaboration audit ledger
 1 file changed, <n> insertions(+)
 create mode 100644 docs/architecture/external-collaboration-audit.md
```

## Task 3: Publish the P2P Knowledge Exchange Track Charter

**Files:**
- Create: `docs/architecture/p2p-knowledge-exchange-track.md`
- Test: `docs/architecture/p2p-knowledge-exchange-track.md`

- [ ] **Step 1: Verify the track document does not already exist**

Run:

```bash
test -f docs/architecture/p2p-knowledge-exchange-track.md && echo "exists" || echo "missing"
```

Expected:

```text
missing
```

- [ ] **Step 2: Create `docs/architecture/p2p-knowledge-exchange-track.md` with the first concrete product-track charter**

Write:

```md
# P2P Knowledge Exchange Track

## Goal

Define the first concrete external market activity for sovereign Lango agents:

- expertise access,
- reviewable deliverables,
- bounded exportability,
- trusted settlement,
- dispute-ready receipts.

## Why This Track Comes First

Lango should reach meaningful external economic activity before it tries to support broad external execution or long-running multi-agent organizations.

Knowledge exchange is the narrowest useful slice because it:

- creates real external value,
- forces trust and exportability boundaries to become explicit,
- produces reviewable artifacts,
- creates a clean bridge toward later team execution.

## In Scope

- pseudonymous but cryptographically continuous identities,
- owner-root trust plus agent-specific reputation,
- allowlist plus explicit exportability policy,
- expertise access and reviewable deliverables,
- small upfront payment plus approval-based final settlement,
- on-chain stablecoin as the trust anchor,
- off-chain accrual only after trust is established,
- dispute handling grounded in signed logs, provenance, acceptance criteria, escrow state, and immutable receipts.

## Out of Scope

- unrestricted remote execution,
- full team-based role orchestration,
- live shared workspaces by default,
- complete reputation formula design,
- final smart-contract design.

## Default Transaction Shape

1. A leader agent discovers or selects an external counterparty.
2. The leader agent confirms the target artifact is tradeable under exportability policy.
3. The parties agree on price, delivery window, and deliverable scope.
4. A small upfront payment or escrowed commitment is created.
5. The external agent delivers a reviewable artifact.
6. The leader agent approves, rejects, or disputes the artifact.
7. Final settlement is released on approval or handled through dispute rules.

## Required Follow-On Plans

1. `P2P identity / trust / reputation` detailed audit
2. `pricing / negotiation / settlement` detailed audit
3. exportability policy and approval flow design
4. first implementation plan for the `knowledge exchange` runtime path
```

- [ ] **Step 3: Verify the track document has the required operational sections**

Run:

```bash
rg -n "^## Goal$|^## Why This Track Comes First$|^## In Scope$|^## Out of Scope$|^## Default Transaction Shape$|^## Required Follow-On Plans$" docs/architecture/p2p-knowledge-exchange-track.md
```

Expected:

```text
3:## Goal
11:## Why This Track Comes First
21:## In Scope
31:## Out of Scope
38:## Default Transaction Shape
48:## Required Follow-On Plans
```

- [ ] **Step 4: Commit the track charter**

Run:

```bash
git add docs/architecture/p2p-knowledge-exchange-track.md
git -c commit.gpgsign=false commit -m "docs: add p2p knowledge exchange track"
```

Expected:

```text
[<branch> <sha>] docs: add p2p knowledge exchange track
 1 file changed, <n> insertions(+)
 create mode 100644 docs/architecture/p2p-knowledge-exchange-track.md
```

## Task 4: Wire the New Documents into the Docs Site

**Files:**
- Modify: `docs/architecture/index.md`
- Modify: `mkdocs.yml`
- Test: `docs/architecture/index.md`
- Test: `mkdocs.yml`

- [ ] **Step 1: Verify the new pages are not already linked**

Run:

```bash
rg -n "Master Document|External Collaboration Audit|P2P Knowledge Exchange Track" docs/architecture/index.md mkdocs.yml
```

Expected:

```text
<no matches>
```

- [ ] **Step 2: Add cards to `docs/architecture/index.md` for the new pages**

Insert after the existing cards:

```md
-   :material-compass-outline: **[Master Document](master-document.md)**

    ---

    Top-level product constitution, product path, capability areas, and execution-track portfolio for Lango.

-   :material-clipboard-search-outline: **[External Collaboration Audit](external-collaboration-audit.md)**

    ---

    The first audit ledger for the product area that most directly defines Lango: trust, pricing, settlement, teams, and shared artifacts.

-   :material-cash-fast: **[P2P Knowledge Exchange Track](p2p-knowledge-exchange-track.md)**

    ---

    The first concrete product track for external sovereign-agent economic activity.
```

- [ ] **Step 3: Add the three new pages to `mkdocs.yml` under `Architecture`**

Update the `Architecture` nav block to:

```yaml
  - Architecture:
    - architecture/index.md
    - System Overview: architecture/overview.md
    - Project Structure: architecture/project-structure.md
    - Data Flow: architecture/data-flow.md
    - Master Document: architecture/master-document.md
    - External Collaboration Audit: architecture/external-collaboration-audit.md
    - P2P Knowledge Exchange Track: architecture/p2p-knowledge-exchange-track.md
```

- [ ] **Step 4: Build the docs site in strict mode**

Run:

```bash
mkdocs build --strict
```

Expected:

```text
INFO    -  Cleaning site directory
INFO    -  Building documentation to directory: site
INFO    -  Documentation built in <time>
```

- [ ] **Step 5: Commit the documentation wiring**

Run:

```bash
git add docs/architecture/index.md mkdocs.yml
git -c commit.gpgsign=false commit -m "docs: wire master docs into architecture nav"
```

Expected:

```text
[<branch> <sha>] docs: wire master docs into architecture nav
 2 files changed, <n> insertions(+)
```

## Task 5: Final Consistency Pass Against the Approved Design

**Files:**
- Modify: `docs/architecture/master-document.md`
- Modify: `docs/architecture/external-collaboration-audit.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Test: `docs/architecture/master-document.md`
- Test: `docs/architecture/external-collaboration-audit.md`
- Test: `docs/architecture/p2p-knowledge-exchange-track.md`

- [ ] **Step 1: Compare the published documents against the approved design spec**

Run:

```bash
rg -n "^## Product Constitution$|^## Primary Product Path$|^## Capability Areas$|^## Feature Audit Framework$|^## Track Portfolio$|^## Near-Term Sequencing$|^## External Collaboration Decisions Captured So Far$" docs/superpowers/specs/2026-04-18-lango-master-document-design.md
```

Expected:

```text
26:## Product Constitution
41:## Primary Product Path
78:## Capability Areas
117:## Feature Audit Framework
157:## Track Portfolio
207:## External Collaboration Audit Order
218:## External Collaboration Decisions Captured So Far
286:## Near-Term Sequencing
```

- [ ] **Step 2: Patch any missing sections into the published docs**

If the published docs are missing ideas from the spec, add them directly. The minimum acceptable fixes are:

```md
## Relationship to Existing Stabilization Work

The current stabilization work remains necessary, but it is subordinate to the master document and should be treated as one execution track among several.
```

and

```md
## Relationship to the Master Document

This audit exists under the top-level Lango master document and should not redefine product constitution on its own.
```

and

```md
## Relationship to Later Team Execution

This track is intentionally narrower than leader-led team execution. It should create the trust, settlement, and deliverable boundaries that later team-based collaboration depends on.
```

- [ ] **Step 3: Re-run strict docs build**

Run:

```bash
mkdocs build --strict
```

Expected:

```text
INFO    -  Cleaning site directory
INFO    -  Building documentation to directory: site
INFO    -  Documentation built in <time>
```

- [ ] **Step 4: Commit the consistency pass**

Run:

```bash
git add docs/architecture/master-document.md docs/architecture/external-collaboration-audit.md docs/architecture/p2p-knowledge-exchange-track.md
git -c commit.gpgsign=false commit -m "docs: align master-doc rollout artifacts"
```

Expected:

```text
[<branch> <sha>] docs: align master-doc rollout artifacts
 3 files changed, <n> insertions(+), <m> deletions(-)
```

## Self-Review

### Spec Coverage

This plan covers the approved design by publishing:

- the top-level master document,
- the first audit ledger,
- the first product-track charter,
- the docs-site navigation needed to expose them.

It intentionally does **not** attempt to complete the detailed audit judgments or implementation work for the product tracks themselves. Those should be follow-on plans.

### Placeholder Scan

This plan contains none of the banned placeholder patterns. Every file path, command, and initial document payload is explicit.

### Type Consistency

Document names and nav labels stay consistent across tasks:

- `Master Document`
- `External Collaboration Audit`
- `P2P Knowledge Exchange Track`
