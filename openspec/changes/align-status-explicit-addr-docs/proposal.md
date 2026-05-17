## Why

`lango status --addr <url>` now probes and reports the same normalized explicit gateway target. The public status command page still explains only the omitted-`--addr` default behavior and its JSON schema example still shows the localhost fallback, so operators cannot see that explicit overrides are reflected in output. The existing docs guard also only protects configured-default wording, not the explicit-target display contract.

## What Changes

- document that `lango status --addr <url>` probes the normalized explicit target and reports that same target in the `gateway` field
- update the status JSON example so the `gateway` value matches the custom address example instead of implying localhost output
- add an executable docs guard that fails if status docs omit the explicit-target display contract

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `downstream-docs-sync`: public status docs describe explicit `--addr` probe/display behavior
- `test-coverage`: docs guard tests enforce status explicit-target docs coverage

## Impact

- docs-only user-facing correction under `docs/cli/status.md`
- repository docs guard update under `internal/testutil`
- no runtime behavior changes
