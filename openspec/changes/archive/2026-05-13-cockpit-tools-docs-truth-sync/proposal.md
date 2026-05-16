## Why

The cockpit Tools page already renders explicit unavailable messaging when no tool catalog is wired, but the public docs still describe it as if it merely shows an empty state or always lists registered tools.

That drifts from the actual degraded-page contract now implemented in the runtime.

## What Changes

- Update the cockpit feature page to describe the Tools page as always available with degraded unavailable messaging when the catalog is absent.
- Update the README cockpit shortcut table with the same contract.
- Extend downstream docs requirements so these public entrypoints stay aligned with the runtime.

## Impact

- Public cockpit docs reflect the current Tools page behavior.
- README and feature docs stop understating or misdescribing the nil-catalog path.
