## ADDED Requirements

### Requirement: Shared guarded confirmation supports TTY-only command flows
The shared CLI prompt package SHALL provide a confirmation helper that rejects non-terminal stdin with a caller-supplied guidance error and treats EOF as a clean denial instead of a hard failure.

#### Scenario: Guarded confirmation rejects non-terminal stdin
- **WHEN** the guarded confirmation helper is called with a non-terminal `*os.File` input stream
- **THEN** it SHALL return a caller-supplied error
- **AND** it SHALL NOT write a confirmation prompt

#### Scenario: Guarded confirmation treats EOF as denial
- **WHEN** the guarded confirmation helper reaches EOF before an approval answer is read
- **THEN** it SHALL return `false` with no error
