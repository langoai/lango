# Design

## Context

The first runtime-oriented knowledge-exchange slice is landed, and release approval outcomes can already be mapped into transaction-level settlement progression state in code. What is still missing is the public architecture page and the OpenSpec closeout that make the slice discoverable and keep the repo-level docs contract aligned.

## Goals / Non-Goals

**Goals**

- publish a bounded public architecture page for settlement progression
- wire the page into the architecture landing page, P2P knowledge-exchange track, and docs navigation
- record the `apply_settlement_progression` meta tool in OpenSpec
- archive the completed change after syncing main specs

**Non-Goals**

- no new settlement executor
- no new dispute engine
- no human adjudication UI
- no broader escrow lifecycle work

## Decisions

### 1. Treat settlement progression as a transaction-level control-plane slice

The public page should describe the landed progression layer without overstating execution completeness.

### 2. Keep the track doc focused on landed slices plus remaining gaps

The P2P knowledge-exchange track should move settlement progression from follow-on design work into landed work and leave actual settlement execution, partial-settlement rules, and dispute engine completion as the remaining gaps.

### 3. Record the settlement progression meta tool directly in `meta-tools`

The OpenSpec update should capture the new receipts-backed tool as part of the existing meta-tools contract rather than inventing a separate capability bucket.

## Risks / Trade-offs

- [Risk] The new page could imply that a full settlement lifecycle is already complete.
  - Mitigation: keep the current limits section explicit.

- [Risk] The track doc could drift from the landed code surface.
  - Mitigation: reference only the currently implemented settlement progression slice and its remaining gaps.

## Migration Plan

1. Add the public settlement progression page.
2. Wire the page into the architecture landing page, track page, and nav.
3. Sync the delta specs into the main specs.
4. Archive the completed change under `openspec/changes/archive/2026-04-22-settlement-progression`.
