# Exportability Policy Design

## Purpose

Define the first explicit `exportability policy` model for `knowledge exchange v1`.

This design answers a narrow but critical question:

- when is an artifact tradeable,
- who decides,
- what evidence is recorded,
- and how human override works without mutating the default policy model.

This document is subordinate to:

- `docs/architecture/master-document.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `docs/architecture/trust-security-policy-audit.md`

## Scope

This design covers:

- artifact-level exportability classification,
- source classes,
- decision flow,
- human-review triggers,
- decision evidence and receipts,
- policy expression model.

This design does not yet cover:

- implementation wiring,
- UI/TUI forms,
- exact storage schema,
- dispute-resolution orchestration,
- full approval-routing implementation.

## Problem Statement

`knowledge exchange v1` requires an explicit answer to whether a deliverable may leave the local trust boundary.

Today, Lango has privacy tooling, output sanitization, approval infrastructure, provenance, and auditability. What it does not yet have is one explicit runtime policy for:

- tradeable vs non-tradeable artifacts,
- source-based restriction,
- user-controlled exportability exceptions,
- and decision records that can later feed approval and dispute handling.

Without this, `knowledge exchange v1` lacks a clear product boundary even if the underlying security subsystems are already real.

## Approaches Considered

### Approach A: Content-Only Exportability

Decide exportability from the final artifact content only.

Pros:

- simple mental model,
- easy to explain,
- naturally fits output sanitization tooling.

Cons:

- ignores confidential provenance,
- allows private-source laundering through summarization,
- too weak for `knowledge exchange v1`.

### Approach B: Strict Source-Based Exportability

Decide exportability entirely from the source lineage of the artifact.

Pros:

- strong confidentiality boundary,
- predictable behavior,
- aligns well with product constitution.

Cons:

- can be rigid,
- may require human review even when final content looks safe,
- needs explicit user-controlled exception handling.

### Approach C: Hybrid With Source-Primary Policy

Use source-based classification as the default rule, while still recording content/policy context and allowing narrow user-authorized exceptions.

Pros:

- preserves strong confidentiality guarantees,
- fits delegated autonomy,
- supports later approval and dispute workflows,
- keeps room for future policy refinement without weakening `v1`.

Cons:

- more policy metadata,
- needs explicit receipt-style decision records.

## Recommendation

Use **Approach C**, but with a strongly source-primary stance.

In practice, this means:

- the default exportability decision is source-based,
- content sanitization is only a supporting safety layer,
- and exceptions require explicit user policy or one-time human override.

## Policy Model

The primary exportability decision unit is the `artifact / deliverable`.

Each source asset belongs to one of three classes:

- `public`
- `user-exportable`
- `private-confidential`

The root authority for these classes belongs to the user.

The user expresses policy through a mixed model:

- default: `asset tagging`
- optional: higher-level `policy rules` for repeated patterns

Example implications:

- `public` sources may contribute to exportable artifacts by default.
- `user-exportable` sources may contribute to exportable artifacts within user policy boundaries.
- `private-confidential` sources are non-exportable by default.

### Mixed-Source Rule

When an artifact is derived from multiple source assets, the decision uses `highest sensitivity wins`.

Therefore:

- if any contributing source is `private-confidential`, the artifact is blocked by default,
- if the sources are only `public` and `user-exportable`, the artifact may be exportable within policy boundaries.

### Derived Knowledge Rule

Generalized know-how derived from `private-confidential` material is still non-exportable by default.

It only becomes eligible for export when the user has explicitly authorized that derived class of knowledge to be exported.

### Derived Artifact Rule

Artifacts derived from `user-exportable` sources inherit `user-exportable` status by default.

However:

- derivation may add stricter controls,
- derivation may not automatically loosen restrictions.

## Decision Flow

The exportability flow is:

1. user policy defines the root boundary
2. leader agent performs the first artifact-level decision
3. only high-risk or ambiguous cases are escalated to human review

### Phase 1: Draft-Stage Advisory Decision

The leader agent performs an early advisory decision during drafting.

Possible states:

- `exportable`
- `blocked`
- `needs-human-review`

If the draft-stage result is `blocked`:

- artifact creation may continue,
- but the artifact is not treated as an external-delivery candidate,
- and the leader agent must produce a separate exportable artifact if needed.

### Phase 2: Final Pre-Export Authoritative Decision

Immediately before external export, the system performs an authoritative decision.

If the final result is:

- `exportable`: the artifact may be exported
- `blocked`: export is forbidden unless a human explicitly overrides
- `needs-human-review`: export pauses until human review resolves it

### Human Override Rule

Human override is allowed only as a one-time exception.

That override:

- applies to one transaction only,
- does not mutate the default policy,
- does not automatically authorize future artifacts with the same lineage.

## Human Review Triggers

`needs-human-review` is used for cases where the leader agent must not make the final decision alone.

At minimum, this state is required when:

- a `private-confidential` source is involved but the user has a relevant exception policy,
- source metadata is incomplete,
- source metadata conflicts or cannot be resolved consistently.

This state is not a soft approval.

It means the system has reached a policy boundary that requires human judgment.

## Evidence and Receipts

Every exportability decision must produce a durable decision record.

At minimum, the record must include:

- `decision state`
- `policy code`
- `human-readable explanation`
- `source lineage summary`

### Decision States

- `exportable`
- `blocked`
- `needs-human-review`

### Policy Code

Examples:

- `allowed_public_only`
- `allowed_user_exportable`
- `blocked_private_source`
- `review_metadata_conflict`
- `review_user_exception_required`

### Human-Readable Explanation

Each record must also include a short explanation suitable for operator review.

Examples:

- `Artifact includes a private-confidential source and no matching export exception exists.`
- `Artifact derives only from user-exportable assets authorized for external delivery.`

### Source Lineage Summary

At minimum, the lineage summary contains:

- source class,
- source asset ID or label,
- applied rule.

This creates a decision trail that later systems can reuse for:

- approval decisions,
- audit review,
- dispute evidence,
- receipt generation.

## Receipt Model

The exportability record is part of a broader future `dispute-ready receipt`.

For now, it should be treated as an `exportability receipt` with three layers of evidence:

- `policy basis`
- `source basis`
- `decision basis`

This means an operator must be able to answer:

- what policy rule applied,
- which source assets mattered,
- why the system allowed, blocked, or escalated the artifact.

## Boundaries and Non-Goals

This design intentionally does not claim that:

- sanitization alone makes an artifact tradeable,
- generalized knowledge from private material is safe by default,
- one-time human overrides should become durable automatic policy,
- or agent autonomy should replace user root authority in `v1`.

## Initial Success Criteria

This design is successful if a future implementation can reliably do all of the following:

- classify source assets into three explicit classes,
- make artifact-level exportability decisions twice,
- block mixed artifacts when a private source is present,
- require human review for policy ambiguity or exception paths,
- and emit durable decision records suitable for later approval and dispute flows.

## Follow-On Planning Inputs

The next implementation plan should define:

1. where source class metadata is stored
2. how asset tagging and policy rules are represented
3. how draft-stage and final-stage decisions are invoked
4. how exportability receipts are persisted and surfaced
5. how this decision record integrates with approval flow
6. how this decision record contributes to dispute-ready receipts
