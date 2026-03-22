## Why

The provenance feature (feature/provenance branch) has two spec violations:
1. Core `redactBundle()` silently falls back invalid redaction values to `content`, violating the spec guarantee of "selected redaction level" (spec.md:155,210).
2. `session` and `bundle` CLI commands lack the `provenance.enabled=false` disabled check required by spec (spec.md:134-136: "any provenance command").

Additionally, several cosmetic issues exist: duplicate map clone functions, missing Ent conversion helpers, duplicated P2P CLI code, and undocumented hook patterns.

## What Changes

- Add `RedactionLevel.Valid()` method and `ErrInvalidRedaction` sentinel error in core
- Add redaction validation in `Export()`, `Verify()`, and HTTP route `decodeProvenanceRequest()`
- Add missing `isProvenanceDisabled` check in `session.go` and `bundle.go` CLI commands
- Add disabled notice to `status` command output (without blocking config display)
- Extract `isProvenanceDisabled()` helper and `dateTimeFormat` constant in CLI common.go
- Replace duplicate `cloneAuthorStats`/`cloneFileStats` with `maps.Clone`
- Extract `entRowToAttribution()` conversion helper
- Extract P2P CLI push/fetch shared helpers (`gatewayAddr`, `provenanceRequestBody`, `addProvenanceExchangeFlags`)
- Document RunLedger hook-outside-lock pattern

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `session-provenance`: Redaction validation now enforced in core Export/Verify (rejects invalid levels with ErrInvalidRedaction); all CLI commands enforce disabled check per spec

## Impact

- `internal/provenance/` — types.go, errors.go, bundle.go, attribution_ent_store.go
- `internal/cli/provenance/` — common.go, checkpoint.go, attribution.go, session.go, bundle.go, provenance.go
- `internal/cli/p2p/provenance.go`
- `internal/app/p2p_routes.go`
- `internal/runledger/store.go`, `internal/runledger/ent_store.go`
- Tests added: bundle_test.go (3 cases), p2p_routes_test.go (2 cases), provenance_test.go (2 cases)
