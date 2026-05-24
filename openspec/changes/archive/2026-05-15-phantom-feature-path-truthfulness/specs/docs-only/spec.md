## ADDED Requirements
### Requirement: Phantom-feature-wiring docs stay truthful about removed companion discovery code
Main specs SHALL not reintroduce deleted companion discovery file references once the runtime moved to the gateway-backed companion model.

#### Scenario: Deleted companion discovery path is rejected
- **WHEN** a maintainer updates the `phantom-feature-wiring` main spec
- **THEN** it SHALL not claim that `internal/companion/discovery.go` is part of the current shipped runtime
