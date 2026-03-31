## MODIFIED Requirements

### Requirement: ImportSchema method
`OntologyService.ImportSchema(ctx, bundle, opts)` SHALL import slim types into the local ontology. After successful import, it SHALL call `refreshPredicateCache()` if predicates were added and `version.Add(n)` where n is the total number of added types and predicates.

#### Scenario: Imported predicate immediately usable
- **WHEN** ImportSchema adds a predicate in shadow mode
- **THEN** the predicate SHALL pass PredicateValidator validation immediately (no restart needed)

#### Scenario: Schema version bumped after import
- **WHEN** ImportSchema adds N types and predicates
- **THEN** SchemaVersion SHALL increase by N
