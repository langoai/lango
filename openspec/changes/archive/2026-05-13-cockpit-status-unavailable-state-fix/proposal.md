## Why

The cockpit Status page is documented as showing placeholder text when the observability collector is unavailable, but the implementation currently renders empty-looking metric sections instead. The feature-status section also stays silent when no provider is wired.

Those are real unavailable states and should be surfaced explicitly.

## What Changes

- Render explicit unavailable messaging when the feature-status provider is absent.
- Render explicit unavailable messaging in the metrics sections when the observability collector is absent.
- Add regressions and sync the cockpit status-page spec.

## Impact

- Operators can tell the difference between missing observability wiring and zero-valued metrics.
- Status page behavior finally matches the documented degraded-state contract.
