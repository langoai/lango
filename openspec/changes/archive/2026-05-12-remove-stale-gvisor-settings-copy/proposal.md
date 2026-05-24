## Why

The settings UI still described `gvisor` as the strongest isolation runtime even though the product currently ships only a gVisor stub. That wording directly misleads operators inside the interactive configuration surface.

## What Changes

- Update the P2P Sandbox settings form runtime description to match the current runtime contract.
- Add a form-level regression for the runtime description.
- Sync the `cli-settings` spec so the settings UI contract stays honest about the gVisor stub.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-settings`: the P2P Sandbox runtime description now reflects the current runtime contract instead of advertising gVisor as the strongest available runtime.

## Impact

- Affected code: `internal/cli/settings/forms_p2p.go`
- Affected tests: `internal/cli/settings/forms_impl_test.go`
- Affected specs: `openspec/specs/cli-settings/spec.md`
