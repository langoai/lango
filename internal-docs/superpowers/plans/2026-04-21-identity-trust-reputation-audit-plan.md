# Identity Trust Reputation Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a detailed `identity / trust / reputation` audit ledger for `knowledge exchange v1`, aligned with the master document and existing audit framework.

**Architecture:** Build the audit as a documentation-first slice under `docs/architecture/`, using the existing audit-ledger style from `external-collaboration-audit.md` and `trust-security-policy-audit.md`. The work stays documentation-only, but it must be grounded in current code and docs surfaces and then closed out through OpenSpec.

**Tech Stack:** Markdown, existing architecture audit docs, current P2P/auth/trust/reputation code, OpenSpec

---

## File Map

- Create: `docs/architecture/identity-trust-reputation-audit.md`
  - New detailed audit ledger for the identity/trust/reputation surface.
- Modify: `docs/architecture/index.md`
  - Add the new audit to the architecture landing page if appropriate.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Update the follow-on plan status if the new audit becomes a landed reference.
- Modify: `mkdocs.yml` or `zensical.toml`
  - Add the new audit page to the public Architecture navigation if the architecture section is intended to surface it publicly.
- Create: `openspec/changes/identity-trust-reputation-audit/**`
  - Proposal, design, tasks, and delta specs for this audit-slice documentation work.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync requirements for the new audit document and any architecture index updates.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync requirements if the public architecture docs/navigation are updated.

### Task 1: Draft The Detailed Audit Ledger

**Files:**
- Create: `docs/architecture/identity-trust-reputation-audit.md`

- [ ] **Step 1: Write the failing audit-outline check as a manual comparison**

Run:

```bash
sed -n '1,260p' docs/architecture/external-collaboration-audit.md
sed -n '1,260p' docs/architecture/trust-security-policy-audit.md
sed -n '1,260p' internal-docs/superpowers/specs/2026-04-21-identity-trust-reputation-audit-design.md
```

Expected:

```text
The two existing audit ledgers define the expected structure, and the new design defines the intended relationship model.
```

- [ ] **Step 2: Create the audit document skeleton**

Create `docs/architecture/identity-trust-reputation-audit.md` with this structure:

```md
# Identity Trust Reputation Audit

## Purpose

## Relationship to the Master Document

## Document Ownership

## Audit Order

1. Identity Continuity
2. Trust Entry
3. Reputation
4. Revocation & Trust Decay

## Audit Method

## Current Surface Map

## Baseline Relationship Model

## Detailed Audit: Identity Continuity

## Detailed Audit: Trust Entry

## Detailed Audit: Reputation

## Detailed Audit: Revocation & Trust Decay

## Assessment

## Follow-On Design Inputs
```

- [ ] **Step 3: Fill in document ownership, method, and current surface map**

Use the same ledger style as the existing audit docs. The ownership section should read:

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
| Identity Continuity | Phase 1-2 | `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `docs/gateway/http-api.md`, `internal/p2p/identity/*`, `internal/p2p/handshake/*`, `internal/app/wiring_p2p.go` | Detailed audit complete (`stabilize`) |
| Trust Entry | Phase 1-2 | `docs/security/authentication.md`, `docs/features/p2p-network.md`, `internal/gateway/auth.go`, `internal/p2p/firewall/*`, `internal/p2p/handshake/security_events.go`, `internal/p2p/paygate/*` | Detailed audit complete (`stabilize`) |
| Reputation | Phase 1-2 | `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `internal/p2p/reputation/*`, `internal/app/p2p_routes.go`, `internal/p2p/team/payment.go` | Detailed audit complete (`stabilize`) |
| Revocation & Trust Decay | Phase 1-2 | `docs/features/p2p-network.md`, `internal/p2p/handshake/security_events.go`, `internal/p2p/discovery/gossip.go`, `internal/p2p/reputation/*` | Detailed audit complete (`stabilize`) |
```

- [ ] **Step 4: Write the relationship model and row-level audit records**

The audit content must explicitly lock these ideas from the design:

```md
- owner-root trust provides bootstrap ceiling/floor
- agent/domain reputation is earned from actual exchange history
- admission trust and payment trust are separate gates
- operational signals and durable negative reputation are separate concepts
- new agents begin under constrained trust even under a trusted owner
```

For each row, include:

- Audit Record
- Findings
- Assessment

Each row should start with:

```md
### Audit Record

- Feature name: `Identity Continuity`
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

All four rows remain `stabilize`: the capability family is real, but the operator-facing relationship model still needs consolidation.

## Follow-On Design Inputs

1. `reputation v2`
2. `pricing / negotiation / settlement` audit
3. `knowledge exchange runtime` end-to-end design
```

- [ ] **Step 6: Verify the document renders and reads coherently**

Run:

```bash
sed -n '1,320p' docs/architecture/identity-trust-reputation-audit.md
```

Expected:

```text
The file reads as a complete audit ledger with no placeholders and a consistent relationship model.
```

- [ ] **Step 7: Commit the audit ledger**

Run:

```bash
git add docs/architecture/identity-trust-reputation-audit.md
git -c commit.gpgsign=false commit -m "docs: add identity trust reputation audit"
```

### Task 2: Wire The Audit Into Public Architecture Docs

**Files:**
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`

- [ ] **Step 1: Add the new audit to the architecture landing page**

Update `docs/architecture/index.md` with a new entry consistent with the existing style:

```md
-   :material-account-check-outline: **[Identity Trust Reputation Audit](identity-trust-reputation-audit.md)**
    Audit ledger for identity continuity, trust entry, reputation, and revocation in `knowledge exchange v1`.
```

- [ ] **Step 2: Update the track doc to reflect the landed audit**

In `docs/architecture/p2p-knowledge-exchange-track.md`, update the follow-on list so the identity/trust/reputation audit is no longer framed as not yet started. Replace the current item with a note that the detailed audit now exists and the follow-on work is `reputation v2` plus runtime integration.

Use wording like:

```md
1. `identity / trust / reputation` detailed audit is now landed; the follow-on work is `reputation v2`, stronger trust-entry contracts, and runtime integration
```

- [ ] **Step 3: Add the new audit page to the public Architecture nav**

In `zensical.toml`, add:

```toml
{ "Identity Trust Reputation Audit" = "architecture/identity-trust-reputation-audit.md" }
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
git -c commit.gpgsign=false commit -m "docs: wire identity trust audit into architecture docs"
```

### Task 3: Final Verification

**Files:**
- No new file changes required unless verification reveals a doc wording problem

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

- [ ] **Step 2: Inspect the architecture docs build result manually**

Run:

```bash
rg -n "Identity Trust Reputation Audit" docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml
```

Expected:

```text
The new audit is referenced consistently in the architecture landing page, the track document, and the site nav.
```

### Task 4: OpenSpec Change, Main Spec Sync, And Archive

**Files:**
- Create: `openspec/changes/identity-trust-reputation-audit/proposal.md`
- Create: `openspec/changes/identity-trust-reputation-audit/design.md`
- Create: `openspec/changes/identity-trust-reputation-audit/tasks.md`
- Create: `openspec/changes/identity-trust-reputation-audit/specs/project-docs/spec.md`
- Create: `openspec/changes/identity-trust-reputation-audit/specs/docs-only/spec.md`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`

- [ ] **Step 1: Write the OpenSpec change artifacts**

Create `openspec/changes/identity-trust-reputation-audit/proposal.md`:

```md
## Why

The P2P knowledge-exchange track still lists `identity / trust / reputation` as a major follow-on area, but the relationship model has not been written down in a dedicated audit ledger. That leaves later pricing, settlement, and runtime design without one explicit audit baseline.

## What Changes

- add a detailed `identity / trust / reputation` audit ledger
- wire the new audit into architecture docs and public site navigation
- update the track document to reflect that this detailed audit now exists

## Impact

- `docs/architecture/identity-trust-reputation-audit.md`
- architecture landing and track docs
- docs navigation
```

Create `openspec/changes/identity-trust-reputation-audit/specs/project-docs/spec.md`:

```md
## ADDED Requirements

### Requirement: Identity trust reputation audit is published as an architecture audit
The architecture docs SHALL include a dedicated audit ledger for identity continuity, trust entry, reputation, and revocation under the P2P knowledge-exchange track.

#### Scenario: Audit ledger exists
- **WHEN** a reader opens the architecture docs
- **THEN** they SHALL find a dedicated `identity-trust-reputation-audit.md` page
```

Create `openspec/changes/identity-trust-reputation-audit/specs/docs-only/spec.md`:

```md
## ADDED Requirements

### Requirement: Architecture landing and track docs reference the identity trust audit
The architecture landing page and P2P knowledge-exchange track document SHALL reference the new identity/trust/reputation audit.

#### Scenario: Architecture landing links the audit
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** the new audit SHALL appear alongside the other architecture audit pages

#### Scenario: Track doc reflects landed audit
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** the identity/trust/reputation audit SHALL be described as landed work with follow-on design still remaining
```

- [ ] **Step 2: Sync the main specs**

Run:

```bash
cp openspec/changes/identity-trust-reputation-audit/specs/project-docs/spec.md openspec/specs/project-docs/spec.md
cp openspec/changes/identity-trust-reputation-audit/specs/docs-only/spec.md openspec/specs/docs-only/spec.md
```

- [ ] **Step 3: Archive the change**

Run:

```bash
mkdir -p openspec/changes/archive
mv openspec/changes/identity-trust-reputation-audit openspec/changes/archive/2026-04-21-identity-trust-reputation-audit
git add openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-21-identity-trust-reputation-audit
git -c commit.gpgsign=false commit -m "specs: archive identity trust reputation audit"
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
  - all internal planning artifacts stay under `internal-docs/`
  - public docs stay under `docs/`
  - the new audit page path is consistently `docs/architecture/identity-trust-reputation-audit.md`
