## Why

The cockpit feature reference already has dedicated sections for Mission Control, Approvals, Sessions, Tasks, and Dead Letters, but Tools is still only described in the top roster table. That leaves an actually usable read-only operator surface under-documented.

## What Changes

- Add a dedicated Tools page section to `docs/features/cockpit.md`.
- Describe category cursor navigation, immediate right-panel updates, and degraded states when the catalog is absent.
- Extend downstream docs-sync requirements so the richer Tools surface stays documented.

## Impact

- Public cockpit docs better match the current Tools page interaction model.
- Future docs drift on the Tools surface becomes easier to catch.
