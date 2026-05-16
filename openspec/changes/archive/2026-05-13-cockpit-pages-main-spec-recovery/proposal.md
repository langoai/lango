## Why

Several archived cockpit-page changes updated the same `cockpit-pages` capability, and the current main spec has dropped previously merged scenarios for Dead Letters and Mission Control. That leaves the authoritative spec weaker than the implemented and tested runtime surface.

## What Changes

- Restore the missing Dead Letters registration, unavailable-state, and row-navigation scenarios in the main `cockpit-pages` capability.
- Restore the missing Mission Control degraded-state, empty/workbench help, single-row navigation, decisions-lane Enter, and dismiss-focus scenarios in the same capability.
- Keep this change spec-only so the main spec once again matches the already-landed runtime and tests.

## Capabilities

### New Capabilities

### Modified Capabilities

- `cockpit-pages`: Recover lost Dead Letters and Mission Control requirement scenarios in the main capability spec.

## Impact

- Affected specs: `openspec/specs/cockpit-pages/spec.md`
- No runtime code changes
