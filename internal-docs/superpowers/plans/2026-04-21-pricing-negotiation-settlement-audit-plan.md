# Pricing Negotiation Settlement Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a detailed `pricing / negotiation / settlement` audit ledger for `knowledge exchange v1`, aligned with the master document and the newly landed payment-control-plane slices.

**Architecture:** Build the audit as a documentation-first slice under `docs/architecture/`, using the existing audit-ledger style from `external-collaboration-audit.md`, `trust-security-policy-audit.md`, and the new `identity-trust-reputation-audit.md`. Keep the work documentation-only, but ground it in the current pricing, negotiation, settlement, and escrow surfaces and close it out through OpenSpec.

**Tech Stack:** Markdown, current pricing/negotiation/settlement docs, P2P/economy/payment control-plane code, OpenSpec

---

## File Map

- Create: `docs/architecture/pricing-negotiation-settlement-audit.md`
  - New detailed audit ledger for pricing, negotiation, settlement, and escrow.
- Modify: `docs/architecture/index.md`
  - Add the new audit to the architecture landing page.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Update the track document so this audit is treated as landed audit work rather than pending.
- Modify: `zensical.toml`
  - Add the new audit page to the public Architecture navigation.
- Create: `openspec/changes/pricing-negotiation-settlement-audit/**`
  - Proposal, design, tasks, and delta specs for this audit-slice documentation work.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync the new architecture audit page requirement.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the landing-page and track-doc reference requirements.

### Task 1: Draft The Detailed Pricing / Negotiation / Settlement Audit Ledger

**Files:**
- Create: `docs/architecture/pricing-negotiation-settlement-audit.md`

- [ ] **Step 1: Gather the current control-plane references**

Run:

```bash
sed -n '1,260p' docs/architecture/external-collaboration-audit.md
sed -n '1,220p' docs/features/economy.md
sed -n '1,220p' docs/security/upfront-payment-approval.md
sed -n '1,220p' docs/security/actual-payment-execution-gating.md
sed -n '1,220p' docs/security/escrow-execution.md
sed -n '1,260p' internal-docs/superpowers/specs/2026-04-21-pricing-negotiation-settlement-audit-design.md
```

Expected:

```text
The current public/operator surfaces and the audit design are available as the source material for the ledger.
```

- [ ] **Step 2: Create the audit document skeleton**

Create `docs/architecture/pricing-negotiation-settlement-audit.md` with this structure:

```md
# Pricing Negotiation Settlement Audit

## Purpose

## Relationship to the Master Document

## Document Ownership

## Audit Order

1. Pricing Surface
2. Negotiation
3. Settlement
4. Escrow

## Audit Method

## Current Surface Map

## Baseline Control-Plane Model

## Detailed Audit: Pricing Surface

## Detailed Audit: Negotiation

## Detailed Audit: Settlement

## Detailed Audit: Escrow

## Assessment

## Follow-On Design Inputs
```

- [ ] **Step 3: Fill in ownership, method, and current surface map**

Use this ownership section:

```md
- Primary capability area: `External Collaboration & Economic Exchange`
- Primary execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Stabilization Track`
```

The current surface map should include at least:

```md
| Feature family | Primary phase | Current surface clues | Audit status |
| --- | --- | --- | --- |
| Pricing Surface | Phase 1-2 | `docs/features/p2p-network.md`, `docs/features/economy.md`, `docs/cli/p2p.md`, `internal/cli/p2p/pricing.go`, `internal/economy/pricing/*`, `internal/app/p2p_routes.go` | Detailed audit complete (`stabilize`) |
| Negotiation | Phase 1-2 | `docs/features/economy.md`, `internal/economy/negotiation/*`, `internal/economy/tools.go`, `internal/app/wiring_economy.go` | Detailed audit complete (`stabilize`) |
| Settlement | Phase 1-2 | `docs/security/upfront-payment-approval.md`, `docs/security/actual-payment-execution-gating.md`, `internal/paymentapproval/*`, `internal/paymentgate/*`, `internal/tools/payment/*`, `internal/app/tools_p2p.go` | Detailed audit complete (`stabilize`) |
| Escrow | Phase 1-2 | `docs/security/escrow-execution.md`, `docs/features/economy.md`, `internal/economy/escrow/*`, `internal/escrowexecution/*`, `internal/app/tools_escrow.go` | Detailed audit complete (`stabilize`) |
```

- [ ] **Step 4: Lock the control-plane model and write row-level audit records**

The audit must explicitly lock these ideas from the design:

```md
- `p2p.pricing` = provider-side public quote surface
- `economy.pricing` = local pricing / policy engine
- negotiation is real but under-surfaced
- settlement and escrow are distinct rows
- off-chain accrual / postpay is Phase 2, trust-conditional, and still limited
```

For each row, include:

- Audit Record
- Findings
- Assessment

Each row should start with:

```md
### Audit Record

- Feature name: `Pricing Surface`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: ...
- Core value: ...
- Current problem: ...
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`
```

- [ ] **Step 5: Add final assessment and follow-on design inputs**

End the file with:

```md
## Assessment

All four rows remain `stabilize`: the capability family is real, but the control-plane and progression model still need consolidation.

## Follow-On Design Inputs

1. `knowledge exchange runtime` end-to-end design
2. settlement follow-on work
3. escrow lifecycle completion
```

- [ ] **Step 6: Verify the document reads coherently**

Run:

```bash
sed -n '1,360p' docs/architecture/pricing-negotiation-settlement-audit.md
```

Expected:

```text
The file reads as a complete audit ledger with no placeholders and a consistent control-plane model.
```

- [ ] **Step 7: Commit the audit ledger**

Run:

```bash
git add docs/architecture/pricing-negotiation-settlement-audit.md
git -c commit.gpgsign=false commit -m "docs: add pricing negotiation settlement audit"
```

### Task 2: Wire The Audit Into Public Architecture Docs

**Files:**
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`

- [ ] **Step 1: Add the new audit to the architecture landing page**

Update `docs/architecture/index.md` with a new entry consistent with the existing style:

```md
-   :material-cash-refund: **[Pricing Negotiation Settlement Audit](pricing-negotiation-settlement-audit.md)**
    Audit ledger for pricing surfaces, negotiation, settlement, and escrow in `knowledge exchange v1`.
```

- [ ] **Step 2: Update the track doc to reflect the landed audit**

In `docs/architecture/p2p-knowledge-exchange-track.md`, replace the current audit follow-on item with wording that marks the audit as landed and shifts the remaining work to runtime and lifecycle completion.

Use wording like:

```md
2. `pricing / negotiation / settlement` detailed audit is now landed; the follow-on work is runtime integration, final settlement progression, and escrow lifecycle completion
```

- [ ] **Step 3: Add the new audit page to the public Architecture nav**

In `zensical.toml`, add:

```toml
{ "Pricing Negotiation Settlement Audit" = "architecture/pricing-negotiation-settlement-audit.md" }
```

under the `Architecture` section near the other audit pages.

- [ ] **Step 4: Run the docs build**

Run:

```bash
.venv/bin/zensical build
```

Expected:

```text
Build finished successfully and includes the new architecture audit page.
```

- [ ] **Step 5: Commit the public wiring**

Run:

```bash
git add docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml
git -c commit.gpgsign=false commit -m "docs: wire pricing settlement audit into architecture docs"
```

### Task 3: Final Verification

**Files:**
- No new file changes required unless verification reveals a minimal doc wording issue

- [ ] **Step 1: Run full repository verification**

Run:

```bash
.venv/bin/zensical build
go build ./...
go test ./...
```

Expected:

```text
All commands exit 0.
```

- [ ] **Step 2: Inspect the architecture docs references**

Run:

```bash
rg -n "Pricing Negotiation Settlement Audit" docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml
```

Expected:

```text
The new audit is referenced consistently in the architecture landing page, the track document, and the site nav.
```

### Task 4: OpenSpec Change, Main Spec Sync, And Archive

**Files:**
- Create: `openspec/changes/pricing-negotiation-settlement-audit/proposal.md`
- Create: `openspec/changes/pricing-negotiation-settlement-audit/design.md`
- Create: `openspec/changes/pricing-negotiation-settlement-audit/tasks.md`
- Create: `openspec/changes/pricing-negotiation-settlement-audit/specs/project-docs/spec.md`
- Create: `openspec/changes/pricing-negotiation-settlement-audit/specs/docs-only/spec.md`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`

- [ ] **Step 1: Write the OpenSpec change artifacts**

Create `openspec/changes/pricing-negotiation-settlement-audit/proposal.md`:

```md
## Why

The P2P knowledge-exchange track still treats pricing, negotiation, and settlement as major follow-on work, but the current control-plane relationship between public quotes, local pricing policy, direct settlement, and escrow progression has not been written down in a dedicated audit ledger.

## What Changes

- add a detailed `pricing / negotiation / settlement` audit ledger
- wire the new audit into architecture docs and public site navigation
- update the track document to reflect that this audit now exists

## Impact

- `docs/architecture/pricing-negotiation-settlement-audit.md`
- architecture landing and track docs
- docs navigation
```

Create `openspec/changes/pricing-negotiation-settlement-audit/specs/project-docs/spec.md`:

```md
## ADDED Requirements

### Requirement: Pricing negotiation settlement audit is published as an architecture audit
The architecture docs SHALL include a dedicated audit ledger for pricing surface, negotiation, settlement, and escrow under the P2P knowledge-exchange track.

#### Scenario: Audit ledger exists
- **WHEN** a reader opens the architecture docs
- **THEN** they SHALL find a dedicated `pricing-negotiation-settlement-audit.md` page
```

Create `openspec/changes/pricing-negotiation-settlement-audit/specs/docs-only/spec.md`:

```md
## ADDED Requirements

### Requirement: Architecture landing and track docs reference the pricing settlement audit
The architecture landing page and P2P knowledge-exchange track document SHALL reference the new pricing/negotiation/settlement audit.

#### Scenario: Architecture landing links the audit
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** the new audit SHALL appear alongside the other architecture audit pages

#### Scenario: Track doc reflects landed audit
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** the pricing/negotiation/settlement audit SHALL be described as landed work with follow-on design still remaining
```

- [ ] **Step 2: Sync the main specs**

Run:

```bash
cp openspec/changes/pricing-negotiation-settlement-audit/specs/project-docs/spec.md openspec/specs/project-docs/spec.md
cp openspec/changes/pricing-negotiation-settlement-audit/specs/docs-only/spec.md openspec/specs/docs-only/spec.md
```

- [ ] **Step 3: Archive the change**

Run:

```bash
mkdir -p openspec/changes/archive
mv openspec/changes/pricing-negotiation-settlement-audit openspec/changes/archive/2026-04-21-pricing-negotiation-settlement-audit
git add openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-21-pricing-negotiation-settlement-audit
git -c commit.gpgsign=false commit -m "specs: archive pricing settlement audit"
```

## Self-Review

- Spec coverage:
  - audit ledger creation: Task 1
  - public architecture wiring: Task 2
  - verification: Task 3
  - OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or deferred implementation markers remain
- Type/path consistency:
  - internal planning artifacts stay under `internal-docs/`
  - public docs stay under `docs/`
  - the new audit page path is consistently `docs/architecture/pricing-negotiation-settlement-audit.md`
