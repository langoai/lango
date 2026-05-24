## ADDED Requirements
### Requirement: README P2P git wording guard stays executable
Repository-level regressions that reintroduce stale mixed wording for the README `lango p2p git` family SHALL be enforced by an executable test.

#### Scenario: Stale README git wording is rejected
- **WHEN** the README quick reference uses the `lango p2p git` family
- **THEN** an executable repository test SHALL fail if it reintroduces the older mixed wording
