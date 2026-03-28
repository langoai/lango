## ADDED Requirements

### Requirement: Config deep copy
Config SHALL expose a Clone() method that returns a deep copy via JSON roundtrip. Clone() SHALL be nil-safe.

#### Scenario: Clone produces independent copy
- **WHEN** Clone() is called on a Config
- **THEN** the returned copy SHALL have identical values but mutating the copy SHALL NOT affect the original

#### Scenario: Nil config clone
- **WHEN** Clone() is called on a nil *Config
- **THEN** it SHALL return nil without panic
