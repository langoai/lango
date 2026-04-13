## 1. Core Redaction Validation

- [x] 1.1 Add `RedactionLevel.Valid()` method to `internal/provenance/types.go`
- [x] 1.2 Add `ErrInvalidRedaction` sentinel error to `internal/provenance/errors.go`
- [x] 1.3 Add redaction validation in `Export()` entry point (`internal/provenance/bundle.go`)
- [x] 1.4 Add redaction validation in `Verify()` for import path (`internal/provenance/bundle.go`)
- [x] 1.5 Add route-level validation in `decodeProvenanceRequest()` (`internal/app/p2p_routes.go`)
- [x] 1.6 Add tests: Export invalid redaction, Verify invalid redaction, valid redaction levels (`internal/provenance/bundle_test.go`)
- [x] 1.7 Add tests: push/fetch with invalid redaction return 400 (`internal/app/p2p_routes_test.go`)

## 2. CLI Disabled Consistency

- [x] 2.1 Add `isProvenanceDisabled()` helper and `dateTimeFormat` constant to `internal/cli/provenance/common.go`
- [x] 2.2 Add disabled check in `session.go` (`newSessionTreeCmd`, `newSessionListCmd`)
- [x] 2.3 Add disabled check in `bundle.go` (`newBundleExportCmd`, `newBundleImportCmd`)
- [x] 2.4 Replace inline disabled checks in `checkpoint.go` with `isProvenanceDisabled()` helper
- [x] 2.5 Replace inline disabled checks in `attribution.go` with `isProvenanceDisabled()` helper
- [x] 2.6 Replace inline date format strings in `checkpoint.go` and `attribution.go` with `dateTimeFormat` constant
- [x] 2.7 Add disabled notice at end of `status` command output in `provenance.go`
- [x] 2.8 Add CLI disabled tests (`internal/cli/provenance/provenance_test.go`)

## 3. Cosmetic Cleanup

- [x] 3.1 Replace `cloneAuthorStats`/`cloneFileStats` with `maps.Clone` in `bundle.go`
- [x] 3.2 Extract `entRowToAttribution()` helper in `attribution_ent_store.go`
- [x] 3.3 Extract `gatewayAddr`, `provenanceRequestBody`, `addProvenanceExchangeFlags` helpers in `internal/cli/p2p/provenance.go`
- [x] 3.4 Document hook-outside-lock pattern in `internal/runledger/store.go` and `ent_store.go`

## 4. Verification

- [x] 4.1 Run `go build ./...` — passes with no errors
- [x] 4.2 Run `go test ./...` — all tests pass
- [x] 4.3 Verify README/docs redaction documentation matches new validation behavior
