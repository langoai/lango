# Settlement Progression

This page describes the current transaction-level settlement progression slice for `knowledge exchange v1`.

## Purpose

Settlement progression turns artifact release outcomes and renewed disagreement into canonical transaction-level state that downstream escrow, adjudication, and recovery flows consume.

The slice is intentionally narrow:

- `transaction receipt` owns canonical settlement progression state
- `submission receipt` contributes evidence, reasons, hints, and the event trail
- release approval outcomes map into progression states
- progression updates require a current submission receipt
- actual money movement remains separate
- escalation can re-enter `dispute-ready` from `review-needed`, `approved-for-settlement`, `partially-settled`, or an already disputed transaction

## What Ships

- transaction-level settlement progression state
- release outcome mapping for `approve`, `request-revision`, `reject`, and `escalate`
- canonical `approved-for-settlement`, `review-needed`, `dispute-ready`, `partially-settled`, and `settled` progression
- `partial_hint` plumbing for bounded partial-settlement guidance
- preserved `partial_settlement_hint` when renewed disagreement re-escalates from `partially-settled`
- submission-bound progression writes that append settlement events to the current submission receipt
- disputed-event appends whenever progression re-enters `dispute-ready`
- a receipts-backed `apply_settlement_progression` meta tool
- tool results that include:
  - `settlement_progression_reason_code`
  - `settlement_progression_reason`
  - `partial_hint`
  - `dispute_lifecycle_status`

## Canonical State

The current progression states are:

- `pending`
- `in-progress`
- `review-needed`
- `approved-for-settlement`
- `partially-settled`
- `settled`
- `dispute-ready`

`transaction receipt` keeps the canonical state, reason code, human-readable reason, partial-settlement hint, dispute-ready marker, and dispute lifecycle marker.

`dispute-ready` is now a public canonical path:

- early escalation still maps into `review-needed`
- renewed disagreement from `review-needed`, `approved-for-settlement`, `partially-settled`, or an already disputed receipt maps into `dispute-ready`
- re-escalation from `partially-settled` keeps the current canonical `partial_settlement_hint`

`dispute_lifecycle_status` remains a separate field from settlement progression:

- `hold-active` means dispute hold succeeded and downstream recovery is still on the active held path
- `re-escalated` means post-adjudication recovery exhausted into a canonical dispute-ready re-entry

## Current Limits

This slice does not yet include:

- repeated partial execution from `partially-settled`
- percentage-based or free-form partial hints
- automatic executor selection or a full multi-round settlement orchestrator
- human adjudication or policy UI

The current implementation keeps transaction-level progression canonical first and still leaves broader settlement orchestration to downstream slices.
