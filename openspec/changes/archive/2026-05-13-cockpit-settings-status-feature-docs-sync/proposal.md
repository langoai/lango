## Why

The cockpit feature reference now has dedicated sections for Tools, Approvals, Sessions, Tasks, and Dead Letters, but Settings and Status are still mostly described only in the top roster table. That leaves two always-routable degraded surfaces under-documented despite their concrete runtime behavior.

## What Changes

- Add dedicated Settings and Status sections to `docs/features/cockpit.md`.
- Describe save-unavailable behavior for Settings and unavailable provider/collector states for Status.
- Extend downstream docs-sync requirements so those sections stay aligned.

## Impact

- Public cockpit docs describe the current Settings and Status operator surfaces much more concretely.
- Future docs drift on those degraded page contracts becomes easier to catch.
