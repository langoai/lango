## Why

OntologyService has no way to serialize or deserialize its schema for external consumption. ObjectType and PredicateDefinition contain local-only fields (UUID, timestamps, status, version) that make direct serialization unsuitable for P2P exchange — digests would differ across peers for identical schemas. Stage 3 needs a stable, peer-portable wire format for schema export/import.

## What Changes

- Add `SchemaTypeSlim` and `SchemaPredicateSlim` wire types (no UUID, timestamps, status, version) to `types.go`
- Add `SchemaBundle` type using slim wire types + SHA256 digest for peer-stable identity
- Add `ExportSchema(ctx) → SchemaBundle` (exports active+shadow types/predicates as slim)
- Add `ImportSchema(ctx, bundle, opts) → ImportResult` (imports slim → full, respects governance)
- Add `ImportMode` enum: `shadow` (default), `governed` (proposed via FSM), `dry_run`
- Add `ComputeDigest` function for canonical JSON SHA256

## Capabilities

### New Capabilities
- `ontology-schema-codec`: Versioned schema export/import with slim wire types, digest stability, governance-aware import modes.

### Modified Capabilities
(none — no existing spec changes)

## Impact

- `internal/ontology/types.go` — slim wire types, SchemaBundle, ImportMode, ImportOptions, ImportResult
- `internal/ontology/exchange.go` — ExportSchema, ImportSchema, ComputeDigest, full↔slim converters
- `internal/ontology/service.go` — +2 interface methods + ServiceImpl implementations
- No Ent schema changes, no P2P code changes
- Backward compatible: new methods only, no behavior changes to existing methods
