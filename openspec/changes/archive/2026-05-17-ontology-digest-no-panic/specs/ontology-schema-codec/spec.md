## MODIFIED Requirements

### Requirement: ExportSchema method
`OntologyService.ExportSchema(ctx)` SHALL return a SchemaBundle containing only types and predicates with status `active` or `shadow`. It SHALL require `PermRead`. The Digest field SHALL be computed from the canonical JSON of the slim types and slim predicates. Digest computation failures SHALL be returned as export errors rather than panics.

#### Scenario: Export digest computation failure returns error
- **WHEN** schema export cannot compute the canonical digest
- **THEN** `ExportSchema` SHALL return an error identifying digest computation
- **AND** it SHALL NOT panic

### Requirement: ComputeDigest determinism
`ComputeDigest` SHALL produce a SHA256 hash from canonical JSON (sorted keys, no whitespace) of the Types and Predicates arrays. The same logical schema SHALL always produce the same digest regardless of array ordering. The compatibility API SHALL NOT panic during digest computation.

#### Scenario: Digest computation avoids panic
- **WHEN** digest computation encounters an internal marshal failure
- **THEN** production export code SHALL surface the failure as an error
- **AND** `ComputeDigest` SHALL NOT panic
