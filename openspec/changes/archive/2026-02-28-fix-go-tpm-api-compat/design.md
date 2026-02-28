## Context

The project depends on `github.com/google/go-tpm` v0.9.8 for TPM2-based secret sealing in `internal/keyring/tpm_provider.go`. The go-tpm library introduced breaking API changes between versions:
- `TPMTSymDef` was split into `TPMTSymDef` (algorithm-only) and `TPMTSymDefObject` (full symmetric params)
- `tpm2.Marshal` changed from `([]byte, error)` to `[]byte` (panics on error)
- `tpm2.Unmarshal` changed from `(int, error)` with pointer arg to generic `Unmarshal[T]([]byte) (*T, error)`

## Goals / Non-Goals

**Goals:**
- Fix all 5 compilation errors in `tpm_provider.go` to match go-tpm v0.9.8 API
- Unblock Docker builds

**Non-Goals:**
- Changing TPM sealing behavior or key templates
- Upgrading to a different go-tpm major version
- Adding new TPM functionality

## Decisions

1. **Use `TPMTSymDefObject` for SRK template** — The `TPMSECCParms.Symmetric` field requires `TPMTSymDefObject` in v0.9.8. This is the only correct type; no alternative.

2. **Accept `Marshal` panic-on-error semantics** — v0.9.8's `Marshal` panics instead of returning errors. Since we only marshal well-typed TPM structures (`TPM2BPublic`, `TPM2BPrivate`), panics are not expected in practice. The `marshalSealedBlob` function signature changes from `error` return to direct assignment.

3. **Use generic `Unmarshal[T]` with pointer dereference** — v0.9.8's generic signature returns `*T`. We dereference immediately into the existing variables to minimize diff and preserve the existing control flow.

## Risks / Trade-offs

- [Marshal panics] → Structures are always well-formed from TPM responses; panic risk is negligible. If needed, a `recover` wrapper could be added later.
- [API stability] → go-tpm is pre-1.0; future versions may break again. Pinning v0.9.8 in go.mod mitigates this.
