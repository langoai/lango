## ADDED Requirements
### Requirement: Architecture overview companion wording stays truthful
The architecture overview SHALL describe companion support in terms of the current gateway-connected model rather than a stale discovery subsystem.

#### Scenario: Stale overview companion discovery wording is rejected
- **WHEN** a maintainer updates the architecture overview
- **THEN** it SHALL not describe `internal/security/` as owning a shipped companion discovery subsystem
