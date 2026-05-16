## ADDED Requirements
### Requirement: Architecture overview companion wording guard stays executable
Repository-level regressions that reintroduce stale companion discovery wording into the architecture overview SHALL be enforced by an executable test.

#### Scenario: Stale overview companion discovery wording is rejected
- **WHEN** the overview page reintroduces stale companion discovery wording
- **THEN** an executable repository test SHALL fail
