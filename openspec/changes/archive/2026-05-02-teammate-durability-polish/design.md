# Design

## Scope

This change is intentionally narrow. It does not alter the approval-blocked durability data model, mirror wiring, or live read path.

It only:

- clarifies the RunLedger mirror contract for blocked-state replacement
- records the best-effort concurrency caveat for duplicate block events
- adds regression coverage for transition no-op cases

## Contract Clarifications

- When `runtime_condition` stays `blocked_waiting_approval` but `blocked_reason` or `grant_request_id` changes, the mirror appends a fresh approval-block event and refreshes the cached snapshot to the latest blocked values.
- Transition derivation remains best effort. Concurrent `UpdateProjection(...)` calls for the same run may produce duplicate approval-block events, but the latest durable snapshot still converges on the final blocked state.
