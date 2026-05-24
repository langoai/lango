# Pricing Negotiation Settlement Audit Design

## Purpose

This document defines the audit framing for the `pricing / negotiation / settlement` capability family under the `P2P Knowledge Exchange Track`.

Its purpose is not to produce a full runtime implementation plan yet. Instead, it establishes the control-plane baseline for `knowledge exchange v1` by clarifying:

- what the public pricing surface is,
- what role negotiation actually plays,
- what belongs to settlement versus escrow,
- what is already landed and what remains a follow-on gap.

This audit is intended to become the next detailed audit ledger after the new `identity / trust / reputation` audit.

## Scope

This audit is judged against `knowledge exchange v1`.

It directly covers:

- pricing surface,
- negotiation,
- settlement,
- escrow.

It does not directly design:

- a full dispute engine,
- the final smart contract model,
- full team-execution settlement,
- long-running multi-milestone settlement orchestration,
- a complete end-to-end runtime implementation.

Those remain follow-on work. This audit exists to define the operator-facing control-plane model that later runtime and contract work must inherit.

## Recommended Structure

The audit should use four rows:

1. `Pricing Surface`
2. `Negotiation`
3. `Settlement`
4. `Escrow`

This is the preferred structure because it matches the current product boundary and the landed first slices more clearly than an implementation-module breakdown.

## Baseline Control-Plane Model

The following control-plane model should be explicitly locked by the audit.

### 1. `p2p.pricing` vs `economy.pricing`

These are distinct surfaces with distinct responsibilities.

- `p2p.pricing` is the provider-side public quote surface
- `economy.pricing` is the local pricing and policy engine

They should not be treated as duplicate implementations of the same public API.

### 2. Negotiation is Real but Under-Surfaced

Negotiation already exists as a capability.

But it is not yet a strongly surfaced part of the operator-facing `knowledge exchange v1` transaction story.

So the correct current model is:

- negotiation is real,
- negotiation is not yet the primary public transaction surface,
- negotiation still needs stronger operator-facing articulation.

### 3. Settlement and Escrow Are Distinct

The audit should keep `Settlement` and `Escrow` as separate rows.

`Settlement` covers:

- upfront payment approval,
- direct prepay execution gating,
- final settlement semantics.

`Escrow` covers:

- escrow recommendation,
- escrow execution,
- explicit lifecycle gaps after execution.

This distinction is important because the landed slices are different in maturity and progression.

### 4. Off-Chain Accrual / Postpay Is Phase 2 and Trust-Conditional

Postpay and off-chain accrual are not removed, but they should be described as:

- Phase 2 capabilities,
- trust-conditional,
- currently limited rather than fully generalized.

They should not be described as fully mature general defaults for the system.

## Current Surface Map

The current code and documentation already suggest a substantial real surface:

- public quote endpoints and CLI pricing views,
- local economy pricing logic,
- negotiation engines and wiring,
- upfront payment approval and direct prepay execution gating,
- escrow recommendation and first execution slice,
- settlement semantics scattered across docs and control-plane components.

This means the audit is not deciding whether pricing/settlement exists. It is deciding how these pieces should be understood as one control-plane story.

## Detailed Audit Rows

### Pricing Surface

Initial judgment: `stabilize`

Why:

- the public quote surface is already real
- the internal pricing policy engine is also real
- but the relationship between the two is not yet consistently understood as public quote versus local policy

The likely conclusion is that pricing stays as two distinct but legitimate surfaces that need clearer operator framing.

### Negotiation

Initial judgment: `stabilize`

Why:

- negotiation is implemented and wired
- but it remains under-surfaced in the operator-facing transaction story

The likely conclusion is that negotiation should be kept, but better articulated as a real capability that is not yet the dominant public interface.

### Settlement

Initial judgment: `stabilize`

Why:

- multiple first slices already landed:
  - upfront payment approval
  - direct prepay execution gating
- but final settlement progression still has explicit runtime and dispute gaps

The likely conclusion is that settlement is already a real control plane worth keeping, but still incomplete in progression semantics.

### Escrow

Initial judgment: `stabilize`

Why:

- escrow is already a real engine-level capability
- the `knowledge exchange` path now has a landed first slice for `create + fund`
- but `activate`, `release`, `refund`, and `dispute` remain follow-on work

The likely conclusion is that escrow is no longer merely planned, but is still only partially surfaced for the current product path.

## Assessment Shape

The likely assessment pattern for this audit is:

- all four rows are real capabilities,
- none are immediate `remove` or `defer` candidates,
- the main need is control-plane clarity and runtime progression clarity rather than net-new conceptual invention.

That points toward a `stabilize` judgment across the whole surface, unless deeper exploration reveals a row that is actually a merge candidate.

## Follow-On Design Inputs

This audit is expected to feed directly into three follow-on areas.

### 1. Knowledge Exchange Runtime Design

The runtime design should inherit:

- public quote surface versus internal pricing policy,
- negotiation as a real but not yet dominant operator surface,
- settlement progression and escrow progression as separate but connected paths.

### 2. Settlement Follow-On Work

This audit should supply a baseline for later work on:

- final settlement progression,
- revision / reject / escalate aftermath,
- partial settlement behavior,
- dispute handoff.

### 3. Escrow Lifecycle Completion

This audit should supply a baseline for later work on:

- activate,
- release,
- refund,
- dispute,
- fuller escrow lifecycle semantics for `knowledge exchange`.

## Deliverable Expectation

The final audit document that follows from this design should be a detailed audit ledger plus follow-on design inputs.

It should not yet expand into a full implementation plan or a complete transaction runtime design.
