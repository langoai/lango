## Why

The Sessions page now distinguishes unavailable and empty states, but a configured session-list source failure still renders as a generic `Error: ...` string with no explicit product contract. That leaves a user-visible failure mode under-specified and less actionable than the other cockpit degraded states.

## What Changes

- Render configured session-list failures with explicit session-list failure wording.
- Add regression coverage for the load-error view.
- Extend the sessions-page spec and cockpit docs to describe the configured-source failure state.

## Impact

- Operators can distinguish source failure from empty and unavailable states immediately.
- Runtime, tests, docs, and spec describe the same failure contract.
