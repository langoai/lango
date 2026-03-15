## MODIFIED Requirements

### Requirement: Tool output size management
The system SHALL manage tool output size using token-based tiered compression via `WithOutputManager` middleware instead of character-based truncation. The middleware SHALL classify outputs into Small/Medium/Large tiers and apply content-aware compression.

#### Scenario: Output within budget
- **WHEN** a tool returns output within the token budget
- **THEN** the output SHALL pass through with `_meta` metadata injected

#### Scenario: Output exceeding budget
- **WHEN** a tool returns output exceeding the token budget
- **THEN** the output SHALL be compressed using content-type-specific strategies and `_meta.compressed` SHALL be `true`
