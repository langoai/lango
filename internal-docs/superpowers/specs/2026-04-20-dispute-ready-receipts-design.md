# Dispute-Ready Receipts Design

## Purpose

Define the first explicit `dispute-ready receipt` model for `knowledge exchange v1`.

This design answers the next boundary question after exportability and approval flow:

- what evidence bundle represents a submitted artifact,
- how current canonical state and event history coexist,
- how exportability, approval, settlement, and provenance are connected,
- and what minimum structure is needed before a full dispute system exists.

This document is subordinate to:

- `docs/architecture/master-document.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `docs/architecture/trust-security-policy-audit.md`
- `internal-docs/superpowers/specs/2026-04-20-exportability-policy-design.md`
- `internal-docs/superpowers/specs/2026-04-20-approval-flow-design.md`

## Scope

This design covers:

- `submission receipt`
- `transaction receipt`
- canonical state vs append-only event trail
- exportability, approval, settlement, and provenance linkage
- lite provenance summary and external references

This design does not yet cover:

- full dispute adjudication
- human dispute UI
- complete settlement execution orchestration
- full provenance embedding
- generalized evidence graph

## Problem Statement

Lango now has:

- source-primary exportability decisions
- structured artifact release approval decisions

What it still lacks is one durable receipt model that can explain, after the fact:

- what was submitted,
- what policy state applied,
- what approval state applied,
- what settlement state applied,
- and which provenance or audit references support the current canonical view.

Without that, later disputes would have to reconstruct evidence from loosely connected audit rows, provenance records, and settlement data.

## Approaches Considered

### Approach A: Receipt-As-Audit

Treat audit rows themselves as the receipt system.

Pros:

- smallest implementation,
- reuses an existing append-only store.

Cons:

- no clear canonical current state,
- hard to query as a transaction-level object,
- weak foundation for later dispute workflows.

### Approach B: Canonical Receipt + Event Trail Hybrid

Use dedicated receipt records for canonical state while preserving append-only event history.

Pros:

- separates “current truth” from “state transition history”
- naturally fits submission vs transaction hierarchy
- works well with later dispute handling

Cons:

- more model design than audit-only
- requires explicit references to audit/provenance/settlement

### Approach C: Full Evidence Graph

Model receipts, provenance, settlement, and disputes as one unified graph immediately.

Pros:

- strongest long-term shape
- minimal future migration

Cons:

- far too large for the next slice
- forces dispute engine design too early

## Recommendation

Use **Approach B: Canonical Receipt + Event Trail Hybrid**.

For the first slice, make it a `lite` version:

- separate `submission receipt` and `transaction receipt`
- canonical current state on the receipt records
- append-only event trail for state transitions
- references out to audit/provenance/settlement rather than full embedding

## Core Model

### Submission Receipt

The primary unit is the `artifact submission receipt`.

It represents one submitted artifact and its current canonical state.

Minimum fields:

- `submission_receipt_id`
- `transaction_receipt_id`
- `artifact_fingerprint`
- `exportability_decision`
- `approval_outcome`
- `settlement_context`
- `provenance_summary`
- `provenance_reference`
- `event_trail`

### Transaction Receipt

The `transaction receipt` is the higher-level canonical record for the whole exchange.

Minimum fields:

- `transaction_receipt_id`
- `current_submission_receipt_id`
- `canonical_decision`
- `canonical_settlement_status`
- submission summary / aggregation
- counterparty / deal reference

The relationship is bidirectional:

- a submission points to its parent transaction
- a transaction points to its current canonical submission

## Artifact Fingerprint

The receipt must not fingerprint only the payload bytes.

The minimum fingerprint basis is:

- `payload hash`
- `artifact label / scope metadata`
- `source lineage digest`

This allows the system to distinguish:

- same payload with different source lineage
- same label with different lineage
- same transaction with multiple revised submissions

## Exportability Section

The receipt includes the full first-slice exportability decision:

- `state`
- `policy_code`
- `explanation`
- `lineage_summary`

This is rich enough for a first-pass dispute or operator review without requiring a second fetch just to understand the policy outcome.

## Approval Section

The receipt includes the full first-slice approval outcome:

- `decision`
- `reason`
- `issue_class`
- `fulfillment_assessment`
- `settlement_hint`

The canonical submission approval status SHALL be one of:

- `pending`
- `approved`
- `rejected`
- `revision-requested`
- `escalated`

## Settlement Section

The receipt includes settlement context rather than full settlement execution state.

Minimum fields:

- `transaction_or_deal_id`
- `amount`
- `upfront_status`
- `residual_status`
- `counterparty`
- `current_settlement_hint`
- `onchain_or_offchain_reference`

The canonical transaction settlement status SHALL be one of:

- `pending`
- `partially-settled`
- `settled`
- `disputed`

## Provenance Section

The provenance strategy is hybrid.

The receipt embeds only a lightweight summary and links outward for deeper evidence.

Minimum provenance summary:

- `checkpoint_or_provenance_reference_id`
- `config_fingerprint`
- `signer_summary`
- `attribution_summary`

This makes the receipt dispute-ready without turning it into a full provenance bundle.

## Canonical View And Event Trail

The model is explicitly hybrid:

- canonical receipt records are mutable
- state changes are append-only

This gives two views:

1. `canonical view`
   - what the current transaction/submission state is
2. `event trail`
   - how it got there

The event trail must at minimum capture transitions for:

- draft exportability
- final exportability
- approval
- settlement
- escalation / dispute

Draft-state advisory exportability belongs in the trail, but the canonical view should emphasize the final state that actually affected release and settlement.

## Submission Lifecycle

The first slice assumes:

- many historical submissions may exist for one transaction
- only one `current submission` may be canonical at a time

Older submissions remain as superseded history rather than disappearing.

## Transaction Lifecycle

The transaction receipt maintains:

- the canonical current submission pointer
- the current canonical decision
- the current canonical settlement status

The canonical decision is normally updated by the leader agent, but once a human escalation or dispute system takes over, the higher authority becomes the canonical writer.

## Non-Goals

This design intentionally does not claim that:

- the first slice fully resolves disputes
- audit rows alone are enough as receipts
- the receipt embeds all provenance details
- settlement logic is fully automated by this model

## Initial Success Criteria

This design is successful if a future implementation can:

- create separate `submission` and `transaction` receipts
- maintain one current submission pointer per transaction
- persist canonical approval and settlement status
- store event trail entries for policy/approval/settlement transitions
- embed lite provenance summary and connect external references

## Follow-On Planning Inputs

The next implementation plan should define:

1. receipt storage schema
2. submission vs transaction record ownership and linkage
3. canonical receipt mutation rules
4. append-only event trail model
5. reference strategy for audit, provenance, and settlement
6. how later dispute workflows consume these receipts
