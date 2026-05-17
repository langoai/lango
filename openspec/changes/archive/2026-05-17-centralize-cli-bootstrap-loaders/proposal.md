## Why

`cliboot` is now the shared CLI bootstrap boundary that owns lifecycle options such as storage broker startup, but `doctor`, `settings`, and `onboard` still duplicate `bootstrap.Run` calls. That leaves production entry points vulnerable to option drift whenever bootstrap behavior changes.

## What Changes

- Route `doctor`, `settings`, and `onboard` through the shared `cliboot.BootResult` loader.
- Keep command-level test seams while changing them to wrap the shared loader instead of `bootstrap.Run` directly.
- Add an architecture regression test that forbids production CLI packages from calling `bootstrap.Run` outside `internal/cli/cliboot`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-bootstrap-factory`: strengthen the shared loader requirement so all `internal/cli` production packages, not only `cmd/`, avoid direct `bootstrap.Run` calls.

## Impact

- Affected code: `internal/cli/doctor`, `internal/cli/settings`, `internal/cli/onboard`, and `internal/archtest`.
- No user-facing CLI syntax changes.
- No public documentation changes expected; this is lifecycle and architecture enforcement.
