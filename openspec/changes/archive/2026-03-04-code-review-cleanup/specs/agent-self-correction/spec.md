## MODIFIED Requirements

### Requirement: REJECT pattern matching
The system SHALL provide a `containsRejectPattern` function that matches the exact `[REJECT]` text marker using `strings.Contains`. The match SHALL be case-sensitive (lowercase `[reject]` SHALL NOT match).

#### Scenario: Exact REJECT marker matched
- **WHEN** text contains `[REJECT]`
- **THEN** `containsRejectPattern` SHALL return true

#### Scenario: Case-sensitive matching
- **WHEN** text contains `[reject]` (lowercase)
- **THEN** `containsRejectPattern` SHALL return false

#### Scenario: Normal text not matched
- **WHEN** text contains no `[REJECT]` marker
- **THEN** `containsRejectPattern` SHALL return false
