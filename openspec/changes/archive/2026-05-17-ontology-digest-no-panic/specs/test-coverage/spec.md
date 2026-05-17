## ADDED Requirements

### Requirement: Ontology exchange panic guard stays executable

Executable tests SHALL prevent production `panic` calls from being reintroduced in `internal/ontology/exchange.go`.

#### Scenario: Ontology exchange panic regressions are rejected
- **WHEN** `internal/ontology/exchange.go` contains a `panic(` call
- **THEN** an executable package test SHALL fail and identify the offending file and line
