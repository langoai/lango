## Why

Docker build fails because `internal/keyring/tpm_provider.go` uses outdated `go-tpm` API patterns incompatible with v0.9.8. Five compilation errors block the entire build pipeline.

## What Changes

- Fix `TPMTSymDef` → `TPMTSymDefObject` for the SRK template symmetric field
- Fix `tpm2.Marshal` calls to use single-return signature (v0.9.8 panics on error instead of returning it)
- Fix `tpm2.Unmarshal` calls to use generic function signature `Unmarshal[T](data []byte) (*T, error)`

## Capabilities

### New Capabilities

_None — this is a bug fix for API compatibility._

### Modified Capabilities

- `keyring-security-tiering`: TPM provider implementation updated to match go-tpm v0.9.8 API surface

## Impact

- `internal/keyring/tpm_provider.go` — 5 lines changed across 3 call patterns
- Docker builds unblocked (previously failed at compilation)
- No behavioral change — same TPM seal/unseal logic, just correct API usage
