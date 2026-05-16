## Why

The new shared EOF-deny confirmation helper is in place, but a few remaining destructive/commit-affecting commands still call `ConfirmIO(...)` directly. That leaves those flows with a harder failure mode than the rest of the CLI when confirmation input is missing.

## What Changes

- Switch payment send, keyring clear, and secrets delete to the shared EOF-deny confirmation helper
- Add EOF-abort regressions for those command paths
- Extend the existing EOF-deny spec coverage to these remaining commands

## Capabilities

### New Capabilities

### Modified Capabilities
- `payment-service`: payment send treats EOF as a clean denial
- `passphrase-management`: keyring clear treats EOF as a clean denial
- `cli-secrets-management`: secrets delete treats EOF as a clean denial

## Impact

- Affected code: `internal/cli/payment/send.go`, `internal/cli/payment/payment_test.go`, `internal/cli/security/keyring.go`, `internal/cli/security/secrets.go`, `internal/cli/security/security_test.go`
- Affected specs: `openspec/specs/payment-service/spec.md`, `openspec/specs/passphrase-management/spec.md`, `openspec/specs/cli-secrets-management/spec.md`
- No feature expansion; this is a safety and consistency hardening change
