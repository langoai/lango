## ADDED Requirements

### Requirement: README internal tree keeps a single A2A row
The README internal CLI inventory SHALL contain exactly one `a2a/` row for the shipped `lango a2a card/check` surface.

#### Scenario: Duplicate A2A rows stay removed
- **WHEN** a maintainer updates the README internal tree
- **THEN** it SHALL contain exactly one `a2a/` inventory row describing `lango a2a card/check`
