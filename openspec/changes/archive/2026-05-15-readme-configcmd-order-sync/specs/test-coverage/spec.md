## ADDED Requirements

### Requirement: README config inventory order guard stays executable
Repository-level regressions that reintroduce stale config inventory ordering into the README internal tree SHALL be enforced by an executable test.

#### Scenario: Stale config inventory ordering is rejected
- **WHEN** the README internal tree still documents the shipped config management surface
- **THEN** an executable repository test SHALL fail if it places `validate` ahead of `get`, `set`, and `keys`
