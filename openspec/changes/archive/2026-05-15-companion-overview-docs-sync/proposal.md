## Why

The architecture overview still described `internal/security/` as if it owned a companion discovery subsystem even though the current runtime uses gateway-connected companions rather than a shipped discovery feature.

## What Changes

- sync the architecture overview wording to the current gateway-connected companion model
- extend the companion-connectivity docs guard coverage to include the overview page

## Impact

- public architecture docs better reflect the shipped runtime
- less confusion about where companion connectivity logic lives
- stronger regression protection for stale overview wording
