## ADDED Requirements

### Requirement: Payment-send non-interactive confirmation uses a single shared guard
`lango payment send` SHALL enforce its non-interactive confirmation policy through the shared TTY-input guard without requiring an additional process-global interactive check.

#### Scenario: Pipe-based non-interactive send still requires force
- **WHEN** `lango payment send` receives non-terminal input without `--force`
- **THEN** it SHALL return the existing `use --force for non-interactive mode` error through the shared guard path
