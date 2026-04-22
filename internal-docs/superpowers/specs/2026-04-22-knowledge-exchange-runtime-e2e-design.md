# Knowledge Exchange Runtime E2E Design

## Purpose

This document defines the end-to-end runtime framing for `knowledge exchange v1`.

The current system already has multiple landed first slices:

- exportability evaluation,
- artifact release approval,
- upfront payment approval,
- direct prepay execution gating,
- dispute-ready receipts,
- escrow recommendation execution.

What is still missing is a single transaction-oriented runtime model that explains how those slices compose into one coherent exchange path.

This design provides that model.

## Scope

This design covers:

- transaction opening,
- exportability advisory and authoritative checkpoints,
- upfront payment approval,
- direct payment or escrow path selection,
- work start conditions,
- artifact submission,
- artifact release approval,
- settlement progression,
- dispute handoff.

This design does not yet cover:

- a full dispute engine,
- human approval UI,
- full escrow lifecycle beyond `create + fund`,
- long-running multi-session project runtime.

The target remains `knowledge exchange v1`, not a general long-running collaboration runtime.

## Recommended Runtime Model

The preferred model is **Transaction-Orchestrated Runtime**.

The basic idea is:

- the transaction is the top-level runtime unit,
- the transaction receipt is the canonical control-plane state,
- the submission receipt is the canonical deliverable state,
- runtime branches are selected from transaction state,
- review, revision, and dispute handoff are driven through the current submission inside the same transaction.

This is preferred over a submission-first or workflow-first model because the currently landed slices already center around transaction receipts, payment approval, settlement hints, and escrow execution status.

## Canonical Runtime Model

The runtime should use:

- `transaction / deal` as the top-level unit,
- `transaction receipt` as the canonical control-plane state,
- `submission receipt` as the canonical deliverable state.

`transaction receipt` owns:

- transaction-level control state,
- payment approval state,
- settlement hint,
- escrow execution state,
- current submission pointer,
- readiness for settlement progression or dispute handoff.

`submission receipt` owns:

- artifact-level progression,
- exportability evidence,
- approval outcome,
- fulfillment assessment,
- deliverable history and event trail.

This means the runtime is neither purely payment-first nor purely artifact-first. It is transaction-first at the top level, with deliverable progression nested beneath it.

## Transaction Open

The start of the runtime is when the leader opens a transaction with an external counterparty.

At minimum, transaction open should carry canonical inputs for:

- `counterparty`
- `requested scope`
- `price context`
- `trust context`

These inputs establish the control-plane baseline before work begins.

## Exportability Placement

Exportability appears twice in the runtime.

### Advisory phase

Immediately after transaction open, the system should evaluate exportability in advisory form.

Purpose:

- early direction correction,
- avoid drifting into work that cannot later be released.

### Authoritative phase

Immediately before artifact release, the system should evaluate exportability authoritatively.

Purpose:

- determine whether the actual submission may cross the release boundary,
- supply canonical evidence for release approval and later dispute handling.

This keeps exportability as both an early steering mechanism and a final release gate.

## Upfront Payment Approval Placement

Upfront payment approval belongs after transaction open and exportability advisory, but before work starts.

Its role is to determine the initial payment mode for the transaction:

- `prepay`
- `escrow`
- `reject`
- `escalate`

This is a transaction-level decision, not a deliverable-level decision.

## Runtime Branch Selection

After upfront payment approval, the runtime branches by settlement path.

### Prepay branch

- chosen when the transaction is approved for direct prepay
- work may begin only after direct payment execution is authorized

### Escrow branch

- chosen when the transaction is approved for escrow
- work may begin only after the first escrow execution slice completes `create + fund`

This preserves a real financial gate before work starts, rather than treating payment mode as a passive recommendation.

## Work Start Conditions

The work start gate is intentionally strict.

- `prepay` path: start only after payment execution is authorized
- `escrow` path: start only after escrow reaches funded state

This means the runtime does not allow work to start merely because approval recommended a payment mode. The recommendation must be realized into an actual payment or escrow execution state first.

## Artifact Submission

Work output becomes canonical only when it is turned into a submission receipt.

This means artifact submission is not just “a blob exists.” It is:

- the creation of a new submission receipt,
- the recording of submission-level evidence,
- the possible update of the transaction’s current submission pointer.

The submission receipt is therefore the canonical representation of deliverable progression.

## Artifact Release Approval

Artifact release approval drives the next runtime decision.

Its outcomes remain:

- `approve`
- `reject`
- `request-revision`
- `escalate`

These outcomes are not just passive labels. They move the runtime into one of the next branches.

## Revision Branch

When release approval returns `request-revision`, the runtime stays inside the same transaction.

The model is:

- keep the existing transaction,
- create a new submission,
- update the transaction’s current submission pointer,
- preserve older submissions as history.

Revision is therefore a same-transaction refinement path, not a new deal.

## Settlement Progression

When release approval returns `approve`, the runtime enters settlement progression.

This progression must remain distinct from:

- initial payment-mode approval,
- direct payment execution gating,
- escrow execution lifecycle.

The current design does not yet define the final detailed settlement progression mechanics, but it locks the fact that settlement progression is its own end-to-end runtime phase after release approval.

## Dispute Handoff

Dispute handoff should not open merely because a submission was rejected.

It should open only when there is a concrete disagreement that cannot be resolved within normal approval or revision flow.

Examples:

- settlement disagreement,
- fulfillment disagreement,
- policy disagreement.

This keeps dispute handoff narrower than generic failure.

## Runtime Branches Summary

The end-to-end flow is:

1. leader opens transaction
2. exportability advisory
3. upfront payment approval
4. runtime path selection
   - `prepay`
   - `escrow`
5. payment/escrow gate satisfied
6. work execution
7. artifact submission
8. artifact release approval
9. next branch:
   - settlement progression
   - revision
   - dispute handoff

## State Ownership

The state ownership model is:

### Transaction receipt

Owns:

- transaction-level control plane,
- payment mode state,
- settlement hint,
- escrow execution state,
- current submission pointer,
- higher-level progression branch.

### Submission receipt

Owns:

- deliverable truth,
- exportability evidence,
- release approval outcome,
- fulfillment evidence,
- revision/dispute-relevant history.

This separation lets the runtime stay transaction-oriented while still keeping submission-level evidence explicit and durable.

## Follow-On Implementation Inputs

This design should feed directly into:

### 1. Transaction-open canonical model

The system still needs a concrete transaction-open representation that binds:

- counterparty,
- scope,
- price context,
- trust context

into the runtime from the start.

### 2. Runtime orchestration layer

The system still needs the actual orchestration layer that selects:

- prepay path,
- escrow path,
- revision path,
- dispute handoff path.

### 3. Settlement progression design

The system still needs a concrete design for:

- post-approval settlement progression,
- partial settlement,
- revision/reject aftermath,
- dispute transition mechanics.

### 4. Escrow lifecycle completion

The system still needs follow-on work for:

- activate,
- release,
- refund,
- dispute,
- escrow completion semantics.

## Deliverable Expectation

The final document that follows from this design should be a full end-to-end runtime design document for `knowledge exchange v1`.

It should remain design-level and not yet expand into the full implementation plan.
