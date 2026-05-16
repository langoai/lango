## ADDED Requirements
### Requirement: Phantom-feature-wiring broken-path guard stays executable
Repository-level regressions that reintroduce the deleted `internal/companion/discovery.go` path into the `phantom-feature-wiring` main spec SHALL be enforced by an executable test.

#### Scenario: Deleted companion discovery path is rejected
- **WHEN** the current companion model is gateway-backed rather than discovery-backed
- **THEN** an executable repository test SHALL fail if the main spec reintroduces `internal/companion/discovery.go`
