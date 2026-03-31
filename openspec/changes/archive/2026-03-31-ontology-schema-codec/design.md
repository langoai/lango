## Context

Stage 2 complete. OntologyService has 37 methods with ACL, Actions, Governance. Schema types (ObjectType, PredicateDefinition) contain local-only fields: UUID, CreatedAt, UpdatedAt, Status, Version. These must be stripped for P2P exchange to ensure digest stability across peers.

## Goals / Non-Goals

**Goals:**
- Stable wire format: identical schemas on different peers produce identical digests
- Governance-aware import: imported types respect local FSM (shadow or proposed)
- Single-player value: export/import works standalone for backup/restore

**Non-Goals:**
- P2P transport (Change 3-3/3-5)
- Incremental/delta sync (full bundle only for v1)
- Schema migration or version compatibility negotiation

## Decisions

### D1: Slim wire types in ontology/types.go

`SchemaTypeSlim` and `SchemaPredicateSlim` are defined in the ontology package (not protocol package). This allows Change 3-1 to be independent of Change 3-3 (protocol messages).

**Why:** Avoids circular dependency. Protocol package will import ontology types, not the reverse.

### D2: Export filters active + shadow only

`ExportSchema` only includes types/predicates with status `active` or `shadow`. `proposed`, `quarantined`, and `deprecated` are excluded.

**Why:** Only schema elements in operational use should be shared. Proposed/quarantined are still under review; deprecated should not spread.

### D3: Import never creates active status

`ImportShadow` → shadow status. `ImportGoverned` → proposed status (enters governance FSM). `ImportDryRun` → no mutations.

**Why:** External schema must never bypass local governance. Even without governance enabled, `shadow` provides a trial period.

### D4: Digest uses canonical JSON of slim types

`ComputeDigest` serializes `Types` + `Predicates` arrays with sorted keys, no whitespace, then SHA256. This means the same schema on any peer produces the same digest regardless of local UUID/timestamp differences.

**Why:** Digest stability is the foundation for efficient P2P schema matching (Change 3-2 discovery digest).

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Slim type loses property constraints detail | SchemaPropertySlim includes Name, Type, Required — sufficient for schema identity. Full constraints can be added later |
| Import name collision | ImportResult.TypesConflicting reports conflicts, no silent overwrite |
| Large bundles | Export filters to active+shadow only; future: requested types filter in protocol layer |
