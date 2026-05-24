## Why

The security wiring already returns deterministic bootstrap-required errors for local and KMS signer providers, but the tests only checked for a vague substring and the public configuration docs still had a stale `enclave` provider reference. That weakens both operator guidance and trust in the current security contract.

## What Changes

- Tighten `initSecurity` regressions so the local bootstrap error and KMS build-tag errors are asserted directly.
- Update public configuration docs to list only currently supported signer providers.
- Clarify in docs and specs that local depends on bootstrap-backed storage wiring, while KMS-backed providers also require matching build tags.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: security provider wiring errors are now directly verified for local bootstrap and KMS build-tag requirements.
- `downstream-docs-sync`: security provider tables reflect the current provider set and bootstrap requirements.

## Impact

- Affected code: `internal/app/wiring_test.go`
- Affected docs: `README.md`, `docs/configuration.md`
- Affected specs: `openspec/specs/production-readiness/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
