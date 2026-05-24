## Why

`HeaderView.ActiveAgentSummary` storage is now sanitized at projection time, but `buildActiveAgentSummary()` still aggregates raw `OwnerAgent` strings from `MissionView`. That means the builder itself can produce raw text unless every upstream caller has already sanitized the owner field.

## What Changes

- Sanitize owner-agent labels inside the active-agent summary builder.
- Add regression coverage for malformed owner-agent values.
- Record the expanded header-summary replay-safety contract in OpenSpec and downstream docs.

## Impact

- Makes the active-agent summary builder independently replay-safe instead of relying on upstream callers.
- Tightens Mission Control header summary hygiene at the aggregation boundary.
