# Policy-Driven Replay Controls

This page describes the first `policy-driven replay controls` slice for post-adjudication replay in `knowledge exchange v1`.

## Purpose

This slice adds actor- and outcome-aware replay authorization on top of the shared recovery gate.

The slice is intentionally narrow:

- actor is resolved from runtime context
- replay is fail-closed when actor is unresolved
- replay is fail-closed when actor is not allowed
- policy is backed by current config allowlists
- replay authorization still sits behind the canonical dead-letter evidence gate

## What Ships

- replay-service-local policy gate
- actor resolution from runtime context
- config-backed policy shape:
  - `replay.allowed_actors`
  - `replay.release_allowed_actors`
  - `replay.refund_allowed_actors`
- outcome-aware authorization on top of the shared recovery evidence source used by retry, dead-letter, and manual replay

## Current Limits

This slice does not yet include:

- human approval UI
- org-level policy editor
- per-transaction policy snapshots
- amount-tier replay rules
