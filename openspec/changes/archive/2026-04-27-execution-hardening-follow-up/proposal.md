## Why

The earlier stabilization batch closed the most visible dead-letter/runtime wiring gaps, but a smaller execution-safety follow-up still remained:

- concurrent dispute hold, adjudication, and refund requests for the same transaction could enter their execution paths in parallel
- replay/background retry keys were derived canonically but duplicate submissions were not explicitly deduplicated
- partial-commit success-path tests for hold/refund evidence recording were missing
- settlement escalation fallback still silently accepted unknown progression states

## What Changes

- Add service-local per-transaction serialization to dispute hold, escrow adjudication, and escrow refund.
- Add partial-commit success-recording tests for dispute hold and escrow refund.
- Reuse the existing background `retry_key` as an idempotent dedup key for pending, running, and scheduled tasks.
- Make `escalationProgressionStatus` exhaustive over known progression states and panic on unknown internal states.
- Truth-align public architecture docs and docs-only OpenSpec requirements.

## Impact

- Concurrent execution attempts for the same transaction are serialized at the service boundary.
- Duplicate background retry dispatch is reduced to one task identity per canonical retry key while work is still pending/running/scheduled.
- Failure semantics around success-record persistence are now covered by focused tests.
- Unknown settlement progression states no longer silently fall back through escalation mapping.
