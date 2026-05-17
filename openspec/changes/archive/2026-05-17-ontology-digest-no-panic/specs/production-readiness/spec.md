## ADDED Requirements

### Requirement: Ontology schema digest avoids production panic paths

Ontology schema digest computation SHALL avoid production `panic` calls and route digest computation failures through ordinary errors.

#### Scenario: Schema export digest failure fails closed
- **WHEN** ontology schema export computes the bundle digest
- **AND** digest marshaling fails
- **THEN** export SHALL return an error instead of panicking
