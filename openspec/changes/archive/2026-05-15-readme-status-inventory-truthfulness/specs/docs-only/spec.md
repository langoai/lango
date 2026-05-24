## ADDED Requirements

### Requirement: README status inventory uses the current dead-letter command list
The README internal CLI inventory SHALL describe the status family with the current dead-letter command list rather than vague shorthand.

#### Scenario: Vague status shorthand stays removed
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe `lango status/dead-letter-summary/dead-letters/dead-letter/dead-letter retry`
