## ADDED Requirements

### Requirement: Non-generated coverage target
The repository SHALL maintain at least 90% Go statement coverage for repository-owned non-generated code.

Generated code SHALL be excluded from this target when it is under a known generated-code path such as `internal/ent/` or when the file contains standard generated-code markers such as `Code generated` and `DO NOT EDIT`.

#### Scenario: Non-generated coverage is at least 90%
- **WHEN** the non-generated coverage report runs
- **THEN** the reported statement coverage SHALL be at least 90%
- **AND** generated Go code SHALL not contribute to the numerator or denominator

#### Scenario: Coverage report identifies the largest gaps
- **WHEN** the non-generated coverage report runs
- **THEN** it SHALL include covered statements, total statements, uncovered statements, and the overall percentage
- **AND** it SHALL list the largest non-generated files by uncovered statement count

#### Scenario: Coverage threshold gate fails below target
- **WHEN** the non-generated coverage gate runs against a profile below 90%
- **THEN** it SHALL fail with a non-zero exit status
- **AND** it SHALL report the measured percentage and required threshold

#### Scenario: Coverage threshold gate passes at target
- **WHEN** the non-generated coverage gate runs against a profile at or above 90%
- **THEN** it SHALL exit successfully
- **AND** it SHALL report the measured percentage and required threshold

### Requirement: Coverage increases remain behavior-focused
Coverage-improving tests SHALL assert observable behavior rather than only executing code to raise percentages.

#### Scenario: Coverage batch tests assert behavior
- **WHEN** a coverage batch adds or updates tests
- **THEN** the tests SHALL include assertions for expected outputs, errors, state transitions, or side effects
- **AND** review SHALL reject tests that only invoke functions without checking behavior
