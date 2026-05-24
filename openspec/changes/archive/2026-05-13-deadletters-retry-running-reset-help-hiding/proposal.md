## Why

The Dead Letters page currently keeps advertising `Ctrl+R` filter reset even while a retry request is actively running, but the runtime ignores that key in the same state. That leaves the help bar exposing an inert action during one of the page's highest-attention flows.

## What Changes

- Hide the `Ctrl+R` reset binding from Dead Letters help while a retry request is actively running.
- Add regression coverage for the running-retry help state.
- Update cockpit-page spec and feature docs to describe the same actionability rule.

## Capabilities

### New Capabilities

### Modified Capabilities

- `cockpit-pages`: Dead Letters help becomes conditional on retry-running state for the reset binding.

## Impact

- Affected code: `internal/cli/cockpit/pages/deadletters.go`, `internal/cli/cockpit/pages/deadletters_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/cockpit-pages/spec.md`
