# Approval Flow Design

## Purpose

Define the first explicit `approval flow` model for `knowledge exchange v1`.

This design answers the next boundary question after exportability:

- what must be approved,
- who can approve it,
- what states an approval can return,
- how approval results affect settlement,
- and what evidence must be recorded for later settlement or dispute handling.

This document is subordinate to:

- `docs/architecture/master-document.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `docs/architecture/trust-security-policy-audit.md`
- `internal-docs/superpowers/specs/2026-04-20-exportability-policy-design.md`

## Scope

This design covers:

- approval objects for `knowledge exchange v1`,
- structured approval states,
- release-decision inputs,
- escalation boundaries,
- settlement coupling,
- structured release outcome records.

This design does not yet cover:

- human approval UI,
- complete dispute workflow,
- milestone approval for long-running projects,
- smart-contract-native approval execution,
- approval analytics or operator dashboards.

## Problem Statement

`knowledge exchange v1` now has a first exportability slice, but it still needs an explicit approval model.

Without this, Lango cannot clearly answer:

- what gets approved before work starts,
- what gets approved before release,
- when a leader agent may decide alone,
- when a human must be involved,
- and what structured record drives refund, partial settlement, revision, or escalation.

The current tool approval infrastructure is real, but it is still general-purpose tool execution approval. `knowledge exchange v1` needs a product-level approval flow for market artifacts and their release.

## Approaches Considered

### Approach A: Minimal Binary Approval

Use only two decisions:

- approve
- reject

Pros:

- smallest implementation,
- easy to explain,
- easy to connect to release gating.

Cons:

- no revision state,
- no escalation state,
- too weak for settlement and dispute coupling.

### Approach B: Structured Approval

Use a narrow but structured state model:

- approve
- reject
- request-revision
- escalate

Pros:

- strong enough for `knowledge exchange v1`,
- keeps release and settlement aligned,
- supports human escalation without requiring a full dispute engine.

Cons:

- more state and record design than a binary model,
- requires structured outcome records.

### Approach C: Dispute-Native Approval

Treat approval as a thin front-end for a fuller dispute system from the start.

Pros:

- future-ready,
- naturally aligns with long-running collaboration.

Cons:

- too large for the next slice,
- mixes approval, settlement, and dispute too early.

## Recommendation

Use **Approach B: Structured Approval**.

This is the smallest model that still supports:

- delegated approval by the leader agent,
- human escalation for true boundary cases,
- structured revision/rejection handling,
- and later linkage to settlement and dispute-ready receipts.

## Approval Objects

`knowledge exchange v1` has two approval objects:

1. `Upfront Payment Approval`
2. `Artifact Release Approval`

### Upfront Payment Approval

This approval opens the transaction.

Minimum input:

- amount,
- counterparty,
- trust / reputation,
- requested scope,
- current budget / policy limit.

Default approver:

- `leader agent`

Escalation to human happens when:

- amount exceeds configured threshold,
- counterparty trust conditions are not met,
- user policy or budget limits are exceeded,
- or the transaction context is especially sensitive.

### Artifact Release Approval

This is the primary approval object for `knowledge exchange v1`.

Minimum input:

- artifact itself,
- exportability receipt,
- requested deliverable scope,
- payment / transaction context.

Default approver:

- `leader agent`

Escalation to human happens when:

- exportability is `needs-human-review`,
- a blocked artifact is being overridden,
- or the transaction is high-value, high-risk, or highly sensitive.

## Decision Flow

The default flow is:

1. `upfront payment approval`
2. `artifact work`
3. `artifact release approval`
4. `approve` triggers automatic settlement

This means the main decision gate is not “may work start,” but “may this artifact now be released and settled.”

## Decision States

`artifact release approval` returns one of four states:

- `approve`
- `reject`
- `request-revision`
- `escalate`

### Approve

The artifact is approved for release and settlement may proceed automatically.

### Reject

The submitted artifact is not accepted for this transaction in its current form.

This state does not by itself imply:

- full refund,
- no refund,
- or automatic dispute.

Those depend on transaction terms and fulfillment assessment.

### Request-Revision

The artifact is not yet releasable, but the transaction remains alive.

This state is used when the submission is deficient but still within a repairable transaction path.

Typical causes:

- scope mismatch,
- quality issue,
- missing supporting policy or lineage record.

### Escalate

The leader agent must not decide alone.

This state is reserved for:

- exportability `needs-human-review`,
- blocked override attempts,
- high-value / high-risk / high-sensitivity cases.

## Release Outcome Record

Every `reject` or `request-revision` result must emit a structured outcome record.

At minimum, it must include:

- `decision state`,
- `reason`,
- `issue classification`,
- `fulfillment assessment`,
- `settlement hint`.

### Issue Classification

At minimum:

- `scope mismatch`
- `quality issue`
- `policy issue`

### Fulfillment Assessment

Use a mixed model:

- qualitative first:
  - `none`
  - `partial`
  - `substantial`
- optional quantitative ratio:
  - `0.0 .. 1.0`

This keeps `v1` simple but gives later settlement logic structured input.

### Settlement Hint

This is not final settlement execution.

It is a structured recommendation that helps downstream logic decide whether to:

- refund,
- partially release,
- hold,
- or escalate to dispute.

## Settlement Coupling

The approval flow is directly coupled to settlement.

### Approve Path

- `approve` => automatic settlement

### Reject / Revision Path

- no universal default refund rule,
- no universal default forfeiture rule,
- settlement handling depends on transaction terms and fulfillment assessment.

That means approval does more than decide release permission. It also produces structured input for downstream economic handling.

## Escalation Rules

Human escalation is intentionally narrow.

It is allowed when:

- exportability is `needs-human-review`,
- a blocked artifact is being considered for one-time override,
- or the transaction crosses configured risk / value / sensitivity thresholds.

Human escalation is not the default path.

It is the boundary case path.

## Non-Goals

This design intentionally does not claim that:

- every release rejection should trigger dispute,
- settlement rules are fully automated,
- human override becomes a reusable permanent policy,
- or the general-purpose tool approval system is enough by itself for product-level artifact approval.

## Initial Success Criteria

This design is successful if a future implementation can:

- model the two approval objects explicitly,
- treat `artifact release approval` as the primary approval gate,
- emit `approve / reject / request-revision / escalate`,
- record structured release outcomes for reject/revision,
- and connect `approve` to automatic settlement without pretending that the full dispute system is already built.

## Follow-On Planning Inputs

The next implementation plan should define:

1. where approval objects and states are stored
2. how `artifact release approval` consumes exportability receipts
3. what the structured release outcome record looks like in code and storage
4. how `approve` triggers automatic settlement
5. how `reject / request-revision` outcomes provide settlement hints without overbuilding dispute logic
6. how later human escalation hooks in without requiring a full UI in the first slice
