# Identity Trust Reputation Audit Design

## Purpose

This document defines the audit framing for the `identity / trust / reputation` capability family under the `P2P Knowledge Exchange Track`.

Its purpose is not to design a full `reputation v2` system yet. Instead, it establishes the audit baseline for `knowledge exchange v1` by clarifying:

- what identity continuity means in product terms,
- how trust entry should be understood,
- how reputation should be separated from root trust,
- how revocation and trust decay should be interpreted operationally.

This audit is intended to become the next detailed audit ledger after the earlier `External Collaboration & Economic Exchange` and `Trust, Security & Policy` stabilization work.

## Scope

This audit is judged against `knowledge exchange v1`.

It directly covers:

- identity continuity,
- trust entry,
- reputation,
- revocation and trust decay.

It does not directly design:

- a full `reputation v2` scoring formula,
- detailed pricing or negotiation policy,
- dispute engine mechanics,
- long-running Phase 4 accountability systems.

Those remain follow-on work. This audit exists to define the conceptual and operator-facing relationship model they must inherit.

## Recommended Structure

The audit should use four rows:

1. `Identity Continuity`
2. `Trust Entry`
3. `Reputation`
4. `Revocation & Trust Decay`

This is the preferred structure because it maps cleanly to the current product boundary rather than overfitting the current implementation layout.

## Baseline Relationship Model

The audit should explicitly lock the following relationship model.

### 1. Owner-Root Trust vs Agent/Domain Reputation

These are distinct.

- `owner-root trust` provides a bootstrap ceiling/floor
- `agent/domain-specific reputation` accumulates from actual market and collaboration history

So a trusted owner does not imply a fully trusted new agent. It only prevents the agent from starting from complete zero-trust conditions.

### 2. Admission Trust vs Payment Trust

These are also distinct.

- `admission trust` answers whether a peer is allowed across the boundary at all
- `payment trust` answers what payment friction is appropriate once a transaction is considered

The same signals may inform both, but they are different product gates and must remain documented as such.

### 3. Operational Signals vs Durable Negative Reputation

Repeated failures, timeouts, and policy-triggered friction can be used as immediate operational signals.

But durable negative reputation should not be applied with the same semantics.

The model for this audit is:

- operational signals may affect runtime safety and session access immediately
- durable negative reputation requires stronger adjudication, especially when used beyond temporary safety gating

### 4. Bootstrap Trust vs Earned Trust

New agents should begin under constrained trust conditions.

`owner-root trust` may soften the starting floor, but real friction reduction should be earned through transaction history, fulfillment, and stable collaboration outcomes.

This distinction is required so the system can support trusted owners without collapsing the boundary between inherited trust and earned trust.

## Current Surface Map

The current code and documentation already suggest a substantial real surface:

- identity bundles and DID continuity
- gateway auth and protected-route semantics
- handshake-driven trust entry
- session invalidation and security-event-based revocation
- reputation store and operator query surfaces
- payment trust thresholds for prepay/postpay selection

This means the audit is not deciding whether these capabilities exist. It is deciding how they should be understood and judged under the current product path.

## Detailed Audit Rows

### Identity Continuity

Initial judgment: `stabilize`

Why:

- the runtime already has real DID continuity and multiple identity surfaces
- but the product-facing language is still fragmented between DID form, bundle form, session form, and gateway/auth form

The likely conclusion is that identity continuity is a real capability worth keeping, but the operator model needs consolidation.

### Trust Entry

Initial judgment: `stabilize`

Why:

- gateway auth, handshake approval, owner shield, and payment trust gating are all real
- but they still read like adjacent subsystems instead of one entry model for early exchange

This row should explicitly compare `admission trust` and `payment trust`, rather than hiding the distinction.

### Reputation

Initial judgment: `stabilize`

Why:

- the reputation surface is real and queried through real operator paths
- but the system has not yet frozen the relationship between owner-root trust, agent/domain reputation, bootstrap trust, and earned trust

This row is expected to become the conceptual baseline for later `reputation v2` design.

### Revocation & Trust Decay

Initial judgment: `stabilize`

Why:

- session invalidation and security-event revocation are real
- but trust decay is still primarily operational rather than framed as one coherent product boundary

This row should prioritize access revocation and runtime trust decay mechanics over abstract mathematical scoring.

## Assessment Shape

The likely assessment pattern for this audit is:

- all four rows are real capabilities,
- none are immediate `remove` or `defer` candidates,
- the main need is relationship clarity and operator-facing consolidation.

That points toward a `stabilize` judgment across the whole surface, unless deeper exploration reveals a narrower split.

## Follow-On Design Inputs

This audit is expected to feed directly into three follow-on areas:

### 1. Reputation v2

The audit should supply a stable conceptual baseline for:

- owner-root trust,
- agent/domain reputation,
- negative signal adjudication,
- bootstrap trust and earned trust transitions.

### 2. Pricing / Negotiation / Settlement Audit

That audit should inherit:

- the distinction between admission trust and payment trust,
- the idea that payment friction is a policy gate rather than a single scalar consequence of trust.

### 3. Knowledge Exchange Runtime Design

Runtime design should inherit:

- what trust entry can close immediately,
- what reputation changes are merely operational,
- how revocation should interact with transaction and settlement progression.

## Deliverable Expectation

The final audit document that follows from this design should be a detailed audit ledger plus follow-on design inputs.

It should not yet expand into a full implementation plan or a complete reputation redesign.
