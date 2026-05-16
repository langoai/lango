## ADDED Requirements

### Requirement: Shared confirmation wrapper can treat EOF as denial
The shared CLI prompt package SHALL provide a confirmation wrapper that maps EOF to `(false, nil)` while preserving normal approval/denial semantics for explicit input.

#### Scenario: EOF becomes clean denial
- **WHEN** the shared EOF-deny confirmation wrapper reads EOF before an approval answer is received
- **THEN** it SHALL return `false` with no error
