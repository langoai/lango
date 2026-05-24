## Why

The Dead Letters page already lets `Backspace` edit whichever text filter field is active, but the in-product help still labels that key as `query`. That makes the help bar narrower than the actual runtime behavior and can mislead operators once focus moves beyond the first filter field.

## What Changes

- Relabel the Dead Letters `Backspace` help binding to describe editing the active text filter instead of only the query field.
- Add regression coverage for the new help wording.
- Update the cockpit-pages spec and public cockpit docs to describe the same active-field edit contract.

## Capabilities

### New Capabilities

### Modified Capabilities

- `cockpit-pages`: Dead Letters help wording for `Backspace` changes to match active text-field editing behavior.

## Impact

- Affected code: `internal/cli/cockpit/pages/deadletters.go`, `internal/cli/cockpit/pages/deadletters_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/cockpit-pages/spec.md`
