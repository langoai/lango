## Why

The security status docs are now aligned with the command contract, but the tests still do not directly lock the richer field semantics the command exposes for KMS-backed signers and DB status strings. That leaves the output contract vulnerable to quiet regressions.

## What Changes

- Add render-status regressions that lock KMS signer visibility and DB status strings in both text and JSON output.
- Sync the `cli-security-status` spec so those field semantics are part of the explicit contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-security-status`: signer provider, DB status, and KMS fallback output semantics are now directly regression-tested.

## Impact

- Affected code: `internal/cli/security/security_test.go`
- Affected specs: `openspec/specs/cli-security-status/spec.md`
