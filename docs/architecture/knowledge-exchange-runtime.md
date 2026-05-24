# Knowledge Exchange Runtime

## Purpose

This document records the first transaction-oriented runtime control plane for `knowledge exchange v1`.

It is intentionally narrow: the goal is to describe how the landed first slices compose into one runtime story, not to claim a broader runtime implementation than the codebase currently provides.

## Relationship to the Master Document

This page sits underneath `docs/architecture/master-document.md` and follows its constitution, capability taxonomy, and track-routing rules.

It does not create a new top-level product area. It documents the runtime slice that connects the existing knowledge-exchange work into one control-plane narrative for the `P2P Knowledge Exchange Track`.

## Document Ownership

- Primary capability area: `External Collaboration & Economic Exchange`
- Primary execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Stabilization Track`

## Runtime Design Slice

The first runtime slice is transaction-oriented and receipt-centered.

It ties together six steps:

1. Open the transaction and bind canonical inputs.
2. Select the payment path from the current trust and approval state.
3. Gate work start until exportability and payment conditions are satisfied.
4. Create the submission receipt for the deliverable.
5. Approve, reject, revise, or escalate the release.
6. Advance post-approval state after the release outcome is known.

The canonical state model is simple:

- `transaction receipt` is the runtime control-plane record.
- `submission receipt` is the canonical deliverable record.
- existing approval, payment, exportability, and escrow slices remain the source of truth for their own decisions.

The design explicitly reuses the landed first slices instead of inventing parallel logic:

- exportability evaluation,
- artifact release approval,
- upfront payment approval,
- dispute-ready receipt creation,
- direct prepay execution gating,
- escrow recommendation execution.

## Current Limits

This is the first design slice, not a declaration of a finished runtime subsystem.

- No dedicated `internal/knowledgeruntime` package is documented here as a completed implementation.
- No human approval UI is described as part of this slice.
- No dispute orchestration is folded into this page.
- No generalized team execution or shared-workspace runtime is implied.
- No broader settlement lifecycle is claimed beyond the landed direct prepay and first escrow `create + fund` paths.
- No replacement receipt model is introduced; the existing receipt-backed slices remain canonical for their own responsibilities.

## Follow-On Work

The next work after this slice is runtime implementation, not redesign of the receipt model.

- minimal orchestration wiring for transaction open and runtime branching,
- broader progression handling after approval or escalation,
- deeper provenance and dispute integration,
- continued settlement progression beyond the landed first paths.

