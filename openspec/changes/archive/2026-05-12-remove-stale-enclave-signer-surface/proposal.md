## Why

The runtime and production specs already treat `enclave` as unsupported, but several user-facing and validation surfaces still expose it as if it were valid. That mismatch leaks directly into configuration validation, settings UI, and `lango doctor`.

## What Changes

- Remove `enclave` from config validation and settings UI provider lists.
- Treat KMS-backed providers as valid in `lango doctor` security checks.
- Add regressions that `enclave` is rejected as invalid config while KMS providers remain accepted.
- Sync the production-readiness spec with the actual supported signer-provider set.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: signer-provider validation and doctor checks now match the actual supported provider set.

## Impact

- Affected code: `internal/config/constants.go`, `internal/config/loader.go`, `internal/config/types_security.go`, `internal/cli/settings/forms_security.go`, `internal/cli/doctor/checks/security.go`
- Affected tests: `internal/config/types_defaults_test.go`, `internal/cli/doctor/checks/checks_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
