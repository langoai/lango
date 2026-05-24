## Why

The context panel docs say missing observability dependencies should surface unavailable messaging, but the runtime still renders zero-like token and uptime values when no metrics collector is configured. That makes an unavailable metrics backend look like valid zero data.

## What Changes

- Make the Token Usage, Tool Stats, and System sections render explicit unavailable messaging when the metrics collector is absent.
- Add regression coverage for the nil-collector context-panel view.
- Sync the cockpit feature docs and context-panel spec.

## Impact

- The context panel no longer disguises a missing metrics backend as valid zero-valued metrics.
- Operator-facing observability surfaces become more consistent with the other degraded cockpit pages.
