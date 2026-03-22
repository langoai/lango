## Context

The provenance feature on the `feature/provenance` branch introduces session checkpoints, session trees, attribution tracking, and signed provenance bundles. Two spec contract violations and several code quality issues were identified during code review against the `dev` branch.

Current state:
- `redactBundle()` silently falls back invalid redaction values to `content` (bundle.go:224-226)
- `session.go` and `bundle.go` CLI commands skip the `provenance.enabled` disabled check
- Duplicate map clone functions, missing conversion helpers, and undocumented patterns

## Goals / Non-Goals

**Goals:**
- Enforce redaction level validation at core service layer (Export + Verify) so all entry points are covered
- Add route-level early validation for better UX (400 vs 502)
- Ensure all CLI provenance commands respect `provenance.enabled=false` per spec
- Clean up code duplication in bundle.go, attribution_ent_store.go, and P2P CLI

**Non-Goals:**
- Changing the redaction levels themselves (none/content/full remain unchanged)
- Adding new provenance features or modifying the bundle format
- Restructuring store interfaces or the module system
- Adding comprehensive test coverage beyond the specific changes

## Decisions

### 1. Validation placement: core service + route defense-in-depth
Redaction validation is added in `Export()` and `Verify()` (core authority), plus `decodeProvenanceRequest()` (route early-fail). Core catches all paths; route provides better UX for HTTP clients by returning 400 before attempting remote P2P calls that would result in 502.

Alternative: Route-only validation. Rejected because CLI bundle export and P2P wiring exporter bypass the route.

### 2. Single sentinel error (ErrInvalidRedaction) instead of batch
Per go-errors.md rules, sentinel errors are only for cases where callers match via `errors.Is`. Only `ErrInvalidRedaction` has clear matching need (route handlers distinguish validation vs runtime errors). Other bundle errors (`signer DID required`, etc.) stay as `fmt.Errorf`.

### 3. Status command exception for disabled check
`status` is an introspection command that shows config state. Blocking it with the disabled check would prevent users from seeing that provenance is disabled. Instead, `status` shows config and appends the disabled notice at the end.

### 4. maps.Clone over generic helper
Go 1.21+ provides `maps.Clone` in the standard library. Preferred over writing a generic `cloneMap[K,V]` helper since the project uses Go 1.25.4.

## Risks / Trade-offs

- [Existing bundles with invalid redaction] Bundles already stored with invalid `redaction_level` will fail `Verify()` on re-import → Acceptable because no such bundles should exist in production (redaction was always defaulted to `content`)
- [redactBundle default case] The `default` case in `redactBundle()` is kept as fallback to `content` even though `Export()` now validates. This is defense-in-depth, not a spec guarantee → Acceptable for internal safety
